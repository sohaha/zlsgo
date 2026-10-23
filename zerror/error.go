// Package zerror provides error related operations
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

var goROOT = zutil.GOROOT()

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
		return &Error{
			stack:   zutil.Callers(3),
			errText: &(&[1]string{text})[0],
		}
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

// Unwrap returns if err chain has an Error with code == code
func Unwrap(err error, code ErrCode) (error, bool) {
	for err != nil {
		if e, ok := err.(*Error); ok && e.code == code {
			if e.errText != nil {
				return errors.New(*e.errText), true
			}

			return e.err, true
		}

		err = errors.Unwrap(err)
	}
	return nil, false
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

// UnwrapCode returns the first non-zero error code in the chain
func UnwrapCode(err error) (ErrCode, bool) {
	for err != nil {
		if e, ok := err.(*Error); ok && e.code != 0 {
			return e.code, true
		}
		err = errors.Unwrap(err)
	}
	return 0, false
}

// UnwrapCodes returns all error codes in the chain
func UnwrapCodes(err error) (codes []ErrCode) {
	for err != nil {
		if e, ok := err.(*Error); ok {
			codes = append(codes, e.code)
		}
		err = errors.Unwrap(err)
	}
	return
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

func UnwrapFirst(err error) (ferr error) {
	for {
		if err == nil {
			return
		}

		e, ok := err.(*Error)
		if !ok {
			return err
		}

		if e.errText != nil {
			ferr = errors.New(*e.errText)
		} else {
			ferr = e.err
		}

		err = e.Unwrap()
	}
}

func UnwrapFirstCode(err error) (code ErrCode) {
	for {
		if err == nil {
			return
		}

		e, ok := err.(*Error)
		if !ok {
			return
		}

		code = e.code

		err = e.Unwrap()
	}
}
