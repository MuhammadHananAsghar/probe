package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MuhammadHananAsghar/probe/internal/analyze"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// ViewMode tracks which view is currently active.
type ViewMode int

const (
	// ViewList is the default scrollable request list view.
	ViewList ViewMode = iota
	// ViewDetail shows full details for a selected request.
	ViewDetail
)

// RequestMsg is a bubbletea message carrying a new or updated request.
type RequestMsg struct{ Req *store.Request }

// TickMsg is sent periodically for spinner animation updates.
type TickMsg struct{ T time.Time }

// tickInterval controls how frequently the spinner advances.
const tickInterval = 100 * time.Millisecond

// App is the main bubbletea model that composes the list and detail views.
type App struct {
	mode        ViewMode
	list        listModel
	detail      detailModel
	stats       store.SessionStats
	tracker     *analyze.Tracker
	width       int
	height      int
	proxyAddr   string
	dashAddr    string
	spinnerTick int
	reqCh       <-chan *store.Request
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
	return &App{
		mode:      ViewList,
		list:      newListModel(),
		proxyAddr: proxyAddr,
		dashAddr:  dashAddr,
		tracker:   tracker,
		reqCh:     reqCh,
		width:     120,
		height:    40,
	}
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

		statsBarHeight := 1
		headerHeight := 1
		listHeight := a.height - statsBarHeight - headerHeight
		if listHeight < 1 {
			listHeight = 1
		}
		a.list.width = a.width
		a.list.height = listHeight

		a.detail.width = a.width
		a.detail.height = a.height - statsBarHeight
		a.detail.rebuild()

	// ── New / updated request from proxy ─────────────────────────────────────
	case RequestMsg:
		a.list.upsert(m.Req)
		if a.tracker != nil {
			a.stats = a.tracker.Stats()
		}
		// Re-render detail if it's the currently shown request.
		if a.mode == ViewDetail && a.detail.req != nil && a.detail.req.ID == m.Req.ID {
			a.detail.req = m.Req
			a.detail.rebuild()
		}
		return a, ListenForRequests(a.reqCh)

	// ── Periodic tick (spinner) ───────────────────────────────────────────────
	case TickMsg:
		a.spinnerTick++
		a.list.spinnerTick = a.spinnerTick
		return a, tickCmd()

	// ── Keyboard input ────────────────────────────────────────────────────────
	case tea.KeyMsg:
		switch a.mode {
		case ViewList:
			return a.updateList(m)
		case ViewDetail:
			return a.updateDetail(m)
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
	}
	return a, nil
}

// View implements tea.Model. It renders the active view plus the stats bar.
func (a App) View() string {
	statsBar := renderStatsBar(a.stats, a.proxyAddr, a.dashAddr, a.width)

	var body string
	switch a.mode {
	case ViewDetail:
		body = a.detail.View()
	default:
		header := listHeader()
		body = header + "\n" + a.list.View()
	}

	return body + "\n" + statsBar
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
