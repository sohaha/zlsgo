package zcli

import (
	"context"
	"embed"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sohaha/zlsgo/zfile"
)

func TestInitExecTmpDirReturnsExistingDirectory(t *testing.T) {
	path := initExecTmpDir()
	if path == "" {
		t.Fatal("initExecTmpDir() returned empty path")
	}

	if !zfile.DirExist(path) {
		t.Fatalf("initExecTmpDir() = %q, directory does not exist", path)
	}
}

func TestWriteExecFileWritesBinaryToDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tool.bin")
	data := []byte("fixture")

	if err := writeExecFile(path, data); err != nil {
		t.Fatalf("writeExecFile() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(got) != string(data) {
		t.Fatalf("file content = %q, want %q", got, data)
	}
}

func TestWriteExecFileRejectsEmptyBinary(t *testing.T) {
	err := writeExecFile(filepath.Join(t.TempDir(), "tool.bin"), nil)
	if err == nil {
		t.Fatal("writeExecFile() expected error for empty binary")
	}
}

func TestMergeExecCtxNilContextUsesEngineContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	engine := &Exec{ctx: ctx}

	runCtx, stop := mergeExecCtx(engine, nil)
	defer stop()

	if runCtx != ctx {
		t.Fatal("mergeExecCtx() with nil ctx should reuse engine context")
	}

	cancel()
	select {
	case <-runCtx.Done():
	default:
		t.Fatal("engine context should be cancelled")
	}
}

func TestMergeExecCtxCancelsWhenEngineContextEnds(t *testing.T) {
	engineCtx, engineCancel := context.WithCancel(context.Background())
	userCtx, userCancel := context.WithCancel(context.Background())
	defer userCancel()

	engine := &Exec{ctx: engineCtx}
	runCtx, stop := mergeExecCtx(engine, userCtx)
	defer stop()

	engineCancel()

	select {
	case <-runCtx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("merged context was not cancelled in time")
	}

	if err := runCtx.Err(); err == nil {
		t.Fatal("merged context should be cancelled after engine context ends")
	}
}

func TestExecRunSpecUsesOverrideValues(t *testing.T) {
	engine := &Exec{
		options: &ExecOptions{
			Env:        []string{"ENGINE=1"},
			Dir:        "engine-dir",
			CloseStdin: true,
		},
	}

	spec := engine.runSpec(func(o *ExecOptions) {
		o.Env = []string{"REQUEST=2"}
		o.Dir = "request-dir"
		o.CloseStdin = false
	})

	if got, want := spec.Dir, "request-dir"; got != want {
		t.Fatalf("spec dir = %q, want %q", got, want)
	}

	if !reflect.DeepEqual(spec.Env, []string{"REQUEST=2"}) {
		t.Fatalf("spec env = %#v, want override env", spec.Env)
	}

	if spec.CloseStdin {
		t.Fatal("spec CloseStdin = true, want override false")
	}
}

func TestExecArgvKeepsArgumentBoundaries(t *testing.T) {
	binFile := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(binFile, []byte("fixture"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	engine := &Exec{binFile: binFile}
	args := []string{`C:\Program Files\tool.cmd`, "hello world", `--sql=SELECT E'\n'`, "--quiet"}
	argv, err := engine.argv(args)
	if err != nil {
		t.Fatalf("argv() error = %v", err)
	}

	want := append([]string{binFile}, args...)
	if !reflect.DeepEqual(argv, want) {
		t.Fatalf("argv() = %#v, want %#v", argv, want)
	}
}

func TestExecKeepRunStopsOnContextCancel(t *testing.T) {
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

	ctx, cancel := context.WithCancel(context.Background())
	codeCh, _, err := engine.KeepRun(ctx, []string{"sleep"}, func(string, bool) {})
	if err != nil {
		t.Fatalf("KeepRun() error = %v", err)
	}
	cancel()

	select {
	case code := <-codeCh:
		if code == 0 {
			t.Fatalf("KeepRun() code = %d, want non-zero after cancel", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("KeepRun() did not stop after context cancel")
	}
}

// TestExecKeepRunWaitDelayBoundsOrphanGrandchild locks in the WaitDelay
// guarantee: a grandchild that outlives the child while holding the output
// pipes must not keep KeepRun blocked for the grandchild's lifetime.
func TestExecKeepRunWaitDelayBoundsOrphanGrandchild(t *testing.T) {
	binFile := buildExecTestBinary(t, t.TempDir())
	engine, err := NewExec(embed.FS{}, func(options *ExecOptions) {
		options.BinFile = binFile
		options.WaitDelay = 2 * time.Second
	})
	if err != nil {
		t.Fatalf("NewExec() error = %v", err)
	}
	t.Cleanup(func() {
		_ = engine.Release()
		killExecFixtureGrandchildren(binFile)
	})

	start := time.Now()
	codeCh, _, err := engine.KeepRun(context.Background(), []string{"orphan"}, func(string, bool) {})
	if err != nil {
		t.Fatalf("KeepRun() error = %v", err)
	}

	select {
	case code := <-codeCh:
		if elapsed := time.Since(start); elapsed > 15*time.Second {
			t.Fatalf("KeepRun() returned after %v; WaitDelay did not bound orphan draining", elapsed)
		} else {
			t.Logf("KeepRun() returned code=%d in %v", code, elapsed)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("KeepRun() blocked beyond grandchild-independent deadline; WaitDelay not applied")
	}
}

// killExecFixtureGrandchildren stops any orphaned fixture binary left holding
// the pipes so t.TempDir cleanup does not fail on Windows file locks.
func killExecFixtureGrandchildren(binFile string) {
	if runtime.GOOS != "windows" {
		return
	}
	exe := binFile
	if !strings.HasSuffix(strings.ToLower(exe), ".exe") {
		exe += ".exe"
	}
	cmd := exec.Command("taskkill", "/F", "/IM", filepath.Base(exe))
	_ = cmd.Run()
}
