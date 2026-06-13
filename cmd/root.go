package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aureliushq/ink/internal/config"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
func NewRootCommand() *cobra.Command {
	cfg := config.NewConfig()

	rootCmd := &cobra.Command{
		Use:   "ink",
		Short: "Yet another super simple static site generator",
		Long: `Ink is yet another super simple static site generator.
Write content in markdown, bring your own HTML+CSS templates, deploy anywhere.
Supports CommonMark and GFM. Comes with syntax highlighting, footnotes and margin notes,
and more out-of-the-box.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Setup(cfg); err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		},
	}

	rootCmd.AddCommand(newBuildCommand(cfg))
	rootCmd.AddCommand(newInitCommand(cfg))
	rootCmd.AddCommand(newServeCommand(cfg))
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
