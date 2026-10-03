package commands

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type NearbyOutputWrapper struct {
	Latitude  float64               `json:"latitude"`
	Longitude float64               `json:"longitude"`
	RadiusM   float64               `json:"radius_m"`
	Count     int                   `json:"count"`
	Results   []postcode.NearbyUnit `json:"results"`
}

func (nw NearbyOutputWrapper) CSVHeader() []string {
	return []string{"postcode", "distance_m", "confidence", "state_name", "lga_name", "address"}
}

func (nw NearbyOutputWrapper) CSVRows() [][]string {
	rows := make([][]string, len(nw.Results))
	for i, u := range nw.Results {
		rows[i] = []string{
			u.Postcode,
			fmt.Sprintf("%.1f", u.DistanceM),
			u.Confidence,
			u.StateName,
			u.LGAName,
			u.Address,
		}
	}
	return rows
}

// NewNearbyCmd creates the nearby subcommand.
func NewNearbyCmd(v *viper.Viper) *cobra.Command {
	var lat float64
	var lng float64
	var radius float64

	cmd := &cobra.Command{
		Use:   "nearby",
		Short: "Search for active postcode units within a radius of GPS coordinates",
		Long: `Nearby queries the NIPOST gateway for registered building units within a specified
radius (default 300 meters) of a given geographic coordinate.`,
		Example: `  postcode nearby --lat 6.6018 --lng 3.3515
  postcode nearby --lat 9.0579 --lng 7.4951 --radius 150 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if lat == 0 && lng == 0 {
				return fmt.Errorf("both --lat and --lng must be provided")
			}

			client, err := buildClient(v)
			if err != nil {
				return err
			}

			params := postcode.NearbyParams{
				Latitude:  lat,
				Longitude: lng,
				RadiusM:   radius,
			}

			resp, err := client.Nearby(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("nearby search failed: %w", err)
			}

			out := NearbyOutputWrapper{
				Latitude:  lat,
				Longitude: lng,
				RadiusM:   radius,
				Count:     len(resp.Results),
				Results:   resp.Results,
			}

			return PrintOutput(cmd, v, out, func(w io.Writer) error {
				_, _ = fmt.Fprintf(w, "Nearby units within %.0fm of (%.6f, %.6f) — %d results:\n", radius, lat, lng, len(resp.Results))
				_, _ = fmt.Fprintln(w, "--------------------------------------------------------------------------------")
				for _, u := range resp.Results {
					_, _ = fmt.Fprintf(w, "• %-16s | %5.1fm away [%-6s] | %s, %s | %s\n",
						u.Postcode, u.DistanceM, u.Confidence, u.LGAName, u.StateName, u.Address)
				}
				return nil
			})
		},
	}

	cmd.Flags().Float64Var(&lat, "lat", 0, "geographic latitude (-90 to +90)")
	cmd.Flags().Float64Var(&lng, "lng", 0, "geographic longitude (-180 to +180)")
	cmd.Flags().Float64Var(&radius, "radius", 300, "search radius in meters (max 300)")

	_ = cmd.MarkFlagRequired("lat")
	_ = cmd.MarkFlagRequired("lng")

	return cmd
}
