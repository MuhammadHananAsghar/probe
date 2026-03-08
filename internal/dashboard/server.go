// Package dashboard serves the probe web dashboard — a React SPA bundled
// via go:embed, with a JSON REST API and a WebSocket live-update feed.
package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/cost"
	"github.com/MuhammadHananAsghar/probe/internal/replay"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// Server is the dashboard HTTP server.
type Server struct {
	store     store.Store
	hub       *Hub
	eventCh   <-chan *store.Request
	addr      string
	pricingDB *cost.DB
}

// New creates a new dashboard Server.
func New(addr string, s store.Store, eventCh <-chan *store.Request, pricingDB *cost.DB) *Server {
	return &Server{
		store:     s,
		hub:       newHub(),
		eventCh:   eventCh,
		addr:      addr,
		pricingDB: pricingDB,
	}
}

// Start begins listening on srv.addr. It blocks until ctx is cancelled.
func (srv *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// REST API
	mux.HandleFunc("GET /api/requests", srv.handleGetRequests)
	mux.HandleFunc("GET /api/requests/{id}", srv.handleGetRequest)
	mux.HandleFunc("POST /api/replay/{id}", srv.handleReplay)
	mux.HandleFunc("POST /api/compare", srv.handleCompare)

	// WebSocket
	mux.HandleFunc("/api/ws", srv.handleWS)

	// SPA static files — serve embedded dist, fallback to index.html for client-side routing
	mux.Handle("/", spaHandler(DistFS()))

	httpSrv := &http.Server{
		Addr:         srv.addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // 0 = unlimited for streaming/WS
	}

	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return fmt.Errorf("dashboard: listen %s: %w", srv.addr, err)
	}

	// Start WebSocket broadcast goroutine.
	go srv.hub.ListenAndBroadcast(ctx, srv.eventCh)

	errCh := make(chan error, 1)
	go func() {
		if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

func (srv *Server) handleGetRequests(w http.ResponseWriter, _ *http.Request) {
	all := srv.store.All()
	dtos := make([]apiRequest, len(all))
	for i, r := range all {
		dtos[i] = toDTO(r)
	}
	writeJSON(w, dtos)
}

func (srv *Server) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	req := srv.store.Get(id)
	if req == nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, toDTO(req))
}

func (srv *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	srv.hub.ServeWS(w, r, srv.store.All())
}

// replayRequest is the JSON body for POST /api/replay/{id}.
type replayRequest struct {
	Model        string  `json:"model,omitempty"`
	Provider     string  `json:"provider,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	HasTemp      bool    `json:"has_temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
	HasMaxTokens bool    `json:"has_max_tokens,omitempty"`
	SystemPrompt string  `json:"system_prompt,omitempty"`
}

func (srv *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orig := srv.store.Get(id)
	if orig == nil {
		http.Error(w, "request not found", http.StatusNotFound)
		return
	}

	var req replayRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	opts := replay.Options{
		Model:        req.Model,
		Provider:     store.ProviderName(req.Provider),
		SystemPrompt: req.SystemPrompt,
	}
	if req.HasTemp {
		t := req.Temperature
		opts.Temperature = &t
	}
	if req.HasMaxTokens {
		n := req.MaxTokens
		opts.MaxTokens = &n
	}

	engine := replay.New(srv.store, srv.pricingDB)
	result, err := engine.Replay(r.Context(), orig, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the new replay request to WS clients.
	srv.hub.BroadcastNew(result.Req)

	writeJSON(w, map[string]any{
		"replay_seq":       result.Req.Seq,
		"replay_id":        result.Req.ID,
		"parameter_diffs":  result.ParameterDiffs,
		"original_seq":     result.OriginalSeq,
	})
}

// compareRequest is the JSON body for POST /api/compare.
type compareRequest struct {
	IDA string `json:"id_a"`
	IDB string `json:"id_b"`
}

func (srv *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	var req compareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	a := srv.store.Get(req.IDA)
	b := srv.store.Get(req.IDB)
	if a == nil || b == nil {
		http.Error(w, "one or both requests not found", http.StatusNotFound)
		return
	}
	cmp := replay.Compare(a, b)
	writeJSON(w, cmp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	_ = json.NewEncoder(w).Encode(v)
}
