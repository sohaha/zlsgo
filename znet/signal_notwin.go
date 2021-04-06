// +build !windows

package znet

import (
	"os"
	"syscall"
)

// Restart Restart
func (e *Engine) Restart() error {
	pid := os.Getpid()
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGUSR2)
}
