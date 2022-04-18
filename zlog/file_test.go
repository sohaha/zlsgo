package zlog

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/ztime"
)

func TestMain(m *testing.M) {
	m.Run()
	zfile.Rmdir("tmp2/")
}

func TestLogFile(T *testing.T) {
	t := zlsgo.NewTest(T)
	ResetFlags(BitLevel | BitMicroSeconds)
	defer zfile.Rmdir("tmp2/")
	logPath := "./tmp2/Log.log"
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
	time.Sleep(time.Second * 2)

	t.Equal(true, zfile.FileExist(logPath))

	SetSaveFile("tmp2/ll.log", true)
	Success("ok3")
	Error("err3")
	time.Sleep(time.Second * 2)
	t.EqualTrue(zfile.DirExist("tmp2/ll"))
	Discard()
}

func TestSetSaveFile(t *testing.T) {
	log := New("TestSetSaveFile ")
	log.SetFile("tmp2/test.log")
	defer zfile.Rmdir("tmp2/")
	log.Success("ok")
	go func() {
		log.SetFile("tmp2/test2.log", true)
		for i := 0; i < 100; i++ {
			log.Success("ok2-" + strconv.Itoa(i))
		}
	}()
	time.Sleep(time.Second * 2)
	t.Log(zfile.FileSize("tmp2/test.log"))
	t.Log(zfile.FileSize("tmp2/test2/" + ztime.Now("Y-m-d") + ".log"))
}
