package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/abcubed3/postcode"
)

func main() {
	fmt.Println("==================================================================")
	fmt.Println("  Nigerian Address AI Evaluation Benchmark & Dataset Generator")
	fmt.Println("==================================================================")

	// 1. Generate a realistic synthetic dataset of Nigerian addresses
	// Injects landmarks ("Opposite Central Mosque", "Beside First Bank"), typos,
	// unstructured syntax, and legacy 6-digit postcodes.
	opts := postcode.GeneratorOptions{
		Count:     50,
		Seed:      2026, // Deterministic seed for reproducible benchmarks
		NoiseRate: 0.35, // 35% probability of injecting noise/landmarks/typos
	}

	fmt.Printf("Generating %d synthetic addresses (noise rate: %.1f%%)...\n\n", opts.Count, opts.NoiseRate*100)
	dataset := postcode.GenerateSyntheticDataset(opts)

	// Display sample addresses with various noise patterns
	fmt.Println("Sample Generated Addresses:")
	fmt.Println("------------------------------------------------------------------")
	for i := 0; i < 5 && i < len(dataset); i++ {
		item := dataset[i]
		fmt.Printf("[%d] Type: %-15s | State: %s (%s)\n", i+1, item.NoiseType, item.ExpectedState, item.ExpectedStateCode)
		fmt.Printf("    Raw:      %s\n", item.RawText)
		fmt.Printf("    Expected: Postcode=%s, Lat=%.4f, Lng=%.4f\n\n", item.ExpectedPostcode, item.ExpectedLat, item.ExpectedLng)
	}

	// 2. Evaluate the Baseline Heuristic Extractor against Ground Truth
	fmt.Println("------------------------------------------------------------------")
	fmt.Println("Running Extraction Benchmark against Ground Truth...")
	evalResult := postcode.EvaluateAgent(dataset, postcode.BaselineExtract)

	fmt.Println("\n==================================================================")
	fmt.Println("  Evaluation Scorecard")
	fmt.Println("==================================================================")
	fmt.Printf("  • Total Benchmark Samples: %d\n", evalResult.TotalSamples)
	fmt.Printf("  • State Extraction Accuracy: %.2f%%\n", evalResult.StateAccuracy)
	fmt.Printf("  • Postcode Match Accuracy:   %.2f%%\n", evalResult.PostcodeAccuracy)
	fmt.Printf("  • Valid Postcode Rate:       %.2f%%\n", evalResult.ValidRate)
	fmt.Printf("  • Mean Structural Score:     %.2f / 100.0\n", evalResult.AvgFormatScore)
	fmt.Printf("  • Evaluation Duration:       %d ms\n", evalResult.DurationMs)
	fmt.Println("==================================================================")

	// 3. Export to JSON for external eval frameworks (LangSmith, Promptfoo, Braintrust)
	exportPath := "synthetic_nigerian_addresses.json"
	rawJSON, err := json.MarshalIndent(dataset[:10], "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize dataset: %v", err)
	}

	if err := os.WriteFile(exportPath, rawJSON, 0644); err != nil {
		log.Fatalf("Failed to write export file: %v", err)
	}
	defer func() {
		_ = os.Remove(exportPath) // Clean up sample export file
	}()

	fmt.Printf("\nExported 10 sample records to %s for AI agent prompt-testing.\n", exportPath)
}
