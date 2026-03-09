package timeout

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"net/http"
	"time"

	"github.com/sohaha/zlsgo/znet"
)

type bufferedResponseWriter struct {
	body        bytes.Buffer
	base        http.ResponseWriter
	header      http.Header
	code        int
	wroteHeader bool
}

func newBufferedResponseWriter(base http.ResponseWriter) *bufferedResponseWriter {
	return &bufferedResponseWriter{
		base:   base,
		header: make(http.Header),
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(p)
}

func (w *bufferedResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.code = code
}

func (w *bufferedResponseWriter) Flush() {
	if flusher, ok := w.base.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *bufferedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.base.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func New(waitingTime time.Duration, custom ...znet.HandlerFunc) znet.HandlerFunc {
	return func(c *znet.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), waitingTime)
		defer cancel()

		if c.IsSSE() || c.IsWebsocket() || c.Request.Method == http.MethodConnect {
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		writer := newBufferedResponseWriter(c.Writer)
		child := c.Clone(writer, c.Request.WithContext(ctx))
		done := make(chan struct{}, 1)
		panicErr := make(chan interface{}, 1)

		go func() {
			defer func() {
				if err := recover(); err != nil {
					select {
					case panicErr <- err:
					default:
					}
					return
				}
				select {
				case done <- struct{}{}:
				default:
				}
			}()
			child.Next()
		}()

		select {
		case <-done:
			applyBufferedResponse(c, child, writer)
			c.Abort()
			return
		case err := <-panicErr:
			c.Abort()
			panic(err)
		case <-ctx.Done():
			if len(custom) > 0 {
				custom[0](c)
				c.Abort()
			} else {
				c.Abort(http.StatusGatewayTimeout)
			}
			return
		}
	}
}

func applyBufferedResponse(target, child *znet.Context, writer *bufferedResponseWriter) {
	target.CopyResponse(child)

	for key, values := range writer.Header() {
		for i, value := range values {
			target.SetHeader(key, value, i == 0)
		}
	}

	data := target.PrevContent()
	if len(data.Content) == 0 && writer.body.Len() > 0 {
		data.Content = append(data.Content[:0], writer.body.Bytes()...)
	}
	if data.Code.Load() == 0 && writer.code != 0 {
		data.Code.Store(int32(writer.code))
	}
}
