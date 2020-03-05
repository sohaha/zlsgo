// +build windows

package zshell

import (
	"errors"
)

func RunNewProcess() (pid int, err error) {
	return 0, errors.New("Windows does not support")
}
