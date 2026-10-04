package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DiagnoseOutput struct {
	Total   int                         `json:"total"`
	Valid   int                         `json:"valid"`
	Invalid int                         `json:"invalid"`
	Reports []postcode.DiagnosticReport `json:"reports"`
}

func (do DiagnoseOutput) CSVHeader() []string {
	return []string{"input", "valid", "clean_length", "normalized", "actionable_tip"}
}

func (do DiagnoseOutput) CSVRows() [][]string {
	rows := make([][]string, len(do.Reports))
	for i, r := range do.Reports {
		validStr := "false"
		if r.Valid {
			validStr = "true"
		}
		rows[i] = []string{
			r.Input,
			validStr,
			fmt.Sprintf("%d", r.CleanLength),
			r.Normalized,
			r.ActionableTip,
		}
	}
	return rows
}

// NewDiagnoseCmd creates the diagnose subcommand.
func NewDiagnoseCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnose [postcodes...]",
		Short: "Deeply analyze Nigerian postcode errors with segment suggestions",
		Long: `Diagnose performs segment-by-segment grammar, structural, and administrative
validation of Nigerian postcodes.

It pinpoints exact errors (State, LGA, District, Area, Unit, or Length), provides
closest candidate suggestions for typos, and generates actionable instructions
designed for humans and AI agents.`,
		Example: `  postcode diagnose "ZZ 01 A03 FK 01"
  postcode diagnose "EK 00 A03 FK 00"
  echo "LA 11 W06" | postcode diagnose -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			output := DiagnoseOutput{
				Total:   len(inputs),
				Reports: make([]postcode.DiagnosticReport, 0, len(inputs)),
			}

			for _, raw := range inputs {
				report := postcode.Diagnose(raw)
				if report.Valid {
					output.Valid++
				} else {
					output.Invalid++
				}
				output.Reports = append(output.Reports, report)
			}

			err = PrintOutput(cmd, v, output, func(w io.Writer) error {
				for _, r := range output.Reports {
					if r.Valid {
						_, _ = fmt.Fprintf(w, "✓ %-16s -> VALID: %s\n", r.Input, r.Normalized)
					} else {
						_, _ = fmt.Fprintf(w, "✗ %-16s -> INVALID (%d/11 chars)\n", r.Input, r.CleanLength)
						for _, d := range r.Diagnoses {
							_, _ = fmt.Fprintf(w, "    [%s] %s\n", d.Segment, d.Message)
							if len(d.Suggestions) > 0 {
								_, _ = fmt.Fprintf(w, "      Suggestions: %s\n", strings.Join(d.Suggestions, ", "))
							}
						}
						_, _ = fmt.Fprintf(w, "    Action: %s\n", r.ActionableTip)
					}
				}
				if len(output.Reports) > 1 {
					_, _ = fmt.Fprintf(w, "\nSummary: %d total, %d valid, %d invalid\n", output.Total, output.Valid, output.Invalid)
				}
				return nil
			})
			if err != nil {
				return err
			}

			if output.Invalid > 0 {
				return fmt.Errorf("diagnose: %d of %d postcodes have errors", output.Invalid, output.Total)
			}
			return nil
		},
	}

	return cmd
}
