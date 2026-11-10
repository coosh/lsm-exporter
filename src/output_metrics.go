package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
)

// loopbackWarned tracks model/proxy pairs already reported as loopback, so the
// notice is logged once rather than on every scrape.
var loopbackWarned sync.Map

// modelMetricsURL returns the URL to scrape for a model's metrics. By default
// this goes through llama-swap's /upstream passthrough; in direct mode it hits
// the model's proxy address itself so llama-swap does not see the request.
func modelMetricsURL(model RunningModel) (string, error) {
	if !directUpstream {
		return llamaSwapURL + "/upstream/" + model.Model + "/metrics", nil
	}

	proxy := strings.TrimRight(model.Proxy, "/")
	if proxy == "" {
		return "", fmt.Errorf("no proxy address reported by /running")
	}

	if u, err := url.Parse(proxy); err == nil && isLoopbackHost(u.Hostname()) {
		if _, seen := loopbackWarned.LoadOrStore(model.Model+"|"+proxy, struct{}{}); !seen {
			log.Printf("model %s proxy address %s is a loopback address; direct scraping only works when lsm-exporter shares a host/network namespace with llama-swap", model.Model, proxy)
		}
	}

	return proxy + "/metrics", nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// addModelLabel injects a model="<name>" label into every Prometheus metric line.
func addModelLabel(metrics string, modelName string) string {
	var sb strings.Builder
	for line := range strings.SplitSeq(metrics, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			sb.WriteString(line)
			sb.WriteByte('\n')
			continue
		}

		braceIdx := strings.Index(trimmed, "{")
		spaceIdx := strings.Index(trimmed, " ")

		if braceIdx != -1 && (spaceIdx == -1 || braceIdx < spaceIdx) {
			closing := strings.Index(trimmed[braceIdx:], "}")
			if closing == -1 {
				sb.WriteString(line)
				sb.WriteByte('\n')
				continue
			}
			labelContent := trimmed[braceIdx+1 : braceIdx+closing]
			rest := trimmed[braceIdx+closing:]
			if labelContent == "" {
				fmt.Fprintf(&sb, "%s{model=%q%s\n", trimmed[:braceIdx], modelName, rest)
			} else {
				fmt.Fprintf(&sb, "%s{model=%q,%s%s\n", trimmed[:braceIdx], modelName, labelContent, rest)
			}
		} else if spaceIdx != -1 {
			rest := trimmed[spaceIdx:]
			fmt.Fprintf(&sb, "%s{model=%q}%s\n", trimmed[:spaceIdx], modelName, rest)
		} else {
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func collectMetrics() string {
	var out strings.Builder

	swapMetrics, err := fetch(llamaSwapURL + "/metrics")
	if err != nil {
		log.Printf("failed to fetch llama-swap /metrics: %v", err)
	} else {
		out.WriteString("# source: llama-swap /metrics\n")
		out.WriteString(swapMetrics)
		out.WriteString("\n")
	}

	models := getActiveModels()
	for _, model := range models {
		if model.Model == "" {
			continue
		}
		metricsURL, err := modelMetricsURL(model)
		if err != nil {
			log.Printf("cannot scrape model %s: %v", model.Model, err)
			continue
		}
		modelMetrics, err := fetch(metricsURL)
		if err != nil {
			log.Printf("failed to fetch metrics for model %s: %v", model.Model, err)
			continue
		}
		if strings.TrimSpace(modelMetrics) == "" {
			log.Printf("empty metrics response for model %s", model.Model)
			continue
		}
		fmt.Fprintf(&out, "# source: upstream model=%s\n", model.Model)
		out.WriteString(addModelLabel(modelMetrics, model.Model))
		out.WriteString("\n")
	}

	return out.String()
}
