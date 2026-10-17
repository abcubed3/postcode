package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FormattedPostcode struct {
	Input     string `json:"input"`
	Formatted string `json:"formatted"`
	Style     string `json:"style"`
	Canonical string `json:"canonical"`
	Spaced    string `json:"spaced"`
	Compact   string `json:"compact"`
	Slug      string `json:"slug"`
}

type FormatResult struct {
	Total   int                 `json:"total"`
	Results []FormattedPostcode `json:"results"`
}

func (fr FormatResult) CSVHeader() []string {
	return []string{"input", "formatted", "style", "canonical", "spaced", "compact", "slug"}
}

func (fr FormatResult) CSVRows() [][]string {
	rows := make([][]string, len(fr.Results))
	for i, r := range fr.Results {
		rows[i] = []string{r.Input, r.Formatted, r.Style, r.Canonical, r.Spaced, r.Compact, r.Slug}
	}
	return rows
}

// NewFormatCmd creates the format subcommand.
func NewFormatCmd(v *viper.Viper) *cobra.Command {
	var style string

	cmd := &cobra.Command{
		Use:   "format [postcodes...]",
		Short: "Format and normalize postcodes into canonical, spaced, compact, or slug styles",
		Long: `Format takes raw, dirty, or differently punctuated postcode strings and normalizes
them into one of four representations:
  - canonical: Hyphenated official standard (e.g. EK-01-A03-FK-01)
  - spaced:    Spaced 5-segment format (e.g. EK 01 A03 FK 01)
  - compact:   11-character unspaced uppercase alphanumeric (e.g. EK01A03FK01)
  - slug:      Lowercase hyphenated URL slug (e.g. ek-01-a03-fk-01)

Ideal for shell scripting and sanitizing address pipelines.`,
		Example: `  postcode format "ek 01 a03 fk 01" --style canonical
  postcode format EK-01-A03-FK-01 --style compact
  postcode format EK-01-A03-FK-01 --style slug
  cat raw_codes.txt | postcode format --style canonical`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			style = strings.ToLower(style)
			if style != "canonical" && style != "spaced" && style != "compact" && style != "slug" {
				return fmt.Errorf("invalid style %q: choose from canonical, spaced, compact, or slug", style)
			}

			res := FormatResult{
				Total:   len(inputs),
				Results: make([]FormattedPostcode, 0, len(inputs)),
			}

			for _, raw := range inputs {
				p, parseErr := postcode.Parse(raw)
				if parseErr != nil {
					return fmt.Errorf("formatting %q: %w", raw, parseErr)
				}

				var out string
				switch style {
				case "canonical":
					out = p.Formatted()
				case "spaced":
					out = p.String()
				case "compact":
					out = p.Raw()
				case "slug":
					out = p.Slug()
				}

				res.Results = append(res.Results, FormattedPostcode{
					Input:     raw,
					Formatted: out,
					Style:     style,
					Canonical: p.Formatted(),
					Spaced:    p.String(),
					Compact:   p.Raw(),
					Slug:      p.Slug(),
				})
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for _, r := range res.Results {
					_, _ = fmt.Fprintln(w, r.Formatted)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&style, "style", "s", "canonical", "target format style: canonical (hyphenated), spaced, compact, or slug")
	_ = cmd.RegisterFlagCompletionFunc("style", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"canonical", "spaced", "compact", "slug"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
