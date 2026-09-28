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

// unsetEnv removes the given environment variables and returns a function that
// restores their original values.
func unsetEnv(keys ...string) func() {
	original := make(map[string]string, len(keys))
	missing := make([]string, 0, len(keys))
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			original[key] = value
		} else {
			missing = append(missing, key)
		}
		_ = os.Unsetenv(key)
	}
	return func() {
		for _, key := range missing {
			_ = os.Unsetenv(key)
		}
		for key, value := range original {
			_ = os.Setenv(key, value)
		}
	}
}

func TestLoadenv(t *testing.T) {
	tt := zlsgo.NewTest(t)
	defer unsetEnv("myos", "name", "time", "comment", "description")()

	_ = zfile.WriteFile(".env", []byte("myos=linux\n name=zls \n\n  time=\"2024-11-14 23:59:01\" \n#comment='comment'\n description=\"hello world\""))
	defer zfile.Rmdir(".env")

	tt.NoError(zutil.Loadenv())

	tt.Equal("linux", zutil.Getenv("myos"))
	tt.Equal("zls", zutil.Getenv("name"))
	tt.Equal("2024-11-14 23:59:01", zutil.Getenv("time"))
	tt.Equal("", zutil.Getenv("comment"))
	tt.Equal("hello world", zutil.Getenv("description"))
}

func TestLoadenvPrecedence(t *testing.T) {
	tt := zlsgo.NewTest(t)
	defer unsetEnv("keepkey", "newkey")()

	_ = zfile.WriteFile(".env", []byte("keepkey=from_file\nnewkey=from_file\n"))
	defer zfile.Rmdir(".env")

	_ = os.Setenv("keepkey", "from_env")

	tt.NoError(zutil.Loadenv())
	tt.Equal("from_env", zutil.Getenv("keepkey"))
	tt.Equal("from_file", zutil.Getenv("newkey"))
}

func TestLoadenvFileOrder(t *testing.T) {
	tt := zlsgo.NewTest(t)
	defer unsetEnv("orderkey")()

	_ = zfile.WriteFile(".env.base", []byte("orderkey=base\n"))
	_ = zfile.WriteFile(".env.local", []byte("orderkey=local\n"))
	defer zfile.Rmdir(".env.base")
	defer zfile.Rmdir(".env.local")

	tt.NoError(zutil.Loadenv(".env.base", ".env.local"))
	tt.Equal("local", zutil.Getenv("orderkey"))
}
