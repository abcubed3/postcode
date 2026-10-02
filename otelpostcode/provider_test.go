package otelpostcode

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	postcode "github.com/abcubed3/postcode"
)

func TestNewTelemetry(t *testing.T) {
	tel, err := NewTelemetry()
	if err != nil {
		t.Fatalf("NewTelemetry() failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"postcode":"EK-01-A03-FK-01","valid":true}}`))
	}))
	defer server.Close()

	client, err := postcode.NewClient(
		postcode.WithBaseURL(server.URL),
		postcode.WithTelemetry(tel),
	)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	res, err := client.Lookup(context.Background(), "EK 01 A03 FK 01", postcode.Level1)
	if err != nil {
		t.Fatalf("Lookup() failed: %v", err)
	}

	if !res.Valid {
		t.Errorf("res.Valid = false, want true")
	}
}
