package server

import (
	"errors"
	"fmt"
	"html/template"
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

func Serve(server *http.Server) error {
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)

	if path == "/" {
		path = "index.html"
	}

	targetFile := filepath.Join("public", "demo", path)
	if filepath.Ext(targetFile) == "" {
		targetFile += ".html"
	}

	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles(targetFile)
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	tmpl.Execute(w, nil)
}
