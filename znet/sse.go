package znet

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

type SSE struct {
	lastID    string
	Comment   []byte
	events    chan *sseEvent
	reply     chan struct{}
	net       *Context
	option    *SSEOption
	ctx       context.Context
	ctxCancel context.CancelFunc
	close     *zutil.Bool
}

type sseEvent struct {
	ID      string
	Data    []byte
	Event   string
	Comment string
}

func (s *SSE) LastEventID() string {
	return s.lastID
}

func (s *SSE) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s *SSE) sendComment() {
	s.events <- &sseEvent{
		Comment: "ping",
	}
}

func (s *SSE) Send(id string, data string, event ...string) error {
	return s.SendByte(id, zstring.String2Bytes(data), event...)
}

func (s *SSE) cancel() {
	zlog.Error(8888)
	s.close.Store(true)
	s.ctxCancel()
	flusher, _ := s.net.Writer.(http.Flusher)

	s.net.SetHeader("Content-Type", "text/event-stream")
	s.net.SetHeader("Cache-Control", "no-cache")
	s.net.SetHeader("Connection", "keep-alive")
	s.net.Abort(http.StatusNoContent)
	s.net.write()

	flusher.Flush()
}

func (s *SSE) start() {
	w := s.net.Writer
	r := s.net.Request
	flusher, _ := w.(http.Flusher)

	s.net.SetHeader("Content-Type", "text/event-stream")
	s.net.SetHeader("Cache-Control", "no-cache")
	s.net.SetHeader("Connection", "keep-alive")
	s.net.Abort(http.StatusOK)
	s.net.write()

	flusher.Flush()

	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	b := zstring.Buffer(7)

sseFor:
	for {
		select {
		case <-ticker.C:
			go s.sendComment()
		case <-r.Context().Done():
			s.close.Store(true)
			s.ctxCancel()
			break sseFor
		case ev := <-s.events:
			if len(ev.Data) > 0 {
				b.WriteString("id: ")
				b.WriteString(ev.ID)
				b.WriteString("\n")

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
			if _, err := w.Write(data); err == nil && r.Context().Err() == nil {
				flusher.Flush()
			}

			b.Reset()
			b.Grow(7)
			s.reply <- struct{}{}
		}
	}

}

func (s *SSE) SendByte(id string, data []byte, event ...string) error {
	if s.close.Load() {
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

	<-s.reply
	return nil
}

type SSEOption struct {
	RetryTime      int
	HeartbeatsTime time.Duration
	Verify         bool
}

func NewSSE(c *Context, opts ...func(lastID string, opts *SSEOption)) *SSE {
	id := c.GetHeader("Last-Event-ID")
	ctx, cancel := context.WithCancel(context.TODO())
	s := &SSE{
		lastID:    id,
		events:    make(chan *sseEvent),
		reply:     make(chan struct{}),
		net:       c,
		ctx:       ctx,
		ctxCancel: cancel,
		close:     zutil.NewBool(false),
		option: &SSEOption{
			Verify:         true,
			RetryTime:      3000,
			HeartbeatsTime: time.Second * 15,
		},
	}

	for _, opt := range opts {
		opt(id, s.option)
	}

	if s.option.Verify {
		go s.start()
	} else {
		s.cancel()
	}

	return s
}
