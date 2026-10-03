package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type GatewayStatus struct {
	BaseURL     string     `json:"base_url"`
	Healthy     bool       `json:"healthy"`
	LatencyMs   float64    `json:"latency_ms"`
	APIKeySet   bool       `json:"api_key_set"`
	RateLimit   *RateLimit `json:"rate_limit,omitempty"`
	ServerError string     `json:"server_error,omitempty"`
}

type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"`
}

// NewStatusCmd creates the status/quota subcommand.
func NewStatusCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"quota", "ping"},
		Short:   "Check NIPOST gateway health, latency, and rate limit quotas",
		Long: `Status probes the configured NIPOST gateway base URL, measures round-trip
latency, validates API key authentication, and reports the current remaining quota
and rate limit reset time.`,
		Example: `  postcode status
  postcode status --base-url http://localhost:8080 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}

			start := time.Now()
			// Probe gateway with a standard public L1 lookup of canonical test code
			_, lookupErr := client.Lookup(cmd.Context(), "EK-01-A03-FK-01", postcode.Level1)
			latency := time.Since(start)

			apiKey := v.GetString("api-key")
			baseURL := v.GetString("base-url")
			if baseURL == "" {
				baseURL = postcode.DefaultBaseURL
			}

			st := GatewayStatus{
				BaseURL:   baseURL,
				Healthy:   lookupErr == nil,
				LatencyMs: float64(latency.Microseconds()) / 1000.0,
				APIKeySet: apiKey != "",
			}

			if lookupErr != nil {
				st.ServerError = lookupErr.Error()
			}

			if rl := client.RateLimit(); rl != nil {
				st.RateLimit = &RateLimit{
					Limit:     rl.Limit,
					Remaining: rl.Remaining,
					ResetAt:   rl.ResetAt,
				}
			}

			return PrintOutput(cmd, v, st, func(w io.Writer) error {
				fmt.Fprintln(w, "============================================================")
				fmt.Fprintf(w, "Gateway Endpoint: %s\n", st.BaseURL)
				if st.Healthy {
					fmt.Fprintf(w, "Status:           HEALTHY (%.2fms latency)\n", st.LatencyMs)
				} else {
					fmt.Fprintf(w, "Status:           UNREACHABLE / ERROR\n")
					fmt.Fprintf(w, "Error:            %s\n", st.ServerError)
				}
				if st.APIKeySet {
					fmt.Fprintf(w, "API Key:          CONFIGURED\n")
				} else {
					fmt.Fprintf(w, "API Key:          NONE (public free endpoints only)\n")
				}
				if st.RateLimit != nil {
					fmt.Fprintf(w, "Rate Quota:       %d / %d remaining\n", st.RateLimit.Remaining, st.RateLimit.Limit)
					fmt.Fprintf(w, "Reset At:         %s\n", st.RateLimit.ResetAt.Format(time.RFC3339))
				}
				fmt.Fprintln(w, "============================================================")
				return nil
			})
		},
	}

	return cmd
}
