package postcode

import (
	"strings"
	"testing"
)

func TestDiagnoseValid(t *testing.T) {
	report := Diagnose("EK 01 A03 FK 01")
	if !report.Valid {
		t.Fatalf("expected valid, got invalid: %+v", report)
	}
	if report.Normalized != "EK-01-A03-FK-01" {
		t.Errorf("expected normalized 'EK-01-A03-FK-01', got %q", report.Normalized)
	}
	if len(report.Diagnoses) != 0 {
		t.Errorf("expected 0 diagnoses for valid code, got %d", len(report.Diagnoses))
	}
	if report.CleanLength != 11 {
		t.Errorf("expected clean length 11, got %d", report.CleanLength)
	}
}

func TestDiagnoseInvalidLength(t *testing.T) {
	// Incomplete
	report := Diagnose("EK 01 A03")
	if report.Valid {
		t.Fatalf("expected invalid for truncated input")
	}
	if report.CleanLength != 7 {
		t.Errorf("expected clean length 7, got %d", report.CleanLength)
	}
	foundLengthDiag := false
	for _, d := range report.Diagnoses {
		if d.Segment == "Length" {
			foundLengthDiag = true
			if !strings.Contains(d.Message, "missing") {
				t.Errorf("expected message to explain missing characters, got %q", d.Message)
			}
		}
	}
	if !foundLengthDiag {
		t.Errorf("expected Length diagnosis for truncated input")
	}

	// Empty
	emptyReport := Diagnose("")
	if emptyReport.Valid || emptyReport.CleanLength != 0 {
		t.Errorf("expected clean length 0 for empty input, got %d", emptyReport.CleanLength)
	}
}

func TestDiagnoseUnknownState(t *testing.T) {
	report := Diagnose("ZZ 01 A03 FK 01")
	if report.Valid {
		t.Fatalf("expected invalid for unknown state ZZ")
	}

	foundStateDiag := false
	for _, d := range report.Diagnoses {
		if d.Segment == "State" {
			foundStateDiag = true
			if !strings.Contains(d.Message, "not a recognized Nigerian state code") {
				t.Errorf("unexpected message: %q", d.Message)
			}
			if len(d.Suggestions) == 0 {
				t.Errorf("expected suggestions for state correction, got none")
			}
		}
	}
	if !foundStateDiag {
		t.Errorf("expected State diagnosis for ZZ")
	}
}

func TestDiagnoseZeroSegments(t *testing.T) {
	// LGA 00
	reportLGA := Diagnose("EK 00 A03 FK 01")
	if reportLGA.Valid {
		t.Fatalf("expected invalid for LGA 00")
	}
	foundLGA := false
	for _, d := range reportLGA.Diagnoses {
		if d.Segment == "LGA" && strings.Contains(d.Message, "00") {
			foundLGA = true
		}
	}
	if !foundLGA {
		t.Errorf("expected LGA diagnosis for '00'")
	}

	// Building Unit 00
	reportUnit := Diagnose("EK 01 A03 FK 00")
	if reportUnit.Valid {
		t.Fatalf("expected invalid for Unit 00")
	}
	foundUnit := false
	for _, d := range reportUnit.Diagnoses {
		if d.Segment == "BuildingUnit" && strings.Contains(d.Message, "00") {
			foundUnit = true
		}
	}
	if !foundUnit {
		t.Errorf("expected BuildingUnit diagnosis for '00'")
	}
}

func TestDiagnoseGrammarErrors(t *testing.T) {
	// Non-alpha state, non-digit LGA, non-alpha Area
	report := Diagnose("12 AB A03 99 01")
	if report.Valid {
		t.Fatalf("expected invalid")
	}
	if len(report.Diagnoses) < 3 {
		t.Errorf("expected multiple diagnoses, got %d", len(report.Diagnoses))
	}
}

func TestValidationError(t *testing.T) {
	report := Diagnose("EK 00 A03 FK 01")
	valErr := &ValidationError{Report: report}
	if !strings.Contains(valErr.Error(), "invalid") {
		t.Errorf("expected error string to mention invalid, got %q", valErr.Error())
	}
}
