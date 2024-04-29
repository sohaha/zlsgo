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

// SetDefault set default logger
func SetDefault(l *Logger) {
	log = NewZLog(l.out, l.prefix, l.flag, l.level, l.color, l.calldDepth+1)
}

// Deprecated: please use SetDefault
// GetFlags Get the tag bits
func GetFlags() int {
	return log.GetFlags()
}

// Deprecated: please use SetDefault
// DisableConsoleColor DisableConsoleColor
func DisableConsoleColor() {
	log.DisableConsoleColor()
}

// Deprecated: please use SetDefault
// ForceConsoleColor ForceConsoleColor
func ForceConsoleColor() {
	log.ForceConsoleColor()
}

// Deprecated: please use SetDefault
// ResetFlags Setting Markup Bits
func ResetFlags(flag int) {
	log.ResetFlags(flag)
}

// Deprecated: please use SetDefault
// AddFlag Set flag Tags
func AddFlag(flag int) {
	log.AddFlag(flag)
}

// Deprecated: please use SetDefault
// SetPrefix Setting log header prefix
func SetPrefix(prefix string) {
	log.SetPrefix(prefix)
}

// Deprecated: please use SetDefault
// SetFile Setting up log files
func SetFile(filepath string, archive ...bool) {
	log.SetFile(filepath, archive...)
}

// Deprecated: please use SetDefault
// SetSaveFile SetSaveFile
func SetSaveFile(filepath string, archive ...bool) {
	log.SetSaveFile(filepath, archive...)
}

// Deprecated: please use SetDefault
// SetLogLevel Setting log display level
func SetLogLevel(level int) {
	log.SetLogLevel(level)
}

// Deprecated: please use SetDefault
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
