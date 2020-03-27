/*
 * @Author: seekwe
 * @Date:   2019-05-17 13:45:52
 * @Last Modified by:   seekwe
 * @Last Modified time: 2020-02-17 12:22:00
 */

package zlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

// Log header information tag bit, using bitmap mode
const (
	BitDate         = 1 << iota                            // Date marker  2019/01/23
	BitTime                                                // Time Label Bit  01:23:12
	BitMicroSeconds                                        // Microsecond label bit 01:23:12.111222
	BitLongFile                                            // Full file name /home/go/src/github.com/sohaha/zlsgo/doc.go
	BitShortFile                                           // Final File name   doc.go
	BitLevel                                               // Current log level
	BitStdFlag      = BitDate | BitTime                    // Standard header log format
	BitDefault      = BitLevel | BitShortFile | BitStdFlag // Default log header format
	// LogMaxBuf LogMaxBuf
	LogMaxBuf = 1024 * 1024
)

// log level
const (
	LogFatal = iota
	LogPanic
	LogError
	LogWarn
	LogSuccess
	LogInfo
	LogDebug
	LogDump
	LogNot = -1
)

var Levels = []string{
	"[Fatal]",
	"[Panic]",
	"[Error]",
	"[Warn] ",
	"[Succe]",
	"[Info] ",
	"[Debug]",
	"[Dump] ",
}

var LevelColous = []Color{
	ColorRed,
	ColorLightRed,
	ColorRed,
	ColorYellow,
	ColorGreen,
	ColorBlue,
	ColorLightCyan,
	ColorCyan,
}

// Logger logger struct
type Logger struct {
	mu            sync.RWMutex
	prefix        string
	flag          int
	out           io.Writer
	buf           bytes.Buffer
	file          *os.File
	calldDepth    int
	level         int
	color         bool
	FileMaxSize   int64
	fileDir       string
	fileName      string
	fileAndStdout bool
}
type (
	formatter struct {
		v     reflect.Value
		force bool
		quote bool
	}
	visit struct {
		v   uintptr
		typ reflect.Type
	}
	zprinter struct {
		io.Writer
		tw      *tabwriter.Writer
		visited map[visit]int
		depth   int
	}
)

// New Initialize a log object
func New(moduleName ...string) *Logger {
	name := ""
	if len(moduleName) > 0 {
		name = moduleName[0]
	}
	return NewZLog(os.Stderr, name, BitDefault, LogDump, true, 2)
}

// NewZLog Create log
func NewZLog(out io.Writer, prefix string, flag int, level int, color bool, calldDepth int) *Logger {
	zlog := &Logger{out: out, prefix: prefix, flag: flag, file: nil, calldDepth: calldDepth, level: level, color: color}
	runtime.SetFinalizer(zlog, CleanLog)
	return zlog
}

// CleanLog CleanLog
func CleanLog(log *Logger) {
	log.CloseFile()
}

// DisableConsoleColor DisableConsoleColor
func (log *Logger) DisableConsoleColor() {
	log.color = false
}

// ForceConsoleColor ForceConsoleColor
func (log *Logger) ForceConsoleColor() {
	log.color = true
}

// ColorTextWrap ColorTextWrap
func (log *Logger) ColorTextWrap(color Color, text string) string {
	if log.color {
		return ColorTextWrap(color, text)
	}
	return text
}

// ColorBackgroundWrap ColorBackgroundWrap
func (log *Logger) ColorBackgroundWrap(color Color, backgroundColor Color, text string) string {
	if log.color {
		return ColorBackgroundWrap(color, backgroundColor, text)
	}
	return text
}

// OpTextWrap OpTextWrap
func (log *Logger) OpTextWrap(color Op, text string) string {
	if log.color {
		return OpTextWrap(color, text)
	}
	return text
}

func (log *Logger) formatHeader(buf *bytes.Buffer, t time.Time, file string, line int, level int) {
	if level == LogNot {
		return
	}
	if log.prefix != "" {
		buf.WriteString(log.prefix)
	}
	if log.flag&(BitDate|BitTime|BitMicroSeconds|BitLevel) != 0 {
		if log.flag&BitDate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			buf.WriteByte('/') // "2019/"
			itoa(buf, int(month), 2)
			buf.WriteByte('/') // "2019/04/"
			itoa(buf, day, 2)
			buf.WriteByte(' ') // "2019/04/11 "
		}

		if log.flag&(BitTime|BitMicroSeconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			buf.WriteByte(':') // "12:"
			itoa(buf, min, 2)
			buf.WriteByte(':') // "12:12:"
			itoa(buf, sec, 2)  // "12:12:59"
			if log.flag&BitMicroSeconds != 0 {
				buf.WriteByte('.')
				itoa(buf, t.Nanosecond()/1e3, 6) // "12:12:59.123456
			}
			buf.WriteByte(' ')
		}

		if log.flag&BitLevel != 0 {
			buf.WriteString(log.ColorTextWrap(LevelColous[level], Levels[level]+" "))
		}

		if log.flag&(BitShortFile|BitLongFile) != 0 {
			if log.flag&BitShortFile != 0 {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				file = short
			}
			buf.WriteString(file)
			buf.WriteByte(':')
			itoa(buf, line, -1)
			buf.WriteString(": ")
		}
	}
}

// Writer Writer
func (log *Logger) Writer() io.Writer {
	return log.out
}

// OutPut Output log
func (log *Logger) OutPut(level int, s string, isWrap bool, prefixText ...string) error {
	log.mu.Lock()
	defer log.mu.Unlock()
	isNotLevel := level == LogNot
	if log.level < level {
		return nil
	}
	if len(prefixText) > 0 {
		s = prefixText[0] + s
	}
	now := time.Now()
	var file string
	var line int
	if !isNotLevel && (log.flag&(BitShortFile|BitLongFile) != 0) {
		var ok bool
		_, file, line, ok = runtime.Caller(log.calldDepth)
		if !ok {
			file = "unknown-file"
			line = 0
		}
	}

	log.buf.Reset()
	log.formatHeader(&log.buf, now, file, line, level)
	log.buf.WriteString(s)
	if isWrap && len(s) > 0 && s[len(s)-1] != '\n' {
		log.buf.WriteByte('\n')
	}
	_, err := log.out.Write(log.buf.Bytes())
	if log.file != nil && log.FileMaxSize > 0 {
		if fileInfo, err := log.file.Stat(); err == nil {
			logSize := fileInfo.Size()
			if logSize > log.FileMaxSize {
				logFile := log.fileDir + "/" + log.fileName
				oldFile := oldLogFile(log.fileDir, log.fileName)
				log.CloseFile()
				_ = os.Rename(logFile, oldFile)
				file, _ := openFile(log.fileDir, log.fileName)
				log.file = file
				if log.fileAndStdout {
					log.out = io.MultiWriter(log.file, os.Stdout)
				} else {
					log.out = file
				}
			}
		}
	}
	return err
}

// Printf Printf
func (log *Logger) Printf(format string, v ...interface{}) {
	_ = log.OutPut(LogNot, fmt.Sprintf(format, v...), false)
}

// Println Println
func (log *Logger) Println(v ...interface{}) {
	_ = log.OutPut(LogNot, fmt.Sprintln(v...), true)
}

// Debugf Debugf
func (log *Logger) Debugf(format string, v ...interface{}) {
	_ = log.OutPut(LogDebug, fmt.Sprintf(format, v...), false)
}

// Debug Debug
func (log *Logger) Debug(v ...interface{}) {
	_ = log.OutPut(LogDebug, fmt.Sprintln(v...), true)
}

// Dump Dump
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

	_ = log.OutPut(LogDump, fmt.Sprintln(args...), true)
}

// Successf Successf
func (log *Logger) Successf(format string, v ...interface{}) {
	_ = log.OutPut(LogSuccess, fmt.Sprintf(format, v...), false)
}

// Success Success
func (log *Logger) Success(v ...interface{}) {
	_ = log.OutPut(LogSuccess, fmt.Sprintln(v...), true)
}

// Infof Infof
func (log *Logger) Infof(format string, v ...interface{}) {
	_ = log.OutPut(LogInfo, fmt.Sprintf(format, v...), false)
}

// Info Info
func (log *Logger) Info(v ...interface{}) {
	_ = log.OutPut(LogInfo, fmt.Sprintln(v...), true)
}

// Warnf Warnf
func (log *Logger) Warnf(format string, v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintf(format, v...), false)
}

// Warn Warn
func (log *Logger) Warn(v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintln(v...), true)
}

// Errorf Errorf
func (log *Logger) Errorf(format string, v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintf(format, v...), false)
}

// Error Error
func (log *Logger) Error(v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintln(v...), true)
}

// Fatalf Fatalf
func (log *Logger) Fatalf(format string, v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintf(format, v...), false)
	osExit(1)
}

// Fatal Fatal
func (log *Logger) Fatal(v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintln(v...), true)
	osExit(1)
}

// Panicf Panicf
func (log *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	_ = log.OutPut(LogPanic, fmt.Sprintf(format, s), false)
	panic(s)
}

// panic panic
func (log *Logger) Panic(v ...interface{}) {
	s := fmt.Sprintln(v...)
	_ = log.OutPut(LogPanic, s, true)
	panic(s)
}

// Stack Stack
func (log *Logger) Stack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, LogMaxBuf)
	n := runtime.Stack(buf, true)
	s += string(buf[:n])
	s += "\n"
	_ = log.OutPut(LogError, s, true)
}

// Track Track
func (log *Logger) Track(logTip string, v ...int) {
	depth := log.calldDepth
	max := 1
	l := len(v)
	if l == 1 {
		max = v[0]
	} else if l > 1 {
		depth = depth + v[1]
		max = v[0]
	}
	if max == 0 {
		max = 9999
	}
	b := zstring.Buffer()
	track := TrackCurrent(max, depth)
	b.WriteString(logTip)
	b.WriteString("\n")
	b.WriteString(strings.Join(track, "\n"))
	_ = log.OutPut(LogDebug, b.String(), true)
}

func TrackCurrent(max, depth int) (track []string) {
	stop := func() bool {
		if max == -1 {
			return false
		}
		max--
		return max <= -1
	}
	for skip := depth; ; skip++ {
		name, file, line, ok := callerName(skip)
		if !ok || stop() {
			break
		}
		track = append(track, fmt.Sprintf("%v:%d %v", file, line, name))
	}
	return
}

func callerName(skip int) (name, file string, line int, ok bool) {
	var pc uintptr
	if pc, file, line, ok = runtime.Caller(skip + 1); !ok {
		return
	}
	name = runtime.FuncForPC(pc).Name()
	return
}

// GetFlags Get the current log bitmap tag
func (log *Logger) GetFlags() int {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.flag
}

// ResetFlags Reset the GetFlags bitMap tag bit in the log
func (log *Logger) ResetFlags(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag = flag
}

// AddFlag Set flag Tags
func (log *Logger) AddFlag(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag |= flag
}

// SetPrefix Setting log prefix
func (log *Logger) SetPrefix(prefix string) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.prefix = prefix
}

// SetLogLevel Setting log display level
func (log *Logger) SetLogLevel(level int) {
	log.level = level
}

// GetLogLevel Get log display level
func (log *Logger) GetLogLevel() int {
	return log.level
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

func formatArgs(args ...interface{}) []interface{} {
	formatted := make([]interface{}, 0, len(args))
	for _, a := range args {
		s := ColorTextWrap(ColorCyan, sprint(a))
		formatted = append(formatted, s)
	}
	return formatted
}

func sprint(a ...interface{}) string {
	return fmt.Sprint(wrap(a, true)...)
}

func wrap(a []interface{}, force bool) []interface{} {
	w := make([]interface{}, len(a))
	for i, x := range a {
		w[i] = formatter{v: reflect.ValueOf(x), force: force}
	}
	return w
}

func writeByte(w io.Writer, b byte) {
	_, _ = w.Write([]byte{b})
}

func prependArgName(names []string, values []interface{}) []interface{} {
	prepended := make([]interface{}, len(values))
	for i, value := range values {
		name := ""
		if i < len(names) {
			name = names[i]
		}
		if name == "" {
			prepended[i] = value
			continue
		}
		name = ColorTextWrap(ColorBlue, OpTextWrap(OpBold, name))
		prepended[i] = fmt.Sprintf("\n%s=%s", name, value)
	}
	return prepended
}
