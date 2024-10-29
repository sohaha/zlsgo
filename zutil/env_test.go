package zutil_test

import (
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
}

func TestEnv(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(zutil.Getenv("HOME"))
	t.Log(zutil.Getenv("myos"))
	t.Log(zutil.Getenv("我不存在", "66"))
}

func TestGOROOT(t *testing.T) {
	t.Log(zutil.GOROOT())
}

func TestLoadenv(t *testing.T) {
	tt := zlsgo.NewTest(t)
	_ = zfile.WriteFile(".env", []byte("myos=linux\n name=zls "))
	defer zfile.Rmdir(".env")

	tt.NoError(zutil.Loadenv())

	tt.Equal("linux", zutil.Getenv("myos"))
	tt.Equal("zls", zutil.Getenv("name"))
}
