package zerror

import (
	"github.com/sohaha/zlsgo/zutil"
)

// TryCatch exception capture
func TryCatch(fn func() error) (err error) {
	return zutil.TryCatch(fn)
}

// Panic  if error is not nil, usually used in conjunction with TryCatch
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}
