package server

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aureliushq/ink/internal/config"
)

func NewServer(cfg *config.Config, host string, port int64) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slug := path.Clean(r.URL.Path)

		if slug == "/" {
			targetFile := path.Join("public", "index.html")
			http.ServeFile(w, r, targetFile)
			return
		}

		slug = strings.Replace(slug, "/", "", 1)
		var targetFile string
		for _, collection := range cfg.Build.Collections {
			match, err := path.Match(collection, slug)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if match {
				targetFile = path.Join("public", slug, "index")
				break
			} else {
				targetFile = path.Join("public", slug)
			}
		}
		if path.Ext(targetFile) == "" {
			targetFile += ".html"
		}

		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, targetFile)
	})

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {

}
