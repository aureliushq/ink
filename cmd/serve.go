package cmd

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve your static site locally",
	Long:  `Serve your static site locally in at http://localhost:8782 with live reloading.`,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", pageHandler)

		if err := http.ListenAndServe(":8782", nil); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
