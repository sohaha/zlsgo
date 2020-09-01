package gzip

import (
	"bytes"
	"compress/gzip"
	"strings"

	"github.com/sohaha/zlsgo/znet"
)

func Default() znet.HandlerFunc {
	return New(Config{
		CompressionLevel: 7,
		PoolMaxSize:      200,
		MinContentLength: 1024,
	})
}

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
			defer pool.Put(g)

			be := &bytes.Buffer{}
			g.Reset(be)
			_, err = g.Write(p.Content)
			if err != nil {
				return
			}
			_ = g.Flush()

			c.SetHeader("Content-Encoding", "gzip")
			c.Byte(p.Code, be.Bytes())
		}
	}
}
