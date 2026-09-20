package zlog

import "fmt"

// Entry is a logger bound to a fixed set of structured Fields.
// It is created via Logger.WithField/WithFields/WithError and exposes the
// same level methods as Logger. Entries are immutable: chaining With* on an
// Entry returns a new Entry and never mutates the parent.
type Entry struct {
	logger *Logger
	fields Fields
}

// WithField returns an Entry with the given key-value pair attached.
// Every record written through the Entry carries this field.
func (log *Logger) WithField(key string, value interface{}) *Entry {
	return log.WithFields(Fields{key: value})
}

// WithFields returns an Entry carrying a copy of the given fields.
func (log *Logger) WithFields(fields Fields) *Entry {
	f := make(Fields, len(fields))
	for k, v := range fields {
		f[k] = v
	}
	return &Entry{logger: log, fields: f}
}

// WithError returns an Entry with the error attached under the "error" key.
func (log *Logger) WithError(err error) *Entry {
	return log.WithField("error", err)
}

// WithField returns a new Entry with the additional key-value pair.
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return e.WithFields(Fields{key: value})
}

// WithFields returns a new Entry merging the given fields into a copy of the
// entry's fields; the entry itself is not modified.
func (e *Entry) WithFields(fields Fields) *Entry {
	f := make(Fields, len(e.fields)+len(fields))
	for k, v := range e.fields {
		f[k] = v
	}
	for k, v := range fields {
		f[k] = v
	}
	return &Entry{logger: e.logger, fields: f}
}

// WithError returns a new Entry with the error attached under the "error" key.
func (e *Entry) WithError(err error) *Entry {
	return e.WithField("error", err)
}

// Logger returns the underlying Logger.
func (e *Entry) Logger() *Logger {
	return e.logger
}

// Data returns a copy of the fields bound to this Entry.
func (e *Entry) Data() Fields {
	f := make(Fields, len(e.fields))
	for k, v := range e.fields {
		f[k] = v
	}
	return f
}

// output writes s at the given level. It must be called directly by each
// public method so the caller depth stays exactly one frame deeper than the
// equivalent Logger method.
func (e *Entry) output(level int, s string, isWrap bool) {
	_ = e.logger.outPutFields(level, s, e.fields, isWrap, e.logger.calldDepth+1)
}

// Log writes a record at the given level. LogPanic panics and LogFatal exits,
// matching the semantics of the dedicated level methods.
func (e *Entry) Log(level int, v ...interface{}) {
	if e.logger.level.Load() < int32(level) {
		return
	}
	s := fmt.Sprintln(v...)
	e.output(level, s, true)
	if level == LogPanic {
		panic(s)
	}
	if level == LogFatal {
		osExit(1)
	}
}

// Logf writes a formatted record at the given level, with the same
// panic/exit semantics as Log.
func (e *Entry) Logf(level int, format string, v ...interface{}) {
	if e.logger.level.Load() < int32(level) {
		return
	}
	s := fmt.Sprintf(format, v...)
	e.output(level, s, true)
	if level == LogPanic {
		panic(s)
	}
	if level == LogFatal {
		osExit(1)
	}
}

// Debugf logs a formatted debug message.
func (e *Entry) Debugf(format string, v ...interface{}) {
	if e.logger.level.Load() < LogDebug {
		return
	}
	e.output(LogDebug, fmt.Sprintf(format, v...), true)
}

// Debug logs a debug message.
func (e *Entry) Debug(v ...interface{}) {
	if e.logger.level.Load() < LogDebug {
		return
	}
	e.output(LogDebug, fmt.Sprintln(v...), true)
}

// Infof logs a formatted informational message.
func (e *Entry) Infof(format string, v ...interface{}) {
	if e.logger.level.Load() < LogInfo {
		return
	}
	e.output(LogInfo, fmt.Sprintf(format, v...), true)
}

// Info logs an informational message.
func (e *Entry) Info(v ...interface{}) {
	if e.logger.level.Load() < LogInfo {
		return
	}
	e.output(LogInfo, fmt.Sprintln(v...), true)
}

// Tipsf logs a formatted tip message.
func (e *Entry) Tipsf(format string, v ...interface{}) {
	if e.logger.level.Load() < LogTips {
		return
	}
	e.output(LogTips, fmt.Sprintf(format, v...), true)
}

// Tips logs a tip message.
func (e *Entry) Tips(v ...interface{}) {
	if e.logger.level.Load() < LogTips {
		return
	}
	e.output(LogTips, fmt.Sprintln(v...), true)
}

// Successf logs a formatted success message.
func (e *Entry) Successf(format string, v ...interface{}) {
	if e.logger.level.Load() < LogSuccess {
		return
	}
	e.output(LogSuccess, fmt.Sprintf(format, v...), true)
}

// Success logs a success message.
func (e *Entry) Success(v ...interface{}) {
	if e.logger.level.Load() < LogSuccess {
		return
	}
	e.output(LogSuccess, fmt.Sprintln(v...), true)
}

// Warnf logs a formatted warning message.
func (e *Entry) Warnf(format string, v ...interface{}) {
	if e.logger.level.Load() < LogWarn {
		return
	}
	e.output(LogWarn, fmt.Sprintf(format, v...), true)
}

// Warn logs a warning message.
func (e *Entry) Warn(v ...interface{}) {
	if e.logger.level.Load() < LogWarn {
		return
	}
	e.output(LogWarn, fmt.Sprintln(v...), true)
}

// Errorf logs a formatted error message.
func (e *Entry) Errorf(format string, v ...interface{}) {
	if e.logger.level.Load() < LogError {
		return
	}
	e.output(LogError, fmt.Sprintf(format, v...), true)
}

// Error logs an error message.
func (e *Entry) Error(v ...interface{}) {
	if e.logger.level.Load() < LogError {
		return
	}
	e.output(LogError, fmt.Sprintln(v...), true)
}

// Printf logs a formatted message with no specific level.
func (e *Entry) Printf(format string, v ...interface{}) {
	e.output(LogNot, fmt.Sprintf(format, v...), false)
}

// Println logs a message with no specific level.
func (e *Entry) Println(v ...interface{}) {
	e.output(LogNot, fmt.Sprintln(v...), true)
}

// Fatalf logs a formatted fatal error message and terminates the program.
func (e *Entry) Fatalf(format string, v ...interface{}) {
	if e.logger.level.Load() < LogFatal {
		return
	}
	e.output(LogFatal, fmt.Sprintf(format, v...), true)
	osExit(1)
}

// Fatal logs a fatal error message and terminates the program.
func (e *Entry) Fatal(v ...interface{}) {
	if e.logger.level.Load() < LogFatal {
		return
	}
	e.output(LogFatal, fmt.Sprintln(v...), true)
	osExit(1)
}

// Panicf logs a formatted error message and panics.
func (e *Entry) Panicf(format string, v ...interface{}) {
	if e.logger.level.Load() < LogPanic {
		return
	}
	s := fmt.Sprintf(format, v...)
	e.output(LogPanic, s, true)
	panic(s)
}

// Panic logs an error message and panics.
func (e *Entry) Panic(v ...interface{}) {
	if e.logger.level.Load() < LogPanic {
		return
	}
	s := fmt.Sprintln(v...)
	e.output(LogPanic, s, true)
	panic(s)
}
