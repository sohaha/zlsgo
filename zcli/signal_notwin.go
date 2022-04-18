//go:build !windows
// +build !windows

package zcli

import (
	"os"
	"os/signal"
	"syscall"
)

func KillSignal() bool {
	sig := <-SignalChan()
	return sig != syscall.SIGUSR2
}

func SignalChan() <-chan os.Signal {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGUSR2)
	return quit
}
