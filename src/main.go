package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	llamaSwapURL = strings.TrimRight(getEnv("LLAMA_SWAP_URL", "http://localhost:8080"), "/")
	listenPort   = getEnv("LISTEN_PORT", "9090")
	scrapeTimeout = func() time.Duration {
		d, err := time.ParseDuration(getEnv("SCRAPE_TIMEOUT", "5s"))
		if err != nil {
			return 5 * time.Second
		}
		return d
	}()
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprint(w, collectMetrics())
}

func main() {
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	addr := ":" + listenPort
	log.Printf("llama-swap metrics proxy listening on %s", addr)
	log.Printf("proxying from: %s", llamaSwapURL)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
