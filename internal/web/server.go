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
		"Version": "0.3 MVP",
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
	OK       bool   `json:"ok"`
	File     string `json:"file,omitempty"`
	Download string `json:"download,omitempty"`
	Text     string `json:"text,omitempty"`
	Error    string `json:"error,omitempty"`
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
	parsed, err := url.ParseRequestURI(req.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("Paste a valid public video URL.")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("Only http and https URLs are supported.")
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
