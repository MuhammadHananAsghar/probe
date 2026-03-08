package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// wsMessage is the envelope sent to connected browser clients.
type wsMessage struct {
	Type string          `json:"type"`
	ID   string          `json:"id,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Hub manages all active WebSocket connections and broadcasts updates.
type Hub struct {
	mu      sync.Mutex
	clients map[*wsClient]struct{}
}

// wsClient wraps a single WebSocket connection.
type wsClient struct {
	conn *websocket.Conn
	send chan wsMessage
}

func newHub() *Hub {
	return &Hub{clients: make(map[*wsClient]struct{})}
}

func (h *Hub) add(c *wsClient) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) remove(c *wsClient) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// broadcast sends a message to all connected clients (non-blocking).
func (h *Hub) broadcast(msg wsMessage) {
	h.mu.Lock()
	clients := make([]*wsClient, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.Unlock()

	for _, c := range clients {
		select {
		case c.send <- msg:
		default:
			// Client send buffer full — drop the message for this client.
		}
	}
}

// marshalReq serialises a Request as an apiRequest DTO for the wire,
// including base64-encoded bodies.
func marshalReq(req *store.Request) json.RawMessage {
	b, _ := json.Marshal(toDTO(req))
	return b
}

// ServeWS handles a single WebSocket upgrade and drives the read/write loops.
// snapshot is the current list of all captured requests (sent on connect).
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, snapshot []*store.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow any Origin for local dashboard
	})
	if err != nil {
		return
	}

	client := &wsClient{conn: conn, send: make(chan wsMessage, 256)}
	h.add(client)
	defer func() {
		h.remove(client)
		conn.CloseNow()
	}()

	ctx := r.Context()

	// Send snapshot of all existing requests on connect.
	dtos := make([]apiRequest, len(snapshot))
	for i, r := range snapshot {
		dtos[i] = toDTO(r)
	}
	snapData, _ := json.Marshal(dtos)
	_ = wsjson.Write(ctx, conn, wsMessage{Type: "snapshot", Data: snapData})

	// Write loop: drain the send channel and forward to client.
	go func() {
		for msg := range client.send {
			if err := wsjson.Write(ctx, conn, msg); err != nil {
				return
			}
		}
	}()

	// Read loop: keep connection alive, detect disconnect.
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			return
		}
	}
}

// BroadcastNew sends a "request" event to all clients.
func (h *Hub) BroadcastNew(req *store.Request) {
	h.broadcast(wsMessage{Type: "request", Data: marshalReq(req)})
}

// BroadcastUpdate sends an "update" event to all clients.
func (h *Hub) BroadcastUpdate(req *store.Request) {
	h.broadcast(wsMessage{Type: "update", ID: req.ID, Data: marshalReq(req)})
}

// ListenAndBroadcast reads from ch and broadcasts new/updated requests to all
// WebSocket clients. It stops when ctx is cancelled or ch is closed.
// The first event for a given request ID is broadcast as "request" (new);
// subsequent events for the same ID are broadcast as "update".
func (h *Hub) ListenAndBroadcast(ctx context.Context, ch <-chan *store.Request) {
	seen := make(map[string]struct{})
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-ch:
			if !ok {
				return
			}
			if _, exists := seen[req.ID]; exists {
				h.BroadcastUpdate(req)
			} else {
				seen[req.ID] = struct{}{}
				h.BroadcastNew(req)
			}
		}
	}
}
