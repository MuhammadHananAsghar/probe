package store

import (
	"fmt"
	"sync"
	"time"
)

// Memory is a thread-safe, in-memory ring buffer implementation of Store.
// When the ring buffer is full, the oldest entry is overwritten.
type Memory struct {
	mu   sync.RWMutex
	size int // maximum number of entries

	// Ring buffer storage.
	buf  []*Request
	head int // index of next write position
	full bool // whether the ring has wrapped

	// Secondary index for O(1) lookup by ID and seq.
	byID  map[string]*Request
	bySeq map[int]*Request

	// Global sequence counter; never resets.
	seq int

	// Session aggregate statistics.
	stats SessionStats

	// Pub/sub fan-out.
	subsMu sync.RWMutex
	subs   map[chan *Request]struct{}
}

// NewMemory creates a new Memory store with the given ring buffer capacity.
// If size <= 0, a default of 1000 is used.
func NewMemory(size int) *Memory {
	if size <= 0 {
		size = 1000
	}
	return &Memory{
		size:  size,
		buf:   make([]*Request, size),
		byID:  make(map[string]*Request),
		bySeq: make(map[int]*Request),
		subs:  make(map[chan *Request]struct{}),
	}
}

// generateID creates a simple, unique request ID without external dependencies.
func generateID(seq int) string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), seq)
}

// Add stores a new request, assigns it an ID and sequence number, and returns
// that sequence number. If the buffer is full, the oldest entry is evicted.
func (m *Memory) Add(r *Request) int {
	m.mu.Lock()

	m.seq++
	seq := m.seq

	r.ID = generateID(seq)
	r.Seq = seq

	// Evict the oldest entry if the ring is full.
	if m.full {
		old := m.buf[m.head]
		if old != nil {
			delete(m.byID, old.ID)
			delete(m.bySeq, old.Seq)
		}
	}

	m.buf[m.head] = r
	m.byID[r.ID] = r
	m.bySeq[r.Seq] = r

	m.head = (m.head + 1) % m.size
	if m.head == 0 {
		m.full = true
	}

	// Update stats.
	m.stats.RequestCount++

	// Snapshot for fan-out (copy pointer; caller owns the struct).
	snapshot := copyRequest(r)

	m.mu.Unlock()

	m.fanout(snapshot)
	return seq
}

// Update replaces a stored request (matched by ID) with the provided value.
// If no request with that ID exists, Update is a no-op.
func (m *Memory) Update(r *Request) {
	m.mu.Lock()

	existing, ok := m.byID[r.ID]
	if !ok {
		m.mu.Unlock()
		return
	}

	// Find and overwrite the slot in the ring buffer.
	for i, slot := range m.buf {
		if slot != nil && slot.ID == r.ID {
			m.buf[i] = r
			break
		}
	}

	// Update secondary indexes — seq stays the same, ID stays the same.
	m.byID[r.ID] = r
	m.bySeq[existing.Seq] = r

	// Recompute aggregate cost delta.
	m.stats.TotalCost += r.TotalCost - existing.TotalCost

	// Track latency and TTFT once the request is finalised.
	if (r.Status == StatusDone || r.Status == StatusError) &&
		existing.Status != StatusDone && existing.Status != StatusError {
		m.stats.TotalLatency += r.Latency
		if r.TTFT > 0 {
			m.stats.TotalTTFT += r.TTFT
			m.stats.TTFTCount++
		}
		if r.Status == StatusError {
			m.stats.ErrorCount++
		}
	}

	snapshot := copyRequest(r)
	m.mu.Unlock()

	m.fanout(snapshot)
}

// Get returns the request with the given ID, or nil if not found.
func (m *Memory) Get(id string) *Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.byID[id]
}

// GetBySeq returns the request with the given sequence number, or nil if not found.
func (m *Memory) GetBySeq(seq int) *Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bySeq[seq]
}

// All returns all stored requests in insertion order (oldest first).
func (m *Memory) All() []*Request {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []*Request
	if m.full {
		// The ring has wrapped; elements from head..end then 0..head-1.
		out = make([]*Request, 0, m.size)
		for i := m.head; i < m.size; i++ {
			if m.buf[i] != nil {
				out = append(out, m.buf[i])
			}
		}
		for i := 0; i < m.head; i++ {
			if m.buf[i] != nil {
				out = append(out, m.buf[i])
			}
		}
	} else {
		out = make([]*Request, 0, m.head)
		for i := 0; i < m.head; i++ {
			if m.buf[i] != nil {
				out = append(out, m.buf[i])
			}
		}
	}
	return out
}

// Count returns the number of requests currently stored.
func (m *Memory) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.byID)
}

// Stats returns a snapshot of current session aggregate statistics.
func (m *Memory) Stats() SessionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// Subscribe returns a buffered channel that receives a shallow copy of each
// request after it is added or updated. The caller must drain the channel
// promptly; slow consumers will have messages dropped (non-blocking send).
func (m *Memory) Subscribe() <-chan *Request {
	ch := make(chan *Request, 64)
	m.subsMu.Lock()
	m.subs[ch] = struct{}{}
	m.subsMu.Unlock()
	return ch
}

// Unsubscribe removes the given subscription channel and closes it.
func (m *Memory) Unsubscribe(ch <-chan *Request) {
	m.subsMu.Lock()
	defer m.subsMu.Unlock()
	// The map key must be the writable chan; recover the concrete type.
	for c := range m.subs {
		if (<-chan *Request)(c) == ch {
			delete(m.subs, c)
			close(c)
			return
		}
	}
}

// fanout sends a copy of r to all current subscribers, dropping the message
// for any subscriber whose channel buffer is full.
func (m *Memory) fanout(r *Request) {
	m.subsMu.RLock()
	defer m.subsMu.RUnlock()
	for ch := range m.subs {
		select {
		case ch <- r:
		default:
			// Subscriber is slow; drop rather than block.
		}
	}
}

// copyRequest makes a shallow copy of r so subscribers receive an immutable
// snapshot (field-level slices/maps are shared, which is acceptable for reads).
func copyRequest(r *Request) *Request {
	cp := *r
	return &cp
}
