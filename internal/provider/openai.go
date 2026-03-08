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
		role := v.Get("role").String()
		msg := store.Message{
			Role:    role,
			Content: v.Get("content").String(),
		}
		if role == "system" && req.SystemPrompt == "" {
			req.SystemPrompt = msg.Content
		}
		// Parse tool results from messages with role "tool".
		if role == "tool" {
			toolCallID := v.Get("tool_call_id").String()
			content := v.Get("content").String()
			if toolCallID != "" {
				req.ToolResults = append(req.ToolResults, store.ToolResult{
					ToolCallID: toolCallID,
					Content:    content,
				})
			}
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

	req.ManyTools = len(req.Tools) > 20

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

// ParseEvent implements StreamParser for OpenAI's SSE format.
// OpenAI events have no "event:" line — eventType is always "".
// data is a JSON string like:
//
//	{"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}
//
// or the final chunk with usage:
//
//	{"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5}}
func (o *OpenAI) ParseEvent(eventType, data string, req *store.Request) string {
	if data == "[DONE]" || len(data) == 0 {
		return ""
	}
	b := []byte(data)

	// Finish reason (may be null mid-stream).
	fr := gjson.GetBytes(b, "choices.0.finish_reason")
	if fr.Exists() && fr.Type != gjson.Null {
		req.FinishReason = mapOpenAIFinishReason(fr.String())
	}

	// Tool call deltas: accumulate all tool calls across SSE chunks.
	// OpenAI sends choices[0].delta.tool_calls as an array with an index field.
	gjson.GetBytes(b, "choices.0.delta.tool_calls").ForEach(func(_, tc gjson.Result) bool {
		idx := int(tc.Get("index").Int())
		// Grow the slice if needed.
		for len(req.ToolCalls) <= idx {
			req.ToolCalls = append(req.ToolCalls, store.ToolCall{})
		}
		// First chunk carries id and function.name.
		if id := tc.Get("id").String(); id != "" {
			req.ToolCalls[idx].ID = id
		}
		if name := tc.Get("function.name").String(); name != "" {
			req.ToolCalls[idx].Name = name
		}
		// Accumulate argument delta.
		if args := tc.Get("function.arguments").String(); args != "" {
			req.ToolCalls[idx].ArgumentsJSON += args
		}
		return true
	})

	// Usage in final chunk (some models include it).
	if pt := gjson.GetBytes(b, "usage.prompt_tokens"); pt.Exists() {
		req.InputTokens = int(pt.Int())
	}
	if ct := gjson.GetBytes(b, "usage.completion_tokens"); ct.Exists() {
		req.OutputTokens = int(ct.Int())
	}

	// Content delta.
	return gjson.GetBytes(b, "choices.0.delta.content").String()
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
