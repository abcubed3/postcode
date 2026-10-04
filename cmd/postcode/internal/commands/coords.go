package commands

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CoordinateEntry struct {
	Input     string  `json:"input"`
	Postcode  string  `json:"postcode"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Precision string  `json:"precision"`
	StateName string  `json:"state_name,omitempty"`
	LGAName   string  `json:"lga_name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type CoordsResult struct {
	Total   int               `json:"total"`
	Results []CoordinateEntry `json:"results"`
}

func (cr CoordsResult) CSVHeader() []string {
	return []string{"input", "postcode", "latitude", "longitude", "precision", "state_name", "lga_name", "address"}
}

func (cr CoordsResult) CSVRows() [][]string {
	rows := make([][]string, len(cr.Results))
	for i, r := range cr.Results {
		rows[i] = []string{
			r.Input, r.Postcode,
			fmt.Sprintf("%.6f", r.Latitude),
			fmt.Sprintf("%.6f", r.Longitude),
			r.Precision, r.StateName, r.LGAName, r.Address,
		}
	}
	return rows
}

// NewCoordsCmd creates the coords subcommand.
func NewCoordsCmd(v *viper.Viper) *cobra.Command {
	var online bool

	cmd := &cobra.Command{
		Use:   "coords [postcodes...]",
		Short: "Extract geographic coordinates (latitude, longitude) and precision tier",
		Long: `Coords resolves the exact or hierarchical geographic centroid for one or more postcodes.
By default, this operates completely offline using the built-in reference database of all
37 Nigerian administrative divisions, LGA centroids, and building unit benchmarks.

With --online, it enriches coordinates via the NIPOST gateway.`,
		Example: `  postcode coords EK-01-A03-FK-01
  postcode coords "LA 11 W06 TC 10" -o json
  postcode coords FC-03-B06-AG-12 --online`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			var client *postcode.Client
			if online {
				client, err = buildClient(v)
				if err != nil {
					return err
				}
			}

			res := CoordsResult{
				Total:   len(inputs),
				Results: make([]CoordinateEntry, 0, len(inputs)),
			}

			for _, raw := range inputs {
				var loc postcode.Location

				if online {
					resolvedLoc, resolveErr := client.ResolveLocation(cmd.Context(), raw)
					if resolveErr != nil {
						// Fallback to offline resolution if online fails
						var offErr error
						loc, offErr = postcode.ResolveLocation(raw)
						if offErr != nil {
							return fmt.Errorf("resolving coords for %q: %w", raw, resolveErr)
						}
					} else {
						loc = *resolvedLoc
					}
				} else {
					var offErr error
					loc, offErr = postcode.ResolveLocation(raw)
					if offErr != nil {
						return fmt.Errorf("resolving coords for %q: %w", raw, offErr)
					}
				}

				res.Results = append(res.Results, CoordinateEntry{
					Input:     raw,
					Postcode:  loc.Postcode,
					Latitude:  loc.Latitude,
					Longitude: loc.Longitude,
					Precision: loc.Precision.String(),
					StateName: loc.StateName,
					LGAName:   loc.LGAName,
					Address:   loc.Address,
				})
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, r := range res.Results {
					if r.Address != "" {
						if _, err := fmt.Fprintf(w, "%-16s -> Lat: %10.6f, Lng: %10.6f [%s] (%s - %s)\n", r.Postcode, r.Latitude, r.Longitude, r.Precision, r.Address, r.StateName); err != nil {
							return err
						}
					} else {
						if _, err := fmt.Fprintf(w, "%-16s -> Lat: %10.6f, Lng: %10.6f [%s] (%s, %s)\n",
							r.Postcode, r.Latitude, r.Longitude, r.Precision, r.LGAName, r.StateName); err != nil {
							return err
						}
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&online, "online", false, "enrich coordinates via remote NIPOST gateway")

	return cmd
}
