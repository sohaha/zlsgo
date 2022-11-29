package zerror

import (
	"errors"

	"github.com/sohaha/zlsgo/zutil"
)

type (
	// ErrCode error code type
	ErrCode int32
	// Error wraps err with code
	Error struct {
		err     error
		wrapErr error
		errText *string
		stack   zutil.Stack
		code    ErrCode
		inner   bool
	}

	External func(err error) error
)

var (
	goROOT = zutil.GOROOT()
)

func New(code ErrCode, text string, w ...External) error {
	var err error = &Error{
		code:    code,
		stack:   zutil.Callers(3),
		errText: &(&[1]string{text})[0],
	}

	for i := range w {
		err = w[i](err)
	}

	return err
}

// Reuse the error
func Reuse(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return &Error{
		err:   err,
		stack: zutil.Callers(3),
	}
}

// Wrap wraps err with code
func Wrap(err error, code ErrCode, text string, w ...External) error {
	if err == nil {
		return nil
	}

	for i := range w {
		err = w[i](err)
	}

	return &Error{
		wrapErr: err,
		code:    code,
		stack:   zutil.Callers(3),
		errText: &(&[1]string{text})[0],
	}
}

// Deprecated: please use zerror.With
// SupText returns the error text
func SupText(err error, text string) error {
	return With(err, text)
}

// With returns the inner error's text
func With(err error, text string, w ...External) error {
	if err == nil {
		return nil
	}

	for i := range w {
		err = w[i](err)
	}

	return &Error{
		wrapErr: err,
		inner:   true,
		stack:   zutil.Callers(3),
		errText: &(&[1]string{text})[0],
	}
}

// Unwrap returns if err is Error and its code == code
func Unwrap(err error, code ErrCode) (error, bool) {
	for {
		if err == nil {
			return nil, false
		}

		e, ok := err.(*Error)
		if !ok {
			return err, false
		}

		if e.code == code {
			if e.errText != nil {
				return errors.New(*e.errText), true
			}

			return e.err, true
		}

		err = e.Unwrap()
	}
}

// Is returns if err is Error and its code == code
func Is(err error, code ...ErrCode) bool {
	for i := range code {
		_, ok := Unwrap(err, code[i])
		if ok {
			return true
		}
	}

	return false
}

// UnwrapCode Returns the current error code
func UnwrapCode(err error) (ErrCode, bool) {
	if err == nil {
		return 0, false
	}

	e, ok := err.(*Error)
	if !ok {
		return 0, false
	}

	return e.code, true
}

// UnwrapCodes Returns the current all error code
func UnwrapCodes(err error) (codes []ErrCode) {
	for {
		if err == nil {
			return
		}

		e, ok := err.(*Error)
		if !ok {
			return
		}

		codes = append(codes, e.code)

		err = e.Unwrap()
	}
}

// UnwrapErrors Returns the current all error text
func UnwrapErrors(err error) (errs []string) {
	for {
		if err == nil {
			return
		}

		e, ok := err.(*Error)
		if !ok {
			errs = append(errs, err.Error())
			return
		}

		if e.errText != nil {
			errs = append(errs, *e.errText)
		} else {
			errs = append(errs, e.err.Error())
		}

		err = e.Unwrap()
	}
}
