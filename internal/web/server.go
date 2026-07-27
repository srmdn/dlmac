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
	"strings"
	"time"
)

// Config keeps the web layer thin: the Bash CLI remains the transcript engine.
type Config struct {
	DLMACPath string
	WorkDir   string
}

type Server struct {
	config   Config
	template *template.Template
}

func NewServer(config Config) *Server {
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
		"Version": "0.3 workbench",
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
	OK       bool         `json:"ok"`
	File     string       `json:"file,omitempty"`
	Download string       `json:"download,omitempty"`
	Text     string       `json:"text,omitempty"`
	Output   string       `json:"output,omitempty"`
	Files    []fileResult `json:"files,omitempty"`
	Error    string       `json:"error,omitempty"`
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

	result, err := s.runConvert(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, transcriptResponse{OK: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func validateConvertRequest(req convertRequest) error {
	if strings.TrimSpace(req.File) == "" {
		return errors.New("Paste a local video file path.")
	}

	switch req.To {
	case "mp3", "m4a", "wav":
	default:
		return errors.New("Choose MP3, M4A, or WAV.")
	}

	return nil
}

var convertedLine = regexp.MustCompile(`(?m)^Converted: (.+)$`)

func (s *Server) runConvert(parent context.Context, req convertRequest) (transcriptResponse, error) {
	output, err := s.runDLMAC(parent, 2*time.Hour, "convert", req.File, "--to", req.To)
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
		Output:   output,
		Files: []fileResult{{
			File:     filepath.ToSlash(relativeFile),
			Download: "/downloads/" + url.PathEscape(filepath.Base(relativeFile)),
		}},
	}, nil
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
