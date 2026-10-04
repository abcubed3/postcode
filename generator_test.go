package postcode

import (
	"testing"
)

func TestSyntheticAddressGenerator(t *testing.T) {
	opts := GeneratorOptions{
		Count:     25,
		Seed:      42,
		NoiseRate: 0.4,
	}

	dataset1 := GenerateSyntheticDataset(opts)
	if len(dataset1) != 25 {
		t.Fatalf("expected 25 samples, got %d", len(dataset1))
	}

	dataset2 := GenerateSyntheticDataset(opts)
	if dataset1[0].RawText != dataset2[0].RawText {
		t.Errorf("expected deterministic output with same seed, got %q vs %q", dataset1[0].RawText, dataset2[0].RawText)
	}

	// Verify fields
	for i, item := range dataset1 {
		if item.RawText == "" {
			t.Errorf("item %d has empty RawText", i)
		}
		if item.ExpectedState == "" {
			t.Errorf("item %d has empty ExpectedState", i)
		}
		if item.ExpectedLat == 0 || item.ExpectedLng == 0 {
			t.Errorf("item %d has invalid centroid coordinates", i)
		}
	}
}

func TestBaselineExtract(t *testing.T) {
	raw := "No. 45, Adetokunbo Ademola Street, Near Shoprite Mall, Victoria Island, Lagos State, LA 01 A01 AA 01"
	state, pc := BaselineExtract(raw)

	if state != "Lagos" {
		t.Errorf("expected Lagos, got %s", state)
	}
	cleanPC := "LA 01 A01 AA 01"
	if pc != cleanPC {
		t.Errorf("expected %s, got %s", cleanPC, pc)
	}
}

func TestEvaluateAgent(t *testing.T) {
	dataset := GenerateSyntheticDataset(GeneratorOptions{
		Count:     30,
		Seed:      1234,
		NoiseRate: 0.2,
	})

	res := EvaluateAgent(dataset, BaselineExtract)
	if res.TotalSamples != 30 {
		t.Fatalf("expected 30 total samples, got %d", res.TotalSamples)
	}
	if res.StateAccuracy < 70.0 {
		t.Errorf("expected state accuracy > 70%%, got %.2f%%", res.StateAccuracy)
	}
	if res.PostcodeAccuracy < 60.0 {
		t.Errorf("expected postcode accuracy > 60%%, got %.2f%%", res.PostcodeAccuracy)
	}
	if res.AvgFormatScore <= 0 {
		t.Errorf("expected positive average format score, got %.2f", res.AvgFormatScore)
	}
}
