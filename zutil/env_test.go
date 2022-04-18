package zutil_test

import (
	"runtime"
	"testing"

	"github.com/sohaha/zlsgo"
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
