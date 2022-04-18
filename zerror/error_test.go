package zerror_test

import (
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zerror"
)

const (
	ErrCode200 = zerror.ErrCode(200)
	ErrCode500 = zerror.ErrCode(500)
	ErrCode404 = zerror.ErrCode(404)
	ErrCode401 = zerror.ErrCode(401)
)

func TestError(t *testing.T) {
	tt := zlsgo.NewTest(t)

	err500 := zerror.New(ErrCode500, "The is 500")

	var (
		err404 error
		err401 error
	)
	tt.Run("Wrap", func(t *testing.T, tt *zlsgo.TestUtil) {
		err404 = zerror.Wrap(err500, ErrCode404, "The is 404")
		t.Log(err404)
		err401 = zerror.Wrap(err404, ErrCode401, "The is 401")
		t.Log(err401)
	})

	tt.Run("Is", func(t *testing.T, tt *zlsgo.TestUtil) {
		tt.EqualTrue(zerror.Is(err401, ErrCode401))
		tt.EqualTrue(zerror.Is(err401, ErrCode404))
		tt.EqualTrue(zerror.Is(err401, ErrCode500))
		tt.EqualTrue(!zerror.Is(err401, ErrCode200))
	})

	var (
		rawErr401     error
		rawErr404     error
		rawErr500     error
		rawErr404T500 error
		ok            bool
	)

	tt.Run("Unwrap", func(t *testing.T, tt *zlsgo.TestUtil) {
		rawErr401, ok = zerror.Unwrap(err401, ErrCode401)
		tt.Equal(err401.Error(), rawErr401.Error())

		rawErr404, ok = zerror.Unwrap(err401, ErrCode404)
		tt.EqualTrue(ok)
		tt.Equal(err404.Error(), rawErr404.Error())

		rawErr500, ok = zerror.Unwrap(err401, ErrCode500)
		tt.EqualTrue(ok)
		tt.Equal(err500.Error(), rawErr500.Error())

		rawErr404T500, ok := zerror.Unwrap(err404, ErrCode500)
		tt.EqualTrue(ok)
		tt.Equal(err500.Error(), rawErr404T500.Error())

		rawErr401T404, ok := zerror.Unwrap(err401, ErrCode404)
		tt.EqualTrue(ok)
		tt.Equal(err404.Error(), rawErr401T404.Error())

		rawErr404T401, ok := zerror.Unwrap(err404, ErrCode401)
		tt.EqualTrue(!ok)
		tt.EqualNil(rawErr404T401)
	})

	tt.Run("UnwrapCode", func(t *testing.T, tt *zlsgo.TestUtil) {
		code, ok := zerror.UnwrapCode(rawErr404T500)
		tt.Equal(zerror.ErrCode(0), code)
		tt.EqualTrue(!ok)

		code, ok = zerror.UnwrapCode(err500)
		tt.Equal(ErrCode500, code)
		tt.EqualTrue(ok)

		code, ok = zerror.UnwrapCode(err404)
		tt.Equal(ErrCode404, code)
		tt.EqualTrue(ok)

		code, ok = zerror.UnwrapCode(err401)
		tt.Equal(ErrCode401, code)
		tt.EqualTrue(ok)
	})

	tt.Run("UnwrapCodes", func(t *testing.T, tt *zlsgo.TestUtil) {
		tt.Equal([]zerror.ErrCode{ErrCode401, ErrCode404, ErrCode500}, zerror.UnwrapCodes(err401))
		tt.Equal([]zerror.ErrCode{ErrCode404, ErrCode500}, zerror.UnwrapCodes(err404))
		tt.Equal([]zerror.ErrCode{ErrCode500}, zerror.UnwrapCodes(err500))
	})

	tt.Run("UnwrapErrors", func(t *testing.T, tt *zlsgo.TestUtil) {
		errors := zerror.UnwrapErrors(err401)
		t.Log(strings.Join(errors, ", "))
		tt.Equal([]string{err401.Error(), err404.Error(), err500.Error()}, errors)
		tt.Equal([]string{err404.Error(), err500.Error()}, zerror.UnwrapErrors(err404))
		tt.Equal([]string{err500.Error()}, zerror.UnwrapErrors(err500))
	})
}
