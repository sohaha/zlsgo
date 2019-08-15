package zlog

import (
	"errors"
	"github.com/sohaha/zlsgo"
	"os"
	"testing"
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
func TestTryError(T *testing.T) {
	testTryErrorErr(T)
	testTryErrorString(T)
	testTryErrorXXX(T)
}

func testTryErrorString(T *testing.T) {
	T.Log("testTryErrorString")
	defer TryError()
	Panic("testTryErrorString")
}

func testTryErrorErr(T *testing.T) {
	T.Log("testTryErrorErr")
	defer TryError(func(err error) {
		T.Log("testTryErrorErr", err)
	})
	Panic(errors.New("testTryErrorErr"))
}

func testTryErrorXXX(T *testing.T) {
	defer TryError(func(err error) {
		T.Log("testTryErrorXXX", err)
	})
	Panic(11)
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
