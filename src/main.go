package main

import (
	"crypto/tls"
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
	apiKey        string
	scrapeTimeout time.Duration
	skipTLSVerify bool
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
	flag.StringVar(&apiKey, "k", envDefault("LLAMASWAP_API_KEY", ""), "llama-swap API key")
	flag.IntVar(&timeoutSecs, "t", envDefaultInt("SCRAPE_TIMEOUT", 5), "scrape timeout in seconds")
	flag.BoolVar(&skipTLSVerify, "s", envDefaultBool("LSM_SKIP_TLS_VERIFY", false), "skip TLS certificate verification (self-signed certs)")
	flag.Parse()

	llamaSwapURL = strings.TrimRight(llamaSwapURL, "/")
	scrapeTimeout = time.Duration(timeoutSecs) * time.Second

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipTLSVerify}
	httpClient = &http.Client{
		Timeout:   scrapeTimeout,
		Transport: transport,
	}

	if skipTLSVerify {
		log.Printf("TLS certificate verification is disabled")
	}

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

func envDefaultBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
