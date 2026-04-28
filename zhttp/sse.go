package zhttp

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	SSEEngine struct {
		ctx          context.Context
		eventCh      chan *SSEEvent
		errCh        chan error
		ctxCancel    context.CancelFunc
		verifyHeader func(http.Header) bool
		option       SSEOption
		readyState   int
	}

	SSEEvent struct {
		ID        string
		Event     string
		Undefined []byte
		Data      []byte
	}
)

var (
	delim   = []byte{':'} // []byte{':', ' '}
	ping    = []byte("ping")
	dataEnd = byte('\n')
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

func (sse *SSEEngine) Error() <-chan error {
	return sse.errCh
}

func (sse *SSEEngine) VerifyHeader(fn func(http.Header) bool) {
	sse.verifyHeader = fn
}

func (sse *SSEEngine) OnMessage(fn func(*SSEEvent)) (<-chan struct{}, error) {
	done := make(chan struct{}, 1)
	select {
	case <-sse.Done():
		done <- struct{}{}
		return done, nil
	case e := <-sse.Error():
		done <- struct{}{}
		return done, e
	case v := <-sse.Event():
		go func() {
			fn(v)
			for {
				select {
				case <-sse.Done():
					done <- struct{}{}
					return
				case <-sse.Error():
					done <- struct{}{}
					return
				case v := <-sse.Event():
					fn(v)
				}
			}
		}()

		return done, nil
	}
}

func SSE(url string, v ...interface{}) (*SSEEngine, error) {
	return std.SSE(url, nil, v...)
}

func (e *Engine) sseReq(method, url string, v ...interface{}) (*Res, error) {
	r, err := e.Do(method, url, v...)
	if err != nil {
		return nil, err
	}
	statusCode := r.resp.StatusCode
	if statusCode == http.StatusNoContent {
		return nil, nil
	}

	if statusCode != http.StatusOK {
		return nil, zerror.With(zerror.New(zerror.ErrCode(statusCode), r.String()), "status code is "+strconv.Itoa(statusCode))
	}
	return r, nil
}

type SSEOption struct {
	Method   string
	RetryNum int
}

func (e *Engine) SSE(url string, opt func(*SSEOption), v ...interface{}) (*SSEEngine, error) {
	var (
		retry     = 3000
		currEvent = &SSEEvent{}
	)
	o := SSEOption{
		Method:   "POST",
		RetryNum: -1,
	}
	if opt != nil {
		opt(&o)
	}
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	for i := range v {
		if c, ok := v[i].(context.Context); ok {
			ctx = c
		}
	}

	if ctx == nil {
		ctx, cancel = context.WithCancel(context.TODO())
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	sse := &SSEEngine{
		readyState: 0,
		ctx:        ctx,
		option:     o,
		ctxCancel:  cancel,
		eventCh:    make(chan *SSEEvent),
		errCh:      make(chan error, 1),
		verifyHeader: func(h http.Header) bool {
			return strings.Contains(h.Get("Content-Type"), "text/event-stream")
		},
	}

	lastID := ""
	data := append(v, Header{"Accept": "text/event-stream", "Connection": "keep-alive"}, sse.ctx)
	r, err := e.sseReq(sse.option.Method, url, data...)
	if err != nil {
		return sse, err
	}

	go func() {
		defer func() {
			if r != nil && r.resp != nil && r.resp.Body != nil {
				_ = r.resp.Body.Close()
			}
			close(sse.eventCh)
		}()

		for {
			if sse.ctx.Err() != nil {
				break
			}

			if err == nil {
				if r != nil {
					if sse.verifyHeader != nil && !sse.verifyHeader(r.Response().Header) {
						select {
						case sse.eventCh <- &SSEEvent{
							Undefined: r.Bytes(),
						}:
						case <-sse.ctx.Done():
							break
						}
						if r.resp != nil && r.resp.Body != nil {
							_ = r.resp.Body.Close()
						}
						r = nil
					}
				}

				if r == nil {
					sse.readyState = 2
					cancel()
					return
				}

				sse.readyState = 1

				isPing := false
				streamErr := r.Stream(func(line []byte, eof bool) error {
					select {
					case <-sse.ctx.Done():
						return sse.ctx.Err()
					default:
					}

					i := len(line)
					if i == 1 && line[0] == dataEnd {
						if !isPing {
							select {
							case sse.eventCh <- currEvent:
							case <-sse.ctx.Done():
								return sse.ctx.Err()
							}
							currEvent = &SSEEvent{}
							isPing = false
						} else {
							currEvent = &SSEEvent{}
						}

						return nil
					}

					if i < 2 {
						return nil
					}

					spl := bytes.SplitN(line, delim, 2)
					if len(spl) < 2 {
						currEvent.Undefined = bytes.TrimSpace(line)
						return nil
					}

					if len(spl[0]) == 0 {
						isPing = bytes.Equal(ping, bytes.TrimSpace(spl[1]))
						if !isPing {
							currEvent.Undefined = bytes.TrimSpace(spl[1])
						}
						return nil
					}

					val := bytes.TrimSuffix(spl[1], []byte{'\n'})
					val = bytes.TrimPrefix(val, []byte{' '})

					switch zstring.Bytes2String(spl[0]) {
					case "id":
						lastID = zstring.Bytes2String(val)
						currEvent.ID = lastID
					case "event":
						currEvent.Event = zstring.Bytes2String(val)
					case "data":
						if len(currEvent.Data) > 0 {
							select {
							case sse.eventCh <- currEvent:
							case <-sse.ctx.Done():
								return sse.ctx.Err()
							}
							currEvent = &SSEEvent{}
							isPing = false
						}
						currEvent.Data = append(currEvent.Data, val...)
					case "retry":
						if t, err := strconv.Atoi(zstring.Bytes2String(val)); err == nil {
							retry = t
						}
					}
					if eof && !isPing {
						select {
						case sse.eventCh <- currEvent:
						case <-sse.ctx.Done():
							return sse.ctx.Err()
						}
						currEvent = &SSEEvent{}
					}
					return nil
				})

				if streamErr != nil {
					err = streamErr
				}
			}

			if sse.option.RetryNum >= 0 {
				if sse.option.RetryNum == 0 {
					cancel()
					return
				}
				sse.option.RetryNum--
			}

			select {
			case <-sse.ctx.Done():
				return
			case <-time.After(time.Millisecond * time.Duration(retry)):
				// Continue to retry
			}

			sse.readyState = 0
			ndata := data
			if lastID != "" {
				ndata = append(ndata, Header{"Last-Event-ID": lastID})
			}

			if r != nil && r.resp != nil && r.resp.Body != nil {
				_ = r.resp.Body.Close()
			}

			r, err = e.sseReq(sse.option.Method, url, ndata...)
		}
	}()

	return sse, nil
}
