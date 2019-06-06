package zlog

import "os"

var stdZLog = NewZLog(os.Stderr, "", BitDefault, 6, true, 3)

// GetFlags Get the tag bits
func GetFlags() int {
	return stdZLog.GetFlags()
}

// DisableConsoleColor DisableConsoleColor
func DisableConsoleColor() {
	stdZLog.DisableConsoleColor()
}

// ForceConsoleColor ForceConsoleColor
func ForceConsoleColor() {
	stdZLog.ForceConsoleColor()
}

// ResetFlags Setting Markup Bits
func ResetFlags(flag int) {
	stdZLog.ResetFlags(flag)
}

// AddFlag Set flag Tags
func AddFlag(flag int) {
	stdZLog.AddFlag(flag)
}

// SetPrefix Setting log header prefix
func SetPrefix(prefix string) {
	stdZLog.SetPrefix(prefix)
}

// SetLogFile Setting up log files
func SetLogFile(fileDir string, fileName string) {
	stdZLog.SetLogFile(fileDir, fileName)
}

// SetLogLevel Setting log display level
func SetLogLevel(level int) {
	stdZLog.SetLogLevel(level)
}

// GetLogLevel Setting log display level
func GetLogLevel() int {
	return stdZLog.level
}

// Debugf Debugf
func Debugf(format string, v ...interface{}) {
	stdZLog.Debugf(format, v...)
}

// Debug Debug
func Debug(v ...interface{}) {
	stdZLog.Debug(v...)
}

// Successf Successf
func Successf(format string, v ...interface{}) {
	stdZLog.Successf(format, v...)
}

// Success Success
func Success(v ...interface{}) {
	stdZLog.Success(v...)
}

// Track Track
func Track(logTip string, v ...int) {
	stdZLog.Track(logTip, v...)
}

// Infof Infof
func Infof(format string, v ...interface{}) {
	stdZLog.Infof(format, v...)
}

// Info Info
func Info(v ...interface{}) {
	stdZLog.Info(v...)
}

// Warnf Warnf
func Warnf(format string, v ...interface{}) {
	stdZLog.Warnf(format, v...)
}

// Warn Warn
func Warn(v ...interface{}) {
	stdZLog.Warn(v...)
}

// Errorf Errorf
func Errorf(format string, v ...interface{}) {
	stdZLog.Errorf(format, v...)
}

// Error Error
func Error(v ...interface{}) {
	stdZLog.Error(v...)
}

// Printf Printf
func Printf(format string, v ...interface{}) {
	stdZLog.Printf(format, v...)
}

// Println Println
func Println(v ...interface{}) {
	stdZLog.Println(v...)
}

// Fatalf Fatalf
func Fatalf(format string, v ...interface{}) {
	stdZLog.Fatalf(format, v...)
}

// Fatal Fatal
func Fatal(v ...interface{}) {
	stdZLog.Fatal(v...)
}

// Panicf Panicf
func Panicf(format string, v ...interface{}) {
	stdZLog.Panicf(format, v...)
}

// panic panic
func Panic(v ...interface{}) {
	stdZLog.Panic(v...)
}

// Stack Stack
func Stack(v ...interface{}) {
	stdZLog.Stack(v...)
}

func init() {
	stdZLog.calldDepth = 3
}
