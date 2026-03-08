// Package intercept provides SSE stream interception with zero-copy forwarding.
package intercept

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/provider"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// StreamInterceptor reads a streaming SSE response, forwards each line
// immediately to the client, and records per-chunk timing and content.
type StreamInterceptor struct {
	src     io.ReadCloser
	dst     io.Writer
	flusher http.Flusher // may be nil
	parser  provider.StreamParser
}

// NewStreamInterceptor creates a StreamInterceptor.
// dst is the http.ResponseWriter (or any io.Writer) to forward to.
// parser may be nil (chunks are still timed but not content-parsed).
func NewStreamInterceptor(src io.ReadCloser, dst io.Writer, parser provider.StreamParser) *StreamInterceptor {
	si := &StreamInterceptor{src: src, dst: dst, parser: parser}
	if f, ok := dst.(http.Flusher); ok {
		si.flusher = f
	}
	return si
}

// Intercept streams SSE from src to dst while collecting timing and content
// metadata into req. It blocks until the stream ends, ctx is cancelled, or an
// error occurs. req.Chunks, req.TTFT, and req.StreamStats are populated.
func (si *StreamInterceptor) Intercept(ctx context.Context, req *store.Request, stallThreshold time.Duration) error {
	if stallThreshold == 0 {
		stallThreshold = 500 * time.Millisecond
	}

	scanner := bufio.NewScanner(si.src)
	// Allow up to 1 MB per line (for large tool result embeds).
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	streamStart := time.Now()
	var prevTime time.Time // zero until first chunk
	var accContent strings.Builder

	var currentEventType string
	var pendingLines []string // lines of the current SSE event block

	flushChunk := func(eventType, data string, arrivedAt time.Time) {
		if data == "[DONE]" {
			return
		}

		// Extract content delta via provider parser.
		var delta string
		if si.parser != nil {
			delta = si.parser.ParseEvent(eventType, data, req)
		}

		if delta != "" {
			accContent.WriteString(delta)
		}

		// Only record a chunk if there was actual content (or it's a tool call event).
		if delta == "" && eventType != "content_block_delta" && eventType != "" {
			// Non-content event (message_start, etc.) — skip recording as a chunk.
			return
		}
		// For OpenAI (no event type), skip empty deltas (heartbeat / finish chunks).
		if eventType == "" && delta == "" {
			return
		}

		now := arrivedAt
		var gap time.Duration
		if !prevTime.IsZero() {
			gap = now.Sub(prevTime)
		}
		isStall := !prevTime.IsZero() && gap > stallThreshold

		chunk := store.StreamChunk{
			Index:     len(req.Chunks),
			Content:   delta,
			ArrivedAt: now,
			Gap:       gap,
			IsStall:   isStall,
		}
		req.Chunks = append(req.Chunks, chunk)

		// Record TTFT on first non-empty content delta.
		if delta != "" && req.TTFT == 0 {
			req.TTFT = now.Sub(req.StartedAt)
		}

		prevTime = now
	}

	for {
		select {
		case <-ctx.Done():
			si.finalize(req, accContent.String(), streamStart, stallThreshold)
			if req.StreamStats != nil {
				req.StreamStats.Interrupted = true
			}
			return ctx.Err()
		default:
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		arrivedAt := time.Now()

		// Forward the line to the client immediately.
		si.writeLine(line)

		// Parse SSE line.
		switch {
		case strings.HasPrefix(line, "event:"):
			currentEventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			pendingLines = append(pendingLines, line)

		case strings.HasPrefix(line, "data:"):
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			pendingLines = append(pendingLines, line)
			// Dispatch the event immediately (data line closes the event).
			flushChunk(currentEventType, data, arrivedAt)

		case line == "":
			// Blank line: event boundary in SSE. Reset state.
			currentEventType = ""
			pendingLines = pendingLines[:0]

		default:
			// id: or comment line — forward but ignore.
			pendingLines = append(pendingLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		si.finalize(req, accContent.String(), streamStart, stallThreshold)
		if req.StreamStats != nil {
			req.StreamStats.Interrupted = true
		}
		return err
	}

	si.finalize(req, accContent.String(), streamStart, stallThreshold)
	return nil
}

// writeLine sends a line followed by a newline to the client and flushes.
func (si *StreamInterceptor) writeLine(line string) {
	_, _ = io.WriteString(si.dst, line+"\n")
	if si.flusher != nil {
		si.flusher.Flush()
	}
}

// finalize computes StreamStats once the stream has ended.
func (si *StreamInterceptor) finalize(req *store.Request, accContent string, streamStart time.Time, stallThreshold time.Duration) {
	if accContent != "" && req.ResponseContent == "" {
		req.ResponseContent = accContent
	}

	streamDuration := time.Since(streamStart)
	chunkCount := len(req.Chunks)

	var stallCount int
	for _, c := range req.Chunks {
		if c.IsStall {
			stallCount++
		}
	}

	var throughput float64
	if secs := streamDuration.Seconds(); secs > 0 && req.OutputTokens > 0 {
		throughput = float64(req.OutputTokens) / secs
	}

	var avgTokensPerChunk float64
	if chunkCount > 0 && req.OutputTokens > 0 {
		avgTokensPerChunk = float64(req.OutputTokens) / float64(chunkCount)
	}

	req.StreamStats = &store.StreamStats{
		ChunkCount:        chunkCount,
		TTFT:              req.TTFT,
		StreamDuration:    streamDuration,
		ThroughputTPS:     throughput,
		AvgTokensPerChunk: avgTokensPerChunk,
		StallCount:        stallCount,
		StallThreshold:    stallThreshold,
	}
}
