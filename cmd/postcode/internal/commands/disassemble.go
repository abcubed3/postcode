package commands

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DisassembleItem struct {
	Input    string `json:"input"`
	Postcode string `json:"postcode"`
	State    string `json:"state"`
	LGA      string `json:"lga"`
	District string `json:"district"`
	Area     string `json:"area"`
	Unit     string `json:"unit"`
}

type DisassembleResult struct {
	Total   int               `json:"total"`
	Results []DisassembleItem `json:"results"`
}

func (dr DisassembleResult) CSVHeader() []string {
	return []string{"input", "postcode", "state", "lga", "district", "area", "unit"}
}

func (dr DisassembleResult) CSVRows() [][]string {
	rows := make([][]string, len(dr.Results))
	for i, r := range dr.Results {
		rows[i] = []string{r.Input, r.Postcode, r.State, r.LGA, r.District, r.Area, r.Unit}
	}
	return rows
}

// NewDisassembleCmd creates the disassemble subcommand.
func NewDisassembleCmd(v *viper.Viper) *cobra.Command {
	var online bool

	cmd := &cobra.Command{
		Use:   "disassemble [postcodes...]",
		Short: "Decompose postcodes into individual administrative segment values",
		Long: `Disassemble extracts the individual State, LGA, District, Area, and Unit components
from one or more postcodes. Useful for relational database loading, ETL column splitting,
and administrative aggregation.

Operates offline with zero allocations by default, or calls the NIPOST gateway with --online.`,
		Example: `  postcode disassemble EK-01-A03-FK-01
  postcode disassemble "LA 11 W06 TC 10" -o json
  postcode disassemble EK01A03FK01 --online
  cat codes.txt | postcode disassemble -o csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			res := DisassembleResult{
				Total:   len(inputs),
				Results: make([]DisassembleItem, 0, len(inputs)),
			}

			if online {
				client, err := buildClient(v)
				if err != nil {
					return err
				}
				for _, raw := range inputs {
					segs, err := client.Disassemble(cmd.Context(), raw)
					if err != nil {
						return fmt.Errorf("online disassemble %q: %w", raw, err)
					}
					res.Results = append(res.Results, DisassembleItem{
						Input:    raw,
						Postcode: raw,
						State:    segs.State,
						LGA:      segs.LGA,
						District: segs.District,
						Area:     segs.Area,
						Unit:     segs.Unit,
					})
				}
			} else {
				for _, raw := range inputs {
					p, parseErr := postcode.Parse(raw)
					if parseErr != nil {
						return fmt.Errorf("disassembling %q: %w", raw, parseErr)
					}

					res.Results = append(res.Results, DisassembleItem{
						Input:    raw,
						Postcode: p.Formatted(),
						State:    p.State(),
						LGA:      p.LGA(),
						District: p.District(),
						Area:     p.Area(),
						Unit:     p.BuildingUnit(),
					})
				}
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, r := range res.Results {
					if _, err := fmt.Fprintf(w, "%-16s -> State: %-2s | LGA: %-2s | District: %-3s | Area: %-2s | Unit: %-2s\n", r.Postcode, r.State, r.LGA, r.District, r.Area, r.Unit); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&online, "online", false, "disassemble via remote NIPOST gateway API")

	return cmd
}
