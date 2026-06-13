package cmd

import (
	"fmt"
	"log"

	"github.com/aureliushq/ink/internal/config"
	"github.com/aureliushq/ink/internal/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
func newServeCommand(cfg *config.Config) *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve your static site locally",
		Long:  `Serve your static site locally in at http://localhost:8782 with live reloading.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(cfg)
			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}
			port, err := cmd.Flags().GetInt64("port")
			if err != nil {
				return nil
			}

			srv := server.NewServer(host, port)
			if err := srv.ListenAndServe(); err != nil {
				log.Fatal(err)
				return err
			}

			return nil
		},
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serveCmd.Flags().String("host", "localhost", "Host for the server, defaults to localhost")
	serveCmd.Flags().Int64("port", 8782, "Port for the server, defaults to 8782")
	return serveCmd
}
