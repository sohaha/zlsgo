// Package zshell use a simple way to execute shell commands
package zshell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

var (
	Debug = false
	Env   []string
	Dir   string
)

type ShellBuffer struct {
	writer io.Writer
	buf    *bytes.Buffer
}

func newShellStdBuffer(writer io.Writer) *ShellBuffer {
	return &ShellBuffer{
		writer: writer,
		buf:    bytes.NewBuffer([]byte{}),
	}
}

func (s *ShellBuffer) Write(p []byte) (n int, err error) {
	n, err = s.buf.Write(p)
	if s.writer != nil {
		n, err = s.writer.Write(p)
	}
	return n, err
}

func (s *ShellBuffer) String() string {
	return zstring.Bytes2String(s.buf.Bytes())
}

func ExecCommandHandle(ctx context.Context, command []string, bef func(cmd *exec.Cmd), aft func(cmd *exec.Cmd)) (code int, err error) {
	var status syscall.WaitStatus
	if len(command) == 0 || (len(command) == 1 && command[0] == "") {
		return 1, errors.New("no such command")
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	if Env == nil {
		cmd.Env = os.Environ()
	} else {
		cmd.Env = Env
		Env = nil
	}
	if Debug {
		fmt.Println(fmt.Sprintf("[Command]: %s", strings.Join(command, " ")))
	}
	bef(cmd)
	err = cmd.Start()
	isSuccess := false
	defer func() {
		if Debug {
			userTime := time.Duration(0)
			if cmd != nil && cmd.ProcessState != nil {
				userTime = cmd.ProcessState.UserTime()
			}
			if isSuccess {
				fmt.Println("[OK]", status.ExitStatus(), " Used Time:", userTime)
			} else {
				fmt.Println("[Fail]", status.ExitStatus(), " Used Time:", userTime)
			}
		}
	}()
	if err != nil {
		return 1, err
	}
	aft(cmd)
	err = cmd.Wait()
	status = cmd.ProcessState.Sys().(syscall.WaitStatus)
	isSuccess = cmd.ProcessState.Success()

	return status.ExitStatus(), err
}

type pipeWork struct {
	cmd *exec.Cmd
	r   *io.PipeReader
	w   *io.PipeWriter
}

func PipeExecCommand(ctx context.Context, commands [][]string) (code int, outStr, errStr string, err error) {
	defer func() {
		Dir = ""
	}()
	var (
		cmds   []*pipeWork
		out    bytes.Buffer
		outErr bytes.Buffer
		set    func(r *io.PipeReader)
	)
	set = func(r *io.PipeReader) {
		if len(commands) == 0 {
			return
		}
		command := commands[0]
		commands = commands[1:]
		cmd := exec.CommandContext(ctx, command[0], command[1:]...)
		if Dir != "" {
			cmd.Dir = Dir
		}
		if r != nil {
			cmd.Stdin = r
		}
		p := &pipeWork{
			cmd: cmd,
		}
		if len(commands) == 0 {
			cmd.Stdout = &out
			cmd.Stderr = &outErr
		} else {
			r2, w2 := io.Pipe()
			cmd.Stdout = w2
			p.w = w2
			set(r2)
		}
		cmds = append([]*pipeWork{p}, cmds...)
	}
	set(nil)

	for _, v := range cmds {
		err := v.cmd.Start()
		if err != nil {
			return 1, "", "", err
		}
	}
	status := 0
	for _, v := range cmds {
		err := v.cmd.Wait()
		if v.w != nil {
			_ = v.w.Close()
		}
		waitStatus, _ := v.cmd.ProcessState.Sys().(syscall.WaitStatus)
		status = waitStatus.ExitStatus()
		if err != nil {
			return status, "", "", err
		}
	}

	return status, out.String(), "", nil
}

func ExecCommand(ctx context.Context, command []string, stdIn io.Reader, stdOut io.Writer,
	stdErr io.Writer) (code int, outStr, errStr string, err error) {
	stdout := newShellStdBuffer(stdOut)
	stderr := newShellStdBuffer(stdErr)
	code, err = ExecCommandHandle(ctx, command, func(cmd *exec.Cmd) {
		cmd.Stdout = stdOut
		cmd.Stdin = stdIn
		cmd.Stderr = stdErr
		if Dir != "" {
			cmd.Dir = Dir
			Dir = ""
		}
	}, func(cmd *exec.Cmd) {})
	if err != nil {
		return code, "", "", err
	}
	outStr = stdout.String()
	errStr = stderr.String()
	return
}

func Run(command string) (code int, outStr, errStr string, err error) {
	return RunContext(context.Background(), command)
}

func RunContext(ctx context.Context, command string) (code int, outStr, errStr string, err error) {
	return ExecCommand(ctx, fixCommand(command), nil, nil, nil)
}

func OutRun(command string, stdIn io.Reader, stdOut io.Writer,
	stdErr io.Writer) (code int, outStr, errStr string, err error) {
	return ExecCommand(context.Background(), fixCommand(command), stdIn, stdOut, stdErr)
}

func BgRun(command string) (err error) {
	if strings.TrimSpace(command) == "" {
		return errors.New("no such command")
	}
	arr := fixCommand(command)
	cmd := exec.Command(arr[0], arr[1:]...)
	err = cmd.Start()
	if Debug {
		fmt.Println(fmt.Sprintf("[Command]: %s",
			command))
		if err != nil {
			fmt.Println("[Fail]", err.Error())
		}
	}
	return err
}

func fixCommand(command string) (runCommand []string) {
	tmp := ""
	for _, v := range strings.Split(command, " ") {
		if strings.HasSuffix(v, "\\") {
			tmp += v[:len(v)-1] + " "
		} else {
			if tmp != "" {
				v = tmp + v
				tmp = ""
			}
			runCommand = append(runCommand, v)
		}
	}
	return
}
