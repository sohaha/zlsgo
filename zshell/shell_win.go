//go:build windows
// +build windows

package zshell

import (
	"context"
	"errors"
	"os/exec"
	"syscall"

	"github.com/sohaha/zlsgo/zutil"
)

var chcp = zutil.Once(func() struct{} {
	_ = exec.Command("chcp", "65001").Run()
	return struct{}{}
})

func RunNewProcess(file string, args []string) (pid int, err error) {
	return 0, errors.New("windows does not support")
}

func RunBash(ctx context.Context, command string) (code int, outStr, errStr string, err error) {
	return ExecCommand(ctx, []string{
		"cmd",
		"/C",
		command,
	}, nil, nil, nil)
}

func sysProcAttr(cmd *exec.Cmd) *exec.Cmd {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			// CreationFlags: 0x08000000,
		}
	}

	cmd.SysProcAttr.HideWindow = true
	return cmd
}
