package cmd

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ParsedPostcode struct {
	Input        string `json:"input"`
	Formatted    string `json:"formatted"`
	Compact      string `json:"compact"`
	Spaced       string `json:"spaced"`
	StateCode    string `json:"state_code"`
	StateName    string `json:"state_name,omitempty"`
	StateCapital string `json:"state_capital,omitempty"`
	LGACode      string `json:"lga_code"`
	LGAName      string `json:"lga_name,omitempty"`
	District     string `json:"district"`
	Area         string `json:"area"`
	Unit         string `json:"unit"`
	Zone         string `json:"zone,omitempty"`
}

type ParseResult struct {
	Total   int              `json:"total"`
	Results []ParsedPostcode `json:"results"`
}

func (pr ParseResult) CSVHeader() []string {
	return []string{"input", "formatted", "compact", "spaced", "state_code", "state_name", "lga_code", "lga_name", "district", "area", "unit", "zone"}
}

func (pr ParseResult) CSVRows() [][]string {
	rows := make([][]string, len(pr.Results))
	for i, r := range pr.Results {
		rows[i] = []string{
			r.Input, r.Formatted, r.Compact, r.Spaced,
			r.StateCode, r.StateName, r.LGACode, r.LGAName,
			r.District, r.Area, r.Unit, r.Zone,
		}
	}
	return rows
}

// NewParseCmd creates the parse subcommand.
func NewParseCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse [postcodes...]",
		Short: "Parse Nigerian postcodes into structural segments and administrative metadata",
		Long: `Parse breaks down one or more postcodes into their 5 official administrative
segments (State, LGA, District, Area, and Building Unit) and enriches them with
State names, LGA names, capitals, and geopolitical zones from the built-in registry.`,
		Example: `  postcode parse EK-01-A03-FK-01
  postcode parse "LA 11 W06 TC 10" -o json
  cat codes.txt | postcode parse -o csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			res := ParseResult{
				Total:   len(inputs),
				Results: make([]ParsedPostcode, 0, len(inputs)),
			}

			for _, raw := range inputs {
				p, parseErr := postcode.Parse(raw)
				if parseErr != nil {
					return fmt.Errorf("parsing %q: %w", raw, parseErr)
				}

				loc := p.Location()

				item := ParsedPostcode{
					Input:        raw,
					Formatted:    p.Formatted(),
					Compact:      p.Raw(),
					Spaced:       p.String(),
					StateCode:    p.State(),
					StateName:    loc.StateName,
					LGACode:      p.LGA(),
					LGAName:      loc.LGAName,
					District:     p.District(),
					Area:         p.Area(),
					Unit:         p.BuildingUnit(),
					Zone:         loc.Zone,
				}

				if st, ok := postcode.NigerianStates[p.State()]; ok {
					item.StateCapital = st.Capital
					if item.Zone == "" {
						item.Zone = st.Zone
					}
				}

				res.Results = append(res.Results, item)
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for i, r := range res.Results {
					if i > 0 {
						fmt.Fprintln(w, "------------------------------------------------------------")
					}
					fmt.Fprintf(w, "Postcode:      %s\n", r.Formatted)
					fmt.Fprintf(w, "Compact:       %s\n", r.Compact)
					fmt.Fprintf(w, "Spaced:        %s\n", r.Spaced)
					fmt.Fprintf(w, "State:         %s (%s)\n", r.StateName, r.StateCode)
					if r.StateCapital != "" {
						fmt.Fprintf(w, "Capital:       %s\n", r.StateCapital)
					}
					if r.LGAName != "" {
						fmt.Fprintf(w, "LGA:           %s (%s)\n", r.LGAName, r.LGACode)
					} else {
						fmt.Fprintf(w, "LGA Code:      %s\n", r.LGACode)
					}
					fmt.Fprintf(w, "District:      %s\n", r.District)
					fmt.Fprintf(w, "Area:          %s\n", r.Area)
					fmt.Fprintf(w, "Building Unit: %s\n", r.Unit)
					if r.Zone != "" {
						fmt.Fprintf(w, "Zone:          %s\n", r.Zone)
					}
				}
				return nil
			})
		},
	}

	return cmd
}
