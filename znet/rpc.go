package znet

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
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
			typ := reflect.TypeOf(v)
			for m := 0; m < typ.NumMethod(); m++ {
				method := typ.Method(m)
				mtype := method.Type
				mname := method.Name
				argType := mtype.In(1).String()
				replyType := mtype.In(2).String()
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
