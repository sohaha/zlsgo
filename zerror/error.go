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
		code    ErrCode
		stack   zutil.Stack
	}
)

var (
	goROOT = zutil.GOROOT()
)

func New(code ErrCode, text string) error {
	return &Error{
		err:   errors.New(text),
		code:  code,
		stack: zutil.Callers(3),
	}
}

// Wrap wraps err with code
func Wrap(err error, code ErrCode, text string) error {
	if err == nil {
		return nil
	}

	return &Error{
		wrapErr: err,
		code:    code,
		stack:   zutil.Callers(3),
		err:     errors.New(text),
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
			return e.err, true
		}

		err = e.Unwrap()
	}
}

// Is returns if err is Error and its code == code
func Is(err error, code ErrCode) bool {
	_, ok := Unwrap(err, code)
	return ok
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
			return
		}

		errs = append(errs, e.err.Error())

		err = e.Unwrap()
	}
}
