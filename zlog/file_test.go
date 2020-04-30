package zlog

import (
	"os"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
)

var ws sync.WaitGroup

func TestLogFile(T *testing.T) {
	t := zlsgo.NewTest(T)
	ResetFlags(BitLevel)
	_ = os.RemoveAll("tmp2/")
	SetSaveLogFile("tmp2/Log.log")
	Success("ok1")

	for i := range make([]uint8, 100) {
		ws.Add(1)
		go func(i int) {
			Info(i)
			ws.Done()
		}(i)
	}
	Success("ok2")
	ws.Wait()
	logPath := "./tmp2/Log.log"
	t.Equal(true, zfile.FileExist(logPath))
	SetSaveLogFile("tmp2/ll.log", true)
	t.Equal(true, zfile.DirExist("tmp2/ll"))
	if err := os.RemoveAll("./tmp2/"); err != nil {
		t.Log(err)
	}
	t.Equal(false, zfile.FileExist(logPath))
	Discard()
}
