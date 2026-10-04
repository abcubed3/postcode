package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// OutputFormat defines supported serialization types.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatYAML OutputFormat = "yaml"
	FormatCSV  OutputFormat = "csv"
)

// PrintOutput formats and writes data to the command's stdout based on the configured format.
func PrintOutput(cmd *cobra.Command, v *viper.Viper, data any, defaultTextPrinter func(w io.Writer) error) error {
	format := strings.ToLower(v.GetString("output"))
	if format == "" {
		format = string(FormatText)
	}

	w := cmd.OutOrStdout()

	switch OutputFormat(format) {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)

	case FormatYAML:
		// Render clean YAML without requiring external YAML dependency
		return renderSimpleYAML(w, data)

	case FormatCSV:
		if csvProvider, ok := data.(CSVData); ok {
			cw := csv.NewWriter(w)
			defer cw.Flush()
			if err := cw.Write(csvProvider.CSVHeader()); err != nil {
				return err
			}
			for _, row := range csvProvider.CSVRows() {
				if err := cw.Write(row); err != nil {
					return err
				}
			}
			return nil
		}
		// Fallback to JSON if not CSV-compatible
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)

	case FormatText:
		fallthrough
	default:
		if defaultTextPrinter != nil {
			return defaultTextPrinter(w)
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
}

// CSVData allows structs or collections to expose tabular rows for CSV serialization.
type CSVData interface {
	CSVHeader() []string
	CSVRows() [][]string
}

// renderSimpleYAML converts JSON data into basic clean YAML.
func renderSimpleYAML(w io.Writer, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	return writeYAMLValue(w, raw, 0)
}

func writeYAMLValue(w io.Writer, v any, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	switch val := v.(type) {
	case map[string]any:
		if indent > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		for _, k := range slices.Sorted(maps.Keys(val)) {
			if _, err := fmt.Fprintf(w, "%s%s: ", indentStr, k); err != nil {
				return err
			}
			if err := writeYAMLValue(w, val[k], indent+1); err != nil {
				return err
			}
		}
	case []any:
		if indent > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		for _, item := range val {
			if _, err := fmt.Fprintf(w, "%s- ", indentStr); err != nil {
				return err
			}
			if err := writeYAMLValue(w, item, indent+1); err != nil {
				return err
			}
		}
	case string:
		if val == "" {
			if _, err := fmt.Fprintln(w, `""`); err != nil {
				return err
			}
		} else if strings.ContainsAny(val, ":#\n\t") || strings.HasPrefix(val, " ") || strings.HasSuffix(val, " ") {
			if _, err := fmt.Fprintf(w, "%q\n", val); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "%s\n", val); err != nil {
				return err
			}
		}
	case nil:
		if _, err := fmt.Fprintln(w, "null"); err != nil {
			return err
		}
	default:
		if _, err := fmt.Fprintf(w, "%v\n", val); err != nil {
			return err
		}
	}
	return nil
}
