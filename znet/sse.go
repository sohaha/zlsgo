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
	"github.com/sohaha/zlsgo/zutil"
)

// ErrSSEClosed is returned by Send/TrySend after the connection has stopped.
var ErrSSEClosed = errors.New("client has been closed")

// ErrSSEStopping is returned when the connection is shutting down and no new
// events are accepted.
var ErrSSEStopping = errors.New("SSE is stopping, no new events accepted")

// ErrSSEFull is returned by TrySend when the pending-event buffer is full,
// meaning the client is not reading fast enough.
var ErrSSEFull = errors.New("SSE buffer is full")

// SSE represents a Server-Sent Events connection.
// It handles the event stream between the server and client.
type SSE struct {
	ctx       context.Context
	events    chan *sseEvent
	net       *Context
	option    *SSEOption
	ctxCancel context.CancelFunc
	flush     func()
	stopping  *zutil.Bool
	lastID    string
	method    string
	Comment   []byte
	// control writes deadlines through Unwrap chains; nil-safe to call on
	// writers that do not support them.
	control *http.ResponseController
	// unregister removes the shutdown hook registered by NewSSE.
	unregister func()
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
	if !s.stopping.CAS(false, true) {
		return
	}
	s.ctxCancel()
	// Without this the engine keeps one closure per connection for the whole
	// process lifetime: NewSSE registers a shutdown hook and never removes it.
	if s.unregister != nil {
		s.unregister()
		s.unregister = nil
	}
}

// sendComment sends a ping comment to keep the connection alive.
func (s *SSE) sendComment() {
	if s.ctx.Err() != nil || s.stopping.Load() {
		return
	}
	select {
	case s.events <- &sseEvent{Comment: "ping"}:
	default:
	}
}

func (s *SSE) Send(id string, data string, event ...string) error {
	return s.SendByte(id, zstring.String2Bytes(data), event...)
}

// TrySend is the non-blocking variant of Send: it drops the event and reports
// ErrSSEFull instead of waiting when the client cannot keep up.
func (s *SSE) TrySend(id string, data string, event ...string) error {
	return s.TrySendByte(id, zstring.String2Bytes(data), event...)
}

// writeFrame writes one prepared frame and flushes it. It returns false when
// the connection is no longer writable, which ends the push loop.
func (s *SSE) writeFrame(w io.Writer, data []byte) bool {
	if s.option.WriteTimeout > 0 && s.control != nil {
		_ = s.control.SetWriteDeadline(time.Now().Add(s.option.WriteTimeout))
	}
	if _, err := w.Write(data); err != nil {
		return false
	}
	s.flush()
	return true
}

func (s *SSE) Push() {
	w := s.net.Writer
	r := s.net.Request

	// Every exit path ends the connection, so deregister the shutdown hook here
	// instead of relying on each individual branch.
	defer s.Stop()

	s.net.Abort(http.StatusOK)
	s.net.write()
	s.flush()

	// 清掉收尾时设置的写超时，避免把 deadline 留给下一个使用者。
	if s.option.WriteTimeout > 0 && s.control != nil {
		defer func() { _ = s.control.SetWriteDeadline(time.Time{}) }()
	}

	heartbeatsTime := s.option.HeartbeatsTime
	if heartbeatsTime == 0 {
		heartbeatsTime = 15000
	}
	ticker := time.NewTicker(time.Duration(heartbeatsTime) * time.Millisecond)

	defer ticker.Stop()

	// Use memory manager's pooled buffer for better performance
	b := zutil.GetBuff(512)
	defer zutil.PutBuff(b)

sseFor:
	for {
		select {
		case <-ticker.C:
			s.sendComment()
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

			data := b.Bytes()
			ok := s.writeFrame(w, data)

			b.Reset()
			if !ok {
				break sseFor
			}
		}
	}
}

// SendByte sends raw byte data as an SSE event.
// It allows specifying an event ID and optional event type.
func (s *SSE) SendByte(id string, data []byte, event ...string) error {
	if s.ctx.Err() != nil {
		return ErrSSEClosed
	}

	if s.stopping.Load() {
		return ErrSSEStopping
	}

	ev := newSSEEvent(id, data, event)
	select {
	case s.events <- ev:
		return nil
	case <-s.ctx.Done():
		return ErrSSEClosed
	}
}

// TrySendByte is the non-blocking variant of SendByte. It returns ErrSSEFull
// when the buffer is full so the caller can drop the slow client instead of
// blocking its own publish path.
func (s *SSE) TrySendByte(id string, data []byte, event ...string) error {
	if s.ctx.Err() != nil {
		return ErrSSEClosed
	}

	if s.stopping.Load() {
		return ErrSSEStopping
	}

	select {
	case s.events <- newSSEEvent(id, data, event):
		return nil
	default:
		return ErrSSEFull
	}
}

func newSSEEvent(id string, data []byte, event []string) *sseEvent {
	ev := &sseEvent{
		ID:   id,
		Data: data,
	}
	if len(event) > 0 {
		ev.Event = event[0]
	}
	return ev
}

// SSEOption defines configuration options for an SSE connection.
type SSEOption struct {
	// RetryTime is the client reconnection delay, in milliseconds.
	RetryTime int
	// HeartbeatsTime is the heartbeat interval in milliseconds (the value is
	// multiplied by time.Millisecond when used).
	HeartbeatsTime int
	// WriteTimeout bounds a single frame write. Zero means no deadline, in
	// which case a client that stops reading can pin the push goroutine.
	WriteTimeout time.Duration
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
		stopping:  zutil.NewBool(false),
		option: &SSEOption{
			// RetryTime:      3000,
			HeartbeatsTime: 15000,
		},
	}

	for _, opt := range opts {
		opt(id, s.option)
	}

	// ResponseController goes through Unwrap chains, so a writer wrapped by
	// middleware still flushes. It also reports ErrNotSupported instead of
	// panicking when the writer cannot flush at all.
	control := http.NewResponseController(s.net.Writer)
	s.control = control

	s.flush = func() {
		if c.Request.Context().Err() != nil {
			return
		}
		_ = control.Flush()
	}

	s.net.SetHeader("Content-Type", "text/event-stream", true)
	s.net.SetHeader("Cache-Control", "no-cache", true)
	s.net.SetHeader("Connection", "keep-alive", true)
	c.prevData.Code.Store(http.StatusNoContent)
	// Keep the exact value so Stop can deregister it; otherwise every
	// connection would leave one closure behind for the process lifetime.
	shutdown := func() { s.Stop() }
	s.unregister = func() { s.net.Engine.RemoveShutdown(shutdown) }
	s.net.Engine.AddShutdown(shutdown)
	return s
}

// Stream sends a streaming response to the client.
// The provided step function is called repeatedly until it returns false.
// Each call to step should write data to the provided writer.
func (c *Context) Stream(step func(w io.Writer) bool) bool {
	w := c.Writer
	control := http.NewResponseController(w)
	// write() flushes the headers collected by middleware (SetHeader) before
	// the first frame is written, so CORS and other middleware headers are not
	// lost when the handler writes to the raw ResponseWriter.
	c.write()
	for {
		if c.stopHandle.Load() {
			return false
		}
		keepOpen := step(w)
		_ = control.Flush()
		if !keepOpen {
			return false
		}
	}
}

// FlushHeaders writes the response headers collected by middleware to the
// underlying ResponseWriter and marks the response as committed, without
// ending it.
//
// Handlers that must write to c.Writer directly (SSE, reverse proxies, file
// streaming) call it before their first write. Without it, headers set with
// SetHeader never reach the client, because only the engine's final write()
// copies them over. Context.Stream already does this internally.
func (c *Context) FlushHeaders() {
	c.write()
}
