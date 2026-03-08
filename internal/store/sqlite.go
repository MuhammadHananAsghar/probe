package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// SQLite is a persistent Store backed by SQLite.
// It wraps a Memory store for the in-memory ring buffer (fast reads/writes)
// and persists every Add/Update to SQLite as a background write-through cache.
type SQLite struct {
	*Memory
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS requests (
	id TEXT PRIMARY KEY,
	seq INTEGER NOT NULL,
	data BLOB NOT NULL,
	created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_requests_seq ON requests (seq);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests (created_at);
`

// DefaultDBPath returns the default SQLite database path.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".probe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.db"), nil
}

// NewSQLite opens (or creates) the SQLite database at path, migrates the
// schema, loads recent requests into the memory ring buffer, and returns a
// ready-to-use SQLite store.
func NewSQLite(path string, ringSize int) (*SQLite, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("sqlite open %s: %w", path, err)
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite migrate: %w", err)
	}

	s := &SQLite{
		Memory: NewMemory(ringSize),
		db:     db,
	}

	// Seed memory ring buffer from recent persisted requests.
	if err := s.seed(ringSize); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite seed: %w", err)
	}

	return s, nil
}

// seed loads up to n most recent requests from SQLite into the memory buffer.
func (s *SQLite) seed(n int) error {
	rows, err := s.db.Query(
		`SELECT data FROM requests ORDER BY created_at DESC LIMIT ?`, n)
	if err != nil {
		return err
	}
	defer rows.Close()

	var reqs []*Request
	for rows.Next() {
		var blob []byte
		if err := rows.Scan(&blob); err != nil {
			continue
		}
		var r Request
		if err := json.Unmarshal(blob, &r); err != nil {
			continue
		}
		reqs = append(reqs, &r)
	}

	// Insert in chronological order (oldest first).
	for i := len(reqs) - 1; i >= 0; i-- {
		s.Memory.Add(reqs[i])
	}
	return rows.Err()
}

// Add stores the request in memory and persists it to SQLite.
func (s *SQLite) Add(r *Request) int {
	seq := s.Memory.Add(r)
	s.persist(r)
	return seq
}

// Update updates the request in memory and persists the updated state.
func (s *SQLite) Update(r *Request) {
	s.Memory.Update(r)
	s.persist(r)
}

// persist writes a request to SQLite using an INSERT OR REPLACE.
func (s *SQLite) persist(r *Request) {
	blob, err := json.Marshal(r)
	if err != nil {
		return
	}
	_, _ = s.db.Exec(
		`INSERT OR REPLACE INTO requests (id, seq, data, created_at) VALUES (?, ?, ?, ?)`,
		r.ID, r.Seq, blob, r.StartedAt.Unix(),
	)
}

// Cleanup deletes requests older than retentionDays from SQLite.
// Requests from the current session (seqs present in memory) are never deleted.
func (s *SQLite) Cleanup(retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Unix()
	result, err := s.db.Exec(
		`DELETE FROM requests WHERE created_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Close closes the underlying database connection.
func (s *SQLite) Close() error {
	return s.db.Close()
}

// History returns requests filtered by the given predicate, ordered by
// created_at descending, limited to maxRows rows.
func (s *SQLite) History(maxRows int, filter func(*Request) bool) ([]*Request, error) {
	rows, err := s.db.Query(
		`SELECT data FROM requests ORDER BY created_at DESC LIMIT ?`, maxRows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Request
	for rows.Next() {
		var blob []byte
		if err := rows.Scan(&blob); err != nil {
			continue
		}
		var r Request
		if err := json.Unmarshal(blob, &r); err != nil {
			continue
		}
		if filter == nil || filter(&r) {
			result = append(result, &r)
		}
	}
	return result, rows.Err()
}
