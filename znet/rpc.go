package znet

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type JSONRPCOption struct {
	DisabledHTTP bool
}

func JSONRPC(rcvr map[string]interface{}, opts ...func(o *JSONRPCOption)) func(c *Context) {
	o := JSONRPCOption{}
	if len(opts) > 0 {
		opts[0](&o)
	}

	s := rpc.NewServer()
	for name, v := range rcvr {
		_ = s.RegisterName(name, v)
	}

	return func(c *Context) {
		req := c.Request
		method := req.Method
		var codec rpc.ServerCodec
		if method == "CONNECT" || (method == "POST" && !o.DisabledHTTP) {
			c.stopHandle.Store(true)
			c.write()

			if method == "CONNECT" {
				conn, _, _ := c.Writer.(http.Hijacker).Hijack()
				codec = jsonrpc.NewServerCodec(conn)
				_, _ = io.WriteString(conn, "HTTP/1.0 200 Connected to JSON RPC\n\n")
				s.ServeCodec(codec)
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
