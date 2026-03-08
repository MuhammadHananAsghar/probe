package provider

import (
	"fmt"

	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/tidwall/gjson"
)

// Anthropic implements Provider for the Anthropic Messages API.
type Anthropic struct{}

// Name returns the provider identifier.
func (a *Anthropic) Name() store.ProviderName {
	return store.ProviderAnthropic
}

// ParseRequest extracts structured fields from an Anthropic messages request
// body into req.
func (a *Anthropic) ParseRequest(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	req.Model = gjson.GetBytes(body, "model").String()
	req.Stream = gjson.GetBytes(body, "stream").Bool()

	// Max tokens.
	if mt := gjson.GetBytes(body, "max_tokens"); mt.Exists() {
		v := int(mt.Int())
		req.MaxTokens = &v
	}

	// System prompt: may be a plain string or an array of content blocks.
	sys := gjson.GetBytes(body, "system")
	if sys.Exists() {
		switch sys.Type {
		case gjson.String:
			req.SystemPrompt = sys.String()
		case gjson.JSON:
			// Array of content blocks — concatenate text blocks.
			var parts []byte
			sys.ForEach(func(_, block gjson.Result) bool {
				if block.Get("type").String() == "text" {
					if len(parts) > 0 {
						parts = append(parts, '\n')
					}
					parts = append(parts, block.Get("text").String()...)
				}
				return true
			})
			req.SystemPrompt = string(parts)
		}
	}

	// Messages.
	gjson.GetBytes(body, "messages").ForEach(func(_, v gjson.Result) bool {
		role := v.Get("role").String()
		// Content may be a string or an array of content blocks.
		content := v.Get("content")
		var text string
		if content.Type == gjson.String {
			text = content.String()
		} else {
			// Concatenate text blocks.
			content.ForEach(func(_, block gjson.Result) bool {
				if block.Get("type").String() == "text" {
					text += block.Get("text").String()
				}
				return true
			})
		}
		req.Messages = append(req.Messages, store.Message{
			Role:    role,
			Content: text,
		})
		return true
	})

	// Tools.
	gjson.GetBytes(body, "tools").ForEach(func(_, v gjson.Result) bool {
		tool := store.ToolDefinition{
			Name:        v.Get("name").String(),
			Description: v.Get("description").String(),
			Schema:      v.Get("input_schema").Raw,
		}
		req.Tools = append(req.Tools, tool)
		return true
	})

	return nil
}

// ParseResponse extracts structured fields from an Anthropic messages response
// body into req.
func (a *Anthropic) ParseResponse(body []byte, req *store.Request) error {
	if len(body) == 0 {
		return nil
	}

	// HTTP error responses.
	if req.StatusCode >= 400 {
		msg := gjson.GetBytes(body, "error.message").String()
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", req.StatusCode)
		}
		req.ErrorMessage = msg
		req.Status = store.StatusError
		return nil
	}

	// Token usage.
	req.InputTokens = int(gjson.GetBytes(body, "usage.input_tokens").Int())
	req.OutputTokens = int(gjson.GetBytes(body, "usage.output_tokens").Int())

	// Finish reason.
	stopReason := gjson.GetBytes(body, "stop_reason").String()
	req.FinishReason = mapAnthropicStopReason(stopReason)

	// Content blocks.
	gjson.GetBytes(body, "content").ForEach(func(_, block gjson.Result) bool {
		blockType := block.Get("type").String()
		switch blockType {
		case "text":
			if req.ResponseContent == "" {
				req.ResponseContent = block.Get("text").String()
			}
		case "tool_use":
			tc := store.ToolCall{
				ID:            block.Get("id").String(),
				Name:          block.Get("name").String(),
				ArgumentsJSON: block.Get("input").Raw,
			}
			req.ToolCalls = append(req.ToolCalls, tc)
		}
		return true
	})

	return nil
}

// ParseEvent implements StreamParser for Anthropic's SSE format.
// Anthropic sends named events:
//
//	event: message_start      → usage.input_tokens
//	event: content_block_delta → delta.text
//	event: message_delta      → usage.output_tokens, stop_reason
func (a *Anthropic) ParseEvent(eventType, data string, req *store.Request) string {
	if len(data) == 0 {
		return ""
	}
	b := []byte(data)

	switch eventType {
	case "message_start":
		it := gjson.GetBytes(b, "message.usage.input_tokens")
		if it.Exists() {
			req.InputTokens = int(it.Int())
		}
	case "content_block_delta":
		deltaType := gjson.GetBytes(b, "delta.type").String()
		if deltaType == "text_delta" {
			return gjson.GetBytes(b, "delta.text").String()
		}
		// tool_use input_json_delta — Phase 3
	case "message_delta":
		ot := gjson.GetBytes(b, "usage.output_tokens")
		if ot.Exists() {
			req.OutputTokens = int(ot.Int())
		}
		if sr := gjson.GetBytes(b, "delta.stop_reason"); sr.Exists() {
			req.FinishReason = mapAnthropicStopReason(sr.String())
		}
	}
	return ""
}

// mapAnthropicStopReason converts Anthropic stop_reason strings to the
// canonical store.FinishReason type.
func mapAnthropicStopReason(s string) store.FinishReason {
	switch s {
	case "end_turn":
		return store.FinishStop
	case "max_tokens":
		return store.FinishLength
	case "tool_use":
		return store.FinishToolCall
	default:
		return store.FinishUnknown
	}
}
