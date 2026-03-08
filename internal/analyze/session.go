package analyze

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// SessionManager groups requests into conversations and builds tool chains.
type SessionManager struct {
	mu    sync.Mutex
	convs map[string]time.Time // convID -> last activity
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		convs: make(map[string]time.Time),
	}
}

// newConvID generates an 8-hex-char conversation ID from a seed string.
func newConvID(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(h[:4])
}

// messagesSharePrefix returns true if all messages in prev appear at the start
// of req (compared by role + first 32 chars of content).
func messagesSharePrefix(prev, req []*store.Message) bool {
	if len(prev) == 0 || len(req) == 0 {
		return false
	}
	limit := len(prev)
	if len(req) < limit {
		return false
	}
	for i := 0; i < limit; i++ {
		pRole := prev[i].Role
		rRole := req[i].Role
		if pRole != rRole {
			return false
		}
		pContent := prev[i].Content
		rContent := req[i].Content
		if len(pContent) > 32 {
			pContent = pContent[:32]
		}
		if len(rContent) > 32 {
			rContent = rContent[:32]
		}
		if pContent != rContent {
			return false
		}
	}
	return true
}

// AssignConversation assigns a ConversationID to req by examining all
// previously captured requests. It mutates req.ConversationID in place.
func (sm *SessionManager) AssignConversation(req *store.Request, all []*store.Request) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Build pointer slices for req messages.
	reqMsgPtrs := make([]*store.Message, len(req.Messages))
	for i := range req.Messages {
		reqMsgPtrs[i] = &req.Messages[i]
	}

	// If this request has tool results, find the most recent prior request
	// whose tool calls match any of the tool result IDs.
	if len(req.ToolResults) > 0 {
		// Build a set of ToolCallIDs in this request's ToolResults.
		resultIDs := make(map[string]struct{}, len(req.ToolResults))
		for _, tr := range req.ToolResults {
			resultIDs[tr.ToolCallID] = struct{}{}
		}

		// Walk all in reverse to find most recent parent.
		for i := len(all) - 1; i >= 0; i-- {
			prev := all[i]
			if prev.ID == req.ID {
				continue
			}
			for _, tc := range prev.ToolCalls {
				if _, ok := resultIDs[tc.ID]; ok {
					req.ConversationID = prev.ConversationID
					sm.convs[req.ConversationID] = req.StartedAt
					return
				}
			}
		}
	}

	// Otherwise check if req is a continuation of any recent request via message
	// prefix matching.
	for i := len(all) - 1; i >= 0; i-- {
		prev := all[i]
		if prev.ID == req.ID {
			continue
		}
		if len(prev.Messages) == 0 || len(req.Messages) <= len(prev.Messages) {
			continue
		}
		prevMsgPtrs := make([]*store.Message, len(prev.Messages))
		for j := range prev.Messages {
			prevMsgPtrs[j] = &prev.Messages[j]
		}
		if messagesSharePrefix(prevMsgPtrs, reqMsgPtrs) {
			req.ConversationID = prev.ConversationID
			sm.convs[req.ConversationID] = req.StartedAt
			return
		}
	}

	// No parent found; generate a new conversation ID from the system prompt
	// and first user message content (first 64 chars).
	seed := req.SystemPrompt
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			content := msg.Content
			if len(content) > 64 {
				content = content[:64]
			}
			seed += content
			break
		}
	}
	convID := newConvID(fmt.Sprintf("%s:%s", req.StartedAt.Format(time.RFC3339Nano), seed))
	req.ConversationID = convID
	sm.convs[convID] = req.StartedAt
}

// BuildToolChain assembles a []store.ToolChainStep from requests belonging to
// the same conversation (already filtered and ordered by time).
func BuildToolChain(reqs []*store.Request) []store.ToolChainStep {
	var steps []store.ToolChainStep
	stepNum := 0

	for i, req := range reqs {
		for _, tc := range req.ToolCalls {
			stepNum++
			step := store.ToolChainStep{
				StepNum:    stepNum,
				ToolCallID: tc.ID,
				ToolName:   tc.Name,
				Arguments:  tc.ArgumentsJSON,
				CallReqID:  req.ID,
			}

			// Find matching ToolResult in the next request.
			if i+1 < len(reqs) {
				next := reqs[i+1]
				for _, tr := range next.ToolResults {
					if tr.ToolCallID == tc.ID {
						step.Result = tr.Content
						step.IsError = tr.IsError
						step.ResultReqID = next.ID
						step.Latency = next.StartedAt.Sub(req.EndedAt)
						break
					}
				}
			}

			steps = append(steps, step)
		}
	}

	return steps
}
