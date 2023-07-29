package zshell

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sohaha/zlsgo/zutil"

	"github.com/sohaha/zlsgo"
)

func TestPipe(t *testing.T) {
	if zutil.IsWin() {
		t.Log("ignore windows")
		return
	}
	tt := zlsgo.NewTest(t)
	ctx := context.Background()

	commands := [][]string{
		{"ls", "-a"},
		{"grep", "go"},
		{"grep", "shell_notwin"},
	}

	code, outStr, errStr, err := PipeExecCommand(ctx, commands)
	t.Log(outStr, errStr, err)
	tt.EqualExit(0, code)
	tt.EqualExit("shell_notwin.go", strings.Trim(outStr, " \n"))

	Dir = "../"
	code, outStr, errStr, err = PipeExecCommand(ctx, [][]string{{"ls"}})
	t.Log(code, outStr, errStr, err)

	code, outStr, errStr, err = PipeExecCommand(ctx, [][]string{})
	t.Log(code, outStr, errStr, err)
}

func TestBash(t *testing.T) {
	Debug = true
	tt := zlsgo.NewTest(t)

	var res string
	var errRes string
	var code int
	var err error

	code, res, errRes, err = Run("")
	tt.EqualExit(1, code)
	tt.EqualExit(true, err != nil)
	t.Log(res, errRes)

	code, _, _, err = Run("lll")
	tt.EqualExit(-1, code)
	tt.EqualExit(true, err != nil)
	t.Log(err)

	if !zutil.IsWin() {
		code, _, _, err = Run("ls")
		tt.EqualExit(0, code)
		tt.EqualExit(true, err == nil)
		t.Log(err)
	}

	_, res, _, err = Run("ls -a /Applications/Google\\ Chrome.app")
	t.Log(res)
	t.Log(err)

	err = BgRun("")
	tt.EqualExit(true, err != nil)
	err = BgRun("lll")
	tt.EqualExit(true, err != nil)
	t.Log(err)

	Dir = "."
	Env = []string{"kkk"}
	code, res, errRes, err = OutRun("ls", os.Stdin, os.Stdout, os.Stdin)
	t.Log(res, errRes, code, err)
}

func TestCallbackRun(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var i = 0
	var code <-chan int
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	code, _, err = CallbackRunContext(ctx, "ping www.npmjs.com", func(out string, isBasic bool) {
		fmt.Println(out)
		i = i + 1
		if i > 3 {
			cancel()
		}
	})
	tt.NoError(err)
	tt.Log("code", <-code)
}
