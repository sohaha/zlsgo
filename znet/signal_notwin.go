//go:build !windows
// +build !windows

package znet

import (
	"errors"
	"os"
	"syscall"

	"github.com/sohaha/zlsgo/zutil"
)

var isRestarting = zutil.NewBool(false)

// Restart triggers a server process restart
func (e *Engine) Restart() error {
	if !isRestarting.CAS(false, true) {
		return errors.New("restart already in progress")
	}

	defer isRestarting.Store(false)

	pid := os.Getpid()
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return proc.Signal(syscall.SIGUSR2)
}

// IsRestarting returns whether the server is currently restarting
func (e *Engine) IsRestarting() bool {
	return isRestarting.Load()
}
