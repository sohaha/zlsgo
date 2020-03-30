// +build windows

package zshell

import (
	"errors"
)

func RunNewProcess(filemd5 string) (pid int, err error) {
	return 0, errors.New("Windows does not support")
}
