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

func (c *Config) execPath() (path string) {
	if len(c.Executable) != 0 {
		path, _ = filepath.Abs(c.Executable)
		return path
	}
	path, _ = os.Executable()
	return
}

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

func IsPermissionError(err error) bool {
	return err == ErrNotAnAdministrator || err == ErrNotAnRootUser
}

func run(command string, args ...string) error {
	var out bytes.Buffer
	var outErr bytes.Buffer
	commands := []string{command}
	commands = append(commands, args...)
	return runcmd(commands, bytes.NewReader([]byte("")), &out, &outErr)
}

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

func isServiceRestart(c *Config) bool {
	load := optionRunAtLoadDefault
	if l, ok := c.Options[optionRunAtLoad]; ok {
		load = ztype.ToBool(l)
	}
	return load
}
