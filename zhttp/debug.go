package zhttp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sohaha/zlsgo/zstring"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	rn = "\r\n\r\n"
)

var (
	Debug = false
)

type dumpConn struct {
	io.Writer
	io.Reader
}

func (c *dumpConn) Close() error                       { return nil }
func (c *dumpConn) LocalAddr() net.Addr                { return nil }
func (c *dumpConn) RemoteAddr() net.Addr               { return nil }
func (c *dumpConn) SetDeadline(t time.Time) error      { return nil }
func (c *dumpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *dumpConn) SetWriteDeadline(t time.Time) error { return nil }

type (
	dummyBody struct {
		N   int
		off int
	}

	delegateReader struct {
		c chan io.Reader
		r io.Reader
	}
)

func (r *delegateReader) Read(p []byte) (int, error) {
	if r.r == nil {
		r.r = <-r.c
	}
	return r.r.Read(p)
}

func (d *dummyBody) Read(p []byte) (n int, err error) {
	if d.N <= 0 {
		err = io.EOF
		return
	}
	left := d.N - d.off
	if left <= 0 {
		err = io.EOF
		return
	}

	if l := len(p); l > 0 {
		if l >= left {
			n = left
			err = io.EOF
		} else {
			n = l
		}
		d.off += n
		for i := 0; i < n; i++ {
			p[i] = '*'
		}
	}

	return
}

func (d *dummyBody) Close() error {
	return nil
}

type dumpBuffer struct {
	bytes.Buffer
}

func (b *dumpBuffer) Write(p []byte) {
	if b.Len() > 0 {
		b.Buffer.WriteString(rn)
	}
	b.Buffer.Write(p)
}

func (b *dumpBuffer) WriteString(s string) {
	b.Write([]byte(s))
}

func (r *Res) dumpRequest(dump *dumpBuffer) {
	head := r.r.flag&BitReqHead != 0
	body := r.r.flag&BitReqBody != 0

	if head {
		r.dumpReqHead(dump)
	}
	if body {
		if r.multipartHelper != nil {
			dump.Write(r.multipartHelper.Dump())
		} else if len(r.requesterBody) > 0 {
			dump.Write(r.requesterBody)
		}
	}
}

func (r *Res) dumpReqHead(dump *dumpBuffer) {
	reqSend := new(http.Request)
	*reqSend = *r.req
	if reqSend.URL.Scheme == "https" {
		reqSend.URL = new(url.URL)
		*reqSend.URL = *r.req.URL
		reqSend.URL.Scheme = "http"
	}

	if reqSend.ContentLength > 0 {
		reqSend.Body = &dummyBody{N: int(reqSend.ContentLength)}
	} else {
		reqSend.Body = &dummyBody{N: 1}
	}

	var buf bytes.Buffer
	pr, pw := io.Pipe()
	defer pw.Close()
	dr := &delegateReader{c: make(chan io.Reader)}

	t := &http.Transport{
		Dial: func(_, _ string) (net.Conn, error) {
			return &dumpConn{io.MultiWriter(&buf, pw), dr}, nil
		},
	}
	defer t.CloseIdleConnections()

	client := new(http.Client)
	*client = *r.client
	client.Transport = t

	go func() {
		req, err := http.ReadRequest(bufio.NewReader(pr))
		if err == nil {
			_, _ = io.Copy(ioutil.Discard, req.Body)
			_ = req.Body.Close()
		}

		dr.c <- strings.NewReader("HTTP/1.1 204 No Content\r\nConnection: close\r\n\r\n")
		_ = pr.Close()
	}()

	_, err := client.Do(reqSend)
	if err != nil {
		dump.WriteString(err.Error())
	} else {
		reqDump := buf.Bytes()
		if i := bytes.Index(reqDump, []byte(rn)); i >= 0 {
			reqDump = reqDump[:i]
		}
		dump.Write(reqDump)
	}
}

func (r *Res) dumpResonse(dump *dumpBuffer) {
	head := r.r.flag&BitRespHead != 0
	body := r.r.flag&BitRespBody != 0
	if head {
		responseBodyDump, err := httputil.DumpResponse(r.resp, false)
		if err != nil {
			dump.WriteString(err.Error())
		} else {
			if i := bytes.Index(responseBodyDump, []byte(rn)); i >= 0 {
				responseBodyDump = responseBodyDump[:i]
			}
			dump.Write(responseBodyDump)
		}
	}
	if body && len(r.Bytes()) > 0 {
		dump.Write(r.Bytes())
	}
}

func (r *Res) Cost() time.Duration {
	return r.cost
}

func (r *Res) Dump() string {
	dump := new(dumpBuffer)
	if r.r.flag&BitTime != 0 {
		dump.WriteString(fmt.Sprint(r.cost))
	}
	r.dumpRequest(dump)
	l := dump.Len()
	if l > 0 {
		dump.WriteString(zstring.Pad("", 30, "=", zstring.PadRight))
		l = dump.Len()
	}

	r.dumpResonse(dump)

	return dump.String()
}
