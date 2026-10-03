package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewAutocompleteCmd creates the autocomplete subcommand.
func NewAutocompleteCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autocomplete <query>",
		Short: "Retrieve segment-aware suggestions for a partial postcode query",
		Long: `Autocomplete queries the NIPOST gateway for suggestions matching a partial postcode.
The response identifies the active segment being completed (e.g. State, LGA, District)
along with valid candidate codes and administrative labels.`,
		Example: `  postcode autocomplete "EK 01 A"
  postcode autocomplete "OG-14"
  postcode autocomplete "LA 11" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			client, err := buildClient(v)
			if err != nil {
				return err
			}

			resp, err := client.Autocomplete(cmd.Context(), query)
			if err != nil {
				return fmt.Errorf("autocomplete failed for %q: %w", query, err)
			}

			return PrintOutput(cmd, v, resp, func(w io.Writer) error {
				fmt.Fprintf(w, "Active Segment: %s (Query: %q, %d suggestions)\n", resp.Segment, query, len(resp.Suggestions))
				fmt.Fprintln(w, "------------------------------------------------------------")
				for _, s := range resp.Suggestions {
					fmt.Fprintf(w, "  %-12s  %s\n", s.Code, s.Label)
				}
				return nil
			})
		},
	}

	return cmd
}
