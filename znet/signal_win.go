// +build windows

package znet

import (
	"os"
	"os/signal"
	"syscall"
)

func isKill() bool {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-quit
	return true
}
