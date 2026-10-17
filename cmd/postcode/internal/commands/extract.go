package commands

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ExtractOutput struct {
	Total   int                          `json:"total"`
	Results []postcode.AddressExtraction `json:"results"`
}

func (eo ExtractOutput) CSVHeader() []string {
	return []string{"raw", "postcode", "postcode_valid", "legacy_postcode", "suggested_prefix", "state_code", "state_name", "lga_code", "lga_name", "locality", "street", "landmark", "latitude", "longitude", "confidence"}
}

func (eo ExtractOutput) CSVRows() [][]string {
	rows := make([][]string, len(eo.Results))
	for i, r := range eo.Results {
		validStr := "false"
		if r.PostcodeValid {
			validStr = "true"
		}
		latStr, lngStr := "", ""
		if r.Latitude != 0 || r.Longitude != 0 {
			latStr = fmt.Sprintf("%.6f", r.Latitude)
			lngStr = fmt.Sprintf("%.6f", r.Longitude)
		}
		rows[i] = []string{
			r.Raw,
			r.Postcode,
			validStr,
			r.LegacyPostcode,
			r.SuggestedPrefix,
			r.StateCode,
			r.StateName,
			r.LGACode,
			r.LGAName,
			r.Locality,
			r.Street,
			r.Landmark,
			latStr,
			lngStr,
			r.Confidence,
		}
	}
	return rows
}

// NewExtractCmd creates the extract subcommand for unstructured Nigerian address disambiguation.
func NewExtractCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract [addresses...]",
		Short: "Extract postcodes, state, LGA, locality, street, and landmarks from unstructured Nigerian addresses",
		Long: `Extract runs the 5-stage Nigerian address disambiguation pipeline on raw, unstructured
address strings. It identifies modern 11-character postcodes (with auto-repair), deprecated
6-digit numeric codes, Nigerian states, LGAs, neighborhoods/estates, streets, and landmarks.`,
		Example: `  postcode extract "No. 12, Adetokunbo Ademola Street, Opposite First Bank, Wuse II, Abuja FCT"
  postcode extract "10 Obafemi Awolowo Way, Ikeja, Lagos 100001" --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			out := ExtractOutput{
				Total:   len(inputs),
				Results: make([]postcode.AddressExtraction, 0, len(inputs)),
			}

			for _, raw := range inputs {
				out.Results = append(out.Results, postcode.ExtractAddress(raw))
			}

			return PrintOutput(cmd, v, out, func(w io.Writer) error {
				for i, r := range out.Results {
					if i > 0 {
						_, _ = fmt.Fprintln(w, "----------------------------------------------------------------")
					}
					_, _ = fmt.Fprintf(w, "Input:            %s\n", r.Raw)
					if r.Postcode != "" {
						_, _ = fmt.Fprintf(w, "Postcode:         %s (valid: %v)\n", r.Postcode, r.PostcodeValid)
					}
					if r.LegacyPostcode != "" {
						_, _ = fmt.Fprintf(w, "Legacy Postcode:  %s\n", r.LegacyPostcode)
					}
					if r.SuggestedPrefix != "" {
						_, _ = fmt.Fprintf(w, "Suggested Prefix: %s\n", r.SuggestedPrefix)
					}
					if r.StateCode != "" {
						_, _ = fmt.Fprintf(w, "State:            %s (%s)\n", r.StateName, r.StateCode)
					}
					if r.LGACode != "" {
						_, _ = fmt.Fprintf(w, "LGA:              %s (%s)\n", r.LGAName, r.LGACode)
					}
					if r.Locality != "" {
						_, _ = fmt.Fprintf(w, "Locality:         %s\n", r.Locality)
					}
					if r.Street != "" {
						_, _ = fmt.Fprintf(w, "Street:           %s\n", r.Street)
					}
					if r.Landmark != "" {
						_, _ = fmt.Fprintf(w, "Landmark:         %s\n", r.Landmark)
					}
					if r.Latitude != 0 || r.Longitude != 0 {
						_, _ = fmt.Fprintf(w, "Coordinates:      %.6f, %.6f\n", r.Latitude, r.Longitude)
					}
					_, _ = fmt.Fprintf(w, "Confidence:       %s\n", r.Confidence)
				}
				return nil
			})
		},
	}

	return cmd
}
