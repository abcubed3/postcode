package commands_test

import (
	"bytes"
	"strings"
	"testing"

	cmd "github.com/abcubed3/postcode/cmd/postcode/internal/commands"
)

func executeCmd(args []string, in string) (stdout string, stderr string, err error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	root := cmd.NewRootCmd()
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs(args)

	if in != "" {
		root.SetIn(bytes.NewBufferString(in))
	}

	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestCLI_RootHelp(t *testing.T) {
	out, _, err := executeCmd([]string{"--help"}, "")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	expectedSubstrings := []string{
		"Usage:",
		"postcode [command]",
		"Offline & Transformation Commands:",
		"validate",
		"diagnose",
		"parse",
		"format",
		"coords",
		"map",
		"assemble",
		"disassemble",
		"Gateway & Geocoding Commands:",
		"lookup",
		"autocomplete",
		"nearby",
		"reverse",
		"status",
		"Operations & Developer Tools:",
		"batch",
		"serve",
		"mcp",
		"eval",
		"version",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(out, sub) {
			t.Errorf("expected root help to contain %q, but got:\n%s", sub, out)
		}
	}
}

func TestCLI_Validate(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		stdin      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "valid single postcode argument",
			args:       []string{"validate", "EK 01 A03 FK 01"},
			wantErr:    false,
			wantOutput: "✓ EK 01 A03 FK 01  -> EK-01-A03-FK-01  (State: EK)",
		},
		{
			name:       "valid multiple postcodes",
			args:       []string{"validate", "EK-01-A03-FK-01", "LA 11 W06 TC 10"},
			wantErr:    false,
			wantOutput: "Summary: 2 total, 2 valid, 0 invalid",
		},
		{
			name:       "invalid LGA 00",
			args:       []string{"validate", "EK 00 A03 FK 01"},
			wantErr:    true,
			wantOutput: "INVALID: postcode: invalid format structure: LGA cannot be 00 (valid range 01-99)",
		},
		{
			name:       "invalid non-existent state with strict validation",
			args:       []string{"validate", "--strict", "XX 01 A03 FK 01"},
			wantErr:    true,
			wantOutput: "unknown state code \"XX\"",
		},
		{
			name:       "stdin streaming input",
			args:       []string{"validate"},
			stdin:      "EK-01-A03-FK-01\nLA-11-W06-TC-10\n",
			wantErr:    false,
			wantOutput: "Summary: 2 total, 2 valid, 0 invalid",
		},
		{
			name:       "quiet flag on valid",
			args:       []string{"validate", "--quiet", "EK-01-A03-FK-01"},
			wantErr:    false,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, err := executeCmd(tt.args, tt.stdin)
			if (err != nil) != tt.wantErr {
				t.Fatalf("executeCmd() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantOutput != "" && !strings.Contains(out, tt.wantOutput) {
				t.Errorf("output %q does not contain expected %q", out, tt.wantOutput)
			}
		})
	}
}

func TestCLI_Parse(t *testing.T) {
	t.Run("human text output", func(t *testing.T) {
		out, _, err := executeCmd([]string{"parse", "EK 01 A03 FK 01"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"Postcode:      EK-01-A03-FK-01",
			"Compact:       EK01A03FK01",
			"Spaced:        EK 01 A03 FK 01",
			"State:         Ekiti (EK)",
			"Capital:       Ado Ekiti",
			"LGA:           Ado Ekiti (01)",
			"District:      A03",
			"Area:          FK",
			"Building Unit: 01",
			"Zone:          SOUTH WEST",
		}
		for _, exp := range expected {
			if !strings.Contains(out, exp) {
				t.Errorf("output missing %q:\n%s", exp, out)
			}
		}
	})

	t.Run("json output", func(t *testing.T) {
		out, _, err := executeCmd([]string{"parse", "LA 11 W06 TC 10", "-o", "json"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, `"state_name": "Lagos"`) {
			t.Errorf("json output missing state_name Lagos: %s", out)
		}
		if !strings.Contains(out, `"compact": "LA11W06TC10"`) {
			t.Errorf("json output missing compact: %s", out)
		}
	})

	t.Run("csv output", func(t *testing.T) {
		out, _, err := executeCmd([]string{"parse", "FC 03 B06 AG 12", "-o", "csv"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "formatted,compact,spaced,state_code,state_name") {
			t.Errorf("csv header missing in: %s", out)
		}
		if !strings.Contains(out, "FC-03-B06-AG-12,FC03B06AG12,FC 03 B06 AG 12,FC,Federal Capital Territory") {
			t.Errorf("csv row missing in: %s", out)
		}
	})
}

func TestCLI_Format(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		input    string
		expected string
	}{
		{"canonical", "canonical", "ek 01 a03 fk 01", "EK-01-A03-FK-01"},
		{"spaced", "spaced", "LA-11-W06-TC-10", "LA 11 W06 TC 10"},
		{"compact", "compact", "FC 03 B06 AG 12", "FC03B06AG12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, err := executeCmd([]string{"format", "--style", tt.style, tt.input}, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestCLI_Coords(t *testing.T) {
	out, _, err := executeCmd([]string{"coords", "EK-01-A03-FK-01"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Lat:   7.621100, Lng:   5.221500 [building]") {
		t.Errorf("expected building coords in output: %s", out)
	}
	if !strings.Contains(out, "NTA Road, Back of Fabian Hotel, Ado Ekiti") {
		t.Errorf("expected street address in output: %s", out)
	}
}

func TestCLI_Map(t *testing.T) {
	t.Run("google maps search url", func(t *testing.T) {
		out, _, err := executeCmd([]string{"map", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "https://www.google.com/maps/search/?api=1&query=7.621100,5.221500"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in %s", expected, out)
		}
	})

	t.Run("google maps directions url", func(t *testing.T) {
		out, _, err := executeCmd([]string{"map", "--directions", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "https://www.google.com/maps/dir/?api=1&destination=7.621100,5.221500"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in %s", expected, out)
		}
	})

	t.Run("apple maps url", func(t *testing.T) {
		out, _, err := executeCmd([]string{"map", "--provider", "apple", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "https://maps.apple.com/?ll=7.621100,5.221500") {
			t.Errorf("expected apple maps url in %s", out)
		}
	})

	t.Run("osm url", func(t *testing.T) {
		out, _, err := executeCmd([]string{"map", "--provider", "osm", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "https://www.openstreetmap.org/?mlat=7.621100&mlon=5.221500") {
			t.Errorf("expected osm url in %s", out)
		}
	})
}

func TestCLI_AssembleAndDisassemble(t *testing.T) {
	t.Run("assemble", func(t *testing.T) {
		out, _, err := executeCmd([]string{
			"assemble",
			"--state", "EK",
			"--lga", "01",
			"--district", "A03",
			"--area", "FK",
			"--unit", "01",
		}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Postcode: EK-01-A03-FK-01") {
			t.Errorf("expected Postcode in output: %s", out)
		}
		if !strings.Contains(out, "Compact:  EK01A03FK01") {
			t.Errorf("expected Compact in output: %s", out)
		}
	})

	t.Run("disassemble", func(t *testing.T) {
		out, _, err := executeCmd([]string{"disassemble", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "State: EK | LGA: 01 | District: A03 | Area: FK | Unit: 01"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in %s", expected, out)
		}
	})
}

func TestCLI_BatchCSV(t *testing.T) {
	csvInput := `id,name,shipping_postcode
1,Amina,EK 01 A03 FK 01
2,Bello,INVALID_CODE
`
	args := []string{"batch", "--input", "-", "--output", "-", "--column", "shipping_postcode"}
	out, errOut, err := executeCmd(args, csvInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "postcode_valid,postcode_canonical") {
		t.Errorf("expected enriched header in: %s", out)
	}
	if !strings.Contains(out, "1,Amina,EK 01 A03 FK 01,true,EK-01-A03-FK-01") {
		t.Errorf("expected valid row in: %s", out)
	}
	if !strings.Contains(out, "2,Bello,INVALID_CODE,false,,,,,,,") {
		t.Errorf("expected invalid row in: %s", out)
	}
	if !strings.Contains(errOut, "Batch Process Complete: 2 rows processed") {
		t.Errorf("expected batch summary in stderr: %s", errOut)
	}
}

func TestCLI_Version(t *testing.T) {
	out, _, err := executeCmd([]string{"version"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "postcode version") {
		t.Errorf("expected version output: %s", out)
	}
}

func TestCLI_RootVersionFlag(t *testing.T) {
	out, _, err := executeCmd([]string{"--version"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "postcode version") {
		t.Errorf("expected version flag output, got: %s", out)
	}
}

func TestCLI_LookupValidation(t *testing.T) {
	t.Run("invalid level rejected", func(t *testing.T) {
		_, _, err := executeCmd([]string{"lookup", "--level", "99", "EK-01-A03-FK-01"}, "")
		if err == nil {
			t.Fatal("expected error for invalid level 99, got nil")
		}
		if !strings.Contains(err.Error(), "invalid lookup level 99") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("offline fallback error display on invalid code", func(t *testing.T) {
		out, _, err := executeCmd([]string{"lookup", "INVALID_CODE_99"}, "")
		if err != nil {
			t.Fatalf("unexpected execute error: %v", err)
		}
		if !strings.Contains(out, "INVALID_CODE_99 [INVALID]") {
			t.Errorf("expected INVALID status for input code: %s", out)
		}
		if !strings.Contains(out, "Error:") {
			t.Errorf("expected Error line in output: %s", out)
		}
	})
}

func TestCLI_YAMLDeterminism(t *testing.T) {
	out1, _, err := executeCmd([]string{"parse", "EK-01-A03-FK-01", "-o", "yaml"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out2, _, err := executeCmd([]string{"parse", "EK-01-A03-FK-01", "-o", "yaml"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out1 != out2 {
		t.Errorf("YAML output is not deterministic across multiple runs:\nRun1:\n%s\nRun2:\n%s", out1, out2)
	}
}

func TestCLI_Diagnose(t *testing.T) {
	t.Run("valid code", func(t *testing.T) {
		out, _, err := executeCmd([]string{"diagnose", "EK 01 A03 FK 01"}, "")
		if err != nil {
			t.Fatalf("expected nil error for valid code, got: %v", err)
		}
		if !strings.Contains(out, "VALID: EK-01-A03-FK-01") {
			t.Errorf("expected VALID output, got: %s", out)
		}
	})

	t.Run("invalid code with suggestions", func(t *testing.T) {
		out, _, err := executeCmd([]string{"diagnose", "ZZ 00 A03 FK 00"}, "")
		if err == nil {
			t.Fatal("expected error exit for invalid postcode")
		}
		if !strings.Contains(out, "INVALID") {
			t.Errorf("expected INVALID in output, got: %s", out)
		}
		if !strings.Contains(out, "Suggestions:") {
			t.Errorf("expected Suggestions in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out, _, err := executeCmd([]string{"diagnose", "EK 01 A03 FK 01", "-o", "json"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, `"valid": true`) {
			t.Errorf("expected valid: true in JSON output, got: %s", out)
		}
	})
}

func TestCLI_MCP(t *testing.T) {
	initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`
	out, errOut, err := executeCmd([]string{"mcp"}, initMsg+"\n")
	if err != nil {
		t.Fatalf("mcp command failed: %v", err)
	}
	if !strings.Contains(errOut, "postcode MCP server running on stdio") {
		t.Errorf("expected startup log on stderr, got: %s", errOut)
	}
	if !strings.Contains(out, `"protocolVersion":"2024-11-05"`) {
		t.Errorf("expected initialize response on stdout, got: %s", out)
	}
}

func TestCLI_Eval(t *testing.T) {
	t.Run("default evaluation", func(t *testing.T) {
		out, _, err := executeCmd([]string{"eval", "--samples", "20", "--noise", "0.2"}, "")
		if err != nil {
			t.Fatalf("eval command failed: %v", err)
		}
		if !strings.Contains(out, "Nigerian Address AI Evaluation Scorecard") {
			t.Errorf("expected scorecard header in output, got: %s", out)
		}
		if !strings.Contains(out, "Total Samples:       20") {
			t.Errorf("expected 20 samples in output, got: %s", out)
		}
	})

	t.Run("json output and export", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test_eval.json"
		out, _, err := executeCmd([]string{"eval", "--samples", "10", "--export", tmpFile, "-o", "json"}, "")
		if err != nil {
			t.Fatalf("eval json failed: %v", err)
		}
		if !strings.Contains(out, `"total_samples": 10`) {
			t.Errorf("expected total_samples: 10 in output, got: %s", out)
		}
	})
}

func TestCLI_NearbyAndReverse_ArgumentValidation(t *testing.T) {
	t.Run("nearby requires reference postcode or coordinates", func(t *testing.T) {
		_, _, err := executeCmd([]string{"nearby"}, "")
		if err == nil {
			t.Fatal("expected error when neither postcode nor coordinates provided")
		}
		if !strings.Contains(err.Error(), "specify a reference postcode") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("reverse requires coordinates", func(t *testing.T) {
		_, _, err := executeCmd([]string{"reverse"}, "")
		if err == nil {
			t.Fatal("expected error when no coordinates provided")
		}
		if !strings.Contains(err.Error(), "specify coordinates via args") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nearby rejects invalid postcode", func(t *testing.T) {
		_, _, err := executeCmd([]string{"nearby", "INVALID-CODE"}, "")
		if err == nil {
			t.Fatal("expected error for invalid postcode")
		}
		if !strings.Contains(err.Error(), "invalid reference postcode") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestCLI_Cache(t *testing.T) {
	t.Run("cache status displays path and count", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"cache", "status"}, "")
		if err != nil {
			t.Fatalf("cache status failed: %v", err)
		}
		if !strings.Contains(stdout, "Local Cache Path") || !strings.Contains(stdout, "Cached Buildings") {
			t.Errorf("unexpected output: %s", stdout)
		}
	})

	t.Run("cache list executes without error", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"cache", "list"}, "")
		if err != nil {
			t.Fatalf("cache list failed: %v", err)
		}
		if len(stdout) == 0 {
			t.Error("expected non-empty output from cache list")
		}
	})
}

func TestCLI_Assemble_SingleDigitPadding(t *testing.T) {
	// Tests official docs example: state: ek, lga: 1, district: a03, area: fk, unit: 1
	stdout, _, err := executeCmd([]string{
		"assemble",
		"--state", "ek",
		"--lga", "1",
		"--district", "a03",
		"--area", "fk",
		"--unit", "1",
	}, "")
	if err != nil {
		t.Fatalf("assemble with single-digit segments failed: %v", err)
	}
	if !strings.Contains(stdout, "EK-01-A03-FK-01") {
		t.Errorf("expected EK-01-A03-FK-01 in output, got: %s", stdout)
	}
}

func TestCLI_Reference(t *testing.T) {
	t.Run("states offline", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"reference", "states", "--offline"}, "")
		if err != nil {
			t.Fatalf("reference states --offline failed: %v", err)
		}
		if !strings.Contains(stdout, "FC") || !strings.Contains(stdout, "Federal Capital Territory") {
			t.Errorf("expected FC in output, got: %s", stdout)
		}
	})

	t.Run("lgas offline", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"reference", "lgas", "--state", "FC", "--offline"}, "")
		if err != nil {
			t.Fatalf("reference lgas --state FC --offline failed: %v", err)
		}
		if !strings.Contains(stdout, "01") || !strings.Contains(stdout, "Abaji") {
			t.Errorf("expected Abaji in output, got: %s", stdout)
		}
	})
}

func TestCLI_OfflineMode(t *testing.T) {
	t.Run("lookup offline with centroid and known buildings", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"--offline", "lookup", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("offline lookup failed: %v", err)
		}
		if !strings.Contains(stdout, "offline") {
			t.Errorf("expected offline indication in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "Ekiti") {
			t.Errorf("expected Ekiti state in output, got: %s", stdout)
		}
	})

	t.Run("status offline", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"--offline", "status"}, "")
		if err != nil {
			t.Fatalf("offline status failed: %v", err)
		}
		if !strings.Contains(stdout, "OFFLINE") {
			t.Errorf("expected OFFLINE mode in status output, got: %s", stdout)
		}
	})

	t.Run("nearby offline", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"--offline", "nearby", "LA-08-A86-RG-01"}, "")
		if err != nil {
			t.Fatalf("offline nearby failed: %v", err)
		}
		if !strings.Contains(stdout, "Nearby units (offline)") {
			t.Errorf("expected offline nearby header, got: %s", stdout)
		}
	})

	t.Run("reverse offline", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"--offline", "reverse", "--lat", "6.476111", "--lng", "3.633990"}, "")
		if err != nil {
			t.Fatalf("offline reverse failed: %v", err)
		}
		if !strings.Contains(stdout, "LA-08-A86-RG-01") {
			t.Errorf("expected LA-08-A86-RG-01 in offline reverse output, got: %s", stdout)
		}
	})

	t.Run("coords offline", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"--offline", "coords", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("offline coords failed: %v", err)
		}
		if !strings.Contains(stdout, "7.621100") {
			t.Errorf("expected 7.621100 in offline coords output, got: %s", stdout)
		}
	})

	t.Run("map offline", func(t *testing.T) {
		stdout, _, err := executeCmd([]string{"--offline", "map", "EK-01-A03-FK-01"}, "")
		if err != nil {
			t.Fatalf("offline map failed: %v", err)
		}
		if !strings.Contains(stdout, "https://www.google.com/maps") {
			t.Errorf("expected google maps url in offline map output, got: %s", stdout)
		}
	})
}





