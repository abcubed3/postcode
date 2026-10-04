package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// EvalCmdOutput represents the CLI output for the eval subcommand.
type EvalCmdOutput struct {
	TotalSamples     int     `json:"total_samples"`
	NoiseRate        float64 `json:"noise_rate"`
	StateAccuracy    float64 `json:"state_accuracy"`
	PostcodeAccuracy float64 `json:"postcode_accuracy"`
	ValidRate        float64 `json:"valid_rate"`
	AvgFormatScore   float64 `json:"avg_format_score"`
	DurationMs       int64   `json:"duration_ms"`
	ExportedPath     string  `json:"exported_path,omitempty"`
}

func (eo EvalCmdOutput) CSVHeader() []string {
	return []string{"total_samples", "noise_rate", "state_accuracy", "postcode_accuracy", "valid_rate", "avg_format_score", "duration_ms"}
}

func (eo EvalCmdOutput) CSVRows() [][]string {
	return [][]string{{
		fmt.Sprintf("%d", eo.TotalSamples),
		fmt.Sprintf("%.2f", eo.NoiseRate),
		fmt.Sprintf("%.2f%%", eo.StateAccuracy),
		fmt.Sprintf("%.2f%%", eo.PostcodeAccuracy),
		fmt.Sprintf("%.2f%%", eo.ValidRate),
		fmt.Sprintf("%.2f", eo.AvgFormatScore),
		fmt.Sprintf("%d", eo.DurationMs),
	}}
}

// NewEvalCmd creates the eval command for benchmarking address models and agents.
func NewEvalCmd(v *viper.Viper) *cobra.Command {
	var (
		samplesCount int
		seedVal      int64
		noiseRate    float64
		exportPath   string
		runBaseline  bool
	)

	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Generate synthetic Nigerian addresses and evaluate extraction accuracy",
		Long: `Eval creates realistic synthetic Nigerian address benchmarks and measures
how effectively LLMs and AI agent workflows extract entities, detect postcodes,
and correct noisy unstructured inputs (landmarks, typos, legacy codes).`,
		Example: `  postcode eval --samples 100 --noise 0.3
  postcode eval --samples 500 --export benchmark.json
  postcode eval --samples 50 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := postcode.GeneratorOptions{
				Count:     samplesCount,
				Seed:      uint64(seedVal),
				NoiseRate: noiseRate,
			}

			dataset := postcode.GenerateSyntheticDataset(opts)

			var exportedFile string
			if exportPath != "" {
				data, err := json.MarshalIndent(dataset, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to encode dataset: %w", err)
				}
				if err := os.WriteFile(exportPath, data, 0644); err != nil {
					return fmt.Errorf("failed to write export file: %w", err)
				}
				exportedFile = exportPath
			}

			res := postcode.EvaluateAgent(dataset, postcode.BaselineExtract)

			output := EvalCmdOutput{
				TotalSamples:     res.TotalSamples,
				NoiseRate:        noiseRate,
				StateAccuracy:    res.StateAccuracy,
				PostcodeAccuracy: res.PostcodeAccuracy,
				ValidRate:        res.ValidRate,
				AvgFormatScore:   res.AvgFormatScore,
				DurationMs:       res.DurationMs,
				ExportedPath:     exportedFile,
			}

			return PrintOutput(cmd, v, output, func(w io.Writer) error {
				_, _ = fmt.Fprintln(w, "Nigerian Address AI Evaluation Scorecard")
				_, _ = fmt.Fprintln(w, "========================================")
				_, _ = fmt.Fprintf(w, "Total Samples:       %d\n", output.TotalSamples)
				_, _ = fmt.Fprintf(w, "Noise Rate:          %.1f%%\n", output.NoiseRate*100)
				_, _ = fmt.Fprintf(w, "State Accuracy:      %.2f%%\n", output.StateAccuracy)
				_, _ = fmt.Fprintf(w, "Postcode Accuracy:   %.2f%%\n", output.PostcodeAccuracy)
				_, _ = fmt.Fprintf(w, "Valid Postcode Rate: %.2f%%\n", output.ValidRate)
				_, _ = fmt.Fprintf(w, "Avg Format Score:    %.2f / 100\n", output.AvgFormatScore)
				_, _ = fmt.Fprintf(w, "Eval Latency:        %d ms\n", output.DurationMs)
				if output.ExportedPath != "" {
					_, _ = fmt.Fprintf(w, "Exported Dataset:    %s\n", output.ExportedPath)
				}
				return nil
			})
		},
	}

	cmd.Flags().IntVarP(&samplesCount, "samples", "s", 50, "Number of synthetic addresses to generate")
	cmd.Flags().Int64Var(&seedVal, "seed", 0, "Seed for deterministic dataset generation (0 for random)")
	cmd.Flags().Float64VarP(&noiseRate, "noise", "n", 0.3, "Noise injection rate for landmarks and typos (0.0 to 1.0)")
	cmd.Flags().StringVarP(&exportPath, "export", "e", "", "File path to export dataset as JSON")
	cmd.Flags().BoolVar(&runBaseline, "baseline", true, "Run baseline heuristic extractor benchmark")

	return cmd
}
