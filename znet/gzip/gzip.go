/*
 * @Author: seekwe
 * @Date:   2019-06-06 19:23:27
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-06 19:35:46
 */

package gzip

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/sohaha/zlsgo/znet"
)

func New(level ...int) znet.HandlerFunc {
	gzipLevel := 7
	if len(level) > 0 {
		gzipLevel = level[0]
	}
	return func(c *znet.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
		} else {
			c.SetHeader("Content-Encoding", "gzip")
			w := c.Writer
			gw, _ := gzip.NewWriterLevel(w, gzipLevel)
			defer gw.Close()
			c.Writer = &gzipResponseWriter{Writer: gw, ResponseWriter: w}
			c.Next()
		}
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
