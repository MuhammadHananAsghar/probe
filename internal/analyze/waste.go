package analyze

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// DuplicateGroup holds a set of requests that are considered identical.
type DuplicateGroup struct {
	Requests    []*store.Request
	PotentialSavings float64 // cost of all but the first request
}

// DetectDuplicates finds exact and near-duplicate requests in a session.
// Two requests are duplicates if they share the same model + normalised messages.
func DetectDuplicates(requests []*store.Request) []DuplicateGroup {
	type key struct{ model, hash string }
	groups := make(map[key][]*store.Request)

	for _, r := range requests {
		h := messagesHash(r)
		k := key{model: r.Model, hash: h}
		groups[k] = append(groups[k], r)
	}

	var result []DuplicateGroup
	for _, reqs := range groups {
		if len(reqs) < 2 {
			continue
		}
		// Savings = cost of all but the first occurrence.
		var savings float64
		for _, r := range reqs[1:] {
			savings += r.TotalCost
		}
		result = append(result, DuplicateGroup{Requests: reqs, PotentialSavings: savings})
	}
	return result
}

// messagesHash returns a SHA-256 fingerprint of the request's messages,
// normalising whitespace so trivially different formatting does not prevent
// duplicate detection.
func messagesHash(r *store.Request) string {
	var parts []string
	if r.SystemPrompt != "" {
		parts = append(parts, "sys:"+normalise(r.SystemPrompt))
	}
	for _, m := range r.Messages {
		parts = append(parts, m.Role+":"+normalise(m.Content))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:])
}

func normalise(s string) string {
	// Collapse all whitespace sequences to a single space and trim.
	return strings.Join(strings.Fields(s), " ")
}

// WasteReport holds token waste analysis results for a session.
type WasteReport struct {
	UnchangedSystemPrompt bool
	SystemPromptSavings   float64 // estimated cost savings with prompt caching
	HighInputRatio        []*store.Request
	LongConversations     []*store.Request
	Suggestions           []string
}

// AnalyzeWaste inspects session requests for inefficient token usage patterns.
func AnalyzeWaste(requests []*store.Request) *WasteReport {
	report := &WasteReport{}

	if len(requests) == 0 {
		return report
	}

	// Check if all requests share the same non-empty system prompt.
	firstSys := requests[0].SystemPrompt
	if firstSys != "" {
		allSame := true
		for _, r := range requests[1:] {
			if r.SystemPrompt != firstSys {
				allSame = false
				break
			}
		}
		if allSame && len(requests) > 1 {
			report.UnchangedSystemPrompt = true
			// Rough savings: all but first request pay for system prompt tokens.
			for _, r := range requests[1:] {
				// Estimate system prompt tokens as fraction of input tokens.
				if r.InputTokens > 0 && r.PricingKnown {
					sysRatio := float64(len(firstSys)) / float64(len(firstSys)+totalMsgLen(r))
					report.SystemPromptSavings += r.InputCost * sysRatio
				}
			}
			report.Suggestions = append(report.Suggestions,
				fmt.Sprintf("System prompt is identical across all %d requests — consider Anthropic/OpenAI prompt caching (estimated savings: $%.6f)", len(requests), report.SystemPromptSavings))
		}
	}

	// High input/output ratio.
	for _, r := range requests {
		if r.OutputTokens > 0 && r.InputTokens > 0 {
			ratio := float64(r.InputTokens) / float64(r.OutputTokens)
			if ratio > 10 {
				report.HighInputRatio = append(report.HighInputRatio, r)
			}
		}
	}
	if len(report.HighInputRatio) > 0 {
		report.Suggestions = append(report.Suggestions,
			fmt.Sprintf("%d requests have input/output token ratio > 10:1 — context may be oversized", len(report.HighInputRatio)))
	}

	// Long conversations (5+ turns with growing context).
	convTurns := make(map[string]int)
	for _, r := range requests {
		if r.ConversationID != "" {
			convTurns[r.ConversationID]++
		}
	}
	for _, r := range requests {
		if r.ConversationID != "" && convTurns[r.ConversationID] >= 5 && len(r.Messages) > 8 {
			report.LongConversations = append(report.LongConversations, r)
		}
	}
	if len(report.LongConversations) > 0 {
		report.Suggestions = append(report.Suggestions,
			fmt.Sprintf("%d requests are part of long conversations (5+ turns) — consider summarising context to reduce input tokens", len(report.LongConversations)))
	}

	return report
}

func totalMsgLen(r *store.Request) int {
	total := 0
	for _, m := range r.Messages {
		total += len(m.Content)
	}
	return total
}
