package analyze

import (
	"encoding/json"
	"fmt"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// DetectAnomalies inspects a single completed request and returns any detected
// anomalies. The caller is responsible for persisting the returned slice on req.
func DetectAnomalies(req *store.Request) []store.Anomaly {
	var anomalies []store.Anomaly

	// ManyTools: more than 20 tool definitions degrades model behavior.
	if len(req.Tools) > 20 {
		anomalies = append(anomalies, store.Anomaly{
			Kind:    store.AnomalyManyTools,
			Message: fmt.Sprintf("%d tools defined (>20 may degrade model behavior)", len(req.Tools)),
		})
	}

	// MalformedArgs: validate JSON arguments for each tool call.
	for i, tc := range req.ToolCalls {
		if tc.ArgumentsJSON == "" {
			continue
		}
		if !json.Valid([]byte(tc.ArgumentsJSON)) {
			req.ToolCalls[i].ParseError = true
			anomalies = append(anomalies, store.Anomaly{
				Kind:    store.AnomalyMalformedArgs,
				Message: fmt.Sprintf("tool %q: malformed JSON arguments", tc.Name),
			})
		}
	}

	// OrphanedCall: tool calls with no results are expected in the current
	// request when finish_reason is tool_call; they'll be paired in the
	// follow-up request. Not an anomaly here — skip.

	// ToolLoop: detect if a single tool is called 5+ times in one response.
	toolCounts := make(map[string]int, len(req.ToolCalls))
	for _, tc := range req.ToolCalls {
		toolCounts[tc.Name]++
	}
	for name, count := range toolCounts {
		if count >= 5 {
			anomalies = append(anomalies, store.Anomaly{
				Kind:    store.AnomalyToolLoop,
				Message: fmt.Sprintf("tool %q called %d times in one response", name, count),
			})
		}
	}

	// SlowTool: latency is computed per-step in BuildToolChain; skip here.

	return anomalies
}
