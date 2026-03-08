package provider

import (
	"strings"

	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/tidwall/gjson"
)

// Google implements Provider for Google's Generative Language (Gemini) API.
// Endpoint: generativelanguage.googleapis.com/v1/models/{model}:generateContent
type Google struct{}

// Name returns the provider identifier.
func (g *Google) Name() store.ProviderName { return store.ProviderGoogle }

// ParseRequest extracts model, messages, and generation config from a Gemini request.
func (g *Google) ParseRequest(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	// Model is extracted from the URL path in the proxy layer; fall back to body if empty.
	if req.Model == "" {
		req.Model = "gemini"
	}

	// Gemini uses "contents" array for conversation.
	gjson.GetBytes(body, "contents").ForEach(func(_, v gjson.Result) bool {
		role := v.Get("role").String()
		if role == "model" {
			role = "assistant"
		}
		// Parts is an array; concatenate text parts.
		var contentParts []string
		v.Get("parts").ForEach(func(_, part gjson.Result) bool {
			if t := part.Get("text").String(); t != "" {
				contentParts = append(contentParts, t)
			}
			return true
		})
		content := strings.Join(contentParts, "")
		req.Messages = append(req.Messages, store.Message{Role: role, Content: content})
		return true
	})

	// System instruction (Gemini's equivalent of a system prompt).
	if si := gjson.GetBytes(body, "system_instruction.parts.0.text"); si.Exists() {
		req.SystemPrompt = si.String()
	}

	// Generation config.
	if t := gjson.GetBytes(body, "generationConfig.temperature"); t.Exists() {
		v := t.Float()
		req.Temperature = &v
	}
	if mt := gjson.GetBytes(body, "generationConfig.maxOutputTokens"); mt.Exists() {
		v := int(mt.Int())
		req.MaxTokens = &v
	}

	// Stream is determined by the endpoint path (:streamGenerateContent).
	req.Stream = strings.Contains(req.Path, "streamGenerateContent")

	// Tools.
	gjson.GetBytes(body, "tools").ForEach(func(_, t gjson.Result) bool {
		t.Get("functionDeclarations").ForEach(func(_, fd gjson.Result) bool {
			req.Tools = append(req.Tools, store.ToolDefinition{
				Name:        fd.Get("name").String(),
				Description: fd.Get("description").String(),
			})
			return true
		})
		return true
	})

	if len(req.Tools) > 20 {
		req.ManyTools = true
	}

	return nil
}

// ParseResponse extracts content, tokens, and finish reason from a Gemini response.
func (g *Google) ParseResponse(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	// Non-streaming: single response object.
	candidate := gjson.GetBytes(body, "candidates.0")
	if candidate.Exists() {
		var parts []string
		candidate.Get("content.parts").ForEach(func(_, p gjson.Result) bool {
			if t := p.Get("text").String(); t != "" {
				parts = append(parts, t)
			}
			return true
		})
		req.ResponseContent = strings.Join(parts, "")
		req.FinishReason = geminiFinishReason(candidate.Get("finishReason").String())
	}

	// Token usage.
	usage := gjson.GetBytes(body, "usageMetadata")
	if usage.Exists() {
		req.InputTokens = int(usage.Get("promptTokenCount").Int())
		req.OutputTokens = int(usage.Get("candidatesTokenCount").Int())
	}

	return nil
}

// ParseEvent processes a single Gemini SSE event for streaming responses.
func (g *Google) ParseEvent(_ string, data string, req *store.Request) string {
	if data == "" || data == "[DONE]" {
		return ""
	}

	var delta string
	result := gjson.Parse(data)

	// Streaming: each chunk has candidates[0].content.parts.
	candidate := result.Get("candidates.0")
	if candidate.Exists() {
		candidate.Get("content.parts").ForEach(func(_, p gjson.Result) bool {
			if t := p.Get("text").String(); t != "" {
				delta += t
				req.ResponseContent += t
			}
			return true
		})
		if fr := candidate.Get("finishReason").String(); fr != "" {
			req.FinishReason = geminiFinishReason(fr)
		}
	}

	// Final chunk may carry token usage.
	usage := result.Get("usageMetadata")
	if usage.Exists() {
		if n := int(usage.Get("promptTokenCount").Int()); n > 0 {
			req.InputTokens = n
		}
		if n := int(usage.Get("candidatesTokenCount").Int()); n > 0 {
			req.OutputTokens = n
		}
	}

	return delta
}

// geminiFinishReason maps Gemini finish reason strings to probe's FinishReason.
func geminiFinishReason(s string) store.FinishReason {
	switch strings.ToUpper(s) {
	case "STOP":
		return store.FinishStop
	case "MAX_TOKENS":
		return store.FinishLength
	case "SAFETY", "RECITATION", "OTHER":
		return store.FinishStop
	default:
		return store.FinishUnknown
	}
}
