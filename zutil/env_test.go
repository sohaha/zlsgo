package zutil_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zutil"
)

func TestOs(T *testing.T) {
	t := zlsgo.NewTest(T)

	osName := runtime.GOOS
	t.Log(osName)
	isWin := zutil.IsWin()
	t.Log("isWin", isWin)
	isLinux := zutil.IsLinux()
	t.Log("isLinux", isLinux)
	isMac := zutil.IsMac()
	t.Log("isMac", isMac)

	is32bit := zutil.Is32BitArch()
	t.Log(is32bit)
}

func TestEnv(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(zutil.Getenv("HOME"))
	t.Log(zutil.Getenv("myos"))
	t.Log(zutil.Getenv("我不存在", "66"))
	_ = os.Setenv("TEST_EMPTY_VAR", "")
	defer os.Unsetenv("TEST_EMPTY_VAR")
	t.Equal("", zutil.Getenv("TEST_EMPTY_VAR", "default"))
	t.Equal("default", zutil.Getenv("NON_EXISTENT_VAR", "default"))
	_ = os.Setenv("TEST_SET_VAR", "actual_value")
	defer os.Unsetenv("TEST_SET_VAR")
	t.Equal("actual_value", zutil.Getenv("TEST_SET_VAR", "default"))
}

func TestGOROOT(t *testing.T) {
	t.Log(zutil.GOROOT())
}

func TestLoadenv(t *testing.T) {
	tt := zlsgo.NewTest(t)
	_ = zfile.WriteFile(".env", []byte("myos=linux\n name=zls \n\n  time=\"2024-11-14 23:59:01\" \n#comment='comment'\n description=\"hello world\""))
	defer zfile.Rmdir(".env")

	tt.NoError(zutil.Loadenv())

	tt.Equal("linux", zutil.Getenv("myos"))
	tt.Equal("zls", zutil.Getenv("name"))
	tt.Equal("2024-11-14 23:59:01", zutil.Getenv("time"))
	tt.Equal("", zutil.Getenv("comment"))
	tt.Equal("hello world", zutil.Getenv("description"))
}
