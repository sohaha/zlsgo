// Package zlog provides a flexible logging service with support for different log levels,
// colored output, file logging, and debugging utilities.
package zlog

import (
	"fmt"
	"os"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	log    = NewZLog(os.Stdout, "", BitDefault, LogDump, true, 4)
	osExit = func(code int) {
		if zutil.IsDoubleClickStartUp() {
			_, _ = fmt.Scanln()
		}
		os.Exit(code)
	}
)

// SetDefault sets the default logger instance used by the package-level logging functions.
// This allows customizing the global logger behavior.
func SetDefault(l *Logger) {
	log.Writer().Reset(l)
}

// GetFlags returns the current flag bits controlling the format of log output.
// Deprecated: please use SetDefault and access the logger's methods directly.
func GetFlags() int {
	return log.GetFlags()
}

// DisableConsoleColor turns off colored output in the console.
// Deprecated: please use SetDefault and access the logger's methods directly.
func DisableConsoleColor() {
	log.DisableConsoleColor()
}

// ForceConsoleColor forces colored output in the console even when output is not a terminal.
// Deprecated: please use SetDefault and access the logger's methods directly.
func ForceConsoleColor() {
	log.ForceConsoleColor()
}

// ResetFlags sets the output flags to control the formatting of log messages.
// Deprecated: please use SetDefault and access the logger's methods directly.
func ResetFlags(flag int) {
	log.ResetFlags(flag)
}

// AddFlag adds the specified flag to the current set of output flags.
// Deprecated: please use SetDefault and access the logger's methods directly.
func AddFlag(flag int) {
	log.AddFlag(flag)
}

// SetPrefix sets the prefix for each log line output.
// Deprecated: please use SetDefault and access the logger's methods directly.
func SetPrefix(prefix string) {
	log.SetPrefix(prefix)
}

// SetFile configures logging to a file at the specified path.
// The archive parameter controls whether to archive old logs.
// Deprecated: please use SetDefault and access the logger's methods directly.
func SetFile(filepath string, archive ...bool) {
	log.SetFile(filepath, archive...)
}

// SetSaveFile configures logging to a file at the specified path.
// The archive parameter controls whether to archive old logs.
// Deprecated: please use SetDefault and access the logger's methods directly.
func SetSaveFile(filepath string, archive ...bool) {
	log.SetSaveFile(filepath, archive...)
}

// SetLogLevel sets the minimum level of messages that will be logged.
// Deprecated: please use SetDefault and access the logger's methods directly.
func SetLogLevel(level int) {
	log.SetLogLevel(level)
}

// GetLogLevel returns the current minimum log level.
// Deprecated: please use SetDefault and access the logger's methods directly.
func GetLogLevel() int {
	return log.level
}

// Debugf logs a formatted debug message.
func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Debug logs a debug message.
func Debug(v ...interface{}) {
	log.Debug(v...)
}

// Dump logs detailed information about variables, including their names when possible.
// This is useful for debugging complex data structures.
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

// Successf logs a formatted success message.
func Successf(format string, v ...interface{}) {
	log.Successf(format, v...)
}

// Success logs a success message.
func Success(v ...interface{}) {
	log.Success(v...)
}

// Infof logs a formatted informational message.
func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Info logs an informational message.
func Info(v ...interface{}) {
	log.Info(v...)
}

// Tipsf logs a formatted tip message.
func Tipsf(format string, v ...interface{}) {
	log.Tipsf(format, v...)
}

// Tips logs a tip message.
func Tips(v ...interface{}) {
	log.Tips(v...)
}

// Warnf logs a formatted warning message.
func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

// Warn logs a warning message.
func Warn(v ...interface{}) {
	log.Warn(v...)
}

// Errorf logs a formatted error message.
func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Error logs an error message.
func Error(v ...interface{}) {
	log.Error(v...)
}

// Printf logs a formatted message with no specific level.
func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Println logs a message with no specific level.
func Println(v ...interface{}) {
	log.Println(v...)
}

// Fatalf logs a formatted fatal error message and terminates the program.
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

// Fatal logs a fatal error message and terminates the program.
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Panicf logs a formatted error message and panics.
func Panicf(format string, v ...interface{}) {
	log.Panicf(format, v...)
}

// Panic logs an error message and panics.
func Panic(v ...interface{}) {
	log.Panic(v...)
}

// Track logs the current function call stack with the given message.
// The optional integer parameter controls how many levels of the stack to skip.
func Track(v string, i ...int) {
	log.Track(v, i...)
}

// Stack logs the full stack trace along with the given value.
// This is useful for debugging complex call paths and understanding execution flow.
func Stack(v interface{}) {
	log.Stack(v)
}

// Discard sets the default logger to discard all output.
// This is useful for silencing logs in tests or when logs are not needed.
func Discard() {
	log.Discard()
}
