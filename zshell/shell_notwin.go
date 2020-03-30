// +build !windows

package zshell

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sohaha/zlsgo/zstring"
)

func RunNewProcess(filemd5 string) (pid int, err error) {
	args := os.Args
	file := args[0]
	if filemd5 != "" {
		currentMd5, err := zstring.Md5File(file)
		if err != nil || filemd5 != currentMd5 {
			return 0, errors.New("md5 verification of the file failed")
		}
	}
	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	if tmp, _ := ioutil.TempDir("", ""); tmp != "" {
		tmp = filepath.Dir(tmp)
		if strings.HasPrefix(file, tmp) {
			return 0, errors.New("temporary program does not support startup")
		}
	}
	return syscall.ForkExec(file, args, execSpec)
}
