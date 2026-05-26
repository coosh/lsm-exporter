package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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

	flag.StringVar(&listenHost, "a", "0.0.0.0", "listen address")
	flag.StringVar(&listenPort, "p", "9090", "listen port")
	flag.StringVar(&llamaSwapURL, "l", "http://localhost:8080", "llama-swap base URL")
	flag.IntVar(&timeoutSecs, "t", 5, "scrape timeout in seconds")
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
