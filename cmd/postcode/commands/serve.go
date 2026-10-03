package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/abcubed3/postcode/simulator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewServeCmd creates the serve subcommand to launch the local simulator.
func NewServeCmd(v *viper.Viper) *cobra.Command {
	var port string
	var host string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the local in-memory NIPOST gateway simulator server",
		Long: `Serve launches an in-memory mock NIPOST gateway server preloaded with all 21 official
test postcodes from docs.postcode.gov.ng.

Provides local REST endpoints identical to production:
  - GET  /v1/lookup
  - GET  /v1/search/autocomplete
  - GET  /v1/search/nearby
  - GET  /v1/search/reverse
  - POST /v1/assembly/assemble
  - GET  /v1/assembly/disassemble
  - GET  /healthz

Ideal for local testing, integration test suites, and offline development.`,
		Example: `  postcode serve
  postcode serve --port 9090 --host 0.0.0.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := fmt.Sprintf("%s:%s", host, port)

			server := &http.Server{
				Addr:         addr,
				Handler:      simulator.NewHandler(),
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  30 * time.Second,
			}

			w := cmd.OutOrStdout()
			fmt.Fprintln(w, "================================================================================")
			fmt.Fprintln(w, "🏛️  NIPOST Digital Postcode Gateway - Local Simulator")
			fmt.Fprintln(w, "📖 Official Docs: https://docs.postcode.gov.ng/")
			fmt.Fprintln(w, "================================================================================")
			fmt.Fprintf(w, "🚀 Server listening at: http://%s\n", addr)
			fmt.Fprintf(w, "🩺 Health endpoint:     http://%s/healthz\n", addr)
			fmt.Fprintln(w, "--------------------------------------------------------------------------------")
			fmt.Fprintf(w, "📋 Loaded %d official reference postcodes across 11 states\n", len(simulator.TestPostcodes))
			fmt.Fprintln(w, "Press Ctrl+C to stop.")

			serverErr := make(chan error, 1)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					serverErr <- err
				}
			}()

			select {
			case <-cmd.Context().Done():
				fmt.Fprintln(w, "\nShutting down simulator server gracefully...")
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				return server.Shutdown(shutdownCtx)
			case err := <-serverErr:
				return fmt.Errorf("simulator server error: %w", err)
			}
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "8080", "port for the mock server")
	cmd.Flags().StringVar(&host, "host", "localhost", "host address to bind to")

	return cmd
}
