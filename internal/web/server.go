package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config keeps the web layer thin: the Bash CLI remains the transcript engine.
type Config struct {
	DLMACPath   string
	FFprobePath string
	WorkDir     string
}

type Server struct {
	config   Config
	template *template.Template
}

func NewServer(config Config) *Server {
	if config.FFprobePath == "" {
		config.FFprobePath = "ffprobe"
	}

	tmpl := template.Must(template.ParseFS(assetFS, "assets/templates/index.html"))
	return &Server{
		config:   config,
		template: tmpl,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("POST /api/transcript", s.handleTranscript)
	mux.HandleFunc("POST /api/download", s.handleDownloadMedia)
	mux.HandleFunc("POST /api/pick-file", s.handlePickFile)
	mux.HandleFunc("POST /api/inspect", s.handleInspect)
	mux.HandleFunc("GET /api/capabilities", s.handleCapabilities)
	mux.HandleFunc("POST /api/convert", s.handleConvert)
	mux.HandleFunc("GET /downloads/", s.handleDownload)

	staticFS, err := fs.Sub(assetFS, "assets/static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.template.ExecuteTemplate(w, "index.html", map[string]string{
		"Version": "0.4 workbench",
	}); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

type transcriptRequest struct {
	URL    string `json:"url"`
	Lang   string `json:"lang"`
	Format string `json:"format"`
}

type transcriptResponse struct {
	OK        bool         `json:"ok"`
	File      string       `json:"file,omitempty"`
	Download  string       `json:"download,omitempty"`
	Text      string       `json:"text,omitempty"`
	Output    string       `json:"output,omitempty"`
	Files     []fileResult `json:"files,omitempty"`
	Media     *mediaInfo   `json:"media,omitempty"`
	Cancelled bool         `json:"cancelled,omitempty"`
	Error     string       `json:"error,omitempty"`
}

type fileResult struct {
	File     string `json:"file"`
	Download string `json:"download"`
}

func (s *Server) handleTranscript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req transcriptRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: "Invalid JSON request."})
		return
	}

	if err := validateTranscriptRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	result, err := s.runTranscript(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func validateTranscriptRequest(req transcriptRequest) error {
	if err := validatePublicURL(req.URL); err != nil {
		return err
	}

	switch req.Lang {
	case "en", "id":
	default:
		return errors.New("Choose English or Indonesian.")
	}

	switch req.Format {
	case "txt", "vtt", "srt":
	default:
		return errors.New("Choose Text, WebVTT, or SRT.")
	}

	return nil
}

func validatePublicURL(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("Paste a valid public video URL.")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("Only http and https URLs are supported.")
	}

	return nil
}

var transcriptLine = regexp.MustCompile(`(?m)^Transcript \([^)]+\): (.+)$`)

func (s *Server) runTranscript(parent context.Context, req transcriptRequest) (transcriptResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.config.DLMACPath, "transcript", req.URL, "--lang", req.Lang, "--format", req.Format)
	cmd.Dir = s.config.WorkDir
	output, err := cmd.CombinedOutput()
	textOutput := strings.TrimSpace(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		return transcriptResponse{}, errors.New("Transcript request timed out.")
	}

	if err != nil {
		if textOutput == "" {
			textOutput = err.Error()
		}
		return transcriptResponse{}, fmt.Errorf("%s", textOutput)
	}

	match := transcriptLine.FindStringSubmatch(textOutput)
	if len(match) != 2 {
		return transcriptResponse{}, errors.New("Transcript finished, but dlmac did not report an output file.")
	}

	relativeFile := filepath.Clean(match[1])
	if filepath.IsAbs(relativeFile) || strings.HasPrefix(relativeFile, ".."+string(filepath.Separator)) || relativeFile == ".." {
		return transcriptResponse{}, errors.New("dlmac reported an unsafe output path.")
	}
	if filepath.Dir(relativeFile) != "downloads" {
		return transcriptResponse{}, errors.New("dlmac reported an output file outside downloads.")
	}

	absoluteFile := filepath.Join(s.config.WorkDir, relativeFile)

	response := transcriptResponse{
		OK:       true,
		File:     filepath.ToSlash(relativeFile),
		Download: "/downloads/" + url.PathEscape(filepath.Base(relativeFile)),
	}

	if req.Format == "txt" {
		content, err := os.ReadFile(absoluteFile)
		if err != nil {
			return transcriptResponse{}, fmt.Errorf("Transcript saved, but could not be read: %w", err)
		}
		response.Text = string(content)
	}

	return response, nil
}

type mediaRequest struct {
	URL         string `json:"url"`
	Kind        string `json:"kind"`
	Quality     string `json:"quality"`
	AudioFormat string `json:"audioFormat"`
}

func (s *Server) handleDownloadMedia(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req mediaRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: "Invalid JSON request."})
		return
	}

	if err := validateMediaRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	result, err := s.runMediaDownload(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func validateMediaRequest(req mediaRequest) error {
	if err := validatePublicURL(req.URL); err != nil {
		return err
	}

	switch req.Kind {
	case "video":
		switch req.Quality {
		case "", "best", "360p", "480p", "720p", "1080p":
		default:
			return errors.New("Choose Best, 360p, 480p, 720p, or 1080p.")
		}
	case "audio":
		switch req.AudioFormat {
		case "", "mp3", "m4a", "wav":
		default:
			return errors.New("Choose MP3, M4A, or WAV.")
		}
	default:
		return errors.New("Choose video or audio download.")
	}

	return nil
}

func (s *Server) runMediaDownload(parent context.Context, req mediaRequest) (transcriptResponse, error) {
	before, err := snapshotDownloads(s.config.WorkDir)
	if err != nil {
		return transcriptResponse{}, err
	}

	args := []string{req.Kind, req.URL}
	if req.Kind == "video" && req.Quality != "" && req.Quality != "best" {
		args = append(args, "--quality", req.Quality)
	}
	if req.Kind == "audio" {
		audioFormat := req.AudioFormat
		if audioFormat == "" {
			audioFormat = "mp3"
		}
		args = append(args, "--format", audioFormat)
	}

	output, err := s.runDLMAC(parent, 2*time.Hour, args...)
	if err != nil {
		return transcriptResponse{}, err
	}

	files, err := changedDownloads(s.config.WorkDir, before)
	if err != nil {
		return transcriptResponse{}, err
	}

	return transcriptResponse{
		OK:     true,
		Output: output,
		Files:  files,
	}, nil
}

type convertRequest struct {
	File string `json:"file"`
	To   string `json:"to"`
}

type inspectRequest struct {
	File string `json:"file"`
}

type capabilitiesResponse struct {
	OK   bool `json:"ok"`
	WebP bool `json:"webp"`
}

type mediaInfo struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Size       int64  `json:"size"`
	SizeLabel  string `json:"sizeLabel"`
	Duration   string `json:"duration,omitempty"`
	Dimensions string `json:"dimensions,omitempty"`
}

type ffprobeResult struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType string `json:"codec_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

var errFilePickerCancelled = errors.New("file picker canceled")

const supportedConvertFormats = "mp4, webm, mkv, mov, mp3, m4a, wav, flac, ogg, opus, jpg, png, gif, tiff, webp"

func normalizeConvertFormat(raw string) string {
	format := strings.ToLower(strings.TrimSpace(raw))
	if format == "jpeg" {
		return "jpg"
	}
	return format
}

func isSupportedConvertFormat(format string) bool {
	switch normalizeConvertFormat(format) {
	case "mp4", "webm", "mkv", "mov", "mp3", "m4a", "wav", "flac", "ogg", "opus", "jpg", "png", "gif", "tiff":
		return true
	case "webp":
		return true
	default:
		return false
	}
}

func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeJSONResponse(w, http.StatusOK, capabilitiesResponse{
		OK:   true,
		WebP: webpAvailable(),
	})
}

func webpAvailable() bool {
	return webpToolAvailable("cwebp", "DLMAC_CWEBP_PATH") &&
		webpToolAvailable("gif2webp", "DLMAC_GIF2WEBP_PATH")
}

func webpToolAvailable(tool string, envName string) bool {
	if override := strings.TrimSpace(os.Getenv(envName)); override != "" {
		info, err := os.Stat(override)
		return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
	}

	_, err := exec.LookPath(tool)
	return err == nil
}

func (s *Server) handlePickFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	path, err := chooseLocalFile(r.Context())
	if errors.Is(err, errFilePickerCancelled) {
		writeJSON(w, http.StatusOK, transcriptResponse{OK: true, Cancelled: true})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	media, err := s.inspectMediaFile(r.Context(), path)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, transcriptResponse{
		OK:    true,
		File:  media.Path,
		Media: &media,
	})
}

func (s *Server) handleInspect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req inspectRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: "Invalid JSON request."})
		return
	}

	media, err := s.inspectMediaFile(r.Context(), req.File)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, transcriptResponse{
		OK:    true,
		File:  media.Path,
		Media: &media,
	})
}

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req convertRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: "Invalid JSON request."})
		return
	}

	if err := validateConvertRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: err.Error()})
		return
	}
	if _, err := s.resolveLocalFile(req.File); err != nil {
		writeJSON(w, http.StatusBadRequest, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	result, err := s.runConvert(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func validateConvertRequest(req convertRequest) error {
	if strings.TrimSpace(req.File) == "" {
		return errors.New("Choose a local media file.")
	}

	if !isSupportedConvertFormat(req.To) {
		return fmt.Errorf("Choose a supported target format: %s.", supportedConvertFormats)
	}
	if normalizeConvertFormat(req.To) == "webp" && !webpAvailable() {
		return errors.New("WebP conversion is unavailable. Install it with: brew install webp.")
	}

	return nil
}

var convertedLine = regexp.MustCompile(`(?m)^Converted: (.+)$`)

func (s *Server) runConvert(parent context.Context, req convertRequest) (transcriptResponse, error) {
	inputPath, err := s.resolveLocalFile(req.File)
	if err != nil {
		return transcriptResponse{}, err
	}

	output, err := s.runDLMAC(parent, 2*time.Hour, "convert", inputPath, "--to", normalizeConvertFormat(req.To))
	if err != nil {
		return transcriptResponse{}, err
	}

	match := convertedLine.FindStringSubmatch(output)
	if len(match) != 2 {
		return transcriptResponse{
			OK:     true,
			Output: output,
		}, nil
	}

	relativeFile := filepath.Clean(match[1])
	if filepath.IsAbs(relativeFile) || strings.HasPrefix(relativeFile, ".."+string(filepath.Separator)) || relativeFile == ".." || filepath.Dir(relativeFile) != "downloads" {
		return transcriptResponse{}, errors.New("dlmac reported an unsafe output path.")
	}

	return transcriptResponse{
		OK:       true,
		File:     filepath.ToSlash(relativeFile),
		Download: "/downloads/" + url.PathEscape(filepath.Base(relativeFile)),
		Output:   fmt.Sprintf("Converted: %s", filepath.ToSlash(relativeFile)),
		Files: []fileResult{{
			File:     filepath.ToSlash(relativeFile),
			Download: "/downloads/" + url.PathEscape(filepath.Base(relativeFile)),
		}},
	}, nil
}

func (s *Server) resolveLocalFile(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("Choose a local media file.")
	}

	path := value
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.config.WorkDir, path)
	}
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("Local file not found: %s", value)
	}
	if err != nil {
		return "", fmt.Errorf("Could not read local file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("Choose a file, not a folder.")
	}

	return path, nil
}

func chooseLocalFile(parent context.Context) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", errors.New("The native file picker is only available on macOS.")
	}

	ctx, cancel := context.WithTimeout(parent, 10*time.Minute)
	defer cancel()

	script := `POSIX path of (choose file with prompt "Choose a local media file")`
	output, err := exec.CommandContext(ctx, "osascript", "-e", script).CombinedOutput()
	textOutput := strings.TrimSpace(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("File picker timed out.")
	}
	if err != nil {
		lowerOutput := strings.ToLower(textOutput)
		if strings.Contains(lowerOutput, "user canceled") || strings.Contains(lowerOutput, "user cancelled") || strings.Contains(lowerOutput, "-128") {
			return "", errFilePickerCancelled
		}
		if textOutput == "" {
			textOutput = err.Error()
		}
		return "", fmt.Errorf("Could not open the file picker: %s", textOutput)
	}
	if textOutput == "" {
		return "", errors.New("The file picker returned no file.")
	}

	return filepath.Clean(textOutput), nil
}

func (s *Server) inspectMediaFile(parent context.Context, rawPath string) (mediaInfo, error) {
	path, err := s.resolveLocalFile(rawPath)
	if err != nil {
		return mediaInfo{}, err
	}

	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		s.config.FFprobePath,
		"-v", "error",
		"-show_entries", "stream=codec_type,width,height:format=duration",
		"-of", "json",
		path,
	)
	output, err := cmd.CombinedOutput()
	textOutput := strings.TrimSpace(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		return mediaInfo{}, errors.New("Media inspection timed out.")
	}
	if err != nil {
		if textOutput == "" {
			textOutput = err.Error()
		}
		return mediaInfo{}, fmt.Errorf("Could not inspect the local media file: %s", textOutput)
	}

	var probe ffprobeResult
	if err := json.Unmarshal(output, &probe); err != nil {
		return mediaInfo{}, fmt.Errorf("Could not read media information: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return mediaInfo{}, fmt.Errorf("Could not read local file metadata: %w", err)
	}

	kind := classifyMedia(path, probe.Streams)
	if kind == "" {
		return mediaInfo{}, errors.New("The selected file does not contain a supported media stream.")
	}

	media := mediaInfo{
		Path:      path,
		Name:      filepath.Base(path),
		Kind:      kind,
		Size:      info.Size(),
		SizeLabel: humanSize(info.Size()),
		Duration:  formatDuration(probe.Format.Duration),
	}

	for _, stream := range probe.Streams {
		if stream.CodecType == "video" && stream.Width > 0 && stream.Height > 0 {
			media.Dimensions = fmt.Sprintf("%d×%d", stream.Width, stream.Height)
			break
		}
	}

	return media, nil
}

func classifyMedia(path string, streams []ffprobeStream) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".tif", ".tiff":
		return "image"
	}

	for _, stream := range streams {
		if stream.CodecType == "video" {
			return "video"
		}
	}
	for _, stream := range streams {
		if stream.CodecType == "audio" {
			return "audio"
		}
	}
	return ""
}

func formatDuration(raw string) string {
	secondsFloat, err := strconv.ParseFloat(raw, 64)
	if err != nil || secondsFloat <= 0 {
		return ""
	}

	seconds := int(secondsFloat + 0.5)
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remaining := seconds % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, remaining)
	}
	return fmt.Sprintf("%d:%02d", minutes, remaining)
}

func humanSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(size)
	for _, unit := range units {
		value /= 1024
		if value < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}
	return fmt.Sprintf("%d B", size)
}

func (s *Server) runDLMAC(parent context.Context, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.config.DLMACPath, args...)
	cmd.Dir = s.config.WorkDir
	output, err := cmd.CombinedOutput()
	textOutput := strings.TrimSpace(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("dlmac request timed out.")
	}

	if err != nil {
		if textOutput == "" {
			textOutput = err.Error()
		}
		return "", fmt.Errorf("%s", textOutput)
	}

	return textOutput, nil
}

type fileStamp struct {
	ModTime time.Time
	Size    int64
}

func snapshotDownloads(workDir string) (map[string]fileStamp, error) {
	downloadsDir := filepath.Join(workDir, "downloads")
	entries, err := os.ReadDir(downloadsDir)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]fileStamp{}, nil
	}
	if err != nil {
		return nil, err
	}

	snapshot := make(map[string]fileStamp, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		snapshot[entry.Name()] = fileStamp{
			ModTime: info.ModTime(),
			Size:    info.Size(),
		}
	}

	return snapshot, nil
}

func changedDownloads(workDir string, before map[string]fileStamp) ([]fileResult, error) {
	downloadsDir := filepath.Join(workDir, "downloads")
	entries, err := os.ReadDir(downloadsDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var results []fileResult
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		old, existed := before[entry.Name()]
		if existed && old.ModTime.Equal(info.ModTime()) && old.Size == info.Size() {
			continue
		}

		file := filepath.ToSlash(filepath.Join("downloads", entry.Name()))
		results = append(results, fileResult{
			File:     file,
			Download: "/downloads/" + url.PathEscape(entry.Name()),
		})
	}

	return results, nil
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	name, err := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/downloads/"))
	if err != nil || name == "" || strings.Contains(name, "/") || strings.Contains(name, `\`) {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(s.config.WorkDir, "downloads", filepath.Base(name))
	http.ServeFile(w, r, path)
}

func writeJSON(w http.ResponseWriter, status int, response transcriptResponse) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func writeJSONResponse(w http.ResponseWriter, status int, response capabilitiesResponse) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
