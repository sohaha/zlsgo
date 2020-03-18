// +build !windows

package zshell

import (
	"os"
	"syscall"
)

func RunNewProcess() (pid int, err error) {
	args := os.Args
	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	return syscall.ForkExec(args[0], args, execSpec)
}
