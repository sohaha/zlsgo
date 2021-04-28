package zcli

import (
	"errors"
	"github.com/sohaha/zlsgo/zutil"
	"os"
	"os/exec"
)

// Daemon Turn the current process into a daemon
func Daemon() (quit bool, err error) {
	if zutil.IsWin() {
		return false, errors.New("the current operating system does not support" +
			" daemon" +
			" execution")
	}
	if zutil.GetGid() != 1 {
		return false, errors.New("can only be used in the main goroutine")
	}
	if os.Getppid() != 1 {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		if err := cmd.Start(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
