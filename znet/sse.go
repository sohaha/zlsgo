package znet

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

// SSE represents a Server-Sent Events connection.
// It handles the event stream between the server and client.
type SSE struct {
	ctx       context.Context
	events    chan *sseEvent
	net       *Context
	option    *SSEOption
	ctxCancel context.CancelFunc
	flush     func()
	lastID    string
	method    string
	Comment   []byte
}

// sseEvent represents a single Server-Sent Event with its components.
type sseEvent struct {
	ID      string // Event identifier
	Event   string // Event type
	Comment string // Event comment
	Data    []byte // Event data payload
}

// LastEventID returns the ID of the last event sent over this SSE connection.
func (s *SSE) LastEventID() string {
	return s.lastID
}

// Done returns a channel that's closed when the SSE connection is terminated.
// This can be used to detect when the client disconnects.
func (s *SSE) Done() <-chan struct{} {
	return s.ctx.Done()
}

// Stop terminates the SSE connection.
// This will close the event stream and release associated resources.
func (s *SSE) Stop() {
	s.ctxCancel()
}

// sendComment sends a ping comment to keep the connection alive.
func (s *SSE) sendComment() {
	s.events <- &sseEvent{
		Comment: "ping",
	}
}

func (s *SSE) Send(id string, data string, event ...string) error {
	return s.SendByte(id, zstring.String2Bytes(data), event...)
}

func (s *SSE) Push() {
	w := s.net.Writer
	r := s.net.Request

	s.net.Abort(http.StatusOK)
	s.net.write()
	s.flush()

	heartbeatsTime := s.option.HeartbeatsTime
	if heartbeatsTime == 0 {
		heartbeatsTime = 15000
	}
	ticker := time.NewTicker(time.Duration(heartbeatsTime) * time.Millisecond)

	defer ticker.Stop()

	b := zstring.Buffer(7)
sseFor:
	for {
		select {
		case <-ticker.C:
			go s.sendComment()
		case <-r.Context().Done():
			s.ctxCancel()
			break sseFor
		case <-s.ctx.Done():
			break sseFor
		case ev := <-s.events:
			if len(ev.Data) > 0 {
				if ev.ID != "" {
					b.WriteString("id: ")
					b.WriteString(ev.ID)
					b.WriteString("\n")
				}

				if bytes.HasPrefix(ev.Data, []byte(":")) {
					b.Write(ev.Data)
					b.WriteString("\n")
				} else {
					if bytes.IndexByte(ev.Data, '\n') > 0 {
						for _, v := range bytes.Split(ev.Data, []byte("\n")) {
							b.WriteString("data: ")
							b.Write(v)
							b.WriteString("\n")
						}
					} else {
						b.WriteString("data: ")
						b.Write(ev.Data)
						b.WriteString("\n")
					}
				}

				if len(ev.Event) > 0 {
					b.WriteString("event: ")
					b.WriteString(ev.Event)
					b.WriteString("\n")
				}

				if s.option.RetryTime > 0 {
					b.WriteString("retry: ")
					b.WriteString(strconv.Itoa(s.option.RetryTime))
					b.WriteString("\n")
				}
			}

			if len(ev.Comment) > 0 {
				b.WriteString(": ")
				b.WriteString(ev.Comment)
				b.WriteString("\n")
			}

			b.WriteString("\n")

			data := zstring.String2Bytes(b.String())
			_, _ = w.Write(data)
			s.flush()

			b.Reset()
			b.Grow(7)
		}
	}
}

// SendByte sends raw byte data as an SSE event.
// It allows specifying an event ID and optional event type.
func (s *SSE) SendByte(id string, data []byte, event ...string) error {
	if s.ctx.Err() != nil {
		return errors.New("client has been closed")
	}

	ev := &sseEvent{
		ID:   id,
		Data: data,
	}

	if len(event) > 0 {
		ev.Event = event[0]
	}

	s.events <- ev

	return nil
}

// SSEOption defines configuration options for an SSE connection.
type SSEOption struct {
	RetryTime      int // Client reconnection time in milliseconds
	HeartbeatsTime int // Heartbeat interval in seconds
}

// NewSSE creates a new Server-Sent Events connection from an HTTP context.
// It configures the connection based on the provided options and starts the event loop.
func NewSSE(c *Context, opts ...func(lastID string, opts *SSEOption)) *SSE {
	id := c.GetHeader("Last-Event-ID")
	ctx, cancel := context.WithCancel(context.TODO())
	s := &SSE{
		lastID:    id,
		events:    make(chan *sseEvent, 1),
		net:       c,
		ctx:       ctx,
		ctxCancel: cancel,
		option: &SSEOption{
			// RetryTime:      3000,
			HeartbeatsTime: 15000,
		},
	}

	for _, opt := range opts {
		opt(id, s.option)
	}

	flusher, _ := s.net.Writer.(http.Flusher)

	s.flush = func() {
		if c.Request.Context().Err() != nil {
			return
		}
		flusher.Flush()
	}

	s.net.SetHeader("Content-Type", "text/event-stream")
	s.net.SetHeader("Cache-Control", "no-cache")
	s.net.SetHeader("Connection", "keep-alive")
	c.prevData.Code.Store(http.StatusNoContent)
	s.net.Engine.shutdowns = append(s.net.Engine.shutdowns, func() {
		s.Stop()
	})
	return s
}

// Stream sends a streaming response to the client.
// The provided step function is called repeatedly until it returns false.
// Each call to step should write data to the provided writer.
func (c *Context) Stream(step func(w io.Writer) bool) bool {
	w := c.Writer
	flusher, _ := w.(http.Flusher)
	c.write()
	for {
		if c.stopHandle.Load() {
			return false
		}
		keepOpen := step(w)
		flusher.Flush()
		if !keepOpen {
			return false
		}
	}
}
