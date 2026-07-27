package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/srmdn/dlmac/internal/web"
)

func main() {
	dlmacPath := flag.String("dlmac", "./dlmac", "path to the dlmac CLI")
	workDir := flag.String("workdir", ".", "working directory for downloads")
	openBrowser := flag.Bool("open", true, "open the local URL in the default browser")
	flag.Parse()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("dlmac web: listen: %v", err)
	}

	server := &http.Server{
		Handler: web.NewServer(web.Config{
			DLMACPath: *dlmacPath,
			WorkDir:   *workDir,
		}).Routes(),
	}

	url := fmt.Sprintf("http://%s", listener.Addr().String())
	fmt.Printf("dlmac web: %s\n", url)
	fmt.Println("Press Ctrl+C to stop.")

	if *openBrowser && runtime.GOOS == "darwin" {
		if err := exec.Command("open", url).Start(); err != nil {
			fmt.Printf("dlmac web: could not open browser: %v\n", err)
		}
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("dlmac web: serve: %v", err)
	}
}
