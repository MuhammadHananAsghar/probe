// Package replay re-sends captured LLM requests, optionally with parameter
// modifications or translated to a different provider.
package replay

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/cost"
	"github.com/MuhammadHananAsghar/probe/internal/provider"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// Options control how a request is replayed.
type Options struct {
	// Model overrides the model field in the replayed request body.
	Model string
	// Provider translates the request to the target provider's format.
	// If empty, the original provider is used.
	Provider store.ProviderName
	// Temperature overrides the temperature parameter (nil = keep original).
	Temperature *float64
	// MaxTokens overrides max_tokens (nil = keep original).
	MaxTokens *int
	// SystemPrompt replaces the system message / system_prompt field.
	SystemPrompt string
}

// Result holds the outcome of a replay.
type Result struct {
	// Req is the new store.Request created for the replay.
	Req *store.Request
	// OriginalSeq is the original request's sequence number.
	OriginalSeq int
	// ParameterDiffs describes what was changed from the original.
	ParameterDiffs []string
}

// Engine executes request replays.
type Engine struct {
	s         store.Store
	pricingDB *cost.DB
}

// New creates a new replay Engine.
func New(s store.Store, pricingDB *cost.DB) *Engine {
	return &Engine{s: s, pricingDB: pricingDB}
}

// Replay re-sends request orig with opts applied, stores the result, and
// returns a Result. The original request is never modified.
func (e *Engine) Replay(ctx context.Context, orig *store.Request, opts Options) (*Result, error) {
	if orig == nil {
		return nil, fmt.Errorf("replay: nil request")
	}
	if len(orig.RequestBody) == 0 {
		return nil, fmt.Errorf("replay: request #%d has no stored body (too old or not captured)", orig.Seq)
	}

	// Determine target provider.
	targetProvider := orig.Provider
	if opts.Provider != "" {
		targetProvider = opts.Provider
	}

	p := provider.Detect(providerHost(targetProvider), pathForProvider(targetProvider))
	if p == nil {
		return nil, fmt.Errorf("replay: unsupported provider %q", targetProvider)
	}

	// Build modified body.
	body, diffs, err := applyModifications(orig, opts, targetProvider)
	if err != nil {
		return nil, fmt.Errorf("replay: modifying body: %w", err)
	}

	// Build the replayed store.Request.
	req := &store.Request{
		ID:           newID(),
		StartedAt:    time.Now(),
		Method:       orig.Method,
		URL:          orig.URL,
		Path:         pathForProvider(targetProvider),
		Provider:     targetProvider,
		ProviderHost: providerHost(targetProvider),
		Status:       store.StatusPending,
		ReplayOf:     orig.ID,
	}
	req.RequestHeaders = buildHeaders(orig, targetProvider)
	req.RequestBody = body

	// Parse the replayed request so model/stream/tools are populated.
	if parseErr := p.ParseRequest(body, req); parseErr != nil {
		// non-fatal
	}

	seq := e.s.Add(req)
	req.Seq = seq
	e.s.Update(req)

	// Send to provider.
	upstreamURL := "https://" + providerHost(targetProvider) + req.Path

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("replay: building request: %w", err)
	}
	for k, v := range req.RequestHeaders {
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		req.Status = store.StatusError
		req.ErrorMessage = err.Error()
		req.Latency = time.Since(req.StartedAt)
		req.EndedAt = time.Now()
		e.s.Update(req)
		return nil, fmt.Errorf("replay: upstream: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("replay: reading response: %w", err)
	}

	req.StatusCode = resp.StatusCode
	req.ResponseBody = respBody
	req.Latency = time.Since(req.StartedAt)
	req.EndedAt = time.Now()

	if parseErr := p.ParseResponse(respBody, req); parseErr != nil {
		// non-fatal
	}
	if e.pricingDB != nil {
		cost.Calculate(e.pricingDB, req)
	}

	req.Status = store.StatusDone
	if resp.StatusCode >= 400 {
		req.Status = store.StatusError
		req.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	e.s.Update(req)

	return &Result{
		Req:            req,
		OriginalSeq:    orig.Seq,
		ParameterDiffs: diffs,
	}, nil
}

// providerHost returns the canonical API hostname for a provider.
func providerHost(p store.ProviderName) string {
	switch p {
	case store.ProviderAnthropic:
		return "api.anthropic.com"
	default:
		return "api.openai.com"
	}
}

// pathForProvider returns the chat completions path for a provider.
func pathForProvider(p store.ProviderName) string {
	switch p {
	case store.ProviderAnthropic:
		return "/v1/messages"
	default:
		return "/v1/chat/completions"
	}
}

// buildHeaders copies the original request headers, updating Host.
// Sensitive auth headers are preserved so the replay authenticates correctly.
func buildHeaders(orig *store.Request, target store.ProviderName) map[string]string {
	h := make(map[string]string, len(orig.RequestHeaders))
	for k, v := range orig.RequestHeaders {
		h[k] = v
	}
	h["Host"] = providerHost(target)
	return h
}

// newID generates a random hex request ID.
func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// applyModifications builds the replay body with optional overrides.
// It also returns a human-readable list of parameter diffs.
func applyModifications(orig *store.Request, opts Options, target store.ProviderName) ([]byte, []string, error) {
	var body []byte
	var err error

	// Cross-provider translation if needed.
	if target != orig.Provider && target != "" {
		body, err = translateBody(orig, target)
		if err != nil {
			return nil, nil, fmt.Errorf("translating to %s: %w", target, err)
		}
	} else {
		body = make([]byte, len(orig.RequestBody))
		copy(body, orig.RequestBody)
	}

	// Parse the body so we can modify fields.
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, nil, fmt.Errorf("parsing body JSON: %w", err)
	}

	var diffs []string

	if opts.Model != "" {
		old, _ := m["model"].(string)
		m["model"] = opts.Model
		diffs = append(diffs, fmt.Sprintf("model: %s → %s", old, opts.Model))
	}
	if opts.Temperature != nil {
		var old string
		if v, ok := m["temperature"]; ok {
			old = fmt.Sprintf("%v", v)
		} else {
			old = "(unset)"
		}
		m["temperature"] = *opts.Temperature
		diffs = append(diffs, fmt.Sprintf("temperature: %s → %.2f", old, *opts.Temperature))
	}
	if opts.MaxTokens != nil {
		key := "max_tokens"
		if target == store.ProviderAnthropic {
			key = "max_tokens"
		}
		var old string
		if v, ok := m[key]; ok {
			old = fmt.Sprintf("%v", v)
		} else {
			old = "(unset)"
		}
		m[key] = *opts.MaxTokens
		diffs = append(diffs, fmt.Sprintf("max_tokens: %s → %d", old, *opts.MaxTokens))
	}
	if opts.SystemPrompt != "" {
		if target == store.ProviderAnthropic {
			m["system"] = opts.SystemPrompt
		} else {
			// OpenAI: replace/insert system message.
			if msgs, ok := m["messages"].([]any); ok {
				newMsgs := []any{map[string]any{"role": "system", "content": opts.SystemPrompt}}
				for _, msg := range msgs {
					if mm, ok := msg.(map[string]any); ok {
						if mm["role"] == "system" {
							continue
						}
					}
					newMsgs = append(newMsgs, msg)
				}
				m["messages"] = newMsgs
			}
		}
		preview := opts.SystemPrompt
		if len(preview) > 40 {
			preview = preview[:40] + "..."
		}
		diffs = append(diffs, fmt.Sprintf("system: → %q", preview))
	}
	if target != orig.Provider && target != "" {
		diffs = append(diffs, fmt.Sprintf("provider: %s → %s", orig.Provider, target))
	}

	out, err := json.Marshal(m)
	if err != nil {
		return nil, nil, err
	}
	return out, diffs, nil
}

// translateBody converts the request body from orig.Provider's format to target.
func translateBody(orig *store.Request, target store.ProviderName) ([]byte, error) {
	var src map[string]any
	if err := json.Unmarshal(orig.RequestBody, &src); err != nil {
		return nil, err
	}

	switch {
	case orig.Provider == store.ProviderAnthropic && (target == store.ProviderOpenAI || target == store.ProviderCompatible):
		return anthropicToOpenAI(src)
	case (orig.Provider == store.ProviderOpenAI || orig.Provider == store.ProviderCompatible) && target == store.ProviderAnthropic:
		return openAIToAnthropic(src)
	default:
		// Same-family or unknown: return as-is with model cleared.
		out, _ := json.Marshal(src)
		return out, nil
	}
}

// anthropicToOpenAI translates an Anthropic request body to OpenAI format.
func anthropicToOpenAI(src map[string]any) ([]byte, error) {
	dst := make(map[string]any)

	// Model
	dst["model"] = src["model"]

	// Messages: Anthropic has top-level system, OpenAI has role:system message
	var msgs []map[string]any
	if sys, ok := src["system"].(string); ok && sys != "" {
		msgs = append(msgs, map[string]any{"role": "system", "content": sys})
	}
	if rawMsgs, ok := src["messages"].([]any); ok {
		for _, m := range rawMsgs {
			if mm, ok := m.(map[string]any); ok {
				role, _ := mm["role"].(string)
				// Anthropic content can be string or []content_block
				content := normalizeContent(mm["content"])
				msgs = append(msgs, map[string]any{"role": role, "content": content})
			}
		}
	}
	dst["messages"] = msgs

	// Max tokens
	if v, ok := src["max_tokens"]; ok {
		dst["max_tokens"] = v
	}
	if v, ok := src["temperature"]; ok {
		dst["temperature"] = v
	}
	if v, ok := src["stream"]; ok {
		dst["stream"] = v
	}

	// Tools: Anthropic tools → OpenAI tools
	if tools, ok := src["tools"].([]any); ok && len(tools) > 0 {
		var oaiTools []map[string]any
		for _, t := range tools {
			if tool, ok := t.(map[string]any); ok {
				oaiTools = append(oaiTools, map[string]any{
					"type": "function",
					"function": map[string]any{
						"name":        tool["name"],
						"description": tool["description"],
						"parameters":  tool["input_schema"],
					},
				})
			}
		}
		if len(oaiTools) > 0 {
			dst["tools"] = oaiTools
		}
	}

	return json.Marshal(dst)
}

// openAIToAnthropic translates an OpenAI request body to Anthropic format.
func openAIToAnthropic(src map[string]any) ([]byte, error) {
	dst := make(map[string]any)

	dst["model"] = src["model"]

	// Separate system message from user/assistant messages.
	var system string
	var msgs []map[string]any
	if rawMsgs, ok := src["messages"].([]any); ok {
		for _, m := range rawMsgs {
			if mm, ok := m.(map[string]any); ok {
				role, _ := mm["role"].(string)
				if role == "system" {
					content := normalizeContent(mm["content"])
					system, _ = content.(string)
					continue
				}
				// Tool result messages
				if role == "tool" {
					msgs = append(msgs, map[string]any{
						"role": "user",
						"content": []map[string]any{{
							"type":        "tool_result",
							"tool_use_id": mm["tool_call_id"],
							"content":     normalizeContent(mm["content"]),
						}},
					})
					continue
				}
				msgs = append(msgs, map[string]any{
					"role":    role,
					"content": normalizeContent(mm["content"]),
				})
			}
		}
	}
	if system != "" {
		dst["system"] = system
	}
	dst["messages"] = msgs

	// Required by Anthropic
	if v, ok := src["max_tokens"]; ok {
		dst["max_tokens"] = v
	} else {
		dst["max_tokens"] = 4096
	}
	if v, ok := src["temperature"]; ok {
		dst["temperature"] = v
	}
	if v, ok := src["stream"]; ok {
		dst["stream"] = v
	}

	// Tools: OpenAI tools → Anthropic tools
	if tools, ok := src["tools"].([]any); ok && len(tools) > 0 {
		var anthropicTools []map[string]any
		for _, t := range tools {
			if tool, ok := t.(map[string]any); ok {
				if fn, ok := tool["function"].(map[string]any); ok {
					anthropicTools = append(anthropicTools, map[string]any{
						"name":         fn["name"],
						"description":  fn["description"],
						"input_schema": fn["parameters"],
					})
				}
			}
		}
		if len(anthropicTools) > 0 {
			dst["tools"] = anthropicTools
		}
	}

	return json.Marshal(dst)
}

// normalizeContent converts Anthropic content blocks or OpenAI content to a
// plain string where possible, otherwise returns the original value.
func normalizeContent(v any) any {
	switch c := v.(type) {
	case string:
		return c
	case []any:
		// Try to flatten a single text block.
		if len(c) == 1 {
			if block, ok := c[0].(map[string]any); ok {
				if block["type"] == "text" {
					if text, ok := block["text"].(string); ok {
						return text
					}
				}
			}
		}
		// Build a concatenated string from all text blocks.
		var parts []string
		for _, item := range c {
			if block, ok := item.(map[string]any); ok {
				if block["type"] == "text" {
					if text, ok := block["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "")
		}
		return v
	default:
		return v
	}
}
