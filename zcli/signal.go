package zcli

import (
	"github.com/sohaha/zlsgo/zutil/daemon"
)

func SingleKillSignal() <-chan bool {
	return daemon.SingleKillSignal()
}
