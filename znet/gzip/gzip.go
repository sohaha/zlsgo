package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/sohaha/zlsgo/znet"
)

func Default() znet.HandlerFunc {
	return New(Config{
		CompressionLevel: 7,
		PoolMaxSize:      1024,
		MinContentLength: 1024,
	})
}

// gzipRespBufSize is the initial capacity of the per-request compressed
// response buffer. Responses larger than this grow the buffer as before.
const gzipRespBufSize = 4096

func New(conf Config) znet.HandlerFunc {
	pool := &poolCap{
		c: make(chan *gzip.Writer, conf.PoolMaxSize),
		l: conf.CompressionLevel,
	}
	return func(c *znet.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
		} else {
			c.Next()
			p := c.PrevContent()
			if len(p.Content) < conf.MinContentLength {
				return
			}

			g, err := pool.Get()
			if err != nil {
				return
			}
			defer func() {
				g.Reset(io.Discard)
				pool.Put(g)
			}()

			// Pre-size the buffer to the typical response so bytes.Buffer does
			// not grow repeatedly while the gzip writer streams into it.
			be := bytes.NewBuffer(make([]byte, 0, gzipRespBufSize))
			g.Reset(be)
			_, err = g.Write(p.Content)
			if err != nil {
				return
			}
			err = g.Close()
			if err != nil {
				return
			}

			c.SetHeader("Content-Encoding", "gzip")
			c.Byte(p.Code.Load(), be.Bytes())
		}
	}
}
