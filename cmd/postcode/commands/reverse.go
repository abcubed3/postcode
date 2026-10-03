package cmd

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewReverseCmd creates the reverse geocoding subcommand.
func NewReverseCmd(v *viper.Viper) *cobra.Command {
	var lat float64
	var lng float64
	var maxDist float64

	cmd := &cobra.Command{
		Use:   "reverse",
		Short: "Reverse geocode GPS coordinates to the nearest active postcode unit",
		Long: `Reverse snaps a geographic coordinate pair (latitude, longitude) to the nearest
registered Nigerian building unit postcode. Useful for emergency dispatch, field surveyors,
and mobile location capture.`,
		Example: `  postcode reverse --lat 6.6018 --lng 3.3515
  postcode reverse --lat 9.0579 --lng 7.4951 --max-dist 50 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if lat == 0 && lng == 0 {
				return fmt.Errorf("both --lat and --lng must be provided")
			}

			client, err := buildClient(v)
			if err != nil {
				return err
			}

			params := postcode.ReverseParams{
				Latitude:     lat,
				Longitude:    lng,
				MaxDistanceM: maxDist,
			}

			resp, err := client.Reverse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("reverse geocode failed: %w", err)
			}

			return PrintOutput(cmd, v, resp, func(w io.Writer) error {
				if !resp.Found {
					fmt.Fprintf(w, "No active postcode unit found within %.1fm of (%.6f, %.6f).\n", resp.RadiusM, lat, lng)
					if resp.Message != "" {
						fmt.Fprintf(w, "Note: %s\n", resp.Message)
					}
					return nil
				}

				fmt.Fprintf(w, "Snapping (%.6f, %.6f) -> %s\n", lat, lng, resp.Unit.Postcode)
				fmt.Fprintln(w, "------------------------------------------------------------")
				fmt.Fprintf(w, "Postcode:      %s\n", resp.Unit.Postcode)
				fmt.Fprintf(w, "Distance:      %.1f meters\n", resp.Unit.DistanceM)
				fmt.Fprintf(w, "Confidence:    %s\n", resp.Unit.Confidence)
				fmt.Fprintf(w, "State:         %s\n", resp.State)
				fmt.Fprintf(w, "LGA:           %s\n", resp.Unit.LGAName)
				fmt.Fprintf(w, "District:      %s\n", resp.District)
				fmt.Fprintf(w, "Area:          %s\n", resp.Area)
				if resp.Unit.Address != "" {
					fmt.Fprintf(w, "Address:       %s\n", resp.Unit.Address)
				}
				return nil
			})
		},
	}

	cmd.Flags().Float64Var(&lat, "lat", 0, "geographic latitude (-90 to +90)")
	cmd.Flags().Float64Var(&lng, "lng", 0, "geographic longitude (-180 to +180)")
	cmd.Flags().Float64Var(&maxDist, "max-dist", 25, "maximum snap distance in meters (max 250)")

	_ = cmd.MarkFlagRequired("lat")
	_ = cmd.MarkFlagRequired("lng")

	return cmd
}
