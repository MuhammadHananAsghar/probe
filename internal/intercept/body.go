// Package intercept provides utilities for reading request/response bodies
// without consuming them — the original stream is preserved for forwarding.
package intercept

import (
	"bytes"
	"io"
)

// DrainAndRestore reads all bytes from r, returns them, and replaces r with
// a new reader containing the same bytes. Used to read request bodies without
// consuming them before forwarding.
func DrainAndRestore(r *io.ReadCloser) ([]byte, error) {
	if *r == nil {
		return nil, nil
	}

	data, err := io.ReadAll(*r)
	if err != nil {
		return nil, err
	}

	// Close the original reader.
	_ = (*r).Close()

	// Replace with a fresh reader wrapping the same bytes.
	*r = io.NopCloser(bytes.NewReader(data))

	return data, nil
}

// TeeReadCloser wraps an io.ReadCloser with an io.Writer so bytes are copied
// to an internal buffer as they are read. The original stream flows unimpeded.
type TeeReadCloser struct {
	io.Reader
	orig io.ReadCloser
	buf  *bytes.Buffer
}

// NewTeeReadCloser creates a TeeReadCloser that tees all bytes read from rc
// into an internal buffer accessible via Bytes.
func NewTeeReadCloser(rc io.ReadCloser) *TeeReadCloser {
	buf := &bytes.Buffer{}
	return &TeeReadCloser{
		Reader: io.TeeReader(rc, buf),
		orig:   rc,
		buf:    buf,
	}
}

// Close closes the underlying ReadCloser.
func (t *TeeReadCloser) Close() error {
	return t.orig.Close()
}

// Bytes returns all bytes that have been read through the TeeReadCloser so far.
func (t *TeeReadCloser) Bytes() []byte {
	return t.buf.Bytes()
}
