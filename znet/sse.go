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

type sseEvent struct {
	ID      string
	Event   string
	Comment string
	Data    []byte
}

func (s *SSE) LastEventID() string {
	return s.lastID
}

func (s *SSE) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s *SSE) Stop() {
	s.ctxCancel()
}

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

type SSEOption struct {
	RetryTime      int
	HeartbeatsTime int
}

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
