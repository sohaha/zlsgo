package zhttp

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"
)

type JSONRPCOption struct {
	Timeout time.Duration
}

func JSONRPC(address string, path string, opts ...func(o *JSONRPCOption)) (*rpc.Client, error) {
	o := JSONRPCOption{
		Timeout: time.Second * 30,
	}
	if len(opts) > 0 {
		opts[0](&o)
	}

	if !strings.ContainsRune(address, ':') {
		address = ":" + address
	}
	conn, err := net.DialTimeout("tcp", address, o.Timeout)
	if err != nil {
		return nil, err
	}

	_, _ = io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")
	_, err = http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		return nil, err
	}

	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))
	return client, err
}
