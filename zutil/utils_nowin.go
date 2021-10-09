//go:build !windows
// +build !windows

package zutil

import (
	"syscall"
)

const (
	darwinOpenMax = 10240
)

func IsDoubleClickStartUp() bool {
	return false
}

func GetParentProcessName() (string, error) {
	return "", nil
}

// MaxRlimit tries to set the resource limit RLIMIT_NOFILE to the max (hard limit)
func MaxRlimit() (int, error) {
	var lim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim); err != nil {
		return 0, err
	}

	if lim.Cur >= lim.Max {
		return int(lim.Cur), nil
	}

	if IsMac() && lim.Max > darwinOpenMax {
		lim.Max = darwinOpenMax
	}

	oldLimit := lim.Cur
	lim.Cur = lim.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &lim); err != nil {
		return int(oldLimit), err
	}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim); err != nil {
		return 0, err
	}

	return int(lim.Cur), nil
}
