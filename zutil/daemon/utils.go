package daemon

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zshell"
	"github.com/sohaha/zlsgo/ztype"
)

// execPath determines the absolute path to the executable for the service.
func (c *Config) execPath() (path string) {
	if len(c.Executable) != 0 {
		path, _ = filepath.Abs(c.Executable)
		return path
	}
	path, _ = os.Executable()
	return
}

// runGrep executes a command and then greps the output for a specific pattern.
// This is useful for filtering command output to find specific information.
func runGrep(grep, command string, args ...string) (res string, err error) {
	var grepout bytes.Buffer
	var out bytes.Buffer
	var outErr bytes.Buffer
	commands := []string{command}
	commands = append(commands, args...)
	err = runcmd(commands, bytes.NewReader([]byte("")), &out, &outErr)
	if err != nil {
		return
	}
	commands = []string{"grep", grep}
	err = runcmd(commands, bytes.NewReader(out.Bytes()), &grepout, &outErr)
	if err != nil {
		return
	}
	res = grepout.String()
	return
}

// isSudo checks if the current process is running with root/sudo privileges.
// This is used to verify if the process has sufficient permissions for
// system-level operations.
func isSudo() error {
	_, id, _, err := zshell.Run("id -u")
	if err != nil {
		return err
	}
	id = strings.Replace(id, "\n", "", -1)
	if id != "0" {
		return ErrNotAnRootUser
	}
	return nil
}

// IsPermissionError checks if an error is related to insufficient permissions.
// It returns true if the error is either ErrNotAnAdministrator or ErrNotAnRootUser.
func IsPermissionError(err error) bool {
	return err == ErrNotAnAdministrator || err == ErrNotAnRootUser
}

// run executes a command with the given arguments.
// It captures the command output but doesn't return it, only returning an error if the command fails.
func run(command string, args ...string) error {
	var out bytes.Buffer
	var outErr bytes.Buffer
	commands := []string{command}
	commands = append(commands, args...)
	return runcmd(commands, bytes.NewReader([]byte("")), &out, &outErr)
}

// runcmd is a low-level function to execute a command with the given input and output buffers.
// It's used internally by run and runGrep to execute commands.
func runcmd(commands []string, in *bytes.Reader, out, outErr *bytes.Buffer) error {
	code, _, _, err := zshell.ExecCommand(context.Background(), commands, in, out, outErr)
	if err != nil {
		return err
	}
	if code != 0 {
		errMsg := outErr.String()
		if errMsg == "" {
			errMsg = out.String()
		}
		err = errors.New(errMsg)
	}
	return err
}

// isServiceRestart determines if the service should be restarted automatically.
// It checks the RunAtLoad option in the Config, defaulting to true if not specified.
func isServiceRestart(c *Config) bool {
	load := optionRunAtLoadDefault
	if l, ok := c.Options[optionRunAtLoad]; ok {
		load = ztype.ToBool(l)
	}
	return load
}
