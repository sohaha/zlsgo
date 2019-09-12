package zls

import (
	"os"
	"os/exec"
	"sync"
	"time"
	"unsafe"
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

func IfVal(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func ExecCommand(commandName string, arg ...string) (string, error) {
	var data string
	c := exec.Command(commandName, arg...)
	c.Env = os.Environ()
	out, err := c.CombinedOutput()
	if out != nil {
		data = *(*string)(unsafe.Pointer(&out))
	}
	if err != nil {
		return data, err
	}
	return data, nil
}
