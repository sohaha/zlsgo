package zcli_test

import (
	"context"
	"embed"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo/zcli"
)

func TestExampleNewExec_localBinary(t *testing.T) {
	binFile, cleanup := writeExecExampleScript()
	defer cleanup()

	engine, err := zcli.NewExec(embed.FS{}, func(options *zcli.ExecOptions) {
		options.BinFile = binFile
	})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Release()

	code, out, _, err := engine.Run(context.Background(), []string{"example"})
	if err != nil {
		t.Fatal(err)
	}

	if code != 0 {
		t.Fatalf("Run() code = %d", code)
	}
	if got := strings.TrimSpace(out); got != "local-example" {
		t.Fatalf("Run() out = %q, want local-example", got)
	}
}

func writeExecExampleScript() (string, func()) {
	dir, err := os.MkdirTemp("", "exec-example-*")
	if err != nil {
		panic(err)
	}

	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, "example.cmd")
		err = os.WriteFile(path, []byte("@echo off\r\necho local-example\r\n"), 0o755)
	} else {
		path = filepath.Join(dir, "example.sh")
		err = os.WriteFile(path, []byte("#!/bin/sh\nprintf 'local-example\\n'\n"), 0o755)
	}
	if err != nil {
		_ = os.RemoveAll(dir)
		panic(err)
	}

	return path, func() {
		_ = os.RemoveAll(dir)
	}
}
