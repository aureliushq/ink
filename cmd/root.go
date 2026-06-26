package cmd

import (
	"context"
	"embed"
	"os"

	"github.com/aureliushq/ink/internal/build"
	"github.com/aureliushq/ink/internal/config"
	"github.com/aureliushq/ink/internal/renderer"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// BuildInfo carries version metadata stamped at build time via -ldflags.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type App struct {
	Config        *config.Config
	Logger        *log.Logger
	TemplateCache *renderer.TemplateCache
	Build         BuildInfo
}

func newApp(buildInfo BuildInfo) *App {
	logger := log.New(os.Stdout)
	return &App{
		Logger: logger,
		Build:  buildInfo,
	}
}

// rootCmd represents the base command when called without any subcommands
func NewRootCommand(themesFS embed.FS, buildInfo BuildInfo) *cobra.Command {
	app := newApp(buildInfo)

	rootCmd := &cobra.Command{
		Use:   "ink",
		Short: "Yet another super simple static site generator",
		Long: `Ink is yet another super simple static site generator.
Write content in markdown, bring your own HTML+CSS templates, deploy anywhere.
Supports CommonMark and GFM. Comes with syntax highlighting, footnotes and margin notes,
and more out-of-the-box.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// The version command reports build metadata and must work
			// outside of a site directory, so skip config/template init.
			if cmd.Name() == "version" {
				return nil
			}

			cfg, err := build.InitConfig()
			if err != nil {
				return err
			}

			templateCache, err := build.InitTemplateCache(cfg, app.Logger, themesFS)
			if err != nil {
				return err
			}

			app.Config = cfg
			app.TemplateCache = templateCache

			return nil
		},
	}

	rootCmd.AddCommand(newBuildCommand(app, themesFS))
	rootCmd.AddCommand(newInitCommand(app))
	rootCmd.AddCommand(newServeCommand(app))
	rootCmd.AddCommand(newVersionCommand(app))
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(themesFS embed.FS, buildInfo BuildInfo) {
	rootCmd := NewRootCommand(themesFS, buildInfo)
	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(buildInfo.Version),
		fang.WithCommit(buildInfo.Commit),
	); err != nil {
		os.Exit(1)
	}
}
