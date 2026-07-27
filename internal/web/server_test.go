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
	if !strings.Contains(response.Body.String(), "dlmac transcript") {
		t.Fatalf("expected transcript page, got %q", response.Body.String())
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
