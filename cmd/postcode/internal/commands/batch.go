package commands

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewBatchCmd creates the batch processing subcommand for ETL workflows.
func NewBatchCmd(v *viper.Viper) *cobra.Command {
	var inputFile string
	var outputFile string
	var outAlias string
	var columnName string
	var workers int
	var autoRepair bool

	cmd := &cobra.Command{
		Use:   "batch [input.csv]",
		Short: "Batch-process and enrich CSV address datasets with coordinates and URLs",
		Long: `Batch reads a CSV file containing postcode records, validates each code, and appends
enrichment columns (canonical postcode, validity, latitude, longitude, precision, and Google Maps URL).

Operates with zero network calls and high throughput, making it ideal for ETL pipelines
and migrating legacy databases.`,
		Example: `  postcode batch --input customers.csv --output enriched.csv --column postcode
  postcode batch input_addresses.csv --column postcode --out cleaned.csv --workers 8
  cat orders.csv | postcode batch --output - --column shipping_postcode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" && len(args) > 0 {
				inputFile = args[0]
			}
			if outputFile == "" && outAlias != "" {
				outputFile = outAlias
			}
			if workers < 1 {
				workers = 1
			}

			r := cmd.InOrStdin()
			if inputFile != "" && inputFile != "-" {
				f, err := os.Open(inputFile)
				if err != nil {
					return fmt.Errorf("opening input file: %w", err)
				}
				defer func() { _ = f.Close() }()
				r = f
			}

			w := cmd.OutOrStdout()
			var outCloser io.Closer
			if outputFile != "" && outputFile != "-" {
				f, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("creating output file: %w", err)
				}
				outCloser = f
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

			enrichRow := func(row []string) ([]string, bool) {
				rawCode := ""
				if colIdx < len(row) {
					rawCode = row[colIdx]
				}
				p, parseErr := postcode.Parse(rawCode)
				if parseErr != nil && autoRepair {
					if repaired, ok := postcode.TryRepair(rawCode); ok {
						p = repaired
						parseErr = nil
					}
				}
				if parseErr != nil {
					return append(row, "false", "", "", "", "", "", "", ""), false
				}
				loc := p.Location()
				latStr := strconv.FormatFloat(loc.Latitude, 'f', 6, 64)
				lngStr := strconv.FormatFloat(loc.Longitude, 'f', 6, 64)
				return append(row,
					"true",
					p.Formatted(),
					loc.StateName,
					loc.LGAName,
					latStr,
					lngStr,
					loc.Precision.String(),
					loc.GoogleMapsURL(),
				), true
			}

			const batchChunkSize = 512
			chunk := make([][]string, 0, batchChunkSize)
			flushChunk := func(rows [][]string) error {
				if len(rows) == 0 {
					return nil
				}
				if workers == 1 || len(rows) == 1 {
					for _, row := range rows {
						enriched, ok := enrichRow(row)
						if ok {
							validCount++
						} else {
							invalidCount++
						}
						if err := writer.Write(enriched); err != nil {
							return err
						}
					}
					return nil
				}

				results := make([][]string, len(rows))
				validFlags := make([]bool, len(rows))
				var wg sync.WaitGroup
				sem := make(chan struct{}, workers)
				for i, rRow := range rows {
					wg.Add(1)
					sem <- struct{}{}
					go func(idx int, inRow []string) {
						defer wg.Done()
						defer func() { <-sem }()
						results[idx], validFlags[idx] = enrichRow(inRow)
					}(i, rRow)
				}
				wg.Wait()

				for i, enriched := range results {
					if validFlags[i] {
						validCount++
					} else {
						invalidCount++
					}
					if err := writer.Write(enriched); err != nil {
						return err
					}
				}
				return nil
			}

			for {
				row, err := reader.Read()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return fmt.Errorf("reading CSV record %d: %w", total+1, err)
				}
				total++
				chunk = append(chunk, row)
				if len(chunk) >= batchChunkSize {
					if err := flushChunk(chunk); err != nil {
						return err
					}
					chunk = chunk[:0]
				}
			}
			if err := flushChunk(chunk); err != nil {
				return err
			}

			writer.Flush()
			if err := writer.Error(); err != nil {
				if outCloser != nil {
					_ = outCloser.Close()
				}
				return fmt.Errorf("flushing CSV output: %w", err)
			}
			if outCloser != nil {
				if err := outCloser.Close(); err != nil {
					return fmt.Errorf("closing output file: %w", err)
				}
			}

			elapsed := time.Since(start)
			throughput := float64(total) / elapsed.Seconds()

			// Report summary to stderr so stdout remains a clean CSV stream
			errW := cmd.ErrOrStderr()
			_, _ = fmt.Fprintf(errW, "\nBatch Process Complete: %d rows processed in %v (%.1f rows/sec)\n", total, elapsed.Round(time.Millisecond), throughput)
			_, _ = fmt.Fprintf(errW, "  ✓ Valid postcodes:   %d (%.1f%%)\n", validCount, float64(validCount)/float64(max(total, 1))*100.0)
			_, _ = fmt.Fprintf(errW, "  ✗ Invalid postcodes: %d (%.1f%%)\n", invalidCount, float64(invalidCount)/float64(max(total, 1))*100.0)

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "path to input CSV file (or '-' for stdin)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "path to output CSV file (or '-' for stdout)")
	cmd.Flags().StringVar(&outAlias, "out", "", "alias for --output")
	cmd.Flags().StringVarP(&columnName, "column", "c", "postcode", "name of the column containing postcodes")
	cmd.Flags().IntVarP(&workers, "workers", "w", 1, "number of concurrent workers")
	cmd.Flags().BoolVarP(&autoRepair, "repair", "r", false, "automatically self-heal recoverable postcode typos before rejecting")

	return cmd
}
