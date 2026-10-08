package commands

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ReferenceResult struct {
	Type     string               `json:"type"`
	State    string               `json:"state,omitempty"`
	LGA      string               `json:"lga,omitempty"`
	District string               `json:"district,omitempty"`
	Total    int                  `json:"total"`
	Items    []postcode.NamedCode `json:"items"`
}

func (rr ReferenceResult) CSVHeader() []string {
	return []string{"code", "name"}
}

func (rr ReferenceResult) CSVRows() [][]string {
	rows := make([][]string, len(rr.Items))
	for i, item := range rr.Items {
		rows[i] = []string{item.Code, item.Name}
	}
	return rows
}

// NewReferenceCmd creates the parent reference command.
func NewReferenceCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reference",
		Short: "Query NIPOST administrative reference catalogs (states, LGAs, districts, areas)",
		Long: `Reference commands inspect precomputed administrative entity catalogs
from the NIPOST Postcode API (or offline reference datasets).
Reference endpoints are free and do not consume credits.`,
		Example: `  postcode reference states
  postcode reference states --offline -o json
  postcode reference lgas --state FC
  postcode reference districts --state FC --lga 01
  postcode reference areas --state FC --lga 01 --district A01`,
	}

	cmd.AddCommand(
		newReferenceStatesCmd(v),
		newReferenceLGAsCmd(v),
		newReferenceDistrictsCmd(v),
		newReferenceAreasCmd(v),
	)

	return cmd
}

func getOfflineStates() []postcode.NamedCode {
	items := make([]postcode.NamedCode, 0, len(postcode.NigerianStates))
	for code, rec := range postcode.NigerianStates {
		items = append(items, postcode.NamedCode{
			Code: code,
			Name: rec.Name,
		})
	}
	slices.SortFunc(items, func(a, b postcode.NamedCode) int {
		return cmp.Compare(a.Code, b.Code)
	})
	return items
}

func getOfflineLGAs(state string) []postcode.NamedCode {
	return postcode.StateLGAs(state)
}

func getOfflineDistricts(state, lga string) []postcode.NamedCode {
	lgaNorm := lga
	if len(lgaNorm) == 1 {
		lgaNorm = "0" + lgaNorm
	}
	seen := make(map[string]bool)
	var items []postcode.NamedCode
	for _, b := range postcode.AllBuildingRecords() {
		if b.StateCode == state && b.LGACode == lgaNorm {
			p, err := postcode.Parse(b.Postcode)
			if err == nil {
				dist := p.District()
				if !seen[dist] {
					seen[dist] = true
					items = append(items, postcode.NamedCode{Code: dist})
				}
			}
		}
	}
	slices.SortFunc(items, func(a, b postcode.NamedCode) int {
		return cmp.Compare(a.Code, b.Code)
	})
	return items
}

func getOfflineAreas(state, lga, district string) []postcode.NamedCode {
	lgaNorm := lga
	if len(lgaNorm) == 1 {
		lgaNorm = "0" + lgaNorm
	}
	seen := make(map[string]bool)
	var items []postcode.NamedCode
	for _, b := range postcode.AllBuildingRecords() {
		if b.StateCode == state && b.LGACode == lgaNorm {
			p, err := postcode.Parse(b.Postcode)
			if err == nil && p.District() == district {
				area := p.Area()
				if !seen[area] {
					seen[area] = true
					items = append(items, postcode.NamedCode{Code: area})
				}
			}
		}
	}
	slices.SortFunc(items, func(a, b postcode.NamedCode) int {
		return cmp.Compare(a.Code, b.Code)
	})
	return items
}

func newReferenceStatesCmd(v *viper.Viper) *cobra.Command {
	var offline bool

	cmd := &cobra.Command{
		Use:   "states",
		Short: "List all 37 Nigerian states and FCT with SSLL codes and full names",
		Example: `  postcode reference states
  postcode reference states --offline -o csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var items []postcode.NamedCode

			if offline || isOffline(v, cmd) {
				items = getOfflineStates()
			} else {
				client, err := buildClient(v)
				if err != nil {
					return err
				}
				var apiErr error
				items, apiErr = client.ReferenceStates(cmd.Context())
				if apiErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: Remote gateway reference states failed (%v). Falling back to offline dataset.\n", apiErr)
					items = getOfflineStates()
				}
			}

			res := ReferenceResult{
				Type:  "states",
				Total: len(items),
				Items: items,
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, item := range res.Items {
					_, _ = fmt.Fprintf(w, "%-4s  %s\n", item.Code, item.Name)
				}
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&offline, "offline", false, "read from offline embedded reference dataset without network call")
	return cmd
}

func newReferenceLGAsCmd(v *viper.Viper) *cobra.Command {
	var state string
	var offline bool

	cmd := &cobra.Command{
		Use:   "lgas",
		Short: "List Local Government Areas (LGAs) within a specified state",
		Example: `  postcode reference lgas --state FC
  postcode reference lgas --state LA --offline -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var items []postcode.NamedCode

			if offline || isOffline(v, cmd) {
				items = getOfflineLGAs(state)
			} else {
				client, err := buildClient(v)
				if err != nil {
					return err
				}
				var apiErr error
				items, apiErr = client.ReferenceLGAs(cmd.Context(), state)
				if apiErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: Remote gateway reference lgas failed (%v). Falling back to offline dataset.\n", apiErr)
					items = getOfflineLGAs(state)
				}
			}

			res := ReferenceResult{
				Type:  "lgas",
				State: state,
				Total: len(items),
				Items: items,
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, item := range res.Items {
					_, _ = fmt.Fprintf(w, "%-4s  %s\n", item.Code, item.Name)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "2-letter state code (e.g. FC, LA, EK)")
	cmd.Flags().BoolVar(&offline, "offline", false, "read from offline embedded reference dataset without network call")
	_ = cmd.MarkFlagRequired("state")

	return cmd
}

func newReferenceDistrictsCmd(v *viper.Viper) *cobra.Command {
	var state, lga string

	cmd := &cobra.Command{
		Use:   "districts",
		Short: "List postal district codes within a state and LGA",
		Example: `  postcode reference districts --state FC --lga 01
  postcode reference districts --state EK --lga 01 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var items []postcode.NamedCode

			if isOffline(v, cmd) {
				items = getOfflineDistricts(state, lga)
			} else {
				client, err := buildClient(v)
				if err != nil {
					return err
				}

				var apiErr error
				items, apiErr = client.ReferenceDistricts(cmd.Context(), state, lga)
				if apiErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: Remote gateway reference districts failed (%v). Falling back to offline database.\n", apiErr)
					items = getOfflineDistricts(state, lga)
				}
			}

			res := ReferenceResult{
				Type:  "districts",
				State: state,
				LGA:   lga,
				Total: len(items),
				Items: items,
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, item := range res.Items {
					if item.Name != "" {
						_, _ = fmt.Fprintf(w, "%-6s  %s\n", item.Code, item.Name)
					} else {
						_, _ = fmt.Fprintf(w, "%s\n", item.Code)
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "2-letter state code (e.g. FC, LA, EK)")
	cmd.Flags().StringVar(&lga, "lga", "", "2-digit LGA code (e.g. 01, 11)")
	_ = cmd.MarkFlagRequired("state")
	_ = cmd.MarkFlagRequired("lga")

	return cmd
}

func newReferenceAreasCmd(v *viper.Viper) *cobra.Command {
	var state, lga, district string

	cmd := &cobra.Command{
		Use:   "areas",
		Short: "List postal area codes within a state, LGA, and district",
		Example: `  postcode reference areas --state FC --lga 01 --district A01
  postcode reference areas --state EK --lga 01 --district A03 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var items []postcode.NamedCode

			if isOffline(v, cmd) {
				items = getOfflineAreas(state, lga, district)
			} else {
				client, err := buildClient(v)
				if err != nil {
					return err
				}

				var apiErr error
				items, apiErr = client.ReferenceAreas(cmd.Context(), state, lga, district)
				if apiErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: Remote gateway reference areas failed (%v). Falling back to offline database.\n", apiErr)
					items = getOfflineAreas(state, lga, district)
				}
			}

			res := ReferenceResult{
				Type:     "areas",
				State:    state,
				LGA:      lga,
				District: district,
				Total:    len(items),
				Items:    items,
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, item := range res.Items {
					if item.Name != "" {
						_, _ = fmt.Fprintf(w, "%-6s  %s\n", item.Code, item.Name)
					} else {
						_, _ = fmt.Fprintf(w, "%s\n", item.Code)
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "2-letter state code (e.g. FC, LA, EK)")
	cmd.Flags().StringVar(&lga, "lga", "", "2-digit LGA code (e.g. 01, 11)")
	cmd.Flags().StringVar(&district, "district", "", "3-character district code (e.g. A01, A03)")
	_ = cmd.MarkFlagRequired("state")
	_ = cmd.MarkFlagRequired("lga")
	_ = cmd.MarkFlagRequired("district")

	return cmd
}
