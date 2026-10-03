package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewBatchCmd creates the batch processing subcommand for ETL workflows.
func NewBatchCmd(v *viper.Viper) *cobra.Command {
	var inputFile string
	var outputFile string
	var columnName string

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch-process and enrich CSV address datasets with coordinates and URLs",
		Long: `Batch reads a CSV file containing postcode records, validates each code, and appends
enrichment columns (canonical postcode, validity, latitude, longitude, precision, and Google Maps URL).

Operates with zero network calls and high throughput, making it ideal for ETL pipelines
and migrating legacy databases.`,
		Example: `  postcode batch --input customers.csv --output enriched.csv --column postcode
  cat orders.csv | postcode batch --output - --column shipping_postcode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var r io.Reader = cmd.InOrStdin()
			if inputFile != "" && inputFile != "-" {
				f, err := os.Open(inputFile)
				if err != nil {
					return fmt.Errorf("opening input file: %w", err)
				}
				defer f.Close()
				r = f
			}

			var w io.Writer = cmd.OutOrStdout()
			if outputFile != "" && outputFile != "-" {
				f, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("creating output file: %w", err)
				}
				defer f.Close()
				w = f
			}

			reader := csv.NewReader(r)
			writer := csv.NewWriter(w)
			defer writer.Flush()

			header, err := reader.Read()
			if err != nil {
				return fmt.Errorf("reading CSV header: %w", err)
			}

			colIdx := -1
			for i, h := range header {
				if strings.EqualFold(strings.TrimSpace(h), columnName) {
					colIdx = i
					break
				}
			}

			if colIdx == -1 {
				return fmt.Errorf("postcode column %q not found in CSV header: %v", columnName, header)
			}

			// Output columns appended
			newHeader := append(header,
				"postcode_valid",
				"postcode_canonical",
				"postcode_state",
				"postcode_lga",
				"latitude",
				"longitude",
				"precision",
				"maps_url",
			)

			if err := writer.Write(newHeader); err != nil {
				return fmt.Errorf("writing header: %w", err)
			}

			start := time.Now()
			total := 0
			validCount := 0
			invalidCount := 0

			for {
				row, err := reader.Read()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return fmt.Errorf("reading CSV record %d: %w", total+1, err)
				}

				total++
				rawCode := ""
				if colIdx < len(row) {
					rawCode = row[colIdx]
				}

				p, parseErr := postcode.Parse(rawCode)
				if parseErr != nil {
					invalidCount++
					enriched := append(row, "false", "", "", "", "", "", "", "")
					if err := writer.Write(enriched); err != nil {
						return err
					}
					continue
				}

				validCount++
				loc := p.Location()
				latStr := fmt.Sprintf("%.6f", loc.Latitude)
				lngStr := fmt.Sprintf("%.6f", loc.Longitude)
				enriched := append(row,
					"true",
					p.Formatted(),
					loc.StateName,
					loc.LGAName,
					latStr,
					lngStr,
					loc.Precision.String(),
					loc.GoogleMapsURL(),
				)

				if err := writer.Write(enriched); err != nil {
					return err
				}
			}

			elapsed := time.Since(start)
			throughput := float64(total) / elapsed.Seconds()

			// Report summary to stderr so stdout remains a clean CSV stream
			errW := cmd.ErrOrStderr()
			fmt.Fprintf(errW, "\nBatch Process Complete: %d rows processed in %v (%.1f rows/sec)\n", total, elapsed.Round(time.Millisecond), throughput)
			fmt.Fprintf(errW, "  ✓ Valid postcodes:   %d (%.1f%%)\n", validCount, float64(validCount)/float64(max(total, 1))*100.0)
			fmt.Fprintf(errW, "  ✗ Invalid postcodes: %d (%.1f%%)\n", invalidCount, float64(invalidCount)/float64(max(total, 1))*100.0)

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "path to input CSV file (or '-' for stdin)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "path to output CSV file (or '-' for stdout)")
	cmd.Flags().StringVarP(&columnName, "column", "c", "postcode", "name of the column containing postcodes")

	return cmd
}
