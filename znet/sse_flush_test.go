package znet

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// plainWriter is a ResponseWriter that is deliberately neither an http.Flusher
// nor an Unwrapper. It models a writer produced by middleware that forgot to
// forward Flush, which used to make Stream and SSE.Push panic on a nil flusher.
type plainWriter struct {
	header http.Header
	body   bytes.Buffer
	code   int
}

func newPlainWriter() *plainWriter {
	return &plainWriter{header: http.Header{}, code: http.StatusOK}
}

func (w *plainWriter) Header() http.Header { return w.header }

func (w *plainWriter) Write(p []byte) (int, error) { return w.body.Write(p) }

func (w *plainWriter) WriteHeader(code int) { w.code = code }

// stalledWriter models a client that stops reading: Write blocks until the
// write deadline passes and then fails, which is what a real net.Conn does.
type stalledWriter struct {
	header http.Header
	mu     sync.Mutex
	// deadline is set through SetWriteDeadline, reached via ResponseController.
	deadline time.Time
}

func newStalledWriter() *stalledWriter {
	return &stalledWriter{header: http.Header{}}
}

func (w *stalledWriter) Header() http.Header { return w.header }

func (w *stalledWriter) WriteHeader(int) {}

func (w *stalledWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	deadline := w.deadline
	w.mu.Unlock()
	if deadline.IsZero() {
		// No deadline configured: block for good so the test fails loudly
		// instead of passing by accident.
		select {}
	}
	if wait := time.Until(deadline); wait > 0 {
		time.Sleep(wait)
	}
	return 0, os.ErrDeadlineExceeded
}

func (w *stalledWriter) SetWriteDeadline(t time.Time) error {
	w.mu.Lock()
	w.deadline = t
	w.mu.Unlock()
	return nil
}

func shutdownCount(e *Engine) int {
	e.shutdownMu.Lock()
	defer e.shutdownMu.Unlock()
	return len(e.shutdowns)
}

// 没有 Flusher 的 writer 上 Stream 必须正常降级，而不是 nil 接口 panic。
func TestStreamWithoutFlusherDoesNotPanic(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	w := newPlainWriter()
	c := r.NewContext(w, req)

	frames := 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Stream(func(out io.Writer) bool {
			frames++
			_, _ = io.WriteString(out, "data: ping\n\n")
			return frames < 3
		})
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stream blocked on a writer without Flush support")
	}
	if frames != 3 {
		t.Fatalf("step called %d times, want 3", frames)
	}
	if !strings.Contains(w.body.String(), "data: ping") {
		t.Fatalf("frames not written: %q", w.body.String())
	}
}

// 没有 Flusher 的 writer 上 SSE.Push 同样必须降级：心跳与事件都不能因为
// flush 不可用而 panic。
func TestSSEPushWithoutFlusherDoesNotPanic(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := newPlainWriter()
	c := r.NewContext(w, req)
	sse := NewSSE(c, func(_ string, opts *SSEOption) {
		opts.HeartbeatsTime = 10
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		sse.Push()
	}()
	if err := sse.Send("1", "hello", "message"); err != nil {
		t.Fatalf("send: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(w.body.String(), "data: hello") {
		if time.Now().After(deadline) {
			sse.Stop()
			t.Fatalf("event never written, body=%q", w.body.String())
		}
		time.Sleep(5 * time.Millisecond)
	}
	sse.Stop()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Push did not return after Stop")
	}
}

// 中间件用 SetHeader 挂的响应头只在引擎统一写响应时才落到 ResponseWriter。
// Stream 必须在写第一帧前刷出它们，否则跨域（CORS）等中间件头全部丢失。
func TestStreamFlushesMiddlewareHeadersBeforeFirstFrame(t *testing.T) {
	r := New("sse-test-" + t.Name())
	r.Use(func(c *Context) {
		c.SetHeader("Access-Control-Allow-Origin", "https://app.example")
		c.SetHeader("X-Middleware", "seen")
	})
	r.GET("/stream", func(c *Context) {
		c.SetHeader("Content-Type", "text/event-stream", true)
		c.SetStatus(http.StatusOK)
		c.Stream(func(out io.Writer) bool {
			_, _ = io.WriteString(out, "event: ready\ndata: {}\n\n")
			return false
		})
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	req.Header.Set("Origin", "https://app.example")
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want the middleware value", got)
	}
	if got := rec.Header().Get("X-Middleware"); got != "seen" {
		t.Fatalf("X-Middleware = %q, want seen", got)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", got)
	}
	if !strings.Contains(rec.Body.String(), "event: ready") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

// 必须直写 c.Writer 的 handler（反代、文件流）用 FlushHeaders 拿到同一份语义：
// 中间件头在第一次写入前落地，且引擎收尾不会重复写响应。
func TestFlushHeadersWritesMiddlewareHeaders(t *testing.T) {
	r := New("sse-test-" + t.Name())
	r.Use(func(c *Context) {
		c.SetHeader("Access-Control-Allow-Origin", "*")
	})
	r.GET("/raw", func(c *Context) {
		c.SetStatus(http.StatusOK)
		c.FlushHeaders()
		_, _ = io.WriteString(c.Writer, "raw-body")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/raw", nil)
	req.Header.Set("Origin", "https://app.example")
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
	if got := rec.Body.String(); got != "raw-body" {
		t.Fatalf("body = %q, want raw-body (engine must not append a second response)", got)
	}
}

// NewSSE 注册的 shutdown 钩子必须在连接结束时摘掉，否则每来一条连接就永久
// 留下一个闭包，进程活多久涨多久。
func TestSSEStopRemovesShutdownHook(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	c := r.NewContext(httptest.NewRecorder(), req)

	before := shutdownCount(r)
	sse := NewSSE(c)
	if got := shutdownCount(r); got != before+1 {
		t.Fatalf("shutdown hooks = %d, want %d", got, before+1)
	}

	sse.Stop()
	if got := shutdownCount(r); got != before {
		t.Fatalf("shutdown hooks after Stop = %d, want %d", got, before)
	}

	// 重复 Stop 必须幂等：不能 panic，也不能重复摘除别人的钩子。
	sse.Stop()
	if got := shutdownCount(r); got != before {
		t.Fatalf("shutdown hooks after double Stop = %d, want %d", got, before)
	}
}

// Push 自己结束时（写失败/客户端断开）同样要摘钩子，而不只是显式 Stop。
func TestSSEPushRemovesShutdownHookOnExit(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	c := r.NewContext(httptest.NewRecorder(), req)

	before := shutdownCount(r)
	sse := NewSSE(c)

	done := make(chan struct{})
	go func() {
		defer close(done)
		sse.Push()
	}()
	sse.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Push did not return after Stop")
	}
	if got := shutdownCount(r); got != before {
		t.Fatalf("shutdown hooks after Push exit = %d, want %d", got, before)
	}
}

// TrySend 在缓冲满时必须立刻返回 ErrSSEFull，而不是阻塞发布方。
func TestSSETrySendDoesNotBlockWhenFull(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	c := r.NewContext(httptest.NewRecorder(), req)
	sse := NewSSE(c)
	defer sse.Stop()

	// 缓冲容量 1：填满后必须报满，且调用本身不阻塞。
	if err := sse.TrySend("1", "first", "message"); err != nil {
		t.Fatalf("first TrySend: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- sse.TrySend("2", "second", "message") }()
	select {
	case err := <-done:
		if !errors.Is(err, ErrSSEFull) {
			t.Fatalf("TrySend on full buffer = %v, want ErrSSEFull", err)
		}
	case <-time.After(time.Second):
		t.Fatal("TrySend blocked on a full buffer")
	}

	// Stop 之后的 Send 必须报错退出，不能永远挂着。
	sse.Stop()
	if err := sse.Send("3", "late", "message"); err == nil {
		t.Fatal("Send after Stop must fail")
	}
}

// WriteTimeout 让卡住不读的客户端最终写失败退出，而不是永久占住推送 goroutine。
func TestSSEPushWriteTimeoutEndsStalledClient(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	c := r.NewContext(newStalledWriter(), req)
	sse := NewSSE(c, func(_ string, opts *SSEOption) {
		opts.HeartbeatsTime = 10
		opts.WriteTimeout = 50 * time.Millisecond
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		sse.Push()
	}()
	if err := sse.TrySend("1", "boom", "message"); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		sse.Stop()
		t.Fatal("Push kept running against a stalled client")
	}
}

// NewSSE 与中间件共存：Push 写第一帧前同样要把中间件（CORS 等）设置的响应头
// 刷到 ResponseWriter，否则跨源调用方会拿到没有 Access-Control-Allow-Origin
// 的 200 响应。
func TestNewSSEPushFlushesMiddlewareHeaders(t *testing.T) {
	r := New("sse-test-" + t.Name())
	r.Use(func(c *Context) {
		c.SetHeader("Access-Control-Allow-Origin", "https://app.example")
		c.SetHeader("X-Middleware", "seen")
	})
	r.GET("/sse", func(c *Context) {
		sse := NewSSE(c, func(_ string, opts *SSEOption) {
			opts.HeartbeatsTime = 60000
			opts.WriteTimeout = time.Second
		})
		done := make(chan struct{})
		go func() {
			defer close(done)
			sse.Push()
		}()
		_ = sse.Send("1", "{}", "ready")
		// 等首帧真的写出去，再结束连接。
		deadline := time.Now().Add(2 * time.Second)
		for !c.IsAbort() && time.Now().Before(deadline) {
			time.Sleep(5 * time.Millisecond)
			break
		}
		sse.Stop()
		<-done
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	req.Header.Set("Origin", "https://app.example")
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want the middleware value", got)
	}
	if got := rec.Header().Get("X-Middleware"); got != "seen" {
		t.Fatalf("X-Middleware = %q, want seen", got)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", got)
	}
}
