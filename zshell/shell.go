// Package zshell use a simple way to execute shell commands
package zshell

import (
	"bufio"
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

func ExecCommandHandle(ctx context.Context, command []string,
	bef func(cmd *exec.Cmd) error, aft func(cmd *exec.Cmd, err error)) (code int,
	err error) {
	var (
		isSuccess bool
		status    syscall.WaitStatus
	)
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
		fmt.Println("[Command]:", strings.Join(command, " "))
	}

	err = bef(cmd)
	if err != nil {
		return -1, err
	}

	err = cmd.Start()
	if Debug {
		defer func() {
			var userTime time.Duration
			if cmd != nil && cmd.ProcessState != nil {
				userTime = cmd.ProcessState.UserTime()
			}
			if isSuccess {
				fmt.Println("[OK]", status.ExitStatus(), " Used Time:", userTime)
			} else {
				fmt.Println("[Fail]", status.ExitStatus(), " Used Time:", userTime)
			}
		}()
	}

	if aft != nil {
		aft(cmd, err)
	}

	if err != nil {
		return -1, err
	}

	err = cmd.Wait()

	code, isSuccess = cmdResult(cmd)
	return code, err
}

func cmdResult(cmd *exec.Cmd) (code int, isSuccess bool) {
	code = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	isSuccess = cmd.ProcessState.Success()
	return
}

type pipeWork struct {
	cmd *exec.Cmd
	// r   *io.PipeReader
	w *io.PipeWriter
}

func PipeExecCommand(ctx context.Context, commands [][]string) (code int, outStr, errStr string, err error) {
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
	code, err = ExecCommandHandle(ctx, command, func(cmd *exec.Cmd) error {
		cmd.Stdout = stdout
		cmd.Stdin = stdIn
		cmd.Stderr = stderr
		if Dir != "" {
			cmd.Dir = Dir
		}
		return nil
	}, nil)
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
	return BgRunContext(context.Background(), command)
}

func BgRunContext(ctx context.Context, command string) (err error) {
	if strings.TrimSpace(command) == "" {
		return errors.New("no such command")
	}
	arr := fixCommand(command)
	cmd := exec.CommandContext(ctx, arr[0], arr[1:]...)
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

func CallbackRun(command string, callback func(out string, isBasic bool)) (<-chan int, func(string), error) {
	return CallbackRunContext(context.Background(), command, callback)
}

type Options struct {
	Dir string
	Env []string
}

func CallbackRunContext(ctx context.Context, command string, callback func(out string, isBasic bool), opt ...func(option *Options)) (<-chan int, func(string), error) {
	var (
		cmd    *exec.Cmd
		err    error
		cancel context.CancelFunc
		code   = make(chan int, 1)
	)

	ctx, cancel = context.WithCancel(ctx)

	var in func(string)
	read := func(stdout io.ReadCloser, isBasic bool) {
		defer cancel()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			callback(scanner.Text(), isBasic)
		}
	}

	_, err = ExecCommandHandle(ctx, fixCommand(command), func(c *exec.Cmd) error {
		o := Options{}
		for _, v := range opt {
			v(&o)
		}
		if len(o.Env) > 0 {
			c.Env = append(c.Env, o.Env...)
		}
		if o.Dir != "" {
			c.Dir = o.Dir
		}
		cmd = c
		stdin, err := c.StdinPipe()
		if err != nil {
			return err
		}
		in = func(s string) {
			io.WriteString(stdin, s)
		}
		stdout, err := c.StdoutPipe()
		if err != nil {
			return err
		}
		go read(stdout, true)
		stderr, err := c.StderrPipe()
		if err != nil {
			return err
		}
		go read(stderr, false)
		return errors.New("")
	}, nil)

	if err.Error() == "" {
		err = cmd.Start()
		if err == nil {
			go func() {
				_ = cmd.Wait()
				c, _ := cmdResult(cmd)
				code <- c
			}()
		} else {
			code <- -1
		}
	}

	return code, in, err
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
