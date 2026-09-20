package zlog

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/ztime"
)

func TestLogFile(T *testing.T) {
	dir := T.TempDir()
	isolateDefaultLogger(T)
	t := zlsgo.NewTest(T)
	ResetFlags(BitLevel | BitMicroSeconds)
	logPath := filepath.Join(dir, "log.log")
	SetSaveFile(logPath)
	Success("ok1")
	var ws sync.WaitGroup
	for i := range make([]uint8, 100) {
		ws.Add(1)
		go func(i int) {
			Info(i)
			ws.Done()
		}(i)
	}

	Success("ok2")
	ws.Wait()
	if err := log.file.Sync(); err != nil {
		T.Fatal(err)
	}

	t.Equal(true, zfile.FileExist(logPath))

	SetSaveFile(filepath.Join(dir, "ll.log"), true)
	Success("ok3")
	Error("err3")
	log.CloseFile()
	t.EqualTrue(zfile.DirExist(filepath.Join(dir, "ll")))
	Discard()
}

func TestSetSaveFile(t *testing.T) {
	dir := t.TempDir()
	log := New("TestSetSaveFile ")
	t.Cleanup(func() { CleanLog(log) })
	log.SetFile(filepath.Join(dir, "test.log"))
	log.Success("ok")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.SetFile(filepath.Join(dir, "test2.log"), true)
		for i := 0; i < 100; i++ {
			log.Success("ok2-" + strconv.Itoa(i))
		}
	}()
	wg.Wait()
	log.CloseFile()
	t.Log(zfile.FileSize(filepath.Join(dir, "test.log")))
	t.Log(zfile.FileSize(filepath.Join(dir, "test2", ztime.Now("Y-m-d")+".log")))
}

func TestLevelFile(t *testing.T) {
	dir := t.TempDir()
	tt := zlsgo.NewTest(t)
	log := New("LevelFile ")
	t.Cleanup(func() { CleanLog(log) })
	infoPath, errorPath := filepath.Join(dir, "info.log"), filepath.Join(dir, "error.log")
	log.SetLevelFile(LogInfo, infoPath)
	log.SetLevelFile(LogError, errorPath)
	log.Info("level-info-1")
	log.Error("level-error-1")
	log.CloseLevelFiles()

	tt.EqualTrue(zfile.FileExist(infoPath))
	tt.EqualTrue(zfile.FileExist(errorPath))

	infoContent, err := zfile.ReadFile(infoPath)
	tt.EqualNil(err)
	errorContent, err := zfile.ReadFile(errorPath)
	tt.EqualNil(err)
	tt.EqualTrue(strings.Contains(string(infoContent), "level-info-1"))
	tt.EqualTrue(strings.Contains(string(errorContent), "level-error-1"))
	tt.EqualTrue(!strings.Contains(string(infoContent), "level-error-1"))
}

func TestFileSwitchStdout(t *testing.T) {
	dir := t.TempDir()
	stdout, err := os.Create(filepath.Join(dir, "stdout"))
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = stdout
	t.Cleanup(func() { os.Stdout = old; stdout.Close() })
	l := NewZLog(io.Discard, "", 0, LogDump, false, 3)
	t.Cleanup(func() { CleanLog(l) })
	l.SetFileErrorHandler(func(err error) { t.Error(err) })
	for i := 0; i < 4; i++ {
		path := filepath.Join(dir, fmt.Sprintf("%d.log", i))
		if i%2 == 0 {
			l.SetSaveFile(path)
		} else {
			l.SetFile(path)
		}
		message := fmt.Sprintf("record-%d\n", i)
		if _, err := l.Out.Write([]byte(message)); err != nil {
			t.Fatal(err)
		}
	}
	l.CloseFile()
	var want string
	for i := 0; i < 4; i++ {
		message := fmt.Sprintf("record-%d\n", i)
		want += message
		data, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%d.log", i)))
		if err != nil || string(data) != message {
			t.Fatalf("file %d: %q, %v", i, data, err)
		}
	}
	data, err := os.ReadFile(stdout.Name())
	if err != nil || string(data) != want {
		t.Fatalf("stdout: %q, %v", data, err)
	}
}

func TestFileCloseErrors(t *testing.T) {
	for _, operation := range []string{"replace", "replace-level", "discard", "close", "close-level", "clean"} {
		t.Run(operation, func(t *testing.T) {
			dir := t.TempDir()
			l := NewZLog(io.Discard, "", 0, LogDump, false, 3)
			t.Cleanup(func() { CleanLog(l) })
			failure := errors.New("flush failed")
			failing := func(name string) *zfile.MemoryFile {
				f := zfile.NewMemoryFile(filepath.Join(dir, name), zfile.MemoryFileFlushBefore(func(*zfile.MemoryFile) error { return failure }))
				_, _ = f.Write([]byte("pending"))
				return f
			}
			var reported []error
			l.SetFileErrorHandler(func(err error) {
				// Reenter a lock-taking method: callbacks must run outside mu.
				l.SetFileErrorHandler(nil)
				reported = append(reported, err)
			})
			if operation == "replace-level" || operation == "close-level" {
				f := failing("level")
				l.levelFiles = map[int]*levelFile{LogInfo: {file: f, out: f}}
			} else {
				l.file = failing("main")
				l.Out = l.file
			}
			switch operation {
			case "replace":
				l.SetFile(filepath.Join(dir, "new.log"))
			case "replace-level":
				l.SetLevelFile(LogInfo, filepath.Join(dir, "new.log"))
			case "discard":
				f := failing("level")
				l.levelFiles = map[int]*levelFile{LogInfo: {file: f, out: f}}
				l.Discard()
			case "close":
				l.CloseFile()
			case "close-level":
				l.CloseLevelFiles()
			case "clean":
				CleanLog(l)
			}
			if len(reported) != 1 || !errors.Is(reported[0], failure) {
				t.Fatalf("errors: %v", reported)
			}
			if operation == "discard" {
				if !strings.Contains(reported[0].Error(), "main") || !strings.Contains(reported[0].Error(), "level") {
					t.Fatalf("missing error: %v", reported)
				}
				if l.Out != io.Discard || l.file != nil || l.levelFiles != nil || l.GetLogLevel() != LogNot {
					t.Fatal("discard left active state")
				}
			}
			if operation == "replace" || operation == "replace-level" {
				l.Info("new output")
				CleanLog(l)
				data, err := os.ReadFile(filepath.Join(dir, "new.log"))
				if err != nil || !strings.Contains(string(data), "new output") {
					t.Fatalf("replacement: %q, %v", data, err)
				}
			}
		})
	}
}
