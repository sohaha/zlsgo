package zutil

import (
	"log"
	"sync"
	"time"
)

func WithLockContext(fn func()) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	fn()
}

func WithRunTimeContext(closer func(), callback func(time.Duration)) {
	start := time.Now()
	closer()
	timeduration := time.Since(start)
	callback(timeduration)
}

// IfVal Simulate ternary calculations, pay attention to handling no variables or indexing problems
func IfVal(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func Try(fn func(), catch func(e interface{}), finally ...func()) {
	if len(finally) > 0 {
		defer func() {
			finally[0]()
		}()
	}
	defer func() {
		if err := recover(); err != nil {
			if catch != nil {
				catch(err)
			} else {
				panic(err)
			}
		}
	}()
	fn()
}

func CheckErr(err error, exit ...bool) {
	if err != nil {
		if len(exit) > 0 && exit[0] {
			log.Fatalln(err)
			return
		}
		panic(err)
	}
}
