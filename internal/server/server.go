package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func NewServer(host string, port int64) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", pageHandler)

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	slug := filepath.Clean(r.URL.Path)

	if slug == "/" {
		slug = "index.html"
	}

	targetFile := filepath.Join("public", slug)
	if filepath.Ext(targetFile) == "" {
		targetFile += ".html"
	}

	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	// tmpl, err := template.ParseFiles(targetFile)
	// if err != nil {
	// 	http.Error(w, "404 page not found", http.StatusNotFound)
	// 	return
	// }
	// tmpl.Execute(w, nil)

	http.ServeFile(w, r, targetFile)
}
