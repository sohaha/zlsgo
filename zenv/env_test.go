package zenv

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"runtime"
)

func TestOs(T *testing.T) {
	t := zlsgo.NewTest(T)
	osName := runtime.GOOS
	t.Log(osName)
	isWin := IsWin()
	t.Log("isWin", isWin)
	isLinux := IsLinux()
	t.Log("isLinux", isLinux)
	isMac := IsMac()
	t.Log("isMac", isMac)
}
func TestEnv(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(Getenv("HOME"))
	t.Log(Getenv("myos"))
	t.Log(Getenv("我不存在", "66"))
}
