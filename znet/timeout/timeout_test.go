package timeout

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
)

func TestWebTimeout(t *testing.T) {
	test := zlsgo.NewTest(t)
	tests := []struct {
		handler      znet.HandlerFunc
		name         string
		expectedBody string
		middleware   []znet.Handler
		expectedCode int
		timeout      time.Duration
		skip         bool
	}{
		{
			name:         "Normal processing, no timeout",
			handler:      func(c *znet.Context) { c.String(200, "ok") },
			middleware:   []znet.Handler{New(2 * time.Second)},
			expectedCode: 200,
			expectedBody: "ok",
			timeout:      2 * time.Second,
		},
		{
			name:         "Timeout with default handler",
			handler:      func(c *znet.Context) { time.Sleep(2 * time.Second); c.String(200, "delayed") },
			middleware:   []znet.Handler{New(1 * time.Second)},
			expectedCode: http.StatusGatewayTimeout,
			expectedBody: "",
			timeout:      1 * time.Second,
		},
		{
			name:         "Timeout with custom handler",
			handler:      func(c *znet.Context) { time.Sleep(2 * time.Second); c.String(200, "delayed") },
			middleware:   []znet.Handler{New(1*time.Second, func(c *znet.Context) { c.String(504, "custom timeout") })},
			expectedCode: 504,
			expectedBody: "custom timeout",
			timeout:      1 * time.Second,
		},
		{
			name:    "Very short timeout should trigger timeout handler",
			handler: func(c *znet.Context) { time.Sleep(50 * time.Millisecond); c.String(200, "should not reach") },
			middleware: []znet.Handler{New(5*time.Millisecond, func(c *znet.Context) {
				c.String(504, "timeout occurred")
			})},
			expectedCode: 504,
			expectedBody: "timeout occurred",
			timeout:      5 * time.Millisecond,
		},
	}

	for i, tc := range tests {
		test.Run(tc.name, func(tt *zlsgo.TestUtil) {
			if tc.skip {
				tt.Log("Skipping test:", tc.name)
				return
			}

			r := newServer()
			path := fmt.Sprintf("/test_%d", i) // Unique path for each test case
			handlers := append([]znet.Handler{tc.handler}, tc.middleware...)
			w := newRequest(r, "GET", path, handlers...)

			tt.Equal(tc.expectedCode, w.Code)
			tt.Equal(tc.expectedBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	test := zlsgo.NewTest(t)
	r := newServer()

	routes := []string{"/concurrent/a", "/concurrent/b", "/concurrent/c", "/concurrent/d", "/concurrent/e"}
	for _, path := range routes {
		handlerPath := path
		r.GET(path, func(c *znet.Context) {
			time.Sleep(100 * time.Millisecond)
			c.String(200, handlerPath)
		}, New(200*time.Millisecond))
	}

	for i := 0; i < 5; i++ {
		path := fmt.Sprintf("/concurrent/%c", 'a'+i)
		w := newRequest(r, "GET", path)
		test.Equal(200, w.Code)
		test.Equal(path, strings.TrimSpace(w.Body.String()))
	}
}

func TestPanicRecovery(t *testing.T) {
	test := zlsgo.NewTest(t)
	r := newServer()

	timeoutHandlerCalled := make(chan bool, 1)

	panicHandler := func(c *znet.Context) {
		panic("test panic")
	}

	timeoutMiddleware := New(1*time.Second, func(c *znet.Context) {
		timeoutHandlerCalled <- true
		c.String(504, "timeout handler called")
	})

	recovery := func(next znet.HandlerFunc) znet.HandlerFunc {
		return func(c *znet.Context) {
			defer func() {
				if r := recover(); r != nil {
					c.String(500, "panic recovered")
				}
			}()
			next(c)
		}
	}

	r.GET("/panic",
		recovery(panicHandler),
		timeoutMiddleware,
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	test.Equal(500, w.Code)
	test.Equal("panic recovered", strings.TrimSpace(w.Body.String()))

	select {
	case <-timeoutHandlerCalled:
		t.Error("Timeout handler was called when it shouldn't have been")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestContextCancellation(t *testing.T) {
	r := newServer()

	handlerDone := make(chan struct{})
	timeoutHandlerCalled := make(chan struct{}, 1)

	r.GET("/cancel",
		func(c *znet.Context) {
			defer close(handlerDone)

			timer := time.NewTimer(2 * time.Second)
			defer timer.Stop()

			select {
			case <-timer.C:
				c.String(200, "should not reach")
			case <-c.Request.Context().Done():
				return
			}
		},
		New(100*time.Millisecond, func(c *znet.Context) {
			select {
			case timeoutHandlerCalled <- struct{}{}:
			default:
			}
			c.String(504, "timeout after cancellation")
		}),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cancel", nil)
	r.ServeHTTP(w, req)
	if w.Code != 504 {
		t.Fatalf("Expected status code 504, got %d", w.Code)
	}

	if body := strings.TrimSpace(w.Body.String()); body != "timeout after cancellation" {
		t.Fatalf("Unexpected response body: %s", body)
	}

	select {
	case <-timeoutHandlerCalled:
	default:
		t.Fatal("Timeout handler was not called")
	}

	select {
	case <-handlerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for handler to complete")
	}
}

func newServer() *znet.Engine {
	return znet.New()
}

func newRequest(r *znet.Engine, method string, path string, handler ...znet.Handler) *httptest.ResponseRecorder {
	method = strings.ToUpper(method)
	if len(handler) > 0 {
		firstHandler := handler[0]
		handlers := handler[1:]
		switch method {
		case "GET":
			r.GET(path, firstHandler, handlers...)
		case "POST":
			r.POST(path, firstHandler, handlers...)
		}
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}
