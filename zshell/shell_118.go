//go:build go1.18
// +build go1.18

package zshell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

func toCommand[T string | []string](command T) []string {
	var commandArgs []string
	switch v := any(command).(type) {
	case []string:
		commandArgs = v
	case string:
		commandArgs = fixCommand(v)
	}
	return commandArgs
}

func CallbackRunContext[T string | []string](ctx context.Context, command T, callback func(str string, isStdout bool), opt ...func(o *Options)) (<-chan int, func(string), error) {
	return callbackRunContext(ctx, toCommand(command), callback, opt...)
}

func CallbackRun[T string | []string](command T, callback func(out string, isBasic bool), opt ...func(o *Options)) (<-chan int, func(string), error) {
	return CallbackRunContext(context.Background(), command, callback, opt...)
}

func Run[T string | []string](command T, opt ...func(o *Options)) (code int, outStr, errStr string, err error) {
	return RunContext(context.Background(), command, opt...)
}

func RunContext[T string | []string](ctx context.Context, command T, opt ...func(o *Options)) (code int, outStr, errStr string, err error) {
	return ExecCommand(ctx, toCommand(command), nil, nil, nil, opt...)
}

func BgRun[T string | []string](command T, opt ...func(o *Options)) (err error) {
	return BgRunContext(context.Background(), command, opt...)
}

func BgRunContext[T string | []string](ctx context.Context, command T, opt ...func(o *Options)) (err error) {
	commandArgs := toCommand(command)
	if len(commandArgs) == 0 {
		return errors.New("no such command")
	}
	cmd := exec.CommandContext(ctx, commandArgs[0], commandArgs[1:]...)
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

func OutRun[T string | []string](command T, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, opt ...func(o Options) Options) (code int, outStr, errStr string, err error) {
	return ExecCommand(context.Background(), toCommand(command), stdIn, stdOut, stdErr)
}
