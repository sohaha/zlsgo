// +build !windows

package znet

import (
	"os"
	"os/signal"
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

func isKill() bool {
	quit := make(chan os.Signal)
	// signal.Notify(quit, os.Interrupt, os.Kill)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGUSR2)
	sig := <-quit
	return sig != syscall.SIGUSR2
}
