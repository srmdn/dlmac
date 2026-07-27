package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIndexRenders(t *testing.T) {
	server := NewServer(Config{DLMACPath: "./dlmac", WorkDir: t.TempDir()})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "dlmac workbench") {
		t.Fatalf("expected workbench page, got %q", response.Body.String())
	}
}

func TestTranscriptRejectsInvalidURL(t *testing.T) {
	server := NewServer(Config{DLMACPath: "./dlmac", WorkDir: t.TempDir()})
	body := bytes.NewBufferString(`{"url":"notaurl","lang":"en","format":"txt"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/transcript", body)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "valid public video URL") {
		t.Fatalf("expected validation error, got %q", response.Body.String())
	}
}

func TestDownloadRunsDLMACAndReturnsChangedFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shell script")
	}

	workDir := t.TempDir()
	fakeDLMAC := filepath.Join(workDir, "dlmac-fake")
	script := `#!/bin/sh
set -eu
mkdir -p downloads
cat > downloads/audio.mp3 <<'EOF'
audio bytes
EOF
printf 'Downloaded audio\n'
`
	if err := os.WriteFile(fakeDLMAC, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	server := NewServer(Config{DLMACPath: fakeDLMAC, WorkDir: workDir})
	body := bytes.NewBufferString(`{"url":"https://www.youtube.com/watch?v=jNQXAC9IVRw","kind":"audio","audioFormat":"mp3"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/download", body)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "downloads/audio.mp3") {
		t.Fatalf("expected changed file, got %q", response.Body.String())
	}
}

func TestConvertRunsDLMACAndReturnsFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shell script")
	}

	workDir := t.TempDir()
	fakeDLMAC := filepath.Join(workDir, "dlmac-fake")
	script := `#!/bin/sh
set -eu
mkdir -p downloads
cat > downloads/clip.m4a <<'EOF'
audio bytes
EOF
printf 'Converted: downloads/clip.m4a\n'
`
	if err := os.WriteFile(fakeDLMAC, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	server := NewServer(Config{DLMACPath: fakeDLMAC, WorkDir: workDir})
	body := bytes.NewBufferString(`{"file":"/tmp/clip.mp4","to":"m4a"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/convert", body)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "/downloads/clip.m4a") {
		t.Fatalf("expected convert download link, got %q", response.Body.String())
	}
}

func TestTranscriptRunsDLMACAndReturnsText(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shell script")
	}

	workDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(workDir, "downloads"), 0o755); err != nil {
		t.Fatal(err)
	}

	fakeDLMAC := filepath.Join(workDir, "dlmac-fake")
	script := `#!/bin/sh
set -eu
cat > downloads/sample.en.txt <<'EOF'
hello from captions
EOF
printf 'Transcript (manual captions): downloads/sample.en.txt\n'
`
	if err := os.WriteFile(fakeDLMAC, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	server := NewServer(Config{DLMACPath: fakeDLMAC, WorkDir: workDir})
	body := bytes.NewBufferString(`{"url":"https://www.youtube.com/watch?v=jNQXAC9IVRw","lang":"en","format":"txt"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/transcript", body)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "hello from captions") {
		t.Fatalf("expected transcript text, got %q", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "/downloads/sample.en.txt") {
		t.Fatalf("expected download link, got %q", response.Body.String())
	}
}
