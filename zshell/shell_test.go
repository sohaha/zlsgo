package zshell

import (
	"os"
	"testing"

	"github.com/sohaha/zlsgo"
)

func Test_bash(T *testing.T) {
	Debug = true
	t := zlsgo.NewTest(T)

	var res string
	var errRes string
	var code int
	var err error

	code, res, errRes, err = Run("")
	t.EqualExit(1, code)
	t.EqualExit(true, err != nil)
	t.Log(res, errRes)

	code, _, _, err = Run("lll")
	t.EqualExit(1, code)
	t.EqualExit(true, err != nil)
	t.Log(err)

	code, _, _, err = Run("ls")
	t.EqualExit(0, code)
	t.EqualExit(true, err == nil)
	t.Log(err)

	code, _, _, err = Run("curl b.c")
	t.Log(err)
	t.EqualExit(6, code)
	t.EqualExit(true, err != nil)

	code, _, _, err = Run("curl qq.com")
	t.Log(err)
	t.EqualExit(0, code)
	t.EqualExit(true, err == nil)

	err = BgRun("")
	t.EqualExit(true, err != nil)
	err = BgRun("lll")
	t.EqualExit(true, err != nil)
	t.Log(err)

	Dir = "."
	Env = []string{"kkk"}
	code, res, errRes, err = OutRun("ls", os.Stdin, os.Stdout, os.Stdin)
	t.Log(res, errRes, code, err)
}
