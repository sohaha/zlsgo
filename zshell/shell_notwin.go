// +build !windows

package zshell

import (
	"os"
	"syscall"
)

func RunNewProcess() (pid int, err error) {
	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	return syscall.ForkExec(os.Args[0], os.Args, execSpec)
}
