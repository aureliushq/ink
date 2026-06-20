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

type App struct {
	Config        *config.Config
	Logger        *log.Logger
	TemplateCache *renderer.TemplateCache
}

func newApp() *App {
	logger := log.NewWithOptions(os.Stdout, log.Options{ReportCaller: true, ReportTimestamp: true})
	return &App{
		Logger: logger,
	}
}

// rootCmd represents the base command when called without any subcommands
func NewRootCommand(themesFS embed.FS) *cobra.Command {
	app := newApp()

	rootCmd := &cobra.Command{
		Use:   "ink",
		Short: "Yet another super simple static site generator",
		Long: `Ink is yet another super simple static site generator.
Write content in markdown, bring your own HTML+CSS templates, deploy anywhere.
Supports CommonMark and GFM. Comes with syntax highlighting, footnotes and margin notes,
and more out-of-the-box.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(themesFS embed.FS) {
	rootCmd := NewRootCommand(themesFS)
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
