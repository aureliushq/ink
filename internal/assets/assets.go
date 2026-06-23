package assets

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/log"
)

// Copy writes static assets into <output>/static. The active theme's bundled
// static/ directory is copied first, then the site's static/ directory (if any)
// is copied on top so a same-named site file overrides the theme's.
func Copy(cfg *config.Config, themesFS embed.FS, logger *log.Logger) error {
	const staticDir = "static"
	publicStaticDir := path.Join(cfg.Build.OutputDir, staticDir)

	themeStaticDir := path.Join("themes", cfg.Theme.Name, staticDir)
	if _, err := fs.Stat(themesFS, themeStaticDir); err == nil {
		if err := copyTree(themesFS, themeStaticDir, publicStaticDir); err != nil {
			return err
		}
	}

	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		if err := copyTree(os.DirFS(staticDir), ".", publicStaticDir); err != nil {
			return err
		}
	} else {
		logger.Info("site static directory not found, using theme's static directory instead")
	}

	return nil
}

// copyTree copies every file under source in fsys into destination, preserving
// the directory structure relative to source and overwriting existing files.
func copyTree(fsys fs.FS, source, destination string) error {
	return fs.WalkDir(fsys, source, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(source, p)
		if err != nil {
			return err
		}
		dest := filepath.Join(destination, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		return copyFile(fsys, p, dest)
	})
}

func copyFile(fsys fs.FS, source, destination string) error {
	in, err := fsys.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
