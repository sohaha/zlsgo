// +build windows

package znet

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// Restart Restart
func (e *Engine) Restart() error {
	return errors.New("Windows does not support")
}

func isKill() bool {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-quit
	return true
}
