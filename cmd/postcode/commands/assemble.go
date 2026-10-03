package commands

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewAssembleCmd creates the assemble subcommand.
func NewAssembleCmd(v *viper.Viper) *cobra.Command {
	var state, lga, district, area, unit string
	var online bool

	cmd := &cobra.Command{
		Use:   "assemble",
		Short: "Assemble 5 administrative segments into a canonical Nigerian postcode",
		Long: `Assemble combines the 5 structural components of a postcode:
  - State (2 letters)
  - LGA (2 digits, 01-99)
  - District (3 alphanumeric characters)
  - Area (2 letters)
  - Building Unit (2 digits, 01-99)

Operates offline with zero allocations by default, or calls the NIPOST gateway with --online.`,
		Example: `  postcode assemble --state EK --lga 01 --district A03 --area FK --unit 01
  postcode assemble --state LA --lga 11 --district W06 --area TC --unit 10 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			segs := postcode.Segments{
				State:    state,
				LGA:      lga,
				District: district,
				Area:     area,
				Unit:     unit,
			}

			if online {
				client, err := buildClient(v)
				if err != nil {
					return err
				}
				resp, err := client.Assemble(cmd.Context(), segs)
				if err != nil {
					return fmt.Errorf("online assembly failed: %w", err)
				}
				return PrintOutput(cmd, v, resp, func(w io.Writer) error {
					_, _ = fmt.Fprintf(w, "Postcode: %s\n", resp.Postcode)
					_, _ = fmt.Fprintf(w, "Display:  %s\n", resp.Display)
					_, _ = fmt.Fprintf(w, "Compact:  %s\n", resp.Compact)
					return nil
				})
			}

			// Offline zero-alloc assembly
			rawStr := fmt.Sprintf("%s%s%s%s%s", state, lga, district, area, unit)
			p, err := postcode.Parse(rawStr)
			if err != nil {
				return fmt.Errorf("assembly failed: %w", err)
			}

			res := postcode.AssembledPostcode{
				Postcode: p.Formatted(),
				Display:  p.String(),
				Compact:  p.Raw(),
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				_, _ = fmt.Fprintf(w, "Postcode: %s\n", res.Postcode)
				_, _ = fmt.Fprintf(w, "Display:  %s\n", res.Display)
				_, _ = fmt.Fprintf(w, "Compact:  %s\n", res.Compact)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "2-letter state code (e.g. EK, LA, FC)")
	cmd.Flags().StringVar(&lga, "lga", "", "2-digit LGA code (e.g. 01, 11, 03)")
	cmd.Flags().StringVar(&district, "district", "", "3-character district code (e.g. A03, W06)")
	cmd.Flags().StringVar(&area, "area", "", "2-letter area code (e.g. FK, TC, AG)")
	cmd.Flags().StringVar(&unit, "unit", "", "2-digit building unit code (e.g. 01, 10, 12)")
	cmd.Flags().BoolVar(&online, "online", false, "assemble via remote NIPOST gateway API")

	_ = cmd.MarkFlagRequired("state")
	_ = cmd.MarkFlagRequired("lga")
	_ = cmd.MarkFlagRequired("district")
	_ = cmd.MarkFlagRequired("area")
	_ = cmd.MarkFlagRequired("unit")

	return cmd
}
