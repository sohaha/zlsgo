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

func TestEnvPrefix(T *testing.T) {
	tt := zlsgo.NewTest(T)
	defer unsetEnv("ZLSGO_TEST_NAME", "ZLSGO_TEST_EMPTY")()
	defer func() { zutil.EnvPrefix = "" }()

	_ = os.Setenv("ZLSGO_TEST_NAME", "prefixed")
	_ = os.Setenv("ZLSGO_TEST_EMPTY", "")

	// Without a prefix Getenv reads the bare name.
	tt.Equal("", zutil.EnvPrefix)
	tt.Equal("prefixed", zutil.Getenv("ZLSGO_TEST_NAME"))
	tt.Equal("fallback", zutil.Getenv("ZLSGO_MISSING", "fallback"))

	zutil.EnvPrefix = "ZLSGO_TEST_"
	tt.Equal("prefixed", zutil.Getenv("NAME"))
	// An empty value is valid and must not fall back to the default.
	tt.Equal("", zutil.Getenv("EMPTY", "fallback"))
	tt.Equal("fallback", zutil.Getenv("MISSING", "fallback"))
	// The prefixed lookup never falls back to the unprefixed variable.
	tt.Equal("", zutil.Getenv("ZLSGO_TEST_NAME"))

	// Clearing the prefix restores the unprefixed behavior.
	zutil.EnvPrefix = ""
	tt.Equal("", zutil.Getenv("NAME"))
}

// TestEnvPrefixConcurrentRead covers the documented usage: EnvPrefix is set once
// during initialization and only read afterwards, so concurrent Getenv calls do
// not race. This test needs -race to be meaningful.
func TestEnvPrefixConcurrentRead(T *testing.T) {
	tt := zlsgo.NewTest(T)
	defer unsetEnv("ZLSGO_TEST_SHARED")()
	defer func() { zutil.EnvPrefix = "" }()

	_ = os.Setenv("ZLSGO_TEST_SHARED", "value")
	zutil.EnvPrefix = "ZLSGO_TEST_"

	done := make(chan struct{})
	for i := 0; i < 16; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				if zutil.Getenv("SHARED") != "value" {
					T.Error("unexpected value from Getenv")
					return
				}
			}
		}()
	}
	for i := 0; i < 16; i++ {
		<-done
	}
	tt.Equal("value", zutil.Getenv("SHARED"))
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

func TestLoadenvDefaults(t *testing.T) {
	tt := zlsgo.NewTest(t)
	defer unsetEnv("defaultkey", "baseonly")()
	defer unsetEnv("ZLSGO_ENV", "APP_ENV", "GO_ENV", "NODE_ENV")()

	_ = zfile.WriteFile(".env", []byte("defaultkey=base\nbaseonly=base\n"))
	_ = zfile.WriteFile(".env.local", []byte("defaultkey=local\n"))
	defer zfile.Rmdir(".env")
	defer zfile.Rmdir(".env.local")

	tt.NoError(zutil.Loadenv())
	tt.Equal("local", zutil.Getenv("defaultkey"))
	tt.Equal("base", zutil.Getenv("baseonly"))
}

func TestLoadenvMode(t *testing.T) {
	tt := zlsgo.NewTest(t)
	defer unsetEnv("modekey")()
	defer unsetEnv("ZLSGO_ENV", "APP_ENV", "GO_ENV", "NODE_ENV")()

	files := map[string]string{
		".env":                  "modekey=base\n",
		".env.local":            "modekey=local\n",
		".env.production":       "modekey=mode\n",
		".env.production.local": "modekey=mode_local\n",
	}
	for name, content := range files {
		_ = zfile.WriteFile(name, []byte(content))
		defer zfile.Rmdir(name)
	}

	_ = os.Setenv("ZLSGO_ENV", "production")

	tt.NoError(zutil.Loadenv())
	tt.Equal("mode_local", zutil.Getenv("modekey"))
}

func TestLoadenvOptionalFiles(t *testing.T) {
	tt := zlsgo.NewTest(t)
	defer unsetEnv("ZLSGO_ENV", "APP_ENV", "GO_ENV", "NODE_ENV")()

	// No .env file exists: the convention set is optional and must not fail.
	tt.NoError(zutil.Loadenv())
	// Explicitly requested files are still reported as missing.
	tt.Error(zutil.Loadenv(".env.does.not.exist"))
}

func BenchmarkGetenvNoPrefix(B *testing.B) {
	_ = os.Setenv("ZLSGO_BENCH_NAME", "value")
	defer os.Unsetenv("ZLSGO_BENCH_NAME")
	defer func() { zutil.EnvPrefix = "" }()
	zutil.EnvPrefix = ""
	B.ReportAllocs()
	B.ResetTimer()
	for i := 0; i < B.N; i++ {
		if zutil.Getenv("ZLSGO_BENCH_NAME") != "value" {
			B.Fatal("unexpected value")
		}
	}
}

func BenchmarkGetenvDefaultPrefix(B *testing.B) {
	_ = os.Setenv("ZLSGO_BENCH_PREFIXED_NAME", "value")
	defer os.Unsetenv("ZLSGO_BENCH_PREFIXED_NAME")
	defer func() { zutil.EnvPrefix = "" }()
	zutil.EnvPrefix = "ZLSGO_BENCH_PREFIXED_"
	B.ReportAllocs()
	B.ResetTimer()
	for i := 0; i < B.N; i++ {
		if zutil.Getenv("NAME") != "value" {
			B.Fatal("unexpected value")
		}
	}
}
