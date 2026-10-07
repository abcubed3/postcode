package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)


type NearbyOutputWrapper struct {
	TargetPostcode string                `json:"target_postcode,omitempty"`
	Latitude       float64               `json:"latitude"`
	Longitude      float64               `json:"longitude"`
	RadiusM        float64               `json:"radius_m"`
	Count          int                   `json:"count"`
	Results        []postcode.NearbyUnit `json:"results"`
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
		Use:   "nearby [postcode]",
		Short: "Search for active postcode units within a radius of GPS coordinates or a postcode",
		Long: `Nearby queries the NIPOST gateway for registered building units within a specified
radius (default 300 meters) of a given geographic coordinate or a reference postcode.`,
		Example: `  postcode nearby LA-08-A86-RG-01 --apikey <key>
  postcode nearby --lat 6.6018 --lng 3.3515 --apikey <key>
  postcode nearby LA-08-A86-RG-01 --radius 150 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetCode string
			var loc postcode.Location
			if len(args) > 0 {
				p, parseErr := postcode.Parse(args[0])
				if parseErr != nil {
					return fmt.Errorf("invalid reference postcode %q: %w", args[0], parseErr)
				}
				targetCode = p.Formatted()
				loc = p.Location()
				if lat == 0 && lng == 0 {
					lat = loc.Latitude
					lng = loc.Longitude
				}
			}

			if isOffline(v, cmd) {
				return runOfflineNearby(cmd, v, targetCode, lat, lng, radius)
			}

			client, err := buildClient(v)
			if err != nil {
				return err
			}

			if (lat == 0 && lng == 0) || (targetCode != "" && loc.Precision > postcode.PrecisionBuilding) {
				if resolved, resErr := client.ResolveLocation(cmd.Context(), targetCode); resErr == nil && (resolved.Latitude != 0 || resolved.Longitude != 0) {
					lat = resolved.Latitude
					lng = resolved.Longitude
				}
			}

			if lat == 0 && lng == 0 {
				return fmt.Errorf("specify a reference postcode (e.g. 'postcode nearby LA-08-A86-RG-01') or coordinates via --lat and --lng")
			}

			params := postcode.NearbyParams{
				Latitude:  lat,
				Longitude: lng,
				RadiusM:   radius,
			}

			resp, err := client.Nearby(cmd.Context(), params)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Remote gateway nearby search failed (%v). Falling back to offline resolution.\n", err)
				return runOfflineNearby(cmd, v, targetCode, lat, lng, radius)
			}

			out := NearbyOutputWrapper{
				TargetPostcode: targetCode,
				Latitude:       lat,
				Longitude:      lng,
				RadiusM:        radius,
				Count:          len(resp.Results),
				Results:        resp.Results,
			}

			return PrintOutput(cmd, v, out, func(w io.Writer) error {
				if targetCode != "" {
					_, _ = fmt.Fprintf(w, "Nearby units within %.0fm of %s (%.6f, %.6f) — %d results:\n", radius, targetCode, lat, lng, len(resp.Results))
				} else {
					_, _ = fmt.Fprintf(w, "Nearby units within %.0fm of (%.6f, %.6f) — %d results:\n", radius, lat, lng, len(resp.Results))
				}
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

	return cmd
}

func runOfflineNearby(cmd *cobra.Command, v *viper.Viper, targetCode string, lat, lng, radius float64) error {
	if lat == 0 && lng == 0 && targetCode != "" {
		p, _ := postcode.Parse(targetCode)
		loc := p.Location()
		lat = loc.Latitude
		lng = loc.Longitude
	}
	if lat == 0 && lng == 0 {
		return fmt.Errorf("specify a reference postcode or coordinates via --lat and --lng")
	}
	units := postcode.SearchNearbyBuildingsOffline(lat, lng, radius)
	out := NearbyOutputWrapper{
		TargetPostcode: targetCode,
		Latitude:       lat,
		Longitude:      lng,
		RadiusM:        radius,
		Count:          len(units),
		Results:        units,
	}
	return PrintOutput(cmd, v, out, func(w io.Writer) error {
		if targetCode != "" {
			_, _ = fmt.Fprintf(w, "Nearby units (offline) within %.0fm of %s (%.6f, %.6f) — %d results:\n", radius, targetCode, lat, lng, len(units))
		} else {
			_, _ = fmt.Fprintf(w, "Nearby units (offline) within %.0fm of (%.6f, %.6f) — %d results:\n", radius, lat, lng, len(units))
		}
		for _, u := range units {
			_, _ = fmt.Fprintf(w, "%-16s  Dist: %-6.1fm  Conf: %-6s  Addr: %s\n", u.Postcode, u.DistanceM, u.Confidence, u.Address)
		}
		return nil
	})
}

