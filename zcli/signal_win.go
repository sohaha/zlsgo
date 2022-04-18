//go:build windows
// +build windows

package zcli

import (
	"os"
	"os/signal"
	"syscall"
)

func KillSignal() bool {
	<-SignalChan()
	return true
}

func SignalChan() <-chan os.Signal {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	return quit
}
