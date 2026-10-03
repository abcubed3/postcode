package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type MapEntry struct {
	Input     string  `json:"input"`
	Postcode  string  `json:"postcode"`
	Provider  string  `json:"provider"`
	URL       string  `json:"url"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

type MapResult struct {
	Total   int        `json:"total"`
	Results []MapEntry `json:"results"`
}

func (mr MapResult) CSVHeader() []string {
	return []string{"input", "postcode", "provider", "url", "latitude", "longitude", "address"}
}

func (mr MapResult) CSVRows() [][]string {
	rows := make([][]string, len(mr.Results))
	for i, r := range mr.Results {
		rows[i] = []string{
			r.Input, r.Postcode, r.Provider, r.URL,
			fmt.Sprintf("%.6f", r.Latitude),
			fmt.Sprintf("%.6f", r.Longitude),
			r.Address,
		}
	}
	return rows
}

// NewMapCmd creates the map subcommand.
func NewMapCmd(v *viper.Viper) *cobra.Command {
	var provider string
	var directions bool
	var openBrowser bool
	var online bool

	cmd := &cobra.Command{
		Use:   "map [postcodes...]",
		Short: "Generate mapping and navigation URLs (Google Maps, Apple Maps, OSM)",
		Long: `Map generates universal web and app deep-links for one or more Nigerian postcodes.
Supported providers include Google Maps, Apple Maps, and OpenStreetMap.

Use --directions to generate turn-by-turn navigation route URLs.
Use --open to automatically launch the generated URL in your default system browser.`,
		Example: `  postcode map EK-01-A03-FK-01
  postcode map "LA 11 W06 TC 10" --directions
  postcode map FC-03-B06-AG-12 --provider apple
  postcode map EK-01-A03-FK-01 --open`,
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

			provider = strings.ToLower(provider)
			if provider != "google" && provider != "apple" && provider != "osm" {
				return fmt.Errorf("invalid provider %q: choose google, apple, or osm", provider)
			}

			res := MapResult{
				Total:   len(inputs),
				Results: make([]MapEntry, 0, len(inputs)),
			}

			for _, raw := range inputs {
				var loc postcode.Location
				if online {
					resolvedLoc, resolveErr := client.ResolveLocation(cmd.Context(), raw)
					if resolveErr == nil {
						loc = *resolvedLoc
					} else {
						loc, _ = postcode.ResolveLocation(raw)
					}
				} else {
					var offErr error
					loc, offErr = postcode.ResolveLocation(raw)
					if offErr != nil {
						return fmt.Errorf("resolving location for %q: %w", raw, offErr)
					}
				}

				var mapURL string
				switch provider {
				case "apple":
					mapURL = loc.AppleMapsURL()
				case "osm":
					mapURL = loc.OpenStreetMapURL()
				case "google":
					fallthrough
				default:
					if directions {
						mapURL = loc.GoogleMapsDirectionsURL()
					} else {
						mapURL = loc.GoogleMapsURL()
					}
				}

				res.Results = append(res.Results, MapEntry{
					Input:     raw,
					Postcode:  loc.Postcode,
					Provider:  provider,
					URL:       mapURL,
					Latitude:  loc.Latitude,
					Longitude: loc.Longitude,
					Address:   loc.Address,
				})

				if openBrowser && mapURL != "" {
					_ = openInBrowser(mapURL)
				}
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, r := range res.Results {
					fmt.Fprintf(w, "%-16s -> %s\n", r.Postcode, r.URL)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "google", "map provider: google, apple, or osm")
	cmd.Flags().BoolVarP(&directions, "directions", "d", false, "generate turn-by-turn navigation / directions URL")
	cmd.Flags().BoolVar(&openBrowser, "open", false, "open generated URL in the default web browser")
	cmd.Flags().BoolVar(&online, "online", false, "enrich location using remote NIPOST gateway before mapping")

	return cmd
}

func openInBrowser(targetURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	return cmd.Start()
}
