package assets

import (
	"embed"
	"io/fs"
	"os"
	"path"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/log"
)

func Copy(cfg *config.Config, themesFS embed.FS, logger *log.Logger) error {
	staticDir := "static"
	themeStaticDir := path.Join("themes", cfg.Theme.Name, staticDir)
	publicStaticDir := path.Join(cfg.Build.OutputDir, staticDir)
	if _, err := fs.Stat(themesFS, themeStaticDir); err == nil {
		if err := copyTree(themesFS, themeStaticDir, publicStaticDir); err != nil {
			return err
		}
	}

	siteStaticDir := "static"
	if _, err := fs.Stat(themesFS, siteStaticDir); err == nil {
		if err := copyTree(themesFS, siteStaticDir, publicStaticDir); err != nil {
			return err
		}
	} else {
		logger.Info("site static directory not found, using theme's static directory instead")
	}
	return nil
}

func copyTree(fsys embed.FS, source, destination string) error {
	sourceFS, err := fs.Sub(fsys, source)
	if err != nil {
		return err
	}

	if err = os.CopyFS(destination, sourceFS); err != nil {
		return err
	}

	return nil
}
