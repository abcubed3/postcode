package commands

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
	var dataFile string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the local in-memory NIPOST API simulator server",
		Long: `Serve launches an in-memory mock NIPOST API server preloaded with all 21 official
test postcodes from docs.postcode.gov.ng, or a custom JSON dataset.

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
  postcode serve --port 2340 --host 0.0.0.0
  postcode serve --data ./custom_postcodes.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataFile != "" {
				if err := simulator.LoadFile(dataFile); err != nil {
					return fmt.Errorf("loading custom postcodes from %s: %w", dataFile, err)
				}
			}

			addr := fmt.Sprintf("%s:%s", host, port)

			server := &http.Server{
				Addr:         addr,
				Handler:      simulator.NewHandler(),
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  30 * time.Second,
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, "================================================================================")
			_, _ = fmt.Fprintln(w, "🏛️ NIPOST Digital Postcode API - Local Simulator")
			_, _ = fmt.Fprintln(w, "📖 Official Docs: https://docs.postcode.gov.ng/")
			_, _ = fmt.Fprintln(w, "================================================================================")
			_, _ = fmt.Fprintf(w, "🚀 Local Server listening at: http://%s\n", addr)
			_, _ = fmt.Fprintf(w, "🩺 Health endpoint:     http://%s/healthz\n", addr)
			_, _ = fmt.Fprintln(w, "--------------------------------------------------------------------------------")
			_, _ = fmt.Fprintf(w, "📋 Loaded %d official reference postcodes across 11 states\n", len(simulator.TestPostcodes))
			_, _ = fmt.Fprintln(w, "Press Ctrl+C to stop.")

			serverErr := make(chan error, 1)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					serverErr <- err
				}
			}()

			select {
			case <-cmd.Context().Done():
				_, _ = fmt.Fprintln(w, "\nShutting down simulator server gracefully...")
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				return server.Shutdown(shutdownCtx)
			case err := <-serverErr:
				return fmt.Errorf("simulator server error: %w", err)
			}
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "2340", "port for the mock server")
	cmd.Flags().StringVar(&host, "host", "localhost", "host address to bind to")
	cmd.Flags().StringVar(&dataFile, "data", "", "path to custom JSON file containing test postcodes")

	return cmd
}
