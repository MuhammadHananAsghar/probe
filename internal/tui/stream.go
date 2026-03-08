package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// streamModel manages the stream view for a single request, showing a
// chunk-by-chunk timing timeline.
type streamModel struct {
	req     *store.Request
	scrollY int
	width   int
	height  int
	lines   []string // pre-rendered lines, rebuilt on each new chunk
}

// newStreamModel creates a stream view for req.
func newStreamModel(req *store.Request) streamModel {
	m := streamModel{req: req, width: 80, height: 24}
	m.lines = m.buildLines()
	return m
}

// update refreshes the model with a new version of the request (called when
// streaming chunks arrive).
func (m *streamModel) update(req *store.Request) {
	m.req = req
	m.lines = m.buildLines()
}

// scrollUp moves the viewport up by one line.
func (m *streamModel) scrollUp() {
	if m.scrollY > 0 {
		m.scrollY--
	}
}

// scrollDown moves the viewport down by one line.
func (m *streamModel) scrollDown() {
	max := len(m.lines) - m.height + 4
	if max < 0 {
		max = 0
	}
	if m.scrollY < max {
		m.scrollY++
	}
}

// View renders the stream view inside a border box.
func (m streamModel) View() string {
	visible := m.lines
	if m.scrollY < len(visible) {
		visible = visible[m.scrollY:]
	}
	availH := m.height - 2 // subtract stats bar + header
	if availH < 1 {
		availH = 1
	}
	if len(visible) > availH {
		visible = visible[:availH]
	}

	content := strings.Join(visible, "\n")
	return borderStyle.Width(m.width - 4).Render(content)
}

// buildLines constructs the display lines from the request's stream data.
func (m streamModel) buildLines() []string {
	req := m.req
	var lines []string

	// Header — summary stats.
	if req.StreamStats != nil {
		ss := req.StreamStats
		ttftStr := "—"
		if req.TTFT > 0 {
			ttftStr = FormatDuration(req.TTFT)
		}
		summary := fmt.Sprintf("TTFT: %s  │  Chunks: %d  │  Duration: %s",
			ttftStr,
			ss.ChunkCount,
			FormatDuration(ss.StreamDuration),
		)
		lines = append(lines, headerStyle.Render(summary))
		lines = append(lines, "")
	} else if req.Status == store.StatusStreaming {
		lines = append(lines, warningStyle.Render("⏳ Streaming in progress..."))
		lines = append(lines, "")
	} else {
		lines = append(lines, dimStyle.Render("(no stream data captured)"))
		return lines
	}

	// Timeline header.
	lines = append(lines, dimStyle.Render("Timeline:"))

	// Connect line.
	connectLine := dimStyle.Render("├── 0ms") + "  " + dimStyle.Render("[connect]")
	lines = append(lines, connectLine)

	chunks := req.Chunks
	const maxShown = 20 // show first 10 + last 10 if > 20 chunks

	showChunk := func(i int, c store.StreamChunk) string {
		elapsed := c.ArrivedAt.Sub(req.StartedAt)
		elapsedStr := FormatDuration(elapsed)

		// Content preview (truncate).
		preview := c.Content
		if len(preview) > 20 {
			preview = preview[:20] + "…"
		}
		preview = strings.ReplaceAll(preview, "\n", "↵")

		tag := ""
		if i == 0 && req.TTFT > 0 {
			tag = successStyle.Render(" (TTFT)")
		}
		if c.IsStall {
			tag += warningStyle.Render(fmt.Sprintf(" ⚠ stall %s", FormatDuration(c.Gap)))
		}

		connector := "├──"
		if i == len(chunks)-1 {
			connector = "└──"
		}

		return fmt.Sprintf("%s %s  chunk %d: %q%s",
			dimStyle.Render(connector),
			lipgloss.NewStyle().Foreground(colorBlue).Render(elapsedStr),
			i+1,
			preview,
			tag,
		)
	}

	if len(chunks) <= maxShown {
		for i, c := range chunks {
			lines = append(lines, showChunk(i, c))
		}
	} else {
		for i := 0; i < 10; i++ {
			lines = append(lines, showChunk(i, chunks[i]))
		}
		skipped := len(chunks) - 20
		if req.StreamStats != nil && req.StreamStats.ChunkCount > 0 {
			avgGap := req.StreamStats.StreamDuration / time.Duration(req.StreamStats.ChunkCount)
			lines = append(lines, dimStyle.Render(fmt.Sprintf("│   ... (%d more chunks, avg %s apart)", skipped, FormatDuration(avgGap))))
		} else {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("│   ... (%d more chunks)", skipped)))
		}
		for i := len(chunks) - 10; i < len(chunks); i++ {
			lines = append(lines, showChunk(i, chunks[i]))
		}
	}

	// Done line.
	if req.Status == store.StatusDone || req.Status == store.StatusError {
		endElapsed := req.EndedAt.Sub(req.StartedAt)
		lines = append(lines, dimStyle.Render("└──")+
			" "+lipgloss.NewStyle().Foreground(colorBlue).Render(FormatDuration(endElapsed))+
			"  "+dimStyle.Render("[done]"))
	}

	lines = append(lines, "")

	// Footer stats.
	if req.StreamStats != nil {
		ss := req.StreamStats
		tpsStr := "—"
		if ss.ThroughputTPS > 0 {
			tpsStr = fmt.Sprintf("%.1f tokens/sec", ss.ThroughputTPS)
		}
		lines = append(lines, fmt.Sprintf("Throughput: %s", successStyle.Render(tpsStr)))

		var stallStr string
		if ss.StallCount == 0 {
			stallStr = successStyle.Render("0 (no gaps > " + FormatDuration(ss.StallThreshold) + ")")
		} else {
			stallStr = warningStyle.Render(fmt.Sprintf("%d stalls detected", ss.StallCount))
		}
		lines = append(lines, fmt.Sprintf("Stalls: %s", stallStr))
	}

	lines = append(lines, "")
	lines = append(lines, dimStyle.Render("q: back   j/k: scroll"))

	return lines
}
