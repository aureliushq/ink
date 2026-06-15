package cmd

import (
	"errors"
	"net/http"

	"github.com/aureliushq/ink/internal/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
func newServeCommand(app *App) *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve your static site locally",
		Long:  `Serve your static site locally in at http://localhost:8782 with live reloading.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}
			port, err := cmd.Flags().GetInt64("port")
			if err != nil {
				return nil
			}

			srv := server.NewServer(host, port)

			app.Logger.Info("Starting server...", "addr", srv.Addr)
			err = srv.ListenAndServe()
			if !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}

	serveCmd.Flags().String("host", "localhost", "Host for the server, defaults to localhost")
	serveCmd.Flags().Int64("port", 8782, "Port for the server, defaults to 8782")
	return serveCmd
}
