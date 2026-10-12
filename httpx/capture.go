package httpx

import (
	"net/http"
	"sync"
)

// capture buffers up to limit bytes of a body; the lock is needed because the transport may still be writing.
type capture struct {
	mu        sync.Mutex
	buf       []byte
	limit     int
	total     int64
	truncated bool
	// binary bodies are only counted, never buffered.
	binaryLabel string
	binary      bool
}

type captureSnapshot struct {
	buf         []byte
	total       int64
	truncated   bool
	binary      bool
	binaryLabel string
}

func (t *LoggingTransport) newCapture(h http.Header, contentLength int64) *capture {
	limit := t.maxCapture
	if t.sink == nil {
		// Without a sink only a preview is ever logged, so buffering more than the budget is wasted memory.
		limit = min(limit, 2*t.budget)
	}

	c := &capture{limit: limit}
	c.binaryLabel, c.binary = isBinaryContent(h)

	if !c.binary && contentLength > 0 {
		c.buf = make([]byte, 0, min(int64(limit), contentLength))
	}

	return c
}

func (c *capture) write(p []byte) {
	if len(p) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.total += int64(len(p))

	if c.binary {
		return
	}

	room := c.limit - len(c.buf)
	if room < len(p) {
		c.truncated = true
	}

	if room > 0 {
		c.buf = append(c.buf, p[:min(room, len(p))]...)
	}
}

func (c *capture) snapshot() captureSnapshot {
	if c == nil {
		return captureSnapshot{}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return captureSnapshot{
		buf: c.buf[:len(c.buf):len(c.buf)], total: c.total, truncated: c.truncated,
		binary: c.binary, binaryLabel: c.binaryLabel,
	}
}
