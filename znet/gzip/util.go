package gzip

import (
	"compress/gzip"
	"io/ioutil"
)

type (
	poolCap struct {
		c chan *gzip.Writer
		l int
	}
	// Config gzip configuration
	Config struct {
		// CompressionLevel gzip compression level to use
		CompressionLevel int
		// PoolMaxSize maximum number of resource pools
		PoolMaxSize int
		// MinContentLength minimum content length to trigger gzip, the unit is in byte.
		MinContentLength int
	}
)

func (bp *poolCap) Get() (g *gzip.Writer, err error) {
	select {
	case g = <-bp.c:
	default:
		g, err = gzip.NewWriterLevel(ioutil.Discard, bp.l)
	}

	return
}

func (bp *poolCap) Put(g *gzip.Writer) {
	select {
	case bp.c <- g:
	default:
	}
}
