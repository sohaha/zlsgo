//go:build windows
// +build windows

package zshell

import (
	"context"
	"errors"
)

func RunNewProcess(filemd5 string) (pid int, err error) {
	return 0, errors.New("windows does not support")
}

func RunBash(ctx context.Context, command string) (code int, outStr, errStr string, err error) {
	return ExecCommand(ctx, []string{
		"cmd",
		"/C",
		command,
	}, nil, nil, nil)
}
