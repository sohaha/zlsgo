//go:build windows
// +build windows

package daemon

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func KillSignal() bool {
	sig, stop := SignalChan()
	<-sig
	stop()
	return true
}

func SignalChan() (<-chan os.Signal, func()) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	return quit, func() {
		signal.Stop(quit)
	}
}

func IsSudo() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}
