package znet

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/sohaha/zlsgo/zreflect"
)

type JSONRPCOption struct {
	DisabledHTTP bool
	Debug        bool
}

func JSONRPC(rcvr map[string]interface{}, opts ...func(o *JSONRPCOption)) func(c *Context) {
	o := JSONRPCOption{}
	if len(opts) > 0 {
		opts[0](&o)
	}

	s := rpc.NewServer()
	methods := make(map[string][]string, 0)
	for name, v := range rcvr {
		err := s.RegisterName(name, v)
		if err == nil && o.Debug {
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
