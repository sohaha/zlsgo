package znet

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/sohaha/zlsgo/zreflect"
)

// JSONRPCOption defines configuration options for the JSON-RPC handler.
type JSONRPCOption struct {
	DisabledHTTP bool // Disables HTTP method handling, only processes RPC calls
	Debug        bool // Enables debug mode with additional logging and method inspection
}

// JSONRPC creates a handler that processes JSON-RPC requests.
// It registers the provided receiver objects with the RPC server and returns a handler function
// that can be used with the router. The handler supports both HTTP and WebSocket transports.
func JSONRPC(rcvr map[string]interface{}, opts ...func(o *JSONRPCOption)) func(c *Context) {
	o := JSONRPCOption{}
	if len(opts) > 0 {
		opts[0](&o)
	}

	s := rpc.NewServer()
	methods := make(map[string][]string, 0)
	for name, v := range rcvr {
		err := s.RegisterName(name, v)
		if err != nil {
			// A failed registration silently removes the service, which would
			// only surface later as "rpc: can't find service" at call time.
			Log.Errorf("JSONRPC: register service [%s] failed: %v", name, err)
			continue
		}
		if o.Debug {
			typ := zreflect.TypeOf(v)
			for m := 0; m < typ.NumMethod(); m++ {
				method := typ.Method(m)
				mtype := method.Type
				mname := method.Name
				l := mtype.NumIn()
				replyType, argType := "-", "-"
				if l > 2 {
					replyType = mtype.In(2).String()
				}
				if l > 1 {
					argType = mtype.In(1).String()
				}
				methods[name+"."+mname] = []string{argType, replyType}
			}
		}
	}

	return func(c *Context) {
		req := c.Request
		method := req.Method
		if o.Debug && method == "GET" {
			c.JSON(200, methods)
			return
		}

		if c.stopHandle.Load() {
			return
		}

		if method == "CONNECT" || (method == "POST" && !o.DisabledHTTP) {
			c.stopHandle.Store(true)
			c.write()

			if method == "CONNECT" {
				hijacker, ok := c.Writer.(http.Hijacker)
				if !ok {
					c.Writer.WriteHeader(http.StatusInternalServerError)
					_, _ = io.WriteString(c.Writer, "response writer does not support hijacking\n")
					return
				}
				conn, _, err := hijacker.Hijack()
				if err != nil {
					c.Writer.WriteHeader(http.StatusInternalServerError)
					_, _ = io.WriteString(c.Writer, "hijack failed: "+err.Error()+"\n")
					return
				}
				defer conn.Close()

				_, _ = io.WriteString(conn, "HTTP/1.0 200 Connected to JSON RPC\n\n")
				s.ServeCodec(jsonrpc.NewServerCodec(conn))
				return
			}

			c.Writer.Header().Set("Content-Type", ContentTypeJSON)
			var conn io.ReadWriteCloser = struct {
				io.Writer
				io.ReadCloser
			}{
				ReadCloser: c.Request.Body,
				Writer:     c.Writer,
			}
			_ = s.ServeRequest(jsonrpc.NewServerCodec(conn))
			return
		}

		c.SetContentType(ContentTypePlain)
		c.String(http.StatusMethodNotAllowed, "405 must CONNECT\n")
	}
}
