package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

const maxContentLen = 120

// detailModel holds the state for the request detail view.
type detailModel struct {
	req     *store.Request
	scrollY int
	width   int
	height  int
	lines   []string // pre-rendered content lines (rebuilt on SetRequest)
}

// newDetailModel creates a detail view for the given request.
func newDetailModel(req *store.Request) detailModel {
	m := detailModel{req: req}
	return m
}

// truncate shortens s to at most n runes, appending "..." if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// sectionHeader renders a horizontal rule with a section label.
func sectionHeader(label string, width int) string {
	line := fmt.Sprintf(" ── %s ", label)
	remaining := width - len(line) - 2
	if remaining > 0 {
		line += strings.Repeat("─", remaining)
	}
	return dimStyle.Render(line)
}

// buildLines constructs the list of content lines for the given request and width.
func buildLines(req *store.Request, width int) []string {
	if req == nil {
		return []string{dimStyle.Render("  No request selected.")}
	}

	// Inner width accounts for border (2) and padding (2 each side)
	inner := width - 6
	if inner < 20 {
		inner = 20
	}

	var lines []string
	add := func(s string) { lines = append(lines, s) }
	blank := func() { add("") }

	blank()

	// Provider / Model
	providerLabel := headerStyle.Render("Provider: ")
	providerVal := ProviderBadge(string(req.Provider))
	modelLabel := headerStyle.Render("  Model: ")
	modelVal := modelStyle.Render(req.Model)
	add(fmt.Sprintf(" %s%s%s%s", providerLabel, providerVal, modelLabel, modelVal))

	// Endpoint / Status
	epLabel := headerStyle.Render("Endpoint: ")
	epVal := fmt.Sprintf("%s %s", req.Method, req.Path)
	statusLabel := headerStyle.Render("  Status: ")

	var statusVal string
	switch req.Status {
	case store.StatusPending, store.StatusStreaming:
		statusVal = warningStyle.Render("⏳ " + string(req.Status))
	case store.StatusDone:
		code := req.StatusCode
		mark := "✓"
		if code >= 400 {
			mark = "✗"
		}
		statusVal = lipgloss.NewStyle().Foreground(StatusColor(code)).Render(fmt.Sprintf("%d %s", code, mark))
	case store.StatusError:
		statusVal = errorStyle.Render("error ✗")
	default:
		statusVal = dimStyle.Render("—")
	}
	add(fmt.Sprintf(" %s%s%s%s", epLabel, epVal, statusLabel, statusVal))

	blank()
	add(sectionHeader("Timing", inner))

	totalStr := FormatDuration(req.Latency)
	ttftStr := dimStyle.Render("—")
	if req.TTFT > 0 {
		ttftStr = FormatDuration(req.TTFT)
	}
	add(fmt.Sprintf(" Total: %s    TTFT: %s", totalStr, ttftStr))

	blank()
	add(sectionHeader("Tokens & Cost", inner))

	inputCostStr := FormatCost(req.InputCost, req.PricingKnown)
	outputCostStr := FormatCost(req.OutputCost, req.PricingKnown)
	totalCostStr := FormatCost(req.TotalCost, req.PricingKnown)
	add(fmt.Sprintf(" Input:  %s tokens (%s)", FormatTokens(req.InputTokens), inputCostStr))
	add(fmt.Sprintf(" Output: %s tokens (%s)", FormatTokens(req.OutputTokens), outputCostStr))
	add(fmt.Sprintf(" Total:  %s", totalCostStr))

	// Messages
	blank()
	msgCount := len(req.Messages)
	if req.SystemPrompt != "" {
		msgCount++ // count system prompt as a message
	}
	add(sectionHeader(fmt.Sprintf("Messages (%d)", msgCount), inner))

	if req.SystemPrompt != "" {
		content := truncate(req.SystemPrompt, maxContentLen)
		add(fmt.Sprintf(" %s  %s",
			dimStyle.Render("[system]"),
			content,
		))
	}
	if len(req.Messages) == 0 && req.SystemPrompt == "" {
		add(dimStyle.Render(" (none)"))
	}
	for _, msg := range req.Messages {
		content := truncate(msg.Content, maxContentLen)
		roleStr := dimStyle.Render(fmt.Sprintf("[%s]", msg.Role))
		add(fmt.Sprintf(" %s  %s", roleStr, content))
	}

	// Response
	blank()
	add(sectionHeader("Response", inner))
	if req.ResponseContent == "" {
		add(dimStyle.Render(" (empty)"))
	} else {
		content := truncate(req.ResponseContent, maxContentLen)
		add(fmt.Sprintf(" %s", content))
	}

	// Tool Calls
	blank()
	add(sectionHeader("Tool Calls", inner))
	if len(req.ToolCalls) == 0 {
		add(dimStyle.Render(" (none)"))
	} else {
		for _, tc := range req.ToolCalls {
			args := truncate(tc.ArgumentsJSON, maxContentLen)
			add(fmt.Sprintf(" %s(%s)", modelStyle.Render(tc.Name), dimStyle.Render(args)))
		}
	}

	// Rate Limits
	blank()
	add(sectionHeader("Rate Limits", inner))
	if req.RateLimitRemaining == 0 && req.RateLimitReset.IsZero() {
		add(dimStyle.Render(" (none captured)"))
	} else {
		if req.RateLimitRemaining > 0 {
			add(fmt.Sprintf(" Remaining: %d", req.RateLimitRemaining))
		}
		if !req.RateLimitReset.IsZero() {
			add(fmt.Sprintf(" Reset at:  %s", req.RateLimitReset.Format("15:04:05")))
		}
	}

	// Error
	if req.ErrorMessage != "" {
		blank()
		add(sectionHeader("Error", inner))
		add(errorStyle.Render(" " + truncate(req.ErrorMessage, maxContentLen)))
	}

	// Streaming stats
	if req.StreamStats != nil {
		ss := req.StreamStats
		blank()
		add(sectionHeader("Stream Stats", inner))
		add(fmt.Sprintf(" Chunks: %d   Throughput: %.1f tok/s   Stalls: %d",
			ss.ChunkCount, ss.ThroughputTPS, ss.StallCount))
	}

	blank()
	// Help line
	add(dimStyle.Render("  q/Esc: back  ↑/↓ or j/k: scroll  r: replay (Phase 5)  e: export (Phase 6)"))
	blank()

	return lines
}

// rebuild regenerates the cached content lines.
func (m *detailModel) rebuild() {
	m.lines = buildLines(m.req, m.width)
}

// scrollUp moves the viewport up by one line.
func (m *detailModel) scrollUp() {
	if m.scrollY > 0 {
		m.scrollY--
	}
}

// scrollDown moves the viewport down by one line.
func (m *detailModel) scrollDown() {
	maxScroll := len(m.lines) - m.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollY < maxScroll {
		m.scrollY++
	}
}

// View renders the full detail view inside a rounded border.
func (m detailModel) View() string {
	if m.req == nil {
		return borderStyle.Render(dimStyle.Render("No request selected."))
	}

	lines := m.lines
	if len(lines) == 0 {
		lines = buildLines(m.req, m.width)
	}

	// Visible window
	visibleHeight := m.height - 2 // subtract border rows
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	start := m.scrollY
	if start > len(lines) {
		start = len(lines)
	}
	end := start + visibleHeight
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[start:end]

	// Pad to visibleHeight so the border box is always full height
	for len(visible) < visibleHeight {
		visible = append(visible, "")
	}

	title := titleStyle.Render(fmt.Sprintf("Request #%d", m.req.Seq))

	innerWidth := m.width - 4 // border(2) + padding(2)
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Pad each line to inner width so borders align.
	var paddedLines []string
	for _, l := range visible {
		raw := lipgloss.Width(l)
		if raw < innerWidth {
			l += strings.Repeat(" ", innerWidth-raw)
		}
		paddedLines = append(paddedLines, l)
	}

	body := strings.Join(paddedLines, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorDim).
		Padding(0, 1).
		Width(m.width - 2).
		Render(title + "\n" + body)
}
