// Package store defines the storage interface and models for probe.
package store

// Store is the interface for request storage backends.
type Store interface {
	// Add stores a new request and returns its assigned sequence number.
	Add(r *Request) int
	// Update replaces the request with matching ID.
	Update(r *Request)
	// Get returns the request with the given ID, or nil.
	Get(id string) *Request
	// GetBySeq returns the request with the given sequence number, or nil.
	GetBySeq(seq int) *Request
	// All returns all stored requests in insertion order.
	All() []*Request
	// Count returns the number of stored requests.
	Count() int
	// Stats returns current session aggregate statistics.
	Stats() SessionStats
	// Subscribe returns a channel that receives a copy of each request
	// after it is added or updated. The caller must drain the channel.
	Subscribe() <-chan *Request
	// Unsubscribe removes the subscription channel.
	Unsubscribe(ch <-chan *Request)
}
