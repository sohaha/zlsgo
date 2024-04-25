// Package zlog provide daily log service
package zlog

import (
	"fmt"
	"os"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	log    = NewZLog(os.Stderr, "", BitDefault, LogDump, true, 4)
	osExit = func(code int) {
		if zutil.IsDoubleClickStartUp() {
			_, _ = fmt.Scanln()
		}
		os.Exit(code)
	}
)

// GetFlags Get the tag bits
func GetFlags() int {
	return log.GetFlags()
}

// DisableConsoleColor DisableConsoleColor
func DisableConsoleColor() {
	log.DisableConsoleColor()
}

// ForceConsoleColor ForceConsoleColor
func ForceConsoleColor() {
	log.ForceConsoleColor()
}

// ResetFlags Setting Markup Bits
func ResetFlags(flag int) {
	log.ResetFlags(flag)
}

// AddFlag Set flag Tags
func AddFlag(flag int) {
	log.AddFlag(flag)
}

// SetPrefix Setting log header prefix
func SetPrefix(prefix string) {
	log.SetPrefix(prefix)
}

// SetFile Setting up log files
func SetFile(filepath string, archive ...bool) {
	log.SetFile(filepath, archive...)
}

// SetSaveFile SetSaveFile
func SetSaveFile(filepath string, archive ...bool) {
	log.SetSaveFile(filepath, archive...)
}

// SetLogLevel Setting log display level
func SetLogLevel(level int) {
	log.SetLogLevel(level)
}

// GetLogLevel Setting log display level
func GetLogLevel() int {
	return log.level
}

// Debugf Debugf
func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Debug Debug
func Debug(v ...interface{}) {
	log.Debug(v...)
}

// Dump Dump
func Dump(v ...interface{}) {
	if log.level < LogDump {
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
	_ = log.outPut(LogDump, fmt.Sprintln(args...), true, log.calldDepth-1)
}

// Successf Successf
func Successf(format string, v ...interface{}) {
	log.Successf(format, v...)
}

// Success Success
func Success(v ...interface{}) {
	log.Success(v...)
}

// Infof Infof
func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Info Info
func Info(v ...interface{}) {
	log.Info(v...)
}

// Tipsf Tipsf
func Tipsf(format string, v ...interface{}) {
	log.Tipsf(format, v...)
}

// Tips Tips
func Tips(v ...interface{}) {
	log.Tips(v...)
}

// Warnf Warnf
func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

// Warn Warn
func Warn(v ...interface{}) {
	log.Warn(v...)
}

// Errorf Errorf
func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Error Error
func Error(v ...interface{}) {
	log.Error(v...)
}

// Printf Printf
func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Println Println
func Println(v ...interface{}) {
	log.Println(v...)
}

// Fatalf Fatalf
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

// Fatal Fatal
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Panicf Panicf
func Panicf(format string, v ...interface{}) {
	log.Panicf(format, v...)
}

// Panic panic
func Panic(v ...interface{}) {
	log.Panic(v...)
}

// Track Track
func Track(v string, i ...int) {
	log.Track(v, i...)
}

// Stack Stack
func Stack(v interface{}) {
	log.Stack(v)
}

// Discard Discard
func Discard() {
	log.Discard()
}
