package zcli

// Exec is a lightweight engine for preparing and running local, embedded, or
// downloaded binaries through a unified Go API.

import (
	"context"
	"embed"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

// ExecOptions controls how Exec resolves, prepares and runs a binary.
type ExecOptions struct {
	// BinFile is either an existing binary path or the target file name after extraction/download.
	BinFile string
	// DownloadHandler keeps backward compatibility for callers that only need the downloaded file path.
	DownloadHandler func(file string) error
	// DownloadHandlerWithDir receives both the downloaded file path and the Exec target directory.
	DownloadHandlerWithDir func(file string, dir string) error
	// DownloadProgress reports current download percentage and total size in bytes.
	DownloadProgress func(percent float64, size int64)
	// DownloadUrl is used when BinFile cannot be resolved locally or from embed.FS.
	DownloadUrl string
	// Env overrides process environment variables for Run and KeepRun.
	Env []string
	// Dir overrides the working directory for Run and KeepRun.
	Dir string
	// CloseStdin closes the child stdin immediately in KeepRun, so commands
	// that read until EOF can finish without waiting for caller input.
	CloseStdin bool
	// WaitDelay bounds how long Run/KeepRun keep draining output pipes after
	// the context is cancelled or the process exits. Orphaned grandchildren
	// that inherit stdout otherwise keep Wait blocked indefinitely. The
	// default is 5s; a negative value disables the bound.
	WaitDelay time.Duration
}

// Exec owns the lifecycle of one prepared binary, including optional extraction,
// download, execution, and cleanup of Exec-owned temporary files.
type Exec struct {
	options *ExecOptions
	tmpDir  string
	ctx     context.Context
	cancel  context.CancelFunc
	fs      embed.FS
	binFile string
	owned   bool
}

// NewExec creates an Exec engine and prepares the target binary from one of three sources:
// an existing BinFile, embedded files in fs, or DownloadUrl.
func NewExec(fs embed.FS, fn ...func(options *ExecOptions)) (*Exec, error) {
	ctx, cancel := context.WithCancel(context.Background())
	options := zutil.Optional(ExecOptions{}, fn...)
	e := &Exec{
		options: &options,
		ctx:     ctx,
		cancel:  cancel,
		fs:      fs,
	}

	if !zfile.FileExist(options.BinFile) {
		if err := e.toDisk(); err != nil {
			cancel()
			return nil, err
		}
		e.binFile = filepath.Join(e.tmpDir, options.BinFile)
	} else {
		e.binFile = zfile.RealPath(options.BinFile)
		e.tmpDir = filepath.Dir(e.binFile)
	}

	return e, nil
}

// Dir returns the working directory that currently stores the prepared binary.
func (e *Exec) Dir() string {
	return e.tmpDir
}

// BinFile returns the resolved binary path used by the current Exec instance.
func (e *Exec) BinFile() string {
	return zfile.RealPath(e.binFile)
}

// toDisk extracts the embedded binary or downloads it into an Exec-owned temp dir.
func (e *Exec) toDisk() error {
	e.tmpDir = filepath.Join(execTmpDir, zstring.UUID())
	e.owned = true
	total, err := embedToDisk(e.fs, e.tmpDir)
	if err != nil {
		return err
	}

	if total > 0 {
		return nil
	}

	return e.downloadToDisk()
}

// runSpec holds the process settings resolved from ExecOptions.
func (e *Exec) runSpec(o ...func(o *ExecOptions)) execRunSpec {
	opt := zutil.Optional(ExecOptions{
		Env:        e.options.Env,
		Dir:        e.options.Dir,
		CloseStdin: e.options.CloseStdin,
		WaitDelay:  e.options.WaitDelay,
	}, o...)
	if opt.WaitDelay == 0 {
		opt.WaitDelay = defaultExecWaitDelay
	}
	return execRunSpec{Env: opt.Env, Dir: opt.Dir, CloseStdin: opt.CloseStdin, WaitDelay: opt.WaitDelay}
}

// defaultExecWaitDelay bounds pipe draining after cancellation so orphaned
// grandchild processes cannot keep Run/KeepRun blocked forever.
const defaultExecWaitDelay = 5 * time.Second

// Run executes the prepared binary synchronously and returns exit code, stdout, stderr and error.
func (e *Exec) Run(ctx context.Context, args []string, opt ...func(o *ExecOptions)) (code int, outStr string, errStr string, err error) {
	argv, err := e.argv(args)
	if err != nil {
		return -1, "", "", err
	}

	runCtx, cancel := mergeExecCtx(e, ctx)
	defer cancel()

	return runExecCommand(runCtx, argv, e.runSpec(opt...))
}

// KeepRun executes the prepared binary in streaming mode.
// It returns a code channel and an input function for writing to stdin.
func (e *Exec) KeepRun(ctx context.Context, args []string, callback func(str string, isStdout bool), opt ...func(o *ExecOptions)) (<-chan int, func(string), error) {
	argv, err := e.argv(args)
	if err != nil {
		return nil, nil, err
	}

	runCtx, cancel := mergeExecCtx(e, ctx)
	codeCh, input, err := callbackRunExecCommand(runCtx, argv, callback, e.runSpec(opt...))
	if err != nil {
		cancel()
		return nil, nil, err
	}

	done := make(chan int, 1)
	go func() {
		defer cancel()
		code, ok := <-codeCh
		if ok {
			done <- code
		}
		close(done)
	}()

	return done, input, nil
}

// Release cancels the engine context and removes temporary files created by the current Exec.
// External BinFile targets are never deleted.
func (e *Exec) Release() error {
	e.cancel()

	if !e.owned || e.tmpDir == "" {
		return nil
	}

	if _, err := os.Stat(e.tmpDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.RemoveAll(e.tmpDir); err != nil {
		return errors.New("remove temp dir " + e.tmpDir + ": " + err.Error())
	}

	return nil
}

// argv returns the exact argument vector for the prepared binary.
// Empty arguments are dropped, matching the historical behaviour.
func (e *Exec) argv(attr []string) ([]string, error) {
	if !zfile.FileExist(e.binFile) {
		return nil, errors.New("BinFile not set")
	}

	args := make([]string, 1, len(attr)+1)
	args[0] = e.binFile
	for _, arg := range attr {
		if arg != "" {
			args = append(args, arg)
		}
	}

	return args, nil
}

// mergeExecCtx derives a run context that ends when either the user context or
// the engine context ends. context.AfterFunc avoids spawning a goroutine per run.
// Both stop and cancel are idempotent, so the returned CancelFunc is safe to call
// more than once.
func mergeExecCtx(e *Exec, ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return e.ctx, func() {}
	}

	runCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(e.ctx, cancel)
	return runCtx, func() {
		stop()
		cancel()
	}
}
