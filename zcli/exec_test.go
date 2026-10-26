package zcli

import (
	"context"
	"embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

//go:embed testdata/embedfs/root/**
var execEmbedFixtureFS embed.FS

func TestNewExecUsesExistingBinFile(t *testing.T) {
	binFile := newExecTestScript(t)

	engine, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = binFile
	})
	if err != nil {
		t.Fatalf("NewExec() error = %v", err)
	}
	t.Cleanup(func() {
		_ = engine.Release()
	})

	if got, want := engine.Dir(), filepath.Dir(binFile); got != want {
		t.Fatalf("Dir() = %q, want %q", got, want)
	}

	if got, want := filepath.Clean(engine.BinFile()), filepath.Clean(binFile); got != want {
		t.Fatalf("BinFile() = %q, want %q", got, want)
	}
}

func TestNewExecRejectsMissingBinarySource(t *testing.T) {
	_, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = "missing-binary"
	})
	if err == nil {
		t.Fatal("NewExec() expected error when no embedded file and no download URL")
	}

	if !strings.Contains(err.Error(), "download url is empty") {
		t.Fatalf("NewExec() error = %q, want download url is empty", err)
	}
}

func TestExecArgvSkipsEmptyArgs(t *testing.T) {
	binFile := newExecTestScript(t)
	engine := &Exec{binFile: binFile}

	argv, err := engine.argv([]string{"--version", "", "--quiet"})
	if err != nil {
		t.Fatalf("argv() error = %v", err)
	}

	want := []string{binFile, "--version", "--quiet"}
	if len(argv) != len(want) {
		t.Fatalf("argv() = %#v, want %#v", argv, want)
	}
	for i := range want {
		if argv[i] != want[i] {
			t.Fatalf("argv()[%d] = %q, want %q", i, argv[i], want[i])
		}
	}
}

func TestExecArgvRequiresExistingBinary(t *testing.T) {
	engine := &Exec{binFile: filepath.Join(t.TempDir(), "missing")}

	_, err := engine.argv(nil)
	if err == nil {
		t.Fatal("argv() expected error when binary does not exist")
	}
}

func TestExecReleaseKeepsExternalBinDirectory(t *testing.T) {
	binFile := newExecTestScript(t)
	dir := filepath.Dir(binFile)

	engine, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = binFile
	})
	if err != nil {
		t.Fatalf("NewExec() error = %v", err)
	}

	if err := engine.Release(); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("external bin directory should remain after Release(): %v", err)
	}
}

func TestExecRunPassesExactArgvAndOverrideOptions(t *testing.T) {
	binFile := buildExecTestBinary(t, filepath.Join(t.TempDir(), "bin space"))
	workDir := filepath.Join(t.TempDir(), "work space")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	engine, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = binFile
	})
	if err != nil {
		t.Fatalf("NewExec() error = %v", err)
	}
	t.Cleanup(func() {
		_ = engine.Release()
	})

	args := []string{
		"value with spaces",
		`C:\Users\x\data`,
		`SELECT E'\\n' ... 'it''s'`,
		`--flag="quoted value"`,
	}
	code, out, errOut, err := engine.Run(nil, args, func(options *ExecOptions) {
		options.Env = []string{"EXEC_TEST_TOKEN=override"}
		options.Dir = workDir
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if code != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %q", code, errOut)
	}

	normalizedDir := strings.ReplaceAll(workDir, "\\", "/")
	realDir, err := filepath.EvalSymlinks(workDir)
	if err == nil {
		normalizedDir = strings.ReplaceAll(realDir, "\\", "/")
	}

	// On Windows the child reports its working directory with backslashes,
	// so normalise both sides before comparing.
	parts := strings.Split(strings.TrimSpace(out), "|")
	if len(parts) != len(args)+2 {
		t.Fatalf("Run() output = %q, want %d fields", out, len(args)+2)
	}
	if parts[0] != "override" {
		t.Fatalf("Run() token = %q, want override", parts[0])
	}
	if got := strings.ReplaceAll(parts[1], "\\", "/"); got != normalizedDir {
		t.Fatalf("Run() dir = %q, want %q", got, normalizedDir)
	}
	for i, arg := range args {
		if parts[i+2] != arg {
			t.Fatalf("Run() arg %d = %q, want %q", i, parts[i+2], arg)
		}
	}
}

func TestExecKeepRunWritesStdin(t *testing.T) {
	binFile := buildExecTestBinary(t, t.TempDir())

	engine, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = binFile
	})
	if err != nil {
		t.Fatalf("NewExec() error = %v", err)
	}
	t.Cleanup(func() {
		_ = engine.Release()
	})

	var lines []string
	codeCh, input, err := engine.KeepRun(context.Background(), []string{"readline"}, func(str string, isStdout bool) {
		lines = append(lines, str)
	})
	if err != nil {
		t.Fatalf("KeepRun() error = %v", err)
	}
	input("hello stdin\n")

	if code := <-codeCh; code != 0 {
		t.Fatalf("KeepRun() exit code = %d", code)
	}
	if got, want := strings.Join(lines, "\n"), "hello stdin"; got != want {
		t.Fatalf("KeepRun() lines = %q, want %q", got, want)
	}
}

func TestExecKeepRunClosesStdinWhenRequested(t *testing.T) {
	binFile := buildExecTestBinary(t, t.TempDir())

	engine, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = binFile
	})
	if err != nil {
		t.Fatalf("NewExec() error = %v", err)
	}
	t.Cleanup(func() {
		_ = engine.Release()
	})

	codeCh, input, err := engine.KeepRun(context.Background(), []string{"cat"}, func(string, bool) {}, func(options *ExecOptions) {
		options.CloseStdin = true
	})
	if err != nil {
		t.Fatalf("KeepRun() error = %v", err)
	}
	input("ignored after CloseStdin")

	if code := <-codeCh; code != 0 {
		t.Fatalf("KeepRun() exit code = %d", code)
	}
}

func TestExecKeepRunStreamsOutput(t *testing.T) {
	binFile := buildExecTestBinary(t, t.TempDir())

	engine, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = binFile
	})
	if err != nil {
		t.Fatalf("NewExec() error = %v", err)
	}
	t.Cleanup(func() {
		_ = engine.Release()
	})

	var lines []string
	codeCh, _, err := engine.KeepRun(nil, []string{"stream"}, func(str string, isStdout bool) {
		channel := "stderr"
		if isStdout {
			channel = "stdout"
		}
		lines = append(lines, channel+":"+str)
	})
	if err != nil {
		t.Fatalf("KeepRun() error = %v", err)
	}

	code := <-codeCh
	if code != 0 {
		t.Fatalf("KeepRun() exit code = %d", code)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		joined := strings.Join(lines, "\n")
		if strings.Contains(joined, "stdout:fixture-stream") && strings.Contains(joined, "stderr:fixture-err") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "stdout:fixture-stream") {
		t.Fatalf("KeepRun() stdout lines = %q", joined)
	}
	if !strings.Contains(joined, "stderr:fixture-err") {
		t.Fatalf("KeepRun() stderr lines = %q", joined)
	}
}

func newExecTestScript(t *testing.T) string {
	t.Helper()

	return newExecTestScriptInDir(t, t.TempDir())
}

func newExecTestScriptInDir(t *testing.T, dir string) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if runtime.GOOS == "windows" {
		path := filepath.Join(dir, "test-bin.cmd")
		content := "@echo off\r\n" +
			"if \"%~1\"==\"stream\" (\r\n" +
			"  echo fixture-stream\r\n" +
			"  echo fixture-err 1>&2\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"echo %EXEC_TEST_TOKEN%^|%CD%^|%~1\r\n"
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatalf("write test script: %v", err)
		}
		return path
	}

	path := filepath.Join(dir, "test-bin.sh")
	content := "#!/bin/sh\n" +
		"if [ \"$1\" = \"stream\" ]; then\n" +
		"  printf 'fixture-stream\\n'\n" +
		"  printf 'fixture-err\\n' >&2\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf '%s|%s|%s\\n' \"$EXEC_TEST_TOKEN\" \"$PWD\" \"$1\"\n"
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write test script: %v", err)
	}
	return path
}

func buildExecTestBinary(t *testing.T, dir string) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	output := filepath.Join(dir, "fixture-bin")
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", output, "./testdata/fixtureexec")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build test binary: %v\n%s", err, out)
	}

	return output
}
