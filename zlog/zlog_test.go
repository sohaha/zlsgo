package zlog

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestLogTrack(T *testing.T) {
	Track("log with Track")
	Stack("log with Stack")
}

func TestLogs(T *testing.T) {
	t := zlsgo.NewTest(T)
	text := "Text"

	log.SetIgnoreLog("test")
	SetLogLevel(LogDump)
	Debug("test")
	Debug("debug")
	Debug("log with Debug")
	Debugf("%s", "log with Debug")
	Info("log with Info")
	Infof("%s", "log with Info")
	Success("log with Success")
	Successf("%s", "log with Success")
	Tips("log with Tips")
	Tipsf("%s", "log with Tips")
	Warn("log with Warn")
	Warnf("%s", "log with Warn")
	Error("log with Error")
	Errorf("%s", "log with Error")
	Println("log with Println")
	Printf("%s\n", "log with Printf")
	Dump("log with Dump", t, T, nil)

	SetLogLevel(LogFatal)
	level := GetLogLevel()
	t.Equal(LogFatal, level)
	ResetFlags(BitLevel | BitShortFile | BitTime)
	flage := GetFlags()
	t.Equal(BitDefault, flage)
	DisableConsoleColor()
	GetFlags()
	ResetFlags(BitDate)
	AddFlag(BitLevel)
	SetPrefix(text)
	ForceConsoleColor()
	ColorBackgroundWrap(ColorBlack, ColorLightGreen, text)
	SetFile("tmp/Log.log")
	CleanLog(log)
	log := New(text)
	log.SetPrefix(text)
	t.EqualExit(log.GetPrefix(), text)
	log.GetLogLevel()
	log.SetSaveFile("tmp/Log.log")
	log.ColorBackgroundWrap(ColorBlack, ColorLightGreen, text)
	log.OpTextWrap(OpBold, text)
	log.Dump(struct {
		S struct {
			N *string
			n string
		}
		M map[string]interface{}
		N string
		I int
		U uint
		F float32
		B bool
	}{N: "test\nyes", M: map[string]interface{}{"s": 1243}, S: struct {
		N *string
		n string
	}{n: ""}})
	CleanLog(log)
	e := os.RemoveAll("tmp/")
	t.Log(e)
}

func TestLogFatal(T *testing.T) {
	ResetFlags(0)
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	Fatal("TestLogFatal")
	Fatalf("%s", "Fatal")
}

func TestLogPanic(T *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			T.Log(err)
		}
	}()
	Panic("log with Panicf")
}

func TestLogPanicf(T *testing.T) {
	t := zlsgo.NewTest(T)
	buf := bytes.NewBuffer(nil)
	oldOut := log.out
	oldLevel := log.level
	log.out = buf
	log.level = LogPanic
	defer func() {
		log.out = oldOut
		log.level = oldLevel
		if err := recover(); err != nil {
			T.Log(err)
			t.Equal("num=1", err)
			out := buf.String()
			T.Log(out)
			t.EqualExit(false, strings.Contains(out, "%!"))
		} else {
			T.Fatal("expected panic")
		}
	}()

	Panicf("num=%d", 1)
}
