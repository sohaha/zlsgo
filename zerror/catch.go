package zerror

import (
	"fmt"
)

// TryCatch exception capture
func TryCatch(fn func() error) (err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			switch e := recoverErr.(type) {
			case error:
				err = Reuse(e)
			case *Error:
				err = e
			default:
				err = Reuse(fmt.Errorf("%v", recoverErr))
			}
		}
	}()
	err = fn()
	return
}

// Panic if error is not nil, usually used in conjunction with TryCatch
func Panic(err error) {
	if err != nil {
		panic(Reuse(err))
	}
}
