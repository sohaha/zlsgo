package zlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"sync"
	"text/tabwriter"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/zutil"
)

// Log header information flag bits, using bitmap mode.
// These constants control what information appears in log message headers.
const (
	// BitDate includes the date in the log header (format: 2019/01/23)
	BitDate int = 1 << iota
	// BitTime includes the time in the log header (format: 01:23:12)
	BitTime
	// BitMicroSeconds includes microseconds in the time (format: 01:23:12.111222)
	BitMicroSeconds
	// BitLongFile includes the full file path in the log header
	// Example: /home/go/src/github.com/sohaha/zlsgo/doc.go
	BitLongFile
	// BitShortFile includes only the file name in the log header (e.g., doc.go)
	BitShortFile
	// BitLevel includes the log level in the log header (e.g., [INFO])
	BitLevel
	// BitStdFlag is the standard header format (date and time)
	BitStdFlag = BitDate | BitTime
	// BitDefault is the default header format (level, short file name, and time)
	BitDefault = BitLevel | BitShortFile | BitTime
	// LogMaxBuf defines the maximum buffer size for log messages in bytes
	LogMaxBuf = 1024 * 1024
)

// Log level constants define the severity levels for log messages.
// Higher values represent less severe levels.
const (
	// LogFatal is for fatal errors that cause the program to exit
	LogFatal = iota
	// LogPanic is for errors that cause a panic
	LogPanic
	// LogTrack is for stack trace information
	LogTrack
	// LogError is for error messages
	LogError
	// LogWarn is for warning messages
	LogWarn
	// LogTips is for tip/hint messages
	LogTips
	// LogSuccess is for success messages
	LogSuccess
	// LogInfo is for informational messages
	LogInfo
	// LogDebug is for debug messages
	LogDebug
	// LogDump is for detailed variable dumps
	LogDump
	// LogNot indicates no logging should occur
	LogNot = -1
)

var Levels = []string{
	"[FATAL]",
	"[PANIC]",
	"[TRACK]",
	"[ERROR]",
	"[WARN] ",
	"[TIPS] ",
	"[SUCCE]",
	"[INFO] ",
	"[DEBUG]",
	"[DUMP] ",
}

var LevelColous = []Color{
	ColorRed,
	ColorLightRed,
	ColorLightYellow,
	ColorRed,
	ColorYellow,
	ColorWhite,
	ColorGreen,
	ColorBlue,
	ColorLightCyan,
	ColorCyan,
}

type (
	// Logger represents a logging object with configurable output destination,
	// formatting options, and log level filtering.
	Logger struct {
		// out is the destination for log output (e.g., os.Stdout)
		out io.Writer
		// file is the memory buffer for file-based logging
		file *zfile.MemoryFile
		// prefix is prepended to each log message
		prefix string
		// fileDir is the directory where log files are stored
		fileDir string
		// fileName is the base name of the log file
		fileName string
		// writeBefore contains functions that are called before writing a log message
		// and can prevent the message from being logged by returning false
		writeBefore []func(level int, log string) bool
		// calldDepth controls how many stack frames to ascend to identify the calling function
		calldDepth int
		// level is the current minimum log level that will be output
		level int
		// flag contains the bitmap of header format options
		flag int
		// mu provides thread safety for the logger
		mu sync.RWMutex
		// color determines whether ANSI color codes are used in output
		color bool
		// fileAndStdout indicates whether to log to both file and standard output
		fileAndStdout bool
	}
	// formatter is an internal type used for formatting values during pretty printing
	formatter struct {
		v     reflect.Value
		force bool
		quote bool
	}
	// visit is an internal type used to track visited objects during recursive pretty printing
	// to prevent infinite loops on circular references
	visit struct {
		typ reflect.Type
		v   uintptr
	}
	// zprinter is an internal type that implements pretty printing of complex data structures
	zprinter struct {
		io.Writer
		tw      *tabwriter.Writer
		visited map[visit]int
		depth   int
	}
)

// New creates a new logger with the given module name.
// The module name is used as a prefix for log messages.
// If no module name is provided, an empty prefix is used.
func New(moduleName ...string) *Logger {
	name := ""
	if len(moduleName) > 0 {
		name = moduleName[0]
	}
	return NewZLog(os.Stdout, name, BitDefault, LogDump, true, 3)
}

// NewZLog creates a new logger with detailed configuration options.
// Parameters:
//   - out: the output destination for log messages
//   - prefix: a prefix for all log messages
//   - flag: bitmap of header format options (see Bit* constants)
//   - level: minimum log level to output
//   - color: whether to use ANSI color codes
//   - calldDepth: how many stack frames to ascend to identify the calling function
func NewZLog(out io.Writer, prefix string, flag int, level int, color bool, calldDepth int) *Logger {
	zlog := &Logger{out: out, prefix: prefix, flag: flag, file: nil, calldDepth: calldDepth, level: level, color: color}
	runtime.SetFinalizer(zlog, CleanLog)
	return zlog
}

// CleanLog performs cleanup operations on a logger, such as closing any open log files.
// This is typically called by the garbage collector when a logger is no longer referenced.
func CleanLog(log *Logger) {
	log.CloseFile()
}

// DisableConsoleColor turns off colored output for this logger instance.
func (log *Logger) DisableConsoleColor() {
	log.color = false
}

// ForceConsoleColor enables colored output for this logger instance,
// even when the output destination is not a terminal.
func (log *Logger) ForceConsoleColor() {
	log.color = true
}

// ColorTextWrap wraps the given text with ANSI color codes for the specified color.
// If colors are disabled for this logger, the original text is returned unchanged.
func (log *Logger) ColorTextWrap(color Color, text string) string {
	if log.color {
		return ColorTextWrap(color, text)
	}
	return text
}

// ColorBackgroundWrap wraps the given text with ANSI color codes for the specified
// text color and background color. If colors are disabled for this logger,
// the original text is returned unchanged.
func (log *Logger) ColorBackgroundWrap(color Color, backgroundColor Color, text string) string {
	if log.color {
		return ColorBackgroundWrap(color, backgroundColor, text)
	}
	return text
}

// OpTextWrap wraps the given text with ANSI codes for the specified text operation
// (like bold, underline, etc.). If colors are disabled for this logger,
// the original text is returned unchanged.
func (log *Logger) OpTextWrap(color Op, text string) string {
	if log.color {
		return OpTextWrap(color, text)
	}
	return text
}

func (log *Logger) formatHeader(buf *bytes.Buffer, file string, line int, level int) {
	if log.flag == 0 {
		return
	}

	t := ztime.Time(log.flag&BitMicroSeconds != 0)

	flags := log.flag

	if flags&BitDate != 0 {
		formatDateAppend(buf, t)
	}

	if flags&(BitTime|BitMicroSeconds) != 0 {
		formatTimeAppend(buf, t)

		if flags&BitMicroSeconds != 0 {
			buf.WriteByte('.')
			itoa(buf, t.Nanosecond()/1e3, 6)
		}
		buf.WriteByte(' ')
	}

	if flags&BitLevel != 0 {
		levelText := Levels[level]
		if log.color {
			buf.WriteString(log.ColorTextWrap(LevelColous[level], levelText))
			buf.WriteByte(' ')
		} else {
			buf.WriteString(levelText)
			buf.WriteByte(' ')
		}
	}

	if flags&(BitShortFile|BitLongFile) != 0 {
		if flags&BitShortFile != 0 {
			lastSlash := -1
			for i := len(file) - 1; i >= 0; i-- {
				if file[i] == '/' {
					lastSlash = i
					break
				}
			}

			if lastSlash >= 0 {
				file = file[lastSlash+1:]
			}
		}

		buf.WriteString(file)
		buf.WriteByte(':')
		itoa(buf, line, -1)
		buf.WriteString(": ")
	}
}

func (log *Logger) outPut(level int, s string, isWrap bool, calldDepth int, prefixText ...string) error {
	if log.writeBefore != nil && len(s) > 0 {
		p := s
		if isWrap && len(p) > 0 && p[len(p)-1] == '\n' {
			p = p[:len(p)-1]
		}
		for i := range log.writeBefore {
			if log.writeBefore[i](level, p) {
				return nil
			}
		}
	}

	buf := zutil.GetBuff(uint(len(s) + 34))
	defer zutil.PutBuff(buf)

	if level != LogNot {
		file, line := log.fileLocation(calldDepth)
		log.formatHeader(buf, file, line, level)
	}

	if log.prefix != "" {
		buf.WriteString(log.prefix)
	}

	if len(prefixText) > 0 {
		buf.WriteString(prefixText[0])
	}

	buf.WriteString(s)

	if isWrap && len(s) > 0 && s[len(s)-1] != '\n' {
		buf.WriteByte('\n')
	}

	_, err := log.out.Write(buf.Bytes())
	return err
}

// Printf formats according to a format specifier and writes to the log output.
// This logs a message with no specific level indicator.
func (log *Logger) Printf(format string, v ...interface{}) {
	_ = log.outPut(LogNot, fmt.Sprintf(format, v...), false, log.calldDepth)
}

// Println writes the arguments to the log output followed by a newline.
// This logs a message with no specific level indicator.
func (log *Logger) Println(v ...interface{}) {
	_ = log.outPut(LogNot, fmt.Sprintln(v...), true, log.calldDepth)
}

// Debugf logs a formatted debug message if the current log level permits debug output.
func (log *Logger) Debugf(format string, v ...interface{}) {
	if log.level < LogDebug {
		return
	}
	_ = log.outPut(LogDebug, fmt.Sprintf(format, v...), false, log.calldDepth)
}

// Debug logs a debug message if the current log level permits debug output.
func (log *Logger) Debug(v ...interface{}) {
	if log.level < LogDebug {
		return
	}
	_ = log.outPut(LogDebug, fmt.Sprintln(v...), true, log.calldDepth)
}

// Dump logs detailed information about variables in a pretty-printed format.
// It attempts to include variable names when possible, making it useful for debugging.
func (log *Logger) Dump(v ...interface{}) {
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

	_ = log.outPut(LogDump, fmt.Sprintln(args...), true, log.calldDepth)
}

// Successf logs a formatted success message if the current log level permits.
func (log *Logger) Successf(format string, v ...interface{}) {
	if log.level < LogSuccess {
		return
	}
	_ = log.outPut(LogSuccess, fmt.Sprintf(format, v...), false, log.calldDepth)
}

// Success logs a success message if the current log level permits.
func (log *Logger) Success(v ...interface{}) {
	if log.level < LogSuccess {
		return
	}
	_ = log.outPut(LogSuccess, fmt.Sprintln(v...), true, log.calldDepth)
}

// Infof logs a formatted informational message if the current log level permits.
func (log *Logger) Infof(format string, v ...interface{}) {
	if log.level < LogInfo {
		return
	}
	_ = log.outPut(LogInfo, fmt.Sprintf(format, v...), false, log.calldDepth)
}

// Info logs an informational message if the current log level permits.
func (log *Logger) Info(v ...interface{}) {
	if log.level < LogInfo {
		return
	}
	_ = log.outPut(LogInfo, fmt.Sprintln(v...), true, log.calldDepth)
}

// Tipsf logs a formatted tip message if the current log level permits.
func (log *Logger) Tipsf(format string, v ...interface{}) {
	if log.level < LogTips {
		return
	}
	_ = log.outPut(LogTips, fmt.Sprintf(format, v...), false, log.calldDepth)
}

// Tips logs a tip message if the current log level permits.
func (log *Logger) Tips(v ...interface{}) {
	if log.level < LogTips {
		return
	}
	_ = log.outPut(LogTips, fmt.Sprintln(v...), true, log.calldDepth)
}

// Warnf logs a formatted warning message if the current log level permits.
func (log *Logger) Warnf(format string, v ...interface{}) {
	if log.level < LogWarn {
		return
	}
	_ = log.outPut(LogWarn, fmt.Sprintf(format, v...), false, log.calldDepth)
}

// Warn logs a warning message if the current log level permits.
func (log *Logger) Warn(v ...interface{}) {
	if log.level < LogWarn {
		return
	}
	_ = log.outPut(LogWarn, fmt.Sprintln(v...), true, log.calldDepth)
}

// Errorf logs a formatted error message if the current log level permits.
func (log *Logger) Errorf(format string, v ...interface{}) {
	if log.level < LogError {
		return
	}
	_ = log.outPut(LogError, fmt.Sprintf(format, v...), false, log.calldDepth)
}

// Error logs an error message if the current log level permits.
func (log *Logger) Error(v ...interface{}) {
	if log.level < LogError {
		return
	}
	_ = log.outPut(LogError, fmt.Sprintln(v...), true, log.calldDepth)
}

// Fatalf logs a formatted fatal error message and terminates the program.
// Before terminating, it ensures all pending log messages are written.
func (log *Logger) Fatalf(format string, v ...interface{}) {
	if log.level < LogFatal {
		return
	}
	_ = log.outPut(LogFatal, fmt.Sprintf(format, v...), false, log.calldDepth)
	osExit(1)
}

// Fatal logs a fatal error message and terminates the program.
// Before terminating, it ensures all pending log messages are written.
func (log *Logger) Fatal(v ...interface{}) {
	if log.level < LogFatal {
		return
	}
	_ = log.outPut(LogFatal, fmt.Sprintln(v...), true, log.calldDepth)
	osExit(1)
}

// Panicf logs a formatted error message and then panics with the same message.
// This is useful for unrecoverable errors that require immediate termination with a stack trace.
func (log *Logger) Panicf(format string, v ...interface{}) {
	if log.level < LogPanic {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = log.outPut(LogPanic, fmt.Sprintf(format, s), false, log.calldDepth)
	panic(s)
}

// Panic logs an error message and then panics with the same message.
// This is useful for unrecoverable errors that require immediate termination with a stack trace.
func (log *Logger) Panic(v ...interface{}) {
	if log.level < LogPanic {
		return
	}
	s := fmt.Sprintln(v...)
	_ = log.outPut(LogPanic, s, true, log.calldDepth)
	panic(s)
}

// Stack logs a stack trace along with the provided value.
// This is useful for debugging to see the call path that led to a particular point in the code.
func (log *Logger) Stack(v interface{}) {
	if log.level < LogTrack {
		return
	}
	var s string
	switch e := v.(type) {
	case error:
		s = fmt.Sprintf("%+v", e)
	case string:
		s = e
	default:
		s = fmt.Sprintf("%v", e)
	}
	_ = log.outPut(LogTrack, s, true, log.calldDepth)
}

// Track logs the current function call stack with the given message.
// The optional integer parameter controls how many levels of the stack to skip.
// This is useful for tracing execution paths through the code.
func (log *Logger) Track(v string, i ...int) {
	if log.level < LogTrack {
		return
	}
	b, skip, max, index := zutil.GetBuff(), 4, 1, 1
	il := len(i)
	if il > 0 {
		max = i[0]
		if il == 2 {
			skip = skip + i[1]
		}
	}
	s := zutil.Callers(skip)
	l := len(s)
	if max >= l {
		max = l
	}
	s = s[:max]
	space := "  "
	b.WriteString(v + "\n")
	s.Format(func(fn *runtime.Func, file string, line int) bool {
		if index > 9 {
			space = " "
		}
		b.WriteString(fmt.Sprintf(
			"   %d).%s%s\n    \t%s:%d\n",
			index, space, fn.Name(), file, line,
		))
		index++
		return true
	})
	text := b.String()
	zutil.PutBuff(b)
	_ = log.outPut(LogTrack, text, true, log.calldDepth)
}

func callerName(skip int) (name, file string, line int, ok bool) {
	var pc uintptr
	if pc, file, line, ok = runtime.Caller(skip + 1); !ok {
		return
	}
	name = runtime.FuncForPC(pc).Name()
	return
}

// GetFlags returns the current flag bits controlling the format of log output.
// These flags determine what information (like date, time, file name) appears in log headers.
func (log *Logger) GetFlags() int {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.flag
}

// ResetFlags sets the output flags to the specified value, replacing any existing flags.
// This controls what information appears in the log message headers.
func (log *Logger) ResetFlags(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag = flag
}

// AddFlag adds the specified flag to the current set of output flags.
// This allows adding individual header elements without affecting existing ones.
func (log *Logger) AddFlag(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag |= flag
}

// SetPrefix sets the prefix for each log line output.
// The prefix appears before any other header information.
func (log *Logger) SetPrefix(prefix string) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.prefix = prefix
}

func (log *Logger) GetPrefix() string {
	return log.prefix
}

// SetLogLevel sets the minimum level of messages that will be logged.
// Messages with a level less than or equal to this value will be output;
// messages with a higher level will be ignored.
func (log *Logger) SetLogLevel(level int) {
	log.level = level
}

// GetLogLevel returns the current minimum log level.
// This indicates what severity of messages are currently being logged.
func (log *Logger) GetLogLevel() int {
	return log.level
}

func (log *Logger) Write(b []byte) (n int, err error) {
	_ = log.outPut(LogWarn, zstring.Bytes2String(b), false, log.calldDepth)
	return len(b), nil
}

func (log *Logger) SetIgnoreLog(logs ...string) {
	log.WriteBefore(func(level int, log string) bool {
		for _, v := range logs {
			if zstring.Match(log, v) {
				return true
			}
		}
		return false
	})
}

func (log *Logger) WriteBefore(fn ...func(level int, log string) bool) {
	log.writeBefore = append(log.writeBefore, fn...)
}

func itoa(buf *bytes.Buffer, i int, wid int) {
	u := uint(i)
	if u == 0 && wid <= 1 {
		buf.WriteByte('0')
		return
	}

	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}

	for bp < len(b) {
		buf.WriteByte(b[bp])
		bp++
	}
}

func (log *Logger) Writer() logWriter {
	return logWriter{log: log}
}

type logWriter struct {
	log *Logger
}

func (wr logWriter) Get() io.Writer {
	return wr.log.out
}

func (wr logWriter) Set(w io.Writer) {
	wr.log.out = w
}

func (wr logWriter) Reset(l *Logger) {
	wr.log.out = l.out
	wr.log.color = l.color
	wr.log.prefix = l.prefix
	wr.log.flag = l.flag
	wr.log.level = l.level
}

// formatArgs formats arguments and optimizes memory usage
func formatArgs(args ...interface{}) []interface{} {
	// Pre-allocate required capacity to avoid dynamic expansion
	formatted := make([]interface{}, 0, len(args))

	// Use temporary buffer to reduce string creation
	buf := zutil.GetBuff(uint(len(args)))
	defer zutil.PutBuff(buf)

	for _, a := range args {
		// Write directly to buffer
		buf.Reset()

		// Use colored formatting
		if a == nil {
			buf.WriteString(ColorTextWrap(ColorCyan, "<nil>"))
		} else {
			// Avoid unnecessary conversions
			switch v := a.(type) {
			case string:
				buf.WriteString(ColorTextWrap(ColorCyan, v))
			case []byte:
				buf.WriteString(ColorTextWrap(ColorCyan, string(v)))
			case error:
				buf.WriteString(ColorTextWrap(ColorCyan, v.Error()))
			default:
				// Use sprint for other types
				buf.WriteString(ColorTextWrap(ColorCyan, sprint(a)))
			}
		}

		// Add buffer content to result
		formatted = append(formatted, buf.String())
	}

	return formatted
}

func sprint(a ...interface{}) string {
	return fmt.Sprint(wrap(a, true)...)
}

func wrap(a []interface{}, force bool) []interface{} {
	w := make([]interface{}, len(a))
	for i, x := range a {
		w[i] = formatter{v: zreflect.ValueOf(x), force: force}
	}
	return w
}

func writeByte(w io.Writer, b byte) {
	_, _ = w.Write([]byte{b})
}

func prependArgName(names []string, values []interface{}) []interface{} {
	vLen := len(values)
	nLen := len(names)
	prepended := make([]interface{}, vLen)
	for i, value := range values {
		name := ""
		if i < nLen {
			name = names[i]
		}
		if name == "" {
			prepended[i] = OpTextWrap(OpBold, value.(string))
			continue
		}
		name = ColorTextWrap(ColorBlue, OpTextWrap(OpBold, name))
		prepended[i] = fmt.Sprintf("%s=%s", name, value)
	}
	return prepended
}

func (log *Logger) fileLocation(calldDepth int) (file string, line int) {
	if log.flag&(BitShortFile|BitLongFile) != 0 {
		var ok bool
		_, file, line, ok = runtime.Caller(calldDepth)
		if !ok {
			file = "unknown-file"
			line = 0
		}
	}
	return
}
