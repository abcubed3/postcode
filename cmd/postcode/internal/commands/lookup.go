package commands

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type LookupOutputItem struct {
	Input                 string                          `json:"input"`
	Postcode              string                          `json:"postcode"`
	Valid                 bool                            `json:"valid"`
	Level                 int                             `json:"level"`
	Source                string                          `json:"source"`
	AdministrativeAddress *postcode.AdministrativeAddress `json:"administrative_address,omitempty"`
	RecentHouseAddress    *postcode.RecentHouseAddress    `json:"recent_house_address,omitempty"`
	BuildingUseStatus     string                          `json:"building_use_status,omitempty"`
	OtherBuildingInfo     map[string]any                  `json:"other_building_info,omitempty"`
	PointGeometry         *postcode.PointGeometry         `json:"point_geometry,omitempty"`
	Error                 string                          `json:"error,omitempty"`
}

type LookupResult struct {
	Total   int                `json:"total"`
	Results []LookupOutputItem `json:"results"`
}

// NewLookupCmd creates the lookup subcommand.
func NewLookupCmd(v *viper.Viper) *cobra.Command {
	var level int
	var offlineFallback bool

	cmd := &cobra.Command{
		Use:   "lookup [postcodes...]",
		Short: "Query NIPOST gateway for graded postcode attributes (Levels 1–5)",
		Long: `Lookup queries the official NIPOST gateway API for graded postcode attributes:
  - Level 1 (Free): Validity status (valid: true/false) and canonical format.
  - Level 2 (Commercial): Administrative boundaries, state, LGA, and street name.
  - Level 3 (Commercial): Official building use status (residential, commercial, mixed).
  - Level 4 (Enterprise): Detailed building metadata and administrative attributes.
  - Level 5 (Enterprise): High-precision GeoJSON point geometry and coordinates.

If the gateway is unreachable or lacks an API key, setting --offline-fallback (default: true)
will gracefully enrich the response using local reference data.`,
		Example: `  postcode lookup EK-01-A03-FK-01 --level 1
  postcode lookup "LA 11 W06 TC 10" --level 2 -o json
  postcode lookup FC-03-B06-AG-12 --level 3 --api-key nipost_live_...
  postcode lookup EK-01-A03-FK-01 --level 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			client, err := buildClient(v)
			if err != nil {
				return err
			}

			if level < 1 || level > 5 {
				return fmt.Errorf("invalid lookup level %d: choose 1 (validity), 2 (address), 3 (building use), 4 (building metadata), or 5 (geometry)", level)
			}

			res := LookupResult{
				Total:   len(inputs),
				Results: make([]LookupOutputItem, 0, len(inputs)),
			}

			offlineMode := isOffline(v, cmd)

			for _, raw := range inputs {
				if offlineMode {
					p, parseErr := postcode.Parse(raw)
					if parseErr != nil {
						res.Results = append(res.Results, LookupOutputItem{
							Input: raw,
							Valid: false,
							Error: parseErr.Error(),
						})
						continue
					}
					loc := p.Location()
					res.Results = append(res.Results, LookupOutputItem{
						Input:    raw,
						Postcode: p.Formatted(),
						Valid:    true,
						Level:    level,
						Source:   "offline",
						AdministrativeAddress: &postcode.AdministrativeAddress{
							State:     p.State(),
							StateName: loc.StateName,
							LGA:       p.LGA(),
							LGAName:   loc.LGAName,
							District:  p.District(),
							Area:      p.Area(),
							Unit:      p.BuildingUnit(),
							Zone:      loc.Zone,
						},
						RecentHouseAddress: &postcode.RecentHouseAddress{
							Address: loc.Address,
						},
						PointGeometry: &postcode.PointGeometry{
							Type:        "Point",
							Coordinates: []float64{loc.Longitude, loc.Latitude},
						},
					})
					continue
				}

				resp, lookupErr := client.Lookup(cmd.Context(), raw, postcode.LookupLevel(level))
				if lookupErr != nil {
					if offlineFallback {
						// Offline graceful fallback
						p, parseErr := postcode.Parse(raw)
						if parseErr != nil {
							res.Results = append(res.Results, LookupOutputItem{
								Input: raw,
								Valid: false,
								Error: parseErr.Error(),
							})
							continue
						}
						loc := p.Location()
						res.Results = append(res.Results, LookupOutputItem{
							Input:    raw,
							Postcode: p.Formatted(),
							Valid:    true,
							Level:    level,
							Source:   "offline_fallback",
							AdministrativeAddress: &postcode.AdministrativeAddress{
								State:     p.State(),
								StateName: loc.StateName,
								LGA:       p.LGA(),
								LGAName:   loc.LGAName,
								District:  p.District(),
								Area:      p.Area(),
								Unit:      p.BuildingUnit(),
								Zone:      loc.Zone,
							},
							RecentHouseAddress: &postcode.RecentHouseAddress{
								Address: loc.Address,
							},
							PointGeometry: &postcode.PointGeometry{
								Type:        "Point",
								Coordinates: []float64{loc.Longitude, loc.Latitude},
							},
						})
						continue
					}
					return fmt.Errorf("lookup failed for %q: %w", raw, lookupErr)
				}

				res.Results = append(res.Results, LookupOutputItem{
					Input:                 raw,
					Postcode:              resp.Postcode,
					Valid:                 resp.Valid,
					Level:                 level,
					Source:                "gateway",
					AdministrativeAddress: resp.AdministrativeAddress,
					RecentHouseAddress:    resp.RecentHouseAddress,
					BuildingUseStatus:     resp.BuildingUseStatus,
					OtherBuildingInfo:     resp.OtherBuildingInfo,
					PointGeometry:         resp.PointGeometry,
				})
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for i, r := range res.Results {
					if i > 0 {
						_, _ = fmt.Fprintln(w, "------------------------------------------------------------")
					}
					status := "VALID"
					if !r.Valid {
						status = "INVALID"
					}
					displayCode := r.Postcode
					if displayCode == "" {
						displayCode = r.Input
					}
					if r.Source != "" {
						_, _ = fmt.Fprintf(w, "Postcode:      %s [%s] (Source: %s, Level: %d)\n", displayCode, status, r.Source, r.Level)
					} else {
						_, _ = fmt.Fprintf(w, "Postcode:      %s [%s]\n", displayCode, status)
					}
					if r.Error != "" {
						_, _ = fmt.Fprintf(w, "Error:         %s\n", r.Error)
					}
					if r.AdministrativeAddress != nil {
						aa := r.AdministrativeAddress
						_, _ = fmt.Fprintf(w, "State:         %s (%s)\n", aa.StateName, aa.State)
						_, _ = fmt.Fprintf(w, "LGA:           %s (%s)\n", aa.LGAName, aa.LGA)
						if aa.DistrictName != "" {
							_, _ = fmt.Fprintf(w, "District:      %s (%s)\n", aa.DistrictName, aa.District)
						} else if aa.District != "" {
							_, _ = fmt.Fprintf(w, "District:      %s\n", aa.District)
						}
						if aa.AreaName != "" {
							_, _ = fmt.Fprintf(w, "Area:          %s (%s)\n", aa.AreaName, aa.Area)
						} else if aa.Area != "" {
							_, _ = fmt.Fprintf(w, "Area:          %s\n", aa.Area)
						}
						if aa.Zone != "" {
							_, _ = fmt.Fprintf(w, "Zone:          %s\n", aa.Zone)
						}
					}
					if r.RecentHouseAddress != nil {
						addr := r.RecentHouseAddress.Address
						if addr == "" {
							addr = r.RecentHouseAddress.Recent
						}
						if addr != "" {
							_, _ = fmt.Fprintf(w, "Address:       %s\n", addr)
						}
					}
					if r.BuildingUseStatus != "" {
						_, _ = fmt.Fprintf(w, "Building Use:  %s\n", r.BuildingUseStatus)
					}
					if len(r.OtherBuildingInfo) > 0 {
						_, _ = fmt.Fprintf(w, "Building Info: %v\n", r.OtherBuildingInfo)
					}
					if r.PointGeometry != nil && len(r.PointGeometry.Coordinates) >= 2 {
						_, _ = fmt.Fprintf(w, "Coordinates:   Lat %.6f, Lng %.6f\n", r.PointGeometry.Coordinates[1], r.PointGeometry.Coordinates[0])
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().IntVarP(&level, "level", "l", 1, "lookup grade depth (1-5; commercial keys support 1-3, 4-5 are enterprise/restricted)")
	cmd.Flags().BoolVar(&offlineFallback, "offline-fallback", true, "fall back to local reference data if gateway is unreachable")

	return cmd
}
