package provider

import (
	"fmt"

	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/tidwall/gjson"
)

// OpenAI implements Provider for the OpenAI API and OpenAI-compatible endpoints.
type OpenAI struct{}

// Name returns the provider identifier.
func (o *OpenAI) Name() store.ProviderName {
	return store.ProviderOpenAI
}

// ParseRequest extracts structured fields from an OpenAI chat completions
// request body into req.
func (o *OpenAI) ParseRequest(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	req.Model = gjson.GetBytes(body, "model").String()
	req.Stream = gjson.GetBytes(body, "stream").Bool()

	// Temperature (optional).
	if t := gjson.GetBytes(body, "temperature"); t.Exists() {
		v := t.Float()
		req.Temperature = &v
	}

	// Max tokens (optional).
	if mt := gjson.GetBytes(body, "max_tokens"); mt.Exists() {
		v := int(mt.Int())
		req.MaxTokens = &v
	}

	// Messages.
	gjson.GetBytes(body, "messages").ForEach(func(_, v gjson.Result) bool {
		msg := store.Message{
			Role:    v.Get("role").String(),
			Content: v.Get("content").String(),
		}
		if msg.Role == "system" && req.SystemPrompt == "" {
			req.SystemPrompt = msg.Content
		}
		req.Messages = append(req.Messages, msg)
		return true
	})

	// Tools (new format).
	gjson.GetBytes(body, "tools").ForEach(func(_, v gjson.Result) bool {
		fn := v.Get("function")
		tool := store.ToolDefinition{
			Name:        fn.Get("name").String(),
			Description: fn.Get("description").String(),
			Schema:      fn.Get("parameters").Raw,
		}
		req.Tools = append(req.Tools, tool)
		return true
	})

	// Legacy functions field → tools.
	if len(req.Tools) == 0 {
		gjson.GetBytes(body, "functions").ForEach(func(_, v gjson.Result) bool {
			tool := store.ToolDefinition{
				Name:        v.Get("name").String(),
				Description: v.Get("description").String(),
				Schema:      v.Get("parameters").Raw,
			}
			req.Tools = append(req.Tools, tool)
			return true
		})
	}

	return nil
}

// ParseResponse extracts structured fields from an OpenAI chat completions
// response body into req. For streaming responses this should be called on the
// final assembled body; individual SSE chunks are handled at the proxy layer.
func (o *OpenAI) ParseResponse(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	// HTTP error responses contain an "error" object.
	if req.StatusCode >= 400 {
		msg := gjson.GetBytes(body, "error.message").String()
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", req.StatusCode)
		}
		req.ErrorMessage = msg
		req.Status = store.StatusError
		return nil
	}

	// Override model from response when present (some providers normalise it).
	if m := gjson.GetBytes(body, "model").String(); m != "" {
		req.Model = m
	}

	// Token usage.
	req.InputTokens = int(gjson.GetBytes(body, "usage.prompt_tokens").Int())
	req.OutputTokens = int(gjson.GetBytes(body, "usage.completion_tokens").Int())

	// Primary content.
	req.ResponseContent = gjson.GetBytes(body, "choices.0.message.content").String()

	// Finish reason.
	fr := gjson.GetBytes(body, "choices.0.finish_reason").String()
	req.FinishReason = mapOpenAIFinishReason(fr)

	// Tool calls.
	gjson.GetBytes(body, "choices.0.message.tool_calls").ForEach(func(_, v gjson.Result) bool {
		tc := store.ToolCall{
			ID:            v.Get("id").String(),
			Name:          v.Get("function.name").String(),
			ArgumentsJSON: v.Get("function.arguments").String(),
		}
		req.ToolCalls = append(req.ToolCalls, tc)
		return true
	})

	return nil
}

// mapOpenAIFinishReason converts OpenAI finish_reason strings to the canonical
// store.FinishReason type.
func mapOpenAIFinishReason(s string) store.FinishReason {
	switch s {
	case "stop":
		return store.FinishStop
	case "length":
		return store.FinishLength
	case "tool_calls", "function_call":
		return store.FinishToolCall
	default:
		return store.FinishUnknown
	}
}
