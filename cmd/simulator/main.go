package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abcubed3/postcode/simulator"
)

func main() {
	port := flag.String("port", getEnvOrDefault("PORT", "8080"), "Port for the mock NIPOST server")
	host := flag.String("host", "localhost", "Host address to bind to")
	flag.Parse()

	addr := fmt.Sprintf("%s:%s", *host, *port)

	server := &http.Server{
		Addr:         addr,
		Handler:      simulator.NewHandler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	fmt.Println("================================================================================")
	fmt.Println("🏛️  NIPOST Digital Postcode Gateway - Live Simulator")
	fmt.Println("📖 Official Reference: https://docs.postcode.gov.ng/concepts/lookup-levels#test-postcodes")
	fmt.Println("================================================================================")
	fmt.Printf("🚀 Server running at: http://%s\n", addr)
	fmt.Printf("🩺 Health check:      http://%s/healthz\n", addr)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("📋 Preloaded Official Test Postcodes:")
	for _, rec := range simulator.TestPostcodes {
		fmt.Printf("  • %-16s | %-12s | %s\n", rec.Canonical, rec.StateName, rec.RecentHouse)
	}
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("💡 Try these sample curl commands:")
	fmt.Printf("  # 1. Level 1 Public Lookup (Free, No Auth):\n")
	fmt.Printf("  curl \"http://%s/v1/lookup?code=EK-01-A03-FK-01&level=1\"\n\n", addr)
	fmt.Printf("  # 2. Level 2 Commercial Lookup (Requires X-API-Key):\n")
	fmt.Printf("  curl \"http://%s/v1/lookup?code=LA-11-W06-TC-10&level=2\" \\\n       -H \"X-API-Key: nipost_live_test\"\n\n", addr)
	fmt.Printf("  # 3. Level 3 Commercial Lookup with Building Use Metadata:\n")
	fmt.Printf("  curl \"http://%s/v1/lookup?code=FC-03-B06-AG-12&level=3\" \\\n       -H \"X-API-Key: nipost_live_test\"\n\n", addr)
	fmt.Printf("  # 4. Segment-Aware Autocomplete:\n")
	fmt.Printf("  curl \"http://%s/v1/search/autocomplete?q=OG-14\"\n\n", addr)
	fmt.Printf("  # 5. Nearby Unit Search (Radius in meters):\n")
	fmt.Printf("  curl \"http://%s/v1/search/nearby?lat=6.6018&lng=3.3515&radius=200\"\n\n", addr)
	fmt.Printf("  # 6. Reverse Geocoding:\n")
	fmt.Printf("  curl \"http://%s/v1/search/reverse?lat=6.6018&lng=3.3515\"\n", addr)
	fmt.Println("================================================================================")
	fmt.Println("Press Ctrl+C to stop.")

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("\nShutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		fmt.Println("Server stopped.")
	case err := <-serverErr:
		log.Fatalf("Server error: %v", err)
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
