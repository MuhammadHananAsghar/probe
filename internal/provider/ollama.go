package provider

import (
	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/tidwall/gjson"
)

// Ollama implements Provider for Ollama's local model API.
// Endpoint: http://localhost:11434/api/chat
type Ollama struct{}

// Name returns the provider identifier.
func (o *Ollama) Name() store.ProviderName { return store.ProviderOllama }

// ParseRequest extracts model, messages, and options from an Ollama chat request.
func (o *Ollama) ParseRequest(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	req.Model = gjson.GetBytes(body, "model").String()
	req.Stream = gjson.GetBytes(body, "stream").Bool()

	// Options block holds temperature and num_predict.
	if t := gjson.GetBytes(body, "options.temperature"); t.Exists() {
		v := t.Float()
		req.Temperature = &v
	}
	if np := gjson.GetBytes(body, "options.num_predict"); np.Exists() {
		v := int(np.Int())
		req.MaxTokens = &v
	}

	// Messages (same role format as OpenAI: user/assistant/system).
	gjson.GetBytes(body, "messages").ForEach(func(_, v gjson.Result) bool {
		role := v.Get("role").String()
		content := v.Get("content").String()
		if role == "system" && req.SystemPrompt == "" {
			req.SystemPrompt = content
		}
		req.Messages = append(req.Messages, store.Message{Role: role, Content: content})
		return true
	})

	return nil
}

// ParseResponse extracts content and token counts from an Ollama response.
// Ollama returns total_duration in nanoseconds; token counts are prompt_eval_count
// and eval_count.
func (o *Ollama) ParseResponse(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	// Non-streaming: single JSON object.
	req.ResponseContent = gjson.GetBytes(body, "message.content").String()

	// Token counts.
	if n := gjson.GetBytes(body, "prompt_eval_count").Int(); n > 0 {
		req.InputTokens = int(n)
	}
	if n := gjson.GetBytes(body, "eval_count").Int(); n > 0 {
		req.OutputTokens = int(n)
	}

	// Done reason.
	if gjson.GetBytes(body, "done").Bool() {
		req.FinishReason = store.FinishStop
	}

	// Cost is always zero for local models.
	req.PricingKnown = true
	req.InputCost = 0
	req.OutputCost = 0
	req.TotalCost = 0

	return nil
}

// ParseEvent processes a single Ollama streaming JSON line.
// Ollama streaming sends newline-delimited JSON (not SSE), but the intercept
// layer handles the raw bytes; we receive individual JSON chunks here.
func (o *Ollama) ParseEvent(_ string, data string, req *store.Request) string {
	if data == "" {
		return ""
	}

	result := gjson.Parse(data)
	delta := result.Get("message.content").String()
	if delta != "" {
		req.ResponseContent += delta
	}

	if result.Get("done").Bool() {
		req.FinishReason = store.FinishStop
		if n := result.Get("prompt_eval_count").Int(); n > 0 {
			req.InputTokens = int(n)
		}
		if n := result.Get("eval_count").Int(); n > 0 {
			req.OutputTokens = int(n)
		}
		// Ensure cost stays zero for local models.
		req.PricingKnown = true
		req.TotalCost = 0
	}

	return delta
}
