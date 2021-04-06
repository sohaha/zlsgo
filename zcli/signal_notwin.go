// +build !windows

package zcli

import (
	"os"
	"os/signal"
	"syscall"
)

func KillSignal() bool {
	quit := make(chan os.Signal)
	// signal.Notify(quit, os.Interrupt, os.Kill)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGUSR2)
	sig := <-quit
	return sig != syscall.SIGUSR2
}
