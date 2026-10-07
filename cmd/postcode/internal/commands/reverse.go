package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

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
		Use:   "reverse [lat lng | lat,lng]",
		Short: "Reverse geocode GPS coordinates to the nearest active postcode unit",
		Long: `Reverse snaps a geographic coordinate pair (latitude, longitude) to the nearest
registered Nigerian building unit postcode. Useful for emergency dispatch, field surveyors,
and mobile location capture.`,
		Example: `  postcode reverse 6.6018 3.3515 --apikey <key>
  postcode reverse --lat 6.6018 --lng 3.3515 --apikey <key>
  postcode reverse 6.6018,3.3515 --max-dist 50 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 2 {
				if pLat, err := strconv.ParseFloat(args[0], 64); err == nil {
					lat = pLat
				}
				if pLng, err := strconv.ParseFloat(args[1], 64); err == nil {
					lng = pLng
				}
			} else if len(args) == 1 && strings.Contains(args[0], ",") {
				parts := strings.Split(args[0], ",")
				if len(parts) == 2 {
					if pLat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
						lat = pLat
					}
					if pLng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						lng = pLng
					}
				}
			}

			if lat == 0 && lng == 0 {
				return fmt.Errorf("specify coordinates via args (e.g. 'postcode reverse 6.6018 3.3515') or flags (--lat and --lng)")
			}

			var resp *postcode.ReverseResponse
			if isOffline(v, cmd) {
				resp = postcode.ReverseCoordinatesOffline(lat, lng, maxDist)
			} else {
				client, err := buildClient(v)
				if err != nil {
					return err
				}

				params := postcode.ReverseParams{
					Latitude:     lat,
					Longitude:    lng,
					MaxDistanceM: maxDist,
				}

				var apiErr error
				resp, apiErr = client.Reverse(cmd.Context(), params)
				if apiErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: Remote gateway reverse geocoding failed (%v). Falling back to offline resolution.\n", apiErr)
					resp = postcode.ReverseCoordinatesOffline(lat, lng, maxDist)
				}
			}

			return PrintOutput(cmd, v, resp, func(w io.Writer) error {
				if !resp.Found {
					_, _ = fmt.Fprintf(w, "No active postcode unit found within %.1fm of (%.6f, %.6f).\n", resp.RadiusM, lat, lng)
					if resp.Message != "" {
						_, _ = fmt.Fprintf(w, "Note: %s\n", resp.Message)
					}
					return nil
				}

				_, _ = fmt.Fprintf(w, "Snapping (%.6f, %.6f) -> %s\n", lat, lng, resp.Unit.Postcode)
				_, _ = fmt.Fprintln(w, "------------------------------------------------------------")
				_, _ = fmt.Fprintf(w, "Postcode:      %s\n", resp.Unit.Postcode)
				_, _ = fmt.Fprintf(w, "Distance:      %.1f meters\n", resp.Unit.DistanceM)
				_, _ = fmt.Fprintf(w, "Confidence:    %s\n", resp.Unit.Confidence)
				_, _ = fmt.Fprintf(w, "State:         %s\n", resp.State)
				_, _ = fmt.Fprintf(w, "LGA:           %s\n", resp.Unit.LGAName)
				_, _ = fmt.Fprintf(w, "District:      %s\n", resp.District)
				_, _ = fmt.Fprintf(w, "Area:          %s\n", resp.Area)
				if resp.Unit.Address != "" {
					_, _ = fmt.Fprintf(w, "Address:       %s\n", resp.Unit.Address)
				}
				return nil
			})
		},
	}

	cmd.Flags().Float64Var(&lat, "lat", 0, "geographic latitude (-90 to +90)")
	cmd.Flags().Float64Var(&lng, "lng", 0, "geographic longitude (-180 to +180)")
	cmd.Flags().Float64Var(&maxDist, "max-dist", 25, "maximum snap distance in meters (max 250)")

	return cmd
}
