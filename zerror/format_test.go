package zerror_test

import (
	"testing"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zlog"
)

func TestFormat(t *testing.T) {
	err := newErr()
	err = wrap500Err(err)
	err = wrap999Err(err)
	zlog.Stack(err)
}

func newErr() error {
	e := func() error {
		return zerror.New(400, "The is 400")
	}
	return e()
}

func wrap500Err(err error) error {
	return zerror.Wrap(err, 500, "Wrap 500 ErrorNil")
}

func wrap999Err(err error) error {
	return zerror.Wrap(err, 999, "Unknown ErrorNil")
}
