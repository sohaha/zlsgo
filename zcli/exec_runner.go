package zcli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// execRunSpec holds the process settings resolved from ExecOptions.
type execRunSpec struct {
	Env        []string
	Dir        string
	CloseStdin bool
	WaitDelay  time.Duration
}

// apply copies the spec onto a command. A nil Env keeps the inherited
// environment; otherwise the entries are appended so the child keeps PATH
// while the override wins for duplicated keys. WaitDelay bounds pipe draining
// after context cancellation: without it, orphaned grandchildren that inherit
// stdout keep cmd.Wait blocked for as long as they live.
func (s execRunSpec) apply(cmd *exec.Cmd) {
	if s.Dir != "" {
		cmd.Dir = s.Dir
	}
	if s.Env != nil {
		cmd.Env = append(os.Environ(), s.Env...)
	}
	if s.WaitDelay > 0 {
		cmd.WaitDelay = s.WaitDelay
	} else if s.WaitDelay < 0 {
		cmd.WaitDelay = 0
	}
}

func newExecCommand(ctx context.Context, argv []string, spec execRunSpec) (*exec.Cmd, error) {
	if len(argv) == 0 || argv[0] == "" {
		return nil, errors.New("no command")
	}

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	spec.apply(cmd)
	return cmd, nil
}

func exitCode(cmd *exec.Cmd, runErr error) int {
	if cmd.ProcessState == nil {
		if runErr != nil {
			return -1
		}
		return 0
	}
	return cmd.ProcessState.ExitCode()
}

// runExecCommand executes argv directly, without any shell-style splitting or quoting.
func runExecCommand(ctx context.Context, argv []string, spec execRunSpec) (code int, outStr string, errStr string, err error) {
	cmd, err := newExecCommand(ctx, argv, spec)
	if err != nil {
		return -1, "", "", err
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return exitCode(cmd, err), stdout.String(), stderr.String(), err
}

// callbackRunExecCommand executes argv directly and streams output line by line.
// The returned writer feeds the process stdin.
func callbackRunExecCommand(ctx context.Context, argv []string, callback func(str string, isStdout bool), spec execRunSpec) (<-chan int, func(string), error) {
	cmd, err := newExecCommand(ctx, argv, spec)
	if err != nil {
		return nil, nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	var (
		stdinMu sync.Mutex
		cbMu    sync.Mutex
	)
	stdinClosed := false

	input := func(s string) {
		stdinMu.Lock()
		defer stdinMu.Unlock()
		if stdinClosed {
			return
		}
		_, _ = io.WriteString(stdin, s)
	}

	if spec.CloseStdin {
		stdinMu.Lock()
		stdinClosed = true
		_ = stdin.Close()
		stdinMu.Unlock()
	}

	emit := func(str string, isStdout bool) {
		// The two pumps run concurrently; serialise user callbacks so callers
		// can append to shared state without extra locking.
		cbMu.Lock()
		defer cbMu.Unlock()
		callback(str, isStdout)
	}

	scan := func(r io.Reader, isStdout bool) {
		scanner := bufio.NewScanner(r)
		buf := execScanBufPool.Get().(*[]byte)
		defer execScanBufPool.Put(buf)
		scanner.Buffer((*buf)[:0], len(*buf))
		for scanner.Scan() {
			emit(execTrimLine(scanner.Bytes()), isStdout)
		}
	}

	codeCh := make(chan int, 1)
	go func() {
		var pumps sync.WaitGroup
		pumps.Add(2)
		go func() {
			defer pumps.Done()
			scan(stdout, true)
		}()
		go func() {
			defer pumps.Done()
			scan(stderr, false)
		}()

		var waitErr error
		if spec.WaitDelay > 0 {
			// WaitDelay makes Wait close the pipes once the delay expires
			// after the process exits, which unblocks pumps stuck on an
			// orphaned grandchild holding the descriptors. Reap the process
			// while draining; sequencing Wait after the pumps would let that
			// grandchild block us for its whole lifetime.
			waitErr = cmd.Wait()
			pumps.Wait()
		} else {
			// Without WaitDelay the StdoutPipe contract requires waiting for
			// both pipes to drain before reaping the process.
			pumps.Wait()
			waitErr = cmd.Wait()
		}

		stdinMu.Lock()
		stdinClosed = true
		_ = stdin.Close()
		stdinMu.Unlock()

		codeCh <- exitCode(cmd, waitErr)
		close(codeCh)
	}()

	return codeCh, input, nil
}

// execScanBufPool lends 1 MiB scan buffers to the two concurrent output pumps.
// Each Scanner owns its buffer for the lifetime of the scan, and the callback
// receives a copy of the line, so reuse is safe without allocating per process.
var execScanBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 1024*1024)
		return &buf
	},
}

// execTrimLine returns a copy of line without the trailing newline, so callers
// keep valid data after the shared scan buffer is reused.
func execTrimLine(line []byte) string {
	line = bytes.TrimSuffix(line, []byte("\r"))
	return string(line)
}
