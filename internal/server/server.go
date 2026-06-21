package server

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/aureliushq/ink/internal/config"
)

func NewServer(cfg *config.Config, host string, port int64) *http.Server {
	mux := http.NewServeMux()
	output := cfg.Build.OutputDir

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqPath := path.Clean(r.URL.Path)

		var target string
		if path.Ext(reqPath) != "" {
			target = path.Join(output, reqPath)
		} else {
			target = path.Join(output, reqPath, "index.html")
		}

		if _, err := os.Stat(target); os.IsNotExist(err) {
			if notFound := path.Join(output, "404.html"); fileExists(notFound) {
				w.WriteHeader(http.StatusNotFound)
				http.Error(w, "404 page not found", http.StatusNotFound)
				return
			}
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		switch path.Ext(target) {
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".html":
			w.Header().Set("Content-Type", "text/html")
		}

		http.ServeFile(w, r, target)
	})

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}
