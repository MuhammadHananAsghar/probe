package tui

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/MuhammadHananAsghar/probe/internal/analyze"
	"github.com/MuhammadHananAsghar/probe/internal/export"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// curlCopiedMsg is emitted when a curl command has been copied to clipboard.
type curlCopiedMsg struct{ ok bool }

// ViewMode tracks which view is currently active.
type ViewMode int

const (
	// ViewList is the default scrollable request list view.
	ViewList ViewMode = iota
	// ViewDetail shows full details for a selected request.
	ViewDetail
	// ViewStream shows the chunk-by-chunk streaming timeline for a request.
	ViewStream
	// ViewTools shows the tool call inspector for a selected request.
	ViewTools
)

// RequestMsg is a bubbletea message carrying a new or updated request.
type RequestMsg struct{ Req *store.Request }

// TickMsg is sent periodically for spinner animation updates.
type TickMsg struct{ T time.Time }

// openStreamMsg is sent when the user presses 's' in the detail view.
type openStreamMsg struct{ Req *store.Request }

// tickInterval controls how frequently the spinner advances.
const tickInterval = 100 * time.Millisecond

// AlertConfig holds thresholds for cost/latency/error alerting.
type AlertConfig struct {
	CostThreshold    float64
	LatencyThreshold time.Duration
	AlertOnError     bool
}

// App is the main bubbletea model that composes the list, detail, stream, and tools views.
type App struct {
	mode        ViewMode
	list        listModel
	detail      detailModel
	stream      streamModel
	tools       toolsModel
	stats       store.SessionStats
	tracker     *analyze.Tracker
	width       int
	height      int
	proxyAddr   string
	dashAddr    string
	spinnerTick int
	reqCh       <-chan *store.Request
	statusMsg   string // transient status shown in the hints bar
	statusAt    time.Time
	alerts      []string    // active alert banners (dismissed with 'd')
	alertCfg    AlertConfig
	sessionCost float64     // running session total for cost alerting
	costAlerted bool        // true once cost alert has fired this session
}

// New creates a new App model ready to be started with bubbletea.
// proxyAddr and dashAddr are shown in the stats bar.
// tracker is used to pull updated session stats after each request.
// reqCh is the channel over which the proxy delivers new/updated requests.
func New(
	proxyAddr, dashAddr string,
	tracker *analyze.Tracker,
	reqCh <-chan *store.Request,
) *App {
	const defaultH = 40
	const statsBarH = 1
	const headerH = 1
	const hintsH = 2
	initialListH := defaultH - statsBarH - headerH - bannerHeight - hintsH
	if initialListH < 1 {
		initialListH = 1
	}
	l := newListModel()
	l.height = initialListH
	l.width = 120

	return &App{
		mode:      ViewList,
		list:      l,
		proxyAddr: proxyAddr,
		dashAddr:  dashAddr,
		tracker:   tracker,
		reqCh:     reqCh,
		width:     120,
		height:    defaultH,
	}
}

// WithAlerts configures alert thresholds on an existing App.
func (a *App) WithAlerts(cfg AlertConfig) {
	a.alertCfg = cfg
}

// Init implements tea.Model. It starts the request listener and the ticker.
func (a App) Init() tea.Cmd {
	return tea.Batch(
		ListenForRequests(a.reqCh),
		tickCmd(),
	)
}

// Update implements tea.Model.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {

	// ── Terminal resize ──────────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height

		const statsBarH = 1
		const headerH = 1
		const hintsH = 2 // hints line + blank line before it
		listHeight := a.height - statsBarH - headerH - bannerHeight - hintsH
		if listHeight < 1 {
			listHeight = 1
		}
		a.list.width = a.width
		a.list.height = listHeight

		a.detail.width = a.width
		a.detail.height = a.height - statsBarH - hintsH
		a.detail.rebuild()

		a.stream.width = a.width
		a.stream.height = a.height - statsBarH - hintsH

	// ── New / updated request from proxy ─────────────────────────────────────
	case RequestMsg:
		a.list.upsert(m.Req)
		if a.tracker != nil {
			a.stats = a.tracker.Stats()
		}
		// Check alerts on completed requests.
		if m.Req.Status == store.StatusDone || m.Req.Status == store.StatusError {
			a.checkAlerts(m.Req)
		}
		// Re-render detail if it's the currently shown request.
		if a.mode == ViewDetail && a.detail.req != nil && a.detail.req.ID == m.Req.ID {
			a.detail.req = m.Req
			a.detail.rebuild()
		}
		// Refresh stream view if the incoming update matches the displayed request.
		if a.mode == ViewStream && a.stream.req != nil && a.stream.req.ID == m.Req.ID {
			a.stream.update(m.Req)
		}
		return a, ListenForRequests(a.reqCh)

	// ── Open stream view (user pressed 's' in detail view) ───────────────────
	case openStreamMsg:
		a.stream = newStreamModel(m.Req)
		a.stream.width = a.width
		a.stream.height = a.height - 1
		a.mode = ViewStream

	// ── Open tools inspector (user pressed 't' in detail view) ───────────────
	case openToolsMsg:
		chain := buildChainForReq(m.Req, a.list.requests)
		a.tools = newToolsModel(m.Req, chain)
		a.tools.width = a.width
		a.tools.height = a.height - 1
		a.mode = ViewTools

	// ── Curl copy result ─────────────────────────────────────────────────────
	case curlCopiedMsg:
		if m.ok {
			a.statusMsg = "curl command copied to clipboard"
		} else {
			a.statusMsg = "curl command printed to stderr (clipboard unavailable)"
		}
		a.statusAt = time.Now()

	// ── Periodic tick (spinner) ───────────────────────────────────────────────
	case TickMsg:
		a.spinnerTick++
		a.list.spinnerTick = a.spinnerTick
		// Clear status message after 3 seconds.
		if a.statusMsg != "" && time.Since(a.statusAt) > 3*time.Second {
			a.statusMsg = ""
		}
		return a, tickCmd()

	// ── Keyboard input ────────────────────────────────────────────────────────
	case tea.KeyMsg:
		// Global: 'd' dismisses the top alert.
		if m.String() == "d" && len(a.alerts) > 0 {
			a.alerts = a.alerts[1:]
			return a, nil
		}
		switch a.mode {
		case ViewList:
			return a.updateList(m)
		case ViewDetail:
			return a.updateDetail(m)
		case ViewStream:
			return a.updateStream(m)
		case ViewTools:
			return a.updateTools(m)
		}
	}

	return a, nil
}

// updateList handles keyboard events in the list view.
func (a App) updateList(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return a, tea.Quit

	case "up", "k":
		a.list.moveUp()

	case "down", "j":
		a.list.moveDown()

	case "enter":
		if req := a.list.selected(); req != nil {
			a.detail = newDetailModel(req)
			a.detail.width = a.width
			a.detail.height = a.height - 1
			a.detail.rebuild()
			a.mode = ViewDetail
		}
	}
	return a, nil
}

// updateDetail handles keyboard events in the detail view.
func (a App) updateDetail(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "esc":
		a.mode = ViewList

	case "up", "k":
		a.detail.scrollUp()

	case "down", "j":
		a.detail.scrollDown()

	case "s":
		if a.detail.req != nil {
			return a, a.detail.openStreamCmd()
		}

	case "t":
		if a.detail.req != nil {
			req := a.detail.req
			return a, func() tea.Msg { return openToolsMsg{Req: req} }
		}

	case "c":
		if a.detail.req != nil {
			req := a.detail.req
			return a, func() tea.Msg {
				curl := export.ToCurl(req, "curl")
				return curlCopiedMsg{ok: copyToClipboard(curl)}
			}
		}
	}
	return a, nil
}

// copyToClipboard writes text to the system clipboard, returning true on success.
func copyToClipboard(text string) bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, then xsel.
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		}
	default:
		return false
	}
	cmd.Stdin = newStringReader(text)
	return cmd.Run() == nil
}

// newStringReader returns an io.Reader over a string without importing extra packages.
func newStringReader(s string) *stringReader { return &stringReader{s: s} }

type stringReader struct {
	s   string
	pos int
}

func (r *stringReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.s) {
		return 0, io.EOF
	}
	n := copy(p, r.s[r.pos:])
	r.pos += n
	return n, nil
}

// checkAlerts evaluates cost, latency, and error thresholds for a completed request.
func (a *App) checkAlerts(req *store.Request) {
	cfg := a.alertCfg

	// Latency alert.
	if cfg.LatencyThreshold > 0 && req.Latency > cfg.LatencyThreshold {
		a.alerts = append(a.alerts, fmt.Sprintf(
			"⚠ Slow request #%d: %s latency (threshold: %s)",
			req.Seq, req.Latency.Round(time.Millisecond), cfg.LatencyThreshold))
	}

	// Error alert.
	if cfg.AlertOnError && (req.StatusCode >= 400 || req.Status == store.StatusError) {
		a.alerts = append(a.alerts, fmt.Sprintf(
			"✗ Error on request #%d: HTTP %d — %s",
			req.Seq, req.StatusCode, req.ErrorMessage))
	}

	// Cost alert (session total, fires once).
	if cfg.CostThreshold > 0 && !a.costAlerted {
		a.sessionCost += req.TotalCost
		if a.sessionCost >= cfg.CostThreshold {
			a.costAlerted = true
			a.alerts = append(a.alerts, fmt.Sprintf(
				"$ Session cost $%.4f has exceeded threshold $%.4f",
				a.sessionCost, cfg.CostThreshold))
		}
	}
}

// updateTools handles keyboard events in the tools inspector view.
func (a App) updateTools(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "esc":
		a.mode = ViewDetail

	case "up", "k":
		a.tools.scrollUp()

	case "down", "j":
		a.tools.scrollDown()
	}
	return a, nil
}

// updateStream handles keyboard events in the stream view.
func (a App) updateStream(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "esc":
		a.mode = ViewDetail

	case "up", "k":
		a.stream.scrollUp()

	case "down", "j":
		a.stream.scrollDown()
	}
	return a, nil
}

// View implements tea.Model. It renders the active view plus the stats bar.
func (a App) View() string {
	statsBar := renderStatsBar(a.stats, a.proxyAddr, a.dashAddr, a.width)
	var hints string
	if a.statusMsg != "" {
		hints = successStyle.Render("  ✓ " + a.statusMsg)
	} else {
		hints = renderHints(a.mode)
	}

	var body string
	switch a.mode {
	case ViewDetail:
		body = a.detail.View()
	case ViewStream:
		body = a.stream.View()
	case ViewTools:
		body = a.tools.View()
	default:
		body = renderBanner() + listHeader() + "\n" + a.list.View()
	}

	// Render active alerts as stacked banners above the hints.
	var alertBanner string
	if len(a.alerts) > 0 {
		alertStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorYellow).
			Foreground(colorYellow).
			Padding(0, 1).
			Width(a.width - 4)
		alertBanner = "\n" + alertStyle.Render(
			fmt.Sprintf("%s  (d to dismiss, %d remaining)", a.alerts[0], len(a.alerts)))
	}

	return body + alertBanner + "\n\n" + hints + "\n" + statsBar
}

// ListenForRequests returns a tea.Cmd that blocks waiting for the next request
// on ch and wraps it as a RequestMsg. Re-issue this command after each message.
func ListenForRequests(ch <-chan *store.Request) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		req, ok := <-ch
		if !ok {
			// Channel closed; return nil so we stop listening.
			return nil
		}
		return RequestMsg{Req: req}
	}
}

// tickCmd returns a tea.Cmd that fires a TickMsg after tickInterval.
func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return TickMsg{T: t}
	})
}

// buildChainForReq builds a tool call chain for req using the conversation
// requests available in the list. Returns nil if there are no tool calls or
// if no conversation context exists.
func buildChainForReq(req *store.Request, all []*store.Request) []store.ToolChainStep {
	if req == nil || req.ConversationID == "" {
		return nil
	}
	var conv []*store.Request
	for _, r := range all {
		if r.ConversationID == req.ConversationID {
			conv = append(conv, r)
		}
	}
	if len(conv) == 0 {
		return nil
	}
	return analyze.BuildToolChain(conv)
}
