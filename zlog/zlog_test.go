/*
 * @Author: seekwe
 * @Date:   2019-06-06 15:21:30
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-06 15:57:34
 */

package zlog

import (
	"os"
	"testing"

	zls "github.com/sohaha/zlsgo"
)

func TestLog(T *testing.T) {
	t := zls.NewTest(T)
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

	Track("log with Track")
	Stack("log with Stack")
	SetLogLevel(LogFatal)
	level := GetLogLevel()
	t.Equal(LogFatal, level)
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
	CleanLog(stdZLog)
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
