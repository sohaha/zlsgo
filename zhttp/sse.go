package zhttp

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	SSEEngine struct {
		readyState int
		eventCh    chan *SSEEvent
		ctx        context.Context
		ctxCancel  context.CancelFunc
	}

	SSEEvent struct {
		ID    string
		Event string
		Data  []byte
	}
)

func (sse *SSEEngine) Event() <-chan *SSEEvent {
	return sse.eventCh
}

func (sse *SSEEngine) Close() {
	sse.ctxCancel()
}

func (sse *SSEEngine) Done() <-chan struct{} {
	return sse.ctx.Done()
}

func SSE(url string, v ...interface{}) *SSEEngine {
	return std.SSE(url, v...)
}

func (e *Engine) sseReq(url string, v ...interface{}) (*Res, error) {
	r, err := e.Get(url, v...)
	if err != nil {
		return nil, err
	}
	statusCode := r.resp.StatusCode
	if statusCode == http.StatusNoContent {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, errors.New("status code is not 200")
	}
	return r, nil
}

func (e *Engine) SSE(url string, v ...interface{}) (sse *SSEEngine) {
	var (
		retry     = 3000
		delim     = []byte{':', ' '}
		currEvent = &SSEEvent{}
	)

	ctx, cancel := context.WithCancel(context.TODO())
	sse = &SSEEngine{
		readyState: 0,
		ctx:        ctx,
		ctxCancel:  cancel,
		eventCh:    make(chan *SSEEvent),
	}

	lastID := ""

	go func() {
		for {
			if sse.ctx.Err() != nil {
				break
			}

			if lastID != "" {
				v = append(v, Header{"Last-Event-ID": lastID})
			}
			v = append(v, Header{"Accept": "text/event-stream"})
			v = append(v, sse.ctx)
			r, err := e.sseReq(url, v...)

			if err == nil {
				if r == nil {
					sse.readyState = 2
					cancel()
					return
				}

				sse.readyState = 1

				err = r.Stream(func(line []byte) error {
					i := len(line)
					if i == 1 && currEvent.ID != "" {
						sse.eventCh <- currEvent
						currEvent = &SSEEvent{}
						return nil
					}

					if i < 2 {
						return nil
					}

					spl := bytes.SplitN(line, delim, 2)
					if len(spl) < 2 {
						return nil
					}

					val := bytes.TrimSuffix(spl[1], []byte{'\n'})

					switch zstring.Bytes2String(spl[0]) {
					case "id":
						lastID = zstring.Bytes2String(val)
						currEvent.ID = lastID
					case "event":
						currEvent.Event = zstring.Bytes2String(val)
					case "data":
						if len(currEvent.Data) > 0 {
							currEvent.Data = append(currEvent.Data, '\n')
						}
						currEvent.Data = append(currEvent.Data, val...)
					case "retry":
						if t, err := strconv.Atoi(zstring.Bytes2String(val)); err == nil {
							retry = t
						}
					}
					return nil
				})
			}

			sse.readyState = 0
			time.Sleep(time.Millisecond * time.Duration(retry))
		}
	}()

	return
}
