package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// spinnerFrames are the characters used to animate pending/streaming requests.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// listModel manages the scrollable request list.
type listModel struct {
	requests    []*store.Request
	cursor      int
	offset      int  // scroll offset (index of first visible row)
	height      int  // available height for the list body
	width       int
	autoScroll  bool // true until user scrolls up manually
	spinnerTick int  // current frame index for spinner animation
}

// newListModel creates an empty list model with auto-scroll enabled.
func newListModel() listModel {
	return listModel{
		autoScroll: true,
	}
}

// upsert adds a new request or replaces an existing one by ID.
func (m *listModel) upsert(req *store.Request) {
	for i, r := range m.requests {
		if r.ID == req.ID {
			m.requests[i] = req
			return
		}
	}
	m.requests = append(m.requests, req)
	if m.autoScroll {
		m.cursor = len(m.requests) - 1
		m.scrollToCursor()
	}
}

// scrollToCursor adjusts the offset so the cursor row is visible.
func (m *listModel) scrollToCursor() {
	if m.height <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
}

// moveUp moves the cursor up by one row.
func (m *listModel) moveUp() {
	if m.cursor > 0 {
		m.cursor--
		m.autoScroll = false
		m.scrollToCursor()
	}
}

// moveDown moves the cursor down by one row.
func (m *listModel) moveDown() {
	if m.cursor < len(m.requests)-1 {
		m.cursor++
		// Re-enable auto-scroll if we've reached the bottom.
		if m.cursor == len(m.requests)-1 {
			m.autoScroll = true
		}
		m.scrollToCursor()
	}
}

// selected returns the currently highlighted request, or nil if the list is empty.
func (m *listModel) selected() *store.Request {
	if len(m.requests) == 0 || m.cursor >= len(m.requests) {
		return nil
	}
	return m.requests[m.cursor]
}

// renderRow renders a single request row.
// Format: " #N  METHOD provider path  model  in→out  $cost  latency  status"
func renderRow(req *store.Request, selected bool, width int, spinnerTick int) string {
	seq := fmt.Sprintf("#%-3d", req.Seq)
	method := req.Method
	if method == "" {
		method = "—"
	}

	provider := ProviderBadge(string(req.Provider))
	path := req.Path
	if path == "" {
		path = "/"
	}
	// Truncate long paths.
	if len(path) > 30 {
		path = path[:27] + "..."
	}

	model := modelStyle.Render(req.Model)
	if req.Model == "" {
		model = dimStyle.Render("—")
	}

	tokens := dimStyle.Render("—")
	if req.InputTokens > 0 || req.OutputTokens > 0 {
		tokens = fmt.Sprintf("%s→%s", FormatTokens(req.InputTokens), FormatTokens(req.OutputTokens))
	}

	cost := FormatCost(req.TotalCost, req.PricingKnown)

	latency := dimStyle.Render("—")
	if req.Latency > 0 {
		latency = FormatDuration(req.Latency)
	}

	// Status indicator
	var statusStr string
	switch req.Status {
	case store.StatusPending, store.StatusStreaming:
		frame := spinnerFrames[spinnerTick%len(spinnerFrames)]
		statusStr = warningStyle.Render(frame)
	case store.StatusDone:
		code := req.StatusCode
		codeColor := StatusColor(code)
		mark := "✓"
		if code >= 400 {
			mark = "✗"
		}
		statusStr = lipgloss.NewStyle().Foreground(codeColor).Render(fmt.Sprintf("%d %s", code, mark))
	case store.StatusError:
		statusStr = errorStyle.Render("ERR ✗")
	default:
		statusStr = dimStyle.Render("—")
	}

	row := fmt.Sprintf(" %s  %s %s %s  %s  %s  %s  %s  %s",
		dimStyle.Render(seq),
		method,
		provider,
		dimStyle.Render(path),
		model,
		tokens,
		cost,
		latency,
		statusStr,
	)

	if selected {
		// Pad to width and apply selected background.
		rawLen := lipgloss.Width(row)
		if rawLen < width {
			row += strings.Repeat(" ", width-rawLen)
		}
		return selectedStyle.Render(row)
	}
	return row
}

// View renders the full list, showing only the visible window of rows.
func (m listModel) View() string {
	if len(m.requests) == 0 {
		return dimStyle.Render("  Waiting for requests…")
	}

	end := m.offset + m.height
	if end > len(m.requests) {
		end = len(m.requests)
	}

	var lines []string
	for i := m.offset; i < end; i++ {
		selected := i == m.cursor
		lines = append(lines, renderRow(m.requests[i], selected, m.width, m.spinnerTick))
	}
	return strings.Join(lines, "\n")
}

// header returns the column header line for the list.
func listHeader() string {
	return headerStyle.Render("  #    METHOD PROVIDER PATH                           MODEL                    TOKENS         COST      LATENCY  STATUS")
}
