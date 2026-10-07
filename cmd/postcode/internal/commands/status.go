package commands

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
			if isOffline(v, cmd) {
				count, cachePath := postcode.GetCacheStats()
				offlineReport := map[string]any{
					"mode":             "offline",
					"cache_path":       cachePath,
					"cached_buildings": count,
				}
				return PrintOutput(cmd, v, offlineReport, func(w io.Writer) error {
					_, _ = fmt.Fprintln(w, "============================================================")
					_, _ = fmt.Fprintln(w, "Mode:             OFFLINE (network calls disabled)")
					_, _ = fmt.Fprintf(w, "Local Cache:      %s\n", cachePath)
					_, _ = fmt.Fprintf(w, "Cached Buildings: %d entries\n", count)
					_, _ = fmt.Fprintln(w, "============================================================")
					return nil
				})
			}

			client, err := buildClient(v)
			if err != nil {
				return err
			}

			start := time.Now()
			// Probe gateway operational health via GET /healthz
			healthErr := client.Health(cmd.Context())
			latency := time.Since(start)

			apiKey := getAPIKey(v)
			baseURL := getBaseURL(v)
			if baseURL == "" {
				baseURL = postcode.DefaultBaseURL
			}

			st := GatewayStatus{
				BaseURL:   baseURL,
				Healthy:   healthErr == nil,
				LatencyMs: float64(latency.Microseconds()) / 1000.0,
				APIKeySet: apiKey != "",
			}

			if healthErr != nil {
				st.ServerError = healthErr.Error()
			} else if apiKey != "" {
				// Probe authentication and retrieve rate quota headers.
				// FC-01-A01-KP-27 is supported in both sandbox and production tiers.
				if _, authErr := client.Lookup(cmd.Context(), "FC-01-A01-KP-27", postcode.Level1); authErr != nil {
					st.ServerError = fmt.Sprintf("API key check failed: %v", authErr)
				}
			}

			if rl := client.RateLimit(); rl != nil {
				st.RateLimit = &RateLimit{
					Limit:     rl.Limit,
					Remaining: rl.Remaining,
					ResetAt:   rl.ResetAt,
				}
			}

			return PrintOutput(cmd, v, st, func(w io.Writer) error {
				_, _ = fmt.Fprintln(w, "============================================================")
				_, _ = fmt.Fprintf(w, "Gateway Endpoint: %s\n", st.BaseURL)
				if st.Healthy {
					_, _ = fmt.Fprintf(w, "Status:           HEALTHY (%.2fms latency)\n", st.LatencyMs)
				} else {
					_, _ = fmt.Fprintf(w, "Status:           UNREACHABLE / ERROR\n")
					_, _ = fmt.Fprintf(w, "Error:            %s\n", st.ServerError)
				}
				if st.APIKeySet {
					_, _ = fmt.Fprintf(w, "API Key:          CONFIGURED\n")
				} else {
					_, _ = fmt.Fprintf(w, "API Key:          NONE (public free endpoints only)\n")
				}
				if st.RateLimit != nil {
					_, _ = fmt.Fprintf(w, "Rate Quota:       %d / %d remaining\n", st.RateLimit.Remaining, st.RateLimit.Limit)
					_, _ = fmt.Fprintf(w, "Reset At:         %s\n", st.RateLimit.ResetAt.Format(time.RFC3339))
				}
				_, _ = fmt.Fprintln(w, "============================================================")
				return nil
			})
		},
	}

	return cmd
}
