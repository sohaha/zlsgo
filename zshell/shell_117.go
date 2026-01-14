//go:build !go1.18
// +build !go1.18

package zshell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func CallbackRunContext(ctx context.Context, command string, callback func(str string, isStdout bool), opt ...func(o *Options)) (<-chan int, func(string), error) {
	return callbackRunContext(ctx, fixCommand(command), callback, opt...)
}

func CallbackRun(command string, callback func(out string, isBasic bool), opt ...func(o *Options)) (<-chan int, func(string), error) {
	return CallbackRunContext(context.Background(), command, callback, opt...)
}

func Run(command string, opt ...func(o *Options)) (code int, outStr, errStr string, err error) {
	return RunContext(context.Background(), command, opt...)
}

func RunContext(ctx context.Context, command string, opt ...func(o *Options)) (code int, outStr, errStr string, err error) {
	return ExecCommand(ctx, fixCommand(command), nil, nil, nil, opt...)
}

func BgRun(command string, opt ...func(o *Options)) (err error) {
	return BgRunContext(context.Background(), command, opt...)
}

func BgRunContext(ctx context.Context, command string, opt ...func(o *Options)) (err error) {
	if strings.TrimSpace(command) == "" {
		return errors.New("no such command")
	}
	arr := fixCommand(command)
	cmd := exec.CommandContext(ctx, arr[0], arr[1:]...)
	wrapOptions(cmd, opt...)
	err = cmd.Start()
	if Debug {
		fmt.Println("[Command]: ", command)
		if err != nil {
			fmt.Println("[Fail]", err.Error())
		}
	}
	go func() {
		_ = cmd.Wait()
	}()
	return err
}

func OutRun(command string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, opt ...func(o Options) Options) (code int, outStr, errStr string, err error) {
	return ExecCommand(context.Background(), fixCommand(command), stdIn, stdOut, stdErr)
}
