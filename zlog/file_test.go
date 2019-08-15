package zlog

import (
	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"os"
	"sync"
	"testing"
)

var ws sync.WaitGroup

func TestLogFile(T *testing.T) {
	t := zlsgo.NewTest(T)
	ResetFlags(BitLevel)
	_ = os.RemoveAll("./tmp2/")
	SetSaveLogFile("tmp2", "Log.log")
	Log.FileMaxSize = 1
	Success("ok1")

	for i := range make([]uint8, 100) {
		ws.Add(1)
		go func(i int) {
			Info(i)
			t.Log(i)
			ws.Done()
		}(i)
	}
	Success("ok2")
	logPath := "./tmp2/Log.log"
	t.Equal(true, zfile.FileExist(logPath))
	ws.Wait()
	Log.CloseFile()
	if err := os.RemoveAll("./tmp2/"); err != nil {
		t.Log(err)
	}
	t.Equal(false, zfile.FileExist(logPath))
}
