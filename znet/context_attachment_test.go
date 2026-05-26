package znet

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/sohaha/zlsgo/zdi"
)

func TestFileAttachmentUsesProvidedPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "attachment.txt")
	if err := os.WriteFile(filePath, []byte("attachment body"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := New()
	r.GET("/download", func(c *Context) {
		c.FileAttachment(filePath, "download.txt")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/download", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != "attachment body" {
		t.Fatalf("expected file body, got %q", got)
	}
	if got := w.Header().Get("Content-Disposition"); got != `attachment; filename="download.txt"` {
		t.Fatalf("unexpected content disposition: %q", got)
	}
}

func TestContextClone(t *testing.T) {
	r := New()
	r.GET("/clone", func(c *Context) {
		c.SetHeader("X-Original", "test")
		c.WithValue("custom", "data")

		clone := c.Clone(nil, nil)
		if clone == nil {
			t.Fatal("clone should not be nil")
		}

		if clone.header["X-Original"][0] != "test" {
			t.Errorf("expected header X-Original to be copied")
		}

		val, _ := clone.Value("custom")
		if val != "data" {
			t.Errorf("expected custom data to be copied")
		}

		c.SetHeader("X-Original", "modified")
		if clone.header["X-Original"][0] == "modified" {
			t.Error("clone should be independent from original")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/clone?key=value", nil)
	r.ServeHTTP(w, req)
}

func TestContextCloneWithCustomRequest(t *testing.T) {
	r := New()
	r.GET("/original", func(c *Context) {
		c.SetHeader("X-Test", "original")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/original", nil)
	r.ServeHTTP(w, req)

	r.GET("/clone", func(c *Context) {
		newReq, _ := http.NewRequest("GET", "/clone", nil)
		newW := httptest.NewRecorder()
		clone := c.Clone(newW, newReq)

		if clone.Request != newReq {
			t.Error("clone should use provided request")
		}
		if clone.Writer != newW {
			t.Error("clone should use provided writer")
		}
	})

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/clone", nil)
	r.ServeHTTP(w2, req2)
}

func TestContextCloneWithInjector(t *testing.T) {
	r := New()
	r.GET("/clone", func(c *Context) {
		injector := zdi.New()
		injector.Maps(c)
		c.injector = injector

		clone := c.Clone(nil, nil)
		if clone.injector == nil {
			t.Error("clone should have injector")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/clone", nil)
	r.ServeHTTP(w, req)
}

func TestCloneValues(t *testing.T) {
	values := url.Values{
		"key1": []string{"value1", "value2"},
		"key2": []string{"value3"},
	}

	cloned := cloneValues(values)

	if cloned == nil {
		t.Fatal("cloned values should not be nil")
	}

	if len(cloned) != len(values) {
		t.Errorf("expected %d keys, got %d", len(values), len(cloned))
	}

	if cloned.Get("key1") != "value1" {
		t.Errorf("expected key1 to be value1, got %s", cloned.Get("key1"))
	}

	values.Set("key1", "modified")
	if cloned.Get("key1") == "modified" {
		t.Error("cloned values should be independent from original")
	}
}

func TestCopyResponse(t *testing.T) {
	r := New()

	r.GET("/target", func(c *Context) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/source", nil)
		sourceCtx := r.NewContext(w, req)
		sourceCtx.SetHeader("X-Source", "true")
		sourceCtx.String(200, "source response")

		c.CopyResponse(sourceCtx)

		if c.header["X-Source"][0] != "true" {
			t.Error("header should be copied")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/target", nil)
	r.ServeHTTP(w, req)
}

func TestCopyResponseWithExistingHeaders(t *testing.T) {
	r := New()

	r.GET("/target", func(c *Context) {
		c.SetHeader("X-Target", "target")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/source", nil)
		sourceCtx := r.NewContext(w, req)
		sourceCtx.SetHeader("X-Source", "true")
		sourceCtx.String(200, "source response")

		c.CopyResponse(sourceCtx)

		if c.header["X-Target"] != nil {
			t.Error("existing headers should be cleared")
		}
		if c.header["X-Source"][0] != "true" {
			t.Error("source header should be copied")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/target", nil)
	r.ServeHTTP(w, req)
}

func TestIsWebsocket(t *testing.T) {
	r := New()

	r.GET("/ws", func(c *Context) {
		if !c.IsWebsocket() {
			t.Error("should detect websocket request")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	r.ServeHTTP(w, req)

	r.GET("/http", func(c *Context) {
		if c.IsWebsocket() {
			t.Error("should not detect regular http as websocket")
		}
	})

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/http", nil)
	r.ServeHTTP(w2, req2)
}

func TestIsSSE(t *testing.T) {
	r := New()

	r.GET("/sse", func(c *Context) {
		if !c.IsSSE() {
			t.Error("should detect SSE request")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)
	req.Header.Set("Accept", "text/event-stream")
	r.ServeHTTP(w, req)

	r.GET("/http", func(c *Context) {
		if c.IsSSE() {
			t.Error("should not detect regular http as SSE")
		}
	})

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/http", nil)
	req2.Header.Set("Accept", "text/html")
	r.ServeHTTP(w2, req2)
}

func TestGetReferer(t *testing.T) {
	r := New()

	r.GET("/test", func(c *Context) {
		referer := c.GetReferer()
		if referer != "http://example.com" {
			t.Errorf("expected referer http://example.com, got %s", referer)
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Referer", "http://example.com")
	r.ServeHTTP(w, req)
}

func TestGetUserAgent(t *testing.T) {
	r := New()

	r.GET("/test", func(c *Context) {
		ua := c.GetUserAgent()
		if ua != "test-agent" {
			t.Errorf("expected user-agent test-agent, got %s", ua)
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	r.ServeHTTP(w, req)
}

func TestMustValue(t *testing.T) {
	r := New()

	r.GET("/test", func(c *Context) {
		c.WithValue("key", "value")
		result := c.MustValue("key")
		if result == nil {
			t.Error("expected value, got nil")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}
