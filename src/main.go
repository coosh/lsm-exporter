package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	llamaSwapURL  string
	scrapeTimeout time.Duration
)

func main() {
	var (
		listenHost  string
		listenPort  string
		timeoutSecs int
	)

	flag.StringVar(&listenHost, "a", envDefault("LSM_LISTEN_ADDR", "0.0.0.0"), "listen address")
	flag.StringVar(&listenPort, "p", envDefault("LSM_LISTEN_PORT", "9090"), "listen port")
	flag.StringVar(&llamaSwapURL, "l", envDefault("LLAMASWAP_URL", "http://localhost:8080"), "llama-swap base URL")
	flag.IntVar(&timeoutSecs, "t", envDefaultInt("SCRAPE_TIMEOUT", 5), "scrape timeout in seconds")
	flag.Parse()

	llamaSwapURL = strings.TrimRight(llamaSwapURL, "/")
	scrapeTimeout = time.Duration(timeoutSecs) * time.Second
	httpClient = &http.Client{Timeout: scrapeTimeout}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<a href="/metrics">Metrics</a>`)
	})
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	addr := listenHost + ":" + listenPort
	log.Printf("llama-swap metrics proxy listening on %s", addr)
	log.Printf("proxying from: %s", llamaSwapURL)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprint(w, collectMetrics())
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
