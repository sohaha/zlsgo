package zhttp

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"
)

type JSONRPC struct {
	client  *rpc.Client
	path    string
	address string
	options JSONRPCOptions
}

func (j *JSONRPC) Call(serviceMethod string, args interface{}, reply interface{}) error {
	r := <-j.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1)).Done

	return r.Error
}

func (j *JSONRPC) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	r := j.client.Go(serviceMethod, args, reply, done)
	if r.Error != nil && r.Error == rpc.ErrShutdown && j.options.Retry {
		if j.options.RetryDelay > 0 {
			time.Sleep(j.options.RetryDelay)
		}
		err := j.connect()
		if err == nil {
			done = make(chan *rpc.Call, 1)
			return j.Go(serviceMethod, args, reply, done)
		}
	}
	return r
}

func (j *JSONRPC) Close() error {
	err := j.client.Close()
	return err
}

func (j *JSONRPC) connect() error {
	var conn net.Conn
	var err error

	d := net.Dialer{}
	if j.options.Timeout > 0 {
		d.Timeout = j.options.Timeout
	}

	if j.options.TlsConfig == nil {
		conn, err = d.Dial("tcp", j.address)
	} else {
		config := j.options.TlsConfig
		if config.RootCAs == nil {
			config.InsecureSkipVerify = true
		}
		conn, err = tls.DialWithDialer(&d, "tcp", j.address, config)
	}

	if err != nil {
		return err
	}

	_, _ = io.WriteString(conn, "CONNECT "+j.path+" HTTP/1.0\n\n")
	_, err = http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		return err
	}

	j.client = rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	return nil
}

type JSONRPCOptions struct {
	TlsConfig  *tls.Config
	Timeout    time.Duration
	Retry      bool
	RetryDelay time.Duration
}

func NewJSONRPC(address string, path string, opts ...func(o *JSONRPCOptions)) (client *JSONRPC, err error) {
	o := JSONRPCOptions{
		Retry:      true,
		RetryDelay: time.Second * 1,
	}
	if len(opts) > 0 {
		opts[0](&o)
	}

	if !strings.ContainsRune(address, ':') {
		address = ":" + address
	}

	client = &JSONRPC{options: o, address: address, path: path}
	err = client.connect()
	return
}
