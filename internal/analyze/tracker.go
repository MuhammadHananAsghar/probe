// Package analyze provides session-level analytics and tracking for probe.
package analyze

import (
	"context"
	"sync"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// Tracker accumulates session statistics from captured requests. It is
// thread-safe and can optionally be driven by a Store subscription via Start.
type Tracker struct {
	mu    sync.RWMutex
	stats store.SessionStats
}

// NewTracker creates a new Tracker with zeroed statistics.
func NewTracker() *Tracker {
	return &Tracker{}
}

// Start subscribes to s and begins updating the Tracker's statistics in a
// background goroutine. The goroutine exits when ctx is cancelled or the
// subscription channel is closed. Only requests in terminal states (Done or
// Error) trigger a Record call.
func (t *Tracker) Start(ctx context.Context, s store.Store) {
	ch := s.Subscribe()
	go func() {
		defer s.Unsubscribe(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case req, ok := <-ch:
				if !ok {
					return
				}
				if req.Status == store.StatusDone || req.Status == store.StatusError {
					t.Record(req)
				}
			}
		}
	}()
}

// Record updates the session stats with data from the given request.
// It should be called once per request when that request reaches a terminal
// state (Done or Error). Calling it multiple times for the same request will
// over-count; use Recompute for a full rebuild from the store.
func (t *Tracker) Record(r *store.Request) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stats.RequestCount++
	t.stats.TotalCost += r.TotalCost
	t.stats.TotalLatency += r.Latency

	if r.Status == store.StatusError || r.ErrorMessage != "" {
		t.stats.ErrorCount++
	}

	if r.TTFT > 0 {
		t.stats.TotalTTFT += r.TTFT
		t.stats.TTFTCount++
	}
}

// Stats returns a snapshot of the current session statistics.
func (t *Tracker) Stats() store.SessionStats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats
}

// Reset clears all accumulated statistics. This is typically called at the
// start of a new probe session.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stats = store.SessionStats{}
}

// Recompute rebuilds the session stats from scratch using all requests
// currently held in the store. This is useful after a ring-buffer eviction
// makes incremental tracking inaccurate.
func (t *Tracker) Recompute(s store.Store) {
	reqs := s.All()
	var stats store.SessionStats
	for _, r := range reqs {
		stats.RequestCount++
		if r.Status == store.StatusError || r.ErrorMessage != "" {
			stats.ErrorCount++
		}
		stats.TotalCost += r.TotalCost
		if r.Latency > 0 {
			stats.TotalLatency += r.Latency
		}
		if r.TTFT > 0 {
			stats.TotalTTFT += r.TTFT
			stats.TTFTCount++
		}
	}

	t.mu.Lock()
	t.stats = stats
	t.mu.Unlock()
}
