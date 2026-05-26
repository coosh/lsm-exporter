package main

import (
	"fmt"
	"log"
	"strings"
)

// addModelLabel injects a model="<name>" label into every Prometheus metric line.
func addModelLabel(metrics string, modelName string) string {
	var sb strings.Builder
	for line := range strings.SplitSeq(metrics, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			sb.WriteString(line + "\n")
			continue
		}

		braceIdx := strings.Index(trimmed, "{")
		spaceIdx := strings.Index(trimmed, " ")

		if braceIdx != -1 && (spaceIdx == -1 || braceIdx < spaceIdx) {
			closing := strings.Index(trimmed[braceIdx:], "}")
			if closing == -1 {
				sb.WriteString(line + "\n")
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
			sb.WriteString(line + "\n")
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
		// Use /upstream passthrough rather than model.
		metricsURL := llamaSwapURL + "/upstream/" + model.Model + "/metrics"
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
