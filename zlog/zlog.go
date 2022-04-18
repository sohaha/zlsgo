// Package zlog provide daily log service
package zlog

import (
	"fmt"
	"os"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	Log    = NewZLog(os.Stderr, "", BitDefault, LogDump, true, 3)
	osExit = func(code int) {
		if zutil.IsDoubleClickStartUp() {
			_, _ = fmt.Scanln()
		}
		os.Exit(code)
	}
)

// GetFlags Get the tag bits
func GetFlags() int {
	return Log.GetFlags()
}

// DisableConsoleColor DisableConsoleColor
func DisableConsoleColor() {
	Log.DisableConsoleColor()
}

// ForceConsoleColor ForceConsoleColor
func ForceConsoleColor() {
	Log.ForceConsoleColor()
}

// ResetFlags Setting Markup Bits
func ResetFlags(flag int) {
	Log.ResetFlags(flag)
}

// AddFlag Set flag Tags
func AddFlag(flag int) {
	Log.AddFlag(flag)
}

// SetPrefix Setting log header prefix
func SetPrefix(prefix string) {
	Log.SetPrefix(prefix)
}

// SetFile Setting up log files
func SetFile(filepath string, archive ...bool) {
	Log.SetFile(filepath, archive...)
}

// SetSaveFile SetSaveFile
func SetSaveFile(filepath string, archive ...bool) {
	Log.SetSaveFile(filepath, archive...)
}

// SetLogLevel Setting log display level
func SetLogLevel(level int) {
	Log.SetLogLevel(level)
}

// GetLogLevel Setting log display level
func GetLogLevel() int {
	return Log.level
}

// Debugf Debugf
func Debugf(format string, v ...interface{}) {
	Log.Debugf(format, v...)
}

// Debug Debug
func Debug(v ...interface{}) {
	Log.Debug(v...)
}

// Dump Dump
func Dump(v ...interface{}) {
	if Log.level < LogDump {
		return
	}
	args := formatArgs(v...)
	_, file, line, ok := callerName(1)
	if ok {
		names, err := argNames(file, line)
		if err == nil {
			args = prependArgName(names, args)
		}
	}
	_ = Log.outPut(LogDump, fmt.Sprintln(args...), true, func() func() {
		Log.calldDepth--
		return func() {
			Log.calldDepth++
		}
	})
}

// Successf Successf
func Successf(format string, v ...interface{}) {
	Log.Successf(format, v...)
}

// Success Success
func Success(v ...interface{}) {
	Log.Success(v...)
}

// Infof Infof
func Infof(format string, v ...interface{}) {
	Log.Infof(format, v...)
}

// Info Info
func Info(v ...interface{}) {
	Log.Info(v...)
}

// Tipsf Tipsf
func Tipsf(format string, v ...interface{}) {
	Log.Tipsf(format, v...)
}

// Tips Tips
func Tips(v ...interface{}) {
	Log.Tips(v...)
}

// Warnf Warnf
func Warnf(format string, v ...interface{}) {
	Log.Warnf(format, v...)
}

// Warn Warn
func Warn(v ...interface{}) {
	Log.Warn(v...)
}

// Errorf Errorf
func Errorf(format string, v ...interface{}) {
	Log.Errorf(format, v...)
}

// Error Error
func Error(v ...interface{}) {
	Log.Error(v...)
}

// Printf Printf
func Printf(format string, v ...interface{}) {
	Log.Printf(format, v...)
}

// Println Println
func Println(v ...interface{}) {
	Log.Println(v...)
}

// Fatalf Fatalf
func Fatalf(format string, v ...interface{}) {
	Log.Fatalf(format, v...)
}

// Fatal Fatal
func Fatal(v ...interface{}) {
	Log.Fatal(v...)
}

// Panicf Panicf
func Panicf(format string, v ...interface{}) {
	Log.Panicf(format, v...)
}

// Panic panic
func Panic(v ...interface{}) {
	Log.Panic(v...)
}

// Track Track
func Track(v string, i ...int) {
	Log.Track(v, i...)
}

// Stack Stack
func Stack(v interface{}) {
	Log.Stack(v)
}

// Discard Discard
func Discard() {
	Log.Discard()
}

func init() {
	Log.calldDepth = 3
}
