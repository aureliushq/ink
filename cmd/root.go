package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type App struct {
	Config *config.Config
	Logger *log.Logger
}

func newApp() *App {
	logger := log.NewWithOptions(os.Stdout, log.Options{ReportCaller: true, ReportTimestamp: true})
	return &App{
		Logger: logger,
	}
}

// rootCmd represents the base command when called without any subcommands
func NewRootCommand() *cobra.Command {
	app := newApp()

	rootCmd := &cobra.Command{
		Use:   "ink",
		Short: "Yet another super simple static site generator",
		Long: `Ink is yet another super simple static site generator.
Write content in markdown, bring your own HTML+CSS templates, deploy anywhere.
Supports CommonMark and GFM. Comes with syntax highlighting, footnotes and margin notes,
and more out-of-the-box.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.NewConfig()
			app.Config = cfg

			if err := config.Setup(app.Config); err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		},
	}

	rootCmd.AddCommand(newBuildCommand(app))
	rootCmd.AddCommand(newInitCommand(app))
	rootCmd.AddCommand(newServeCommand(app))
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewRootCommand()
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ink.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
