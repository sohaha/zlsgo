package zlog

import (
	"os"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestLogTrack(T *testing.T) {
	Track("log with Track")
	Stack("log with Stack")
}

func TestLog(T *testing.T) {
	t := zlsgo.NewTest(T)
	text := "Text"
	Debug("log with Debug")
	Debugf("%s", "log with Debug")
	Info("log with Info")
	Infof("%s", "log with Info")
	Success("log with Success")
	Successf("%s", "log with Success")
	Warn("log with Warn")
	Warnf("%s", "log with Warn")
	Error("log with Error")
	Errorf("%s", "log with Error")
	Println("log with Println")
	Printf("%s", "log with Printf")
	Dump("log with Dump", t, T,nil)

	SetLogLevel(LogFatal)
	level := GetLogLevel()
	t.Equal(LogFatal, level)
	ResetFlags(BitLevel | BitShortFile | BitStdFlag)
	flage := GetFlags()
	t.Equal(BitDefault, flage)
	DisableConsoleColor()
	GetFlags()
	ResetFlags(BitDate)
	AddFlag(BitLevel)
	SetPrefix(text)
	ForceConsoleColor()
	ColorBackgroundWrap(ColorBlack, ColorLightGreen, text)
	SetLogFile("tmp", "Log.log")
	CleanLog(Log)
	log := New(text)
	log.SetPrefix(text)
	log.GetLogLevel()
	log.SetSaveLogFile("tmp", "Log.log")
	log.ColorBackgroundWrap(ColorBlack, ColorLightGreen, text)
	log.OpTextWrap(OpBold, text)
	CleanLog(log)
	e := os.RemoveAll("tmp/")
	t.Log(e)
	Writer()
}

func TestLogFatal(T *testing.T) {
	ResetFlags(0)
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	Fatal("TestLogFatal")
	Fatalf("%s\n", "Fatal")
}

func TestLogPanic(T *testing.T) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	Panic("log with Panicf")
}

func TestLogPanicf(T *testing.T) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	Panicf("%s", "log with Panicf")
}
