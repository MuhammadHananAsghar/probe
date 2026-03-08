package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// exportRequest is a serialisable view of a store.Request that includes the
// raw bodies (which are json:"-" on the store model) as strings.
type exportRequest struct {
	ID              string                 `json:"id"`
	Seq             int                    `json:"seq"`
	StartedAt       time.Time              `json:"started_at"`
	EndedAt         time.Time              `json:"ended_at,omitempty"`
	Method          string                 `json:"method"`
	URL             string                 `json:"url"`
	Path            string                 `json:"path"`
	Provider        store.ProviderName     `json:"provider"`
	ProviderHost    string                 `json:"provider_host"`
	Model           string                 `json:"model"`
	Status          store.RequestStatus    `json:"status"`
	StatusCode      int                    `json:"status_code"`
	Latency         time.Duration          `json:"latency"`
	TTFT            time.Duration          `json:"ttft,omitempty"`
	InputTokens     int                    `json:"input_tokens"`
	OutputTokens    int                    `json:"output_tokens"`
	InputCost       float64                `json:"input_cost"`
	OutputCost      float64                `json:"output_cost"`
	TotalCost       float64                `json:"total_cost"`
	PricingKnown    bool                   `json:"pricing_known"`
	Stream          bool                   `json:"stream"`
	StreamStats     *store.StreamStats     `json:"stream_stats,omitempty"`
	SystemPrompt    string                 `json:"system_prompt,omitempty"`
	Messages        []store.Message        `json:"messages,omitempty"`
	ToolCalls       []store.ToolCall       `json:"tool_calls,omitempty"`
	ToolResults     []store.ToolResult     `json:"tool_results,omitempty"`
	Tools           []store.ToolDefinition `json:"tools,omitempty"`
	Anomalies       []store.Anomaly        `json:"anomalies,omitempty"`
	FinishReason    store.FinishReason     `json:"finish_reason,omitempty"`
	ResponseContent string                 `json:"response_content,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	ConversationID  string                 `json:"conversation_id,omitempty"`
	ReplayOf        string                 `json:"replay_of,omitempty"`
	RequestHeaders  map[string]string      `json:"request_headers,omitempty"`
	ResponseHeaders map[string]string      `json:"response_headers,omitempty"`
	RequestBody     string                 `json:"request_body,omitempty"`
	ResponseBody    string                 `json:"response_body,omitempty"`
}

func toExportRequest(r *store.Request) exportRequest {
	return exportRequest{
		ID:              r.ID,
		Seq:             r.Seq,
		StartedAt:       r.StartedAt,
		EndedAt:         r.EndedAt,
		Method:          r.Method,
		URL:             r.URL,
		Path:            r.Path,
		Provider:        r.Provider,
		ProviderHost:    r.ProviderHost,
		Model:           r.Model,
		Status:          r.Status,
		StatusCode:      r.StatusCode,
		Latency:         r.Latency,
		TTFT:            r.TTFT,
		InputTokens:     r.InputTokens,
		OutputTokens:    r.OutputTokens,
		InputCost:       r.InputCost,
		OutputCost:      r.OutputCost,
		TotalCost:       r.TotalCost,
		PricingKnown:    r.PricingKnown,
		Stream:          r.Stream,
		StreamStats:     r.StreamStats,
		SystemPrompt:    r.SystemPrompt,
		Messages:        r.Messages,
		ToolCalls:       r.ToolCalls,
		ToolResults:     r.ToolResults,
		Tools:           r.Tools,
		Anomalies:       r.Anomalies,
		FinishReason:    r.FinishReason,
		ResponseContent: r.ResponseContent,
		ErrorMessage:    r.ErrorMessage,
		ConversationID:  r.ConversationID,
		ReplayOf:        r.ReplayOf,
		RequestHeaders:  r.RequestHeaders,
		ResponseHeaders: r.ResponseHeaders,
		RequestBody:     string(r.RequestBody),
		ResponseBody:    string(r.ResponseBody),
	}
}

// ToJSON serialises requests as a JSON array. Compact mode skips indentation.
func ToJSON(requests []*store.Request, compact bool) ([]byte, error) {
	exports := make([]exportRequest, len(requests))
	for i, r := range requests {
		exports[i] = toExportRequest(r)
	}
	if compact {
		return json.Marshal(exports)
	}
	return json.MarshalIndent(exports, "", "  ")
}

// ToNDJSON serialises requests as newline-delimited JSON (one object per line).
func ToNDJSON(requests []*store.Request) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, r := range requests {
		if err := enc.Encode(toExportRequest(r)); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// FilterRequests applies simple key=value filters to a slice of requests.
// Supported keys: provider, model, status, min_cost, max_latency_ms.
func FilterRequests(requests []*store.Request, filters []string) []*store.Request {
	if len(filters) == 0 {
		return requests
	}
	result := make([]*store.Request, 0, len(requests))
	for _, r := range requests {
		if matchesAll(r, filters) {
			result = append(result, r)
		}
	}
	return result
}

func matchesAll(r *store.Request, filters []string) bool {
	for _, f := range filters {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch k {
		case "provider":
			if string(r.Provider) != v {
				return false
			}
		case "model":
			if r.Model != v {
				return false
			}
		case "status":
			if string(r.Status) != v {
				return false
			}
		}
	}
	return true
}
