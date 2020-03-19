// +build !windows

package zshell

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func RunNewProcess() (pid int, err error) {
	args := os.Args
	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	file := args[0]
	if tmp, _ := ioutil.TempDir("", ""); tmp != "" {
		tmp = filepath.Dir(tmp)
		if strings.HasPrefix(file, tmp) {
			return 0, errors.New("temporary program does not support startup")
		}
	}
	return syscall.ForkExec(file, args, execSpec)
}
