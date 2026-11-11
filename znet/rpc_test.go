package znet

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type RPCJSONResult struct {
	Num, Ans int
}

type rpcCal int

func (cal *rpcCal) Square(num int, result *RPCJSONResult) error {
	result.Num = num
	result.Ans = num * num
	return nil
}

func newRPCEngine() *Engine {
	e := New()
	e.Log.Discard()
	return e
}

func TestJSONRPCPost(t *testing.T) {
	e := newRPCEngine()
	e.POST("/__rpc", JSONRPC(map[string]interface{}{"Cal": new(rpcCal)}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/__rpc", strings.NewReader(`{"method":"Cal.Square","params":[12],"id":1}`))
	req.Header.Set("Content-Type", ContentTypeJSON)
	e.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d, body: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"Ans":144`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestJSONRPCDebugMethods(t *testing.T) {
	e := newRPCEngine()
	e.Any("/__rpc", JSONRPC(map[string]interface{}{"Cal": new(rpcCal)}, func(o *JSONRPCOption) {
		o.Debug = true
	}))

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/__rpc", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "Cal.Square") {
		t.Fatalf("method list missing Cal.Square: %s", body)
	}
}

// TestJSONRPCHijackUnsupported ensures a CONNECT request served by a
// non-hijackable ResponseWriter reports an error instead of panicking.
func TestJSONRPCHijackUnsupported(t *testing.T) {
	e := newRPCEngine()
	e.Any("/__rpc", JSONRPC(map[string]interface{}{"Cal": new(rpcCal)}))

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodConnect, "/__rpc", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code: %d, body: %s", w.Code, w.Body.String())
	}
}
