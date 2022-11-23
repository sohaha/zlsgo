package zerror_test

import (
	"errors"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zerror"
)

func TestTag(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := errors.New("test")

	zerr := zerror.With(err, "包裹错误", zerror.WrapTag(zerror.NotFound))
	zerr = zerror.With(zerr, "最终错误提示", zerror.WrapTag(zerror.Unauthorized))

	tt.Equal(zerror.Unauthorized, zerror.GetTag(zerr))
}
