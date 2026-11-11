package zhttp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zvalid"
)

type JSONRPC struct {
	client  *rpc.Client
	path    string
	address string
	options JSONRPCOptions
}

// ErrJSONRPCNotConnected is returned when a JSON-RPC call is attempted on a
// client whose connection could not be established (for example because
// NewJSONRPC failed to connect).
var ErrJSONRPCNotConnected = errors.New("zhttp: jsonrpc client is not connected")

func (j *JSONRPC) Call(serviceMethod string, args interface{}, reply interface{}) error {
	r := <-j.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1)).Done

	return r.Error
}

func (j *JSONRPC) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	if j == nil || j.client == nil {
		return j.notConnected(serviceMethod, args, reply, done)
	}

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

// notConnected reports a failed call the same way rpc.Client would, instead of
// dereferencing a nil client and panicking.
func (j *JSONRPC) notConnected(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	call := &rpc.Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Error:         ErrJSONRPCNotConnected,
	}
	if done == nil {
		done = make(chan *rpc.Call, 1)
	}
	call.Done = done
	select {
	case done <- call:
	default:
		go func() { done <- call }()
	}
	return call
}

func (j *JSONRPC) Close() error {
	if j == nil || j.client == nil {
		return nil
	}
	return j.client.Close()
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
		conn, err = tls.DialWithDialer(&d, "tcp", j.address, j.options.TlsConfig.Clone())
	}

	if err != nil {
		return err
	}

	message := zstring.Buffer(2)
	_, _ = message.WriteString("CONNECT " + j.path + " HTTP/1.0\n")

	for k := range j.options.Header {
		_, _ = message.WriteString(k)
		_, _ = message.WriteString(": ")
		_, _ = message.WriteString(j.options.Header.Get(k))
		_, _ = message.WriteString("\n")
	}

	_, _ = message.WriteString("\n\n")
	if _, err = io.WriteString(conn, message.String()); err != nil {
		_ = conn.Close()
		return err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		_ = conn.Close()
		return err
	}

	if response.StatusCode != http.StatusOK {
		_ = response.Body.Close()
		_ = conn.Close()
		return errors.New("Prohibit connection, a status code: " + strconv.Itoa(response.StatusCode))
	}
	j.client = rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	return nil
}

type JSONRPCOptions struct {
	TlsConfig  *tls.Config
	Header     http.Header
	Timeout    time.Duration
	RetryDelay time.Duration
	Retry      bool
}

func NewJSONRPC(address string, path string, opts ...func(o *JSONRPCOptions)) (client *JSONRPC, err error) {
	o := JSONRPCOptions{
		Retry:      true,
		RetryDelay: time.Second * 1,
		Header:     http.Header{},
	}
	if len(opts) > 0 {
		opts[0](&o)
	}

	if !strings.ContainsRune(address, ':') {
		address = ":" + address
	}

	if zvalid.Text(address).IsURL().Ok() {
		s := strings.Split(address, "://")
		if len(s) > 1 {
			address = s[1]
			if s[0] == "https" {
				o.TlsConfig = &tls.Config{}
			}
		}
	}

	client = &JSONRPC{options: o, address: address, path: path}
	err = client.connect()
	return
}
