package commands

import (
	"fmt"
	"io"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ValidationRecord struct {
	Input    string `json:"input"`
	Valid    bool   `json:"valid"`
	Postcode string `json:"postcode,omitempty"`
	State    string `json:"state,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type ValidationResult struct {
	Total   int                `json:"total"`
	Valid   int                `json:"valid"`
	Invalid int                `json:"invalid"`
	Results []ValidationRecord `json:"results"`
}

func (vr ValidationResult) CSVHeader() []string {
	return []string{"input", "valid", "postcode", "state", "reason"}
}

func (vr ValidationResult) CSVRows() [][]string {
	rows := make([][]string, len(vr.Results))
	for i, r := range vr.Results {
		validStr := "false"
		if r.Valid {
			validStr = "true"
		}
		rows[i] = []string{r.Input, validStr, r.Postcode, r.State, r.Reason}
	}
	return rows
}

// NewValidateCmd creates the validate subcommand.
func NewValidateCmd(v *viper.Viper) *cobra.Command {
	var strict bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "validate [postcodes...]",
		Short: "Validate Nigerian postcode grammar, structure, and state codes",
		Long: `Validate checks whether one or more postcodes strictly conform to the official
11-character NIPOST digital alphanumeric standard (State LGA District Area Unit).

Inputs can be supplied as arguments or piped line-by-line via standard input.
Exit code is 0 if all postcodes are valid, or 1 if any invalid code is found.`,
		Example: `  postcode validate EK-01-A03-FK-01
  postcode validate "EK 01 A03 FK 01" "LA 11 W06 TC 10"
  cat postcodes.txt | postcode validate --quiet`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			res := ValidationResult{
				Total:   len(inputs),
				Results: make([]ValidationRecord, 0, len(inputs)),
			}

			for _, raw := range inputs {
				p, parseErr := postcode.Parse(raw)
				if parseErr != nil {
					res.Invalid++
					res.Results = append(res.Results, ValidationRecord{
						Input:  raw,
						Valid:  false,
						Reason: parseErr.Error(),
					})
					continue
				}

				if strict {
					if _, known := postcode.NigerianStates[p.State()]; !known {
						res.Invalid++
						res.Results = append(res.Results, ValidationRecord{
							Input:  raw,
							Valid:  false,
							State:  p.State(),
							Reason: fmt.Sprintf("unknown state code %q (not in 36 States + FCT)", p.State()),
						})
						continue
					}
				}

				res.Valid++
				res.Results = append(res.Results, ValidationRecord{
					Input:    raw,
					Valid:    true,
					Postcode: p.Formatted(),
					State:    p.State(),
				})
			}

			if !quiet {
				err = PrintOutput(cmd, v, res, func(w io.Writer) error {
					for _, r := range res.Results {
						if r.Valid {
							_, _ = fmt.Fprintf(w, "✓ %-16s -> %-16s (State: %s)\n", r.Input, r.Postcode, r.State)
						} else {
							_, _ = fmt.Fprintf(w, "✗ %-16s -> INVALID: %s\n", r.Input, r.Reason)
						}
					}
					if len(res.Results) > 1 {
						_, _ = fmt.Fprintf(w, "\nSummary: %d total, %d valid, %d invalid\n", res.Total, res.Valid, res.Invalid)
					}
					return nil
				})
				if err != nil {
					return err
				}
			}

			if res.Invalid > 0 {
				return fmt.Errorf("validation failed: %d of %d postcodes are invalid", res.Invalid, res.Total)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&strict, "strict", "s", true, "verify state code exists in the official Nigerian 37 state/FCT registry")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output, return non-zero exit code on invalid")

	return cmd
}
