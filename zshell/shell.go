// Package zshell use a simple way to execute shell commands
package zshell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

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

func ExecCommand(command []string, stdIn io.Reader, stdOut io.Writer,
	stdErr io.Writer) (code int, outStr, errStr string, err error) {

	var (
		status syscall.WaitStatus
		stdout *ShellBuffer
		stderr *ShellBuffer
	)

	if len(command) == 0 {
		return 1, "", "", errors.New("no such command")
	}

	if Debug {
		fmt.Println(fmt.Sprintf("[Command]\n%s\n%s",
			command, strings.Repeat("-", len(command))))
	}
	var cmd = exec.Command(command[0], command[1:]...)

	if Env == nil {
		cmd.Env = os.Environ()
	} else {
		cmd.Env = Env
		Env = nil
	}
	if Dir != "" {
		cmd.Dir = Dir
		Dir = ""
	}
	stdout = newShellStdBuffer(stdOut)
	stderr = newShellStdBuffer(stdErr)

	cmd.Stdout = stdout
	cmd.Stdin = stdIn
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		return 1, "", "", err
	}
	err = cmd.Wait()
	status = cmd.ProcessState.Sys().(syscall.WaitStatus)
	isSuccess := cmd.ProcessState.Success()
	if Debug {
		fmt.Println(strings.Repeat("-", len(command)))
		if isSuccess {
			fmt.Println("[OK]", status.ExitStatus(), " Used Time:", cmd.ProcessState.UserTime())
		} else {
			fmt.Println("[Fail]", status.ExitStatus(), " Used Time:", cmd.ProcessState.UserTime())
		}
	}

	outStr = stdout.String()
	errStr = stderr.String()

	return status.ExitStatus(), outStr, errStr, err
}

func Run(command string) (code int, outStr, errStr string, err error) {
	return ExecCommand(strings.Split(command, " "), nil, nil, nil)
}

func OutRun(command string, stdIn io.Reader, stdOut io.Writer,
	stdErr io.Writer) (code int, outStr, errStr string, err error) {
	return ExecCommand(strings.Split(command, " "), stdIn, stdOut, stdErr)
}

func BgRun(command string) (err error) {
	if strings.TrimSpace(command) == "" {
		return errors.New("no such command")
	}

	arr := strings.Split(command, " ")
	cmd := exec.Command(arr[0], arr[1:]...)
	err = cmd.Start()

	if Debug {
		fmt.Println(fmt.Sprintf("[Command]\n%s\n%s",
			command, strings.Repeat("-", len(command))))
		if err != nil {
			fmt.Println("[Error]:", err.Error())
		}
	}
	return err
}
