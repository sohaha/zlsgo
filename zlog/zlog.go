package zlog

import (
	"io"
	"os"
)

var (
	Log    = NewZLog(os.Stderr, "", BitDefault, LogDump, true, 3)
	osExit = os.Exit
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

// SetLogFile Setting up log files
func SetLogFile(fileDir string, fileName string) {
	Log.SetLogFile(fileDir, fileName)
}

// SetSaveLogFile SetSaveLogFile
func SetSaveLogFile(fileDir string, fileName string) {
	Log.SetSaveLogFile(fileDir, fileName)
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
	Log.Dump(v...)
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

// Writer Writer
func Writer() io.Writer {
	return Log.Writer()
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
func Track(logTip string, v ...int) {
	Log.Track(logTip, v...)
}

// Stack Stack
func Stack(v ...interface{}) {
	Log.Stack(v...)
}

func init() {
	Log.calldDepth = 3
}
