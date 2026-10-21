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

func TestDiagnoseLegacy6DigitAndSingleDigitPadding(t *testing.T) {
	// 1. Legacy 6-digit Nigerian postal code (e.g. 100001 for Lagos Ikeja)
	legReport := Diagnose("100001")
	if legReport.Valid {
		t.Fatalf("expected legacy 6-digit code 100001 to be invalid")
	}
	foundLegacy := false
	for _, d := range legReport.Diagnoses {
		if d.Segment == "LegacyFormat" && strings.Contains(d.Message, "LA-11") {
			foundLegacy = true
		}
	}
	if !foundLegacy {
		t.Errorf("expected LegacyFormat diagnosis suggesting LA-11 for '100001', got %+v", legReport.Diagnoses)
	}

	// 2. Single-digit unpadded LGA & Unit in delimited input
	padReport := Diagnose("EK-1-A03-FK-1")
	if padReport.Valid {
		t.Fatalf("expected unpadded 'EK-1-A03-FK-1' to be invalid before repair")
	}
	if padReport.SuggestedFix != "EK-01-A03-FK-01" {
		t.Errorf("expected SuggestedFix 'EK-01-A03-FK-01', got %q", padReport.SuggestedFix)
	}
	for _, d := range padReport.Diagnoses {
		if d.Segment == "District" || d.Segment == "Area" {
			t.Errorf("unexpected false cascading diagnosis on %s: %+v", d.Segment, d)
		}
	}
}

func TestTryRepairAndRepair(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"EK-1-A03-FK-1", "EK-01-A03-FK-01"},
		{"LA 00 W06 TC 00", "LA-01-W06-TC-01"},
		{"BE-01-A01-AA-01", "BN-01-A01-AA-01"}, // ISO 'BE' -> NIPOST 'BN' (Benue)
		{"FCT-03-B06-AG-12", "FC-03-B06-AG-12"},
		{"LA-II-W06-TC-I0", "LA-11-W06-TC-10"}, // OCR 'I' -> '1' in numeric segments
		{"EK-O1-A03-FK-O1", "EK-01-A03-FK-01"}, // OCR 'O' -> '0' in numeric segments
	}

	for _, tc := range cases {
		p, ok := TryRepair(tc.input)
		if !ok {
			t.Errorf("TryRepair(%q) failed, expected %q", tc.input, tc.want)
			continue
		}
		if p.Formatted() != tc.want {
			t.Errorf("TryRepair(%q) = %q, want %q", tc.input, p.Formatted(), tc.want)
		}
		repaired, err := Repair(tc.input)
		if err != nil || repaired.Formatted() != tc.want {
			t.Errorf("Repair(%q) = %q (err: %v), want %q", tc.input, repaired.Formatted(), err, tc.want)
		}
	}

	if _, err := Repair("COMPLETELY_UNRECOVERABLE"); err == nil {
		t.Errorf("expected error from Repair on unrecoverable input")
	}
}

func TestDiagnoseTemporaryCode(t *testing.T) {
	report := Diagnose("TP_01_A01_01")
	if report.Valid {
		t.Errorf("expected temporary code to have Valid=false, got true")
	}
	if !report.IsTemporary {
		t.Errorf("expected IsTemporary=true, got false")
	}
	if report.Normalized != "TP-01-A01-01" {
		t.Errorf("expected Normalized='TP-01-A01-01', got %q", report.Normalized)
	}
	if report.FormatScore != 50.0 {
		t.Errorf("expected FormatScore=50.0, got %f", report.FormatScore)
	}
	if len(report.Diagnoses) == 0 || report.Diagnoses[0].Segment != "TemporaryCode" {
		t.Fatalf("expected Segment='TemporaryCode', got %+v", report.Diagnoses)
	}
	if !strings.Contains(report.ActionableTip, "GET /v1/lookup") {
		t.Errorf("expected ActionableTip to mention 'GET /v1/lookup', got %q", report.ActionableTip)
	}

	report2 := Diagnose("tp-123456")
	if !report2.IsTemporary || report2.Valid {
		t.Errorf("unexpected report for 'tp-123456': %+v", report2)
	}
	if report2.Normalized != "TP-123456" {
		t.Errorf("expected Normalized='TP-123456', got %q", report2.Normalized)
	}
}
