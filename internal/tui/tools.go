package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// toolsModel renders the tool call tree view for a single request.
type toolsModel struct {
	req     *store.Request
	chain   []store.ToolChainStep // cross-request chain (may be empty for single-request view)
	scrollY int
	width   int
	height  int
	lines   []string
}

// openToolsMsg is sent when the user presses 't' in the detail view.
type openToolsMsg struct{ Req *store.Request }

// newToolsModel creates a tools view for req with an optional cross-request chain.
func newToolsModel(req *store.Request, chain []store.ToolChainStep) toolsModel {
	m := toolsModel{req: req, chain: chain, width: 80, height: 24}
	m.lines = m.buildLines()
	return m
}

func (m *toolsModel) scrollUp() {
	if m.scrollY > 0 {
		m.scrollY--
	}
}

func (m *toolsModel) scrollDown() {
	max := len(m.lines) - m.height + 2
	if max < 0 {
		max = 0
	}
	if m.scrollY < max {
		m.scrollY++
	}
}

// buildLines constructs the pre-rendered lines for the tools view.
func (m *toolsModel) buildLines() []string {
	req := m.req
	if req == nil {
		return []string{dimStyle.Render("  No request selected.")}
	}

	var lines []string
	add := func(s string) { lines = append(lines, s) }
	blank := func() { add("") }

	inner := m.width - 6
	if inner < 20 {
		inner = 20
	}

	blank()
	// Header
	title := titleStyle.Render(fmt.Sprintf("Request #%d — Tool Call Inspector", req.Seq))
	add(" " + title)
	blank()

	// Anomaly warnings
	if len(req.Anomalies) > 0 {
		add(sectionHeader("Warnings", inner))
		for _, a := range req.Anomalies {
			icon := "⚠"
			style := warningStyle
			if a.Kind == store.AnomalyMalformedArgs || a.Kind == store.AnomalyToolLoop {
				icon = "✗"
				style = errorStyle
			}
			add(fmt.Sprintf("  %s %s", style.Render(icon), a.Message))
		}
		blank()
	}

	// Tool definitions summary
	toolCount := len(req.Tools)
	manyToolsNote := ""
	if req.ManyTools {
		manyToolsNote = warningStyle.Render("  ⚠ >20 tools may degrade model behavior")
	}
	add(sectionHeader(fmt.Sprintf("Tool Definitions (%d)%s", toolCount, manyToolsNote), inner))
	if toolCount == 0 {
		add(dimStyle.Render("  (none defined in this request)"))
	} else {
		for i, t := range req.Tools {
			desc := t.Description
			if desc == "" {
				desc = "(no description)"
			}
			desc = truncate(desc, 80)
			add(fmt.Sprintf("  %s  %s  %s",
				dimStyle.Render(fmt.Sprintf("%2d.", i+1)),
				modelStyle.Render(t.Name),
				dimStyle.Render(desc),
			))
		}
	}
	blank()

	// Use chain if available, otherwise render tool calls from this request only
	if len(m.chain) > 0 {
		totalLatency := int64(0)
		for _, s := range m.chain {
			totalLatency += s.Latency.Milliseconds()
		}
		summary := fmt.Sprintf("%d tool call(s) | %dms total execution", len(m.chain), totalLatency)
		add(sectionHeader("Tool Call Chain  "+dimStyle.Render(summary), inner))
		blank()
		renderChain(m.chain, inner, add)
	} else {
		// Single-request view
		calls := req.ToolCalls
		results := req.ToolResults

		resultMap := make(map[string]store.ToolResult, len(results))
		for _, r := range results {
			resultMap[r.ToolCallID] = r
		}

		summary := fmt.Sprintf("%d tool call(s)", len(calls))
		add(sectionHeader("Tool Calls  "+dimStyle.Render(summary), inner))
		blank()

		if len(calls) == 0 {
			add(dimStyle.Render("  (no tool calls in this response)"))
		} else {
			for i, tc := range calls {
				renderToolCall(i+1, tc, resultMap[tc.ID], 0, inner, add)
				blank()
			}
		}
	}

	// Tool Results in this request (what was sent back to model)
	if len(req.ToolResults) > 0 {
		blank()
		add(sectionHeader(fmt.Sprintf("Tool Results Sent (%d)", len(req.ToolResults)), inner))
		blank()
		for i, tr := range req.ToolResults {
			errMark := ""
			if tr.IsError {
				errMark = " " + errorStyle.Render("(error)")
			}
			add(fmt.Sprintf("  %s  id: %s%s",
				dimStyle.Render(fmt.Sprintf("%2d.", i+1)),
				dimStyle.Render(truncate(tr.ToolCallID, 20)),
				errMark,
			))
			content := truncate(tr.Content, 120)
			add(fmt.Sprintf("       %s", dimStyle.Render(content)))
		}
	}

	blank()
	add(dimStyle.Render("  q/Esc: back    ↑/↓ or j/k: scroll"))
	blank()

	return lines
}

// renderChain renders cross-request tool chain steps.
func renderChain(chain []store.ToolChainStep, width int, add func(string)) {
	for _, step := range chain {
		tc := store.ToolCall{
			ID:            step.ToolCallID,
			Name:          step.ToolName,
			ArgumentsJSON: step.Arguments,
		}
		result := store.ToolResult{
			ToolCallID: step.ToolCallID,
			Content:    step.Result,
			IsError:    step.IsError,
		}
		renderToolCall(step.StepNum, tc, result, step.Latency.Milliseconds(), width, add)
		add("")
	}
}

// renderToolCall renders a single tool call + optional result block.
func renderToolCall(step int, tc store.ToolCall, result store.ToolResult, latencyMs int64, width int, add func(string)) {
	// Step header
	errBadge := ""
	if tc.ParseError {
		errBadge = " " + errorStyle.Render("[malformed JSON]")
	}
	add(fmt.Sprintf("  %s  %s%s",
		dimStyle.Render(fmt.Sprintf("Step %d ──", step)),
		modelStyle.Render("call: "+tc.Name),
		errBadge,
	))

	// Arguments
	args := tc.ArgumentsJSON
	if args == "" {
		args = "{}"
	}
	prettyArgs := prettyJSON(args, 80)
	for _, line := range strings.Split(prettyArgs, "\n") {
		add(fmt.Sprintf("          %s", dimStyle.Render(line)))
	}

	// Result (if any)
	if result.Content != "" || result.ToolCallID != "" {
		errNote := ""
		if result.IsError {
			errNote = errorStyle.Render(" [error]")
		}
		latNote := ""
		if latencyMs > 0 {
			latNote = fmt.Sprintf("  %s", dimStyle.Render(fmt.Sprintf("(%dms)", latencyMs)))
		}
		add(fmt.Sprintf("          %s%s%s",
			successStyle.Render("result:"),
			errNote,
			latNote,
		))
		content := truncate(result.Content, 120)
		add(fmt.Sprintf("          %s", dimStyle.Render(content)))
	} else {
		add(fmt.Sprintf("          %s", warningStyle.Render("⏳ awaiting result...")))
	}
}

// prettyJSON attempts to pretty-print JSON. Falls back to raw string on error.
func prettyJSON(s string, maxLen int) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return truncate(s, maxLen)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return truncate(s, maxLen)
	}
	out := string(b)
	if len(out) > maxLen {
		return out[:maxLen-3] + "..."
	}
	return out
}

// View renders the tools view.
func (m toolsModel) View() string {
	if m.req == nil {
		return borderStyle.Render(dimStyle.Render("No request selected."))
	}

	lines := m.lines
	if len(lines) == 0 {
		lines = m.buildLines()
	}

	visibleHeight := m.height - 2
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
	for len(visible) < visibleHeight {
		visible = append(visible, "")
	}

	innerWidth := m.width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

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
		Render(body)
}
