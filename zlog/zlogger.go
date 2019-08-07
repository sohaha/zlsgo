/*
 * @Author: seekwe
 * @Date:   2019-05-17 13:45:52
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-06 15:37:24
 */

package zlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
	
	"github.com/sohaha/zlsgo/zls"
)

const (
	// LogMaxBuf LogMaxBuf
	LogMaxBuf = 1024 * 1024
)

// Log header information tag bit, using bitmap mode
const (
	BitDate         = 1 << iota                            // Date marker  2019/01/23
	BitTime                                                // Time Label Bit  01:23:12
	BitMicroSeconds                                        // Microsecond label bit 01:23:12.111222
	BitLongFile                                            // Full file name /home/go/src/github.com/sohaha/zlsgo/doc.go
	BitShortFile                                           // Final File Name   doc.go
	BitLevel                                               // Current log level
	BitStdFlag      = BitDate | BitTime                    // Standard header log format
	BitDefault      = BitLevel | BitShortFile | BitStdFlag // Default log header format
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
	LogNot = -1
)

var levels = []string{
	"[FATAL]",
	"[PANIC]",
	"[ERROR]",
	"[WARN] ",
	"[SUCCE]",
	"[INFO] ",
	"[DEBUG]",
}

var levelColous = []Color{
	ColorRed,
	ColorLightRed,
	ColorRed,
	ColorYellow,
	ColorGreen,
	ColorBlue,
	ColorCyan,
}

// Logger logger struct
type Logger struct {
	mu         sync.Mutex
	prefix     string
	flag       int
	out        io.Writer
	buf        bytes.Buffer
	file       *os.File
	calldDepth int
	level      int
	color      bool
}

// New Initialize a log object
func New(moduleName ...string) *Logger {
	name := ""
	if len(moduleName) > 0 {
		name = moduleName[0]
	}
	return NewZLog(os.Stderr, name, BitDefault, 6, true, 2)
}

// NewZLog Create log
func NewZLog(out io.Writer, prefix string, flag int, level int, color bool, calldDepth int) *Logger {
	zlog := &Logger{out: out, prefix: prefix, flag: flag, file: nil, calldDepth: calldDepth, level: level, color: color}
	runtime.SetFinalizer(zlog, CleanLog)
	return zlog
}

// CleanLog CleanLog
func CleanLog(log *Logger) {
	log.closeFile()
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
			buf.WriteString(log.ColorTextWrap(levelColous[level], levels[level]+" "))
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
func (log *Logger) OutPut(level int, s string, prefixText ...string) error {
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
	log.mu.Lock()
	defer log.mu.Unlock()
	if !isNotLevel && (log.flag&(BitShortFile|BitLongFile) != 0) {
		log.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(log.calldDepth)
		if !ok {
			file = "unknown-file"
			line = 0
		}
		log.mu.Lock()
	}
	
	log.buf.Reset()
	log.formatHeader(&log.buf, now, file, line, level)
	log.buf.WriteString(s)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		log.buf.WriteByte('\n')
	}
	_, err := log.out.Write(log.buf.Bytes())
	return err
}

// Printf Printf
func (log *Logger) Printf(format string, v ...interface{}) {
	_ = log.OutPut(LogNot, fmt.Sprintf(format, v...))
}

// Println Println
func (log *Logger) Println(v ...interface{}) {
	_ = log.OutPut(LogNot, fmt.Sprintln(v...))
}

// Debugf Debugf
func (log *Logger) Debugf(format string, v ...interface{}) {
	_ = log.OutPut(LogDebug, fmt.Sprintf(format, v...))
}

// Debug Debug
func (log *Logger) Debug(v ...interface{}) {
	_ = log.OutPut(LogDebug, fmt.Sprintln(v...))
}

// Successf Successf
func (log *Logger) Successf(format string, v ...interface{}) {
	_ = log.OutPut(LogSuccess, fmt.Sprintf(format, v...))
}

// Success Success
func (log *Logger) Success(v ...interface{}) {
	_ = log.OutPut(LogSuccess, fmt.Sprintln(v...))
}

// Infof Infof
func (log *Logger) Infof(format string, v ...interface{}) {
	_ = log.OutPut(LogInfo, fmt.Sprintf(format, v...))
}

// Info Info
func (log *Logger) Info(v ...interface{}) {
	_ = log.OutPut(LogInfo, fmt.Sprintln(v...))
}

// Warnf Warnf
func (log *Logger) Warnf(format string, v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintf(format, v...))
}

// Warn Warn
func (log *Logger) Warn(v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintln(v...))
}

// Errorf Errorf
func (log *Logger) Errorf(format string, v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintf(format, v...))
}

// Error Error
func (log *Logger) Error(v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintln(v...))
}

// Fatalf Fatalf
func (log *Logger) Fatalf(format string, v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatal Fatal
func (log *Logger) Fatal(v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panicf Panicf
func (log *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	_ = log.OutPut(LogPanic, fmt.Sprintf(format, s))
	panic(s)
}

// panic panic
func (log *Logger) Panic(v ...interface{}) {
	s := fmt.Sprintln(v...)
	_ = log.OutPut(LogPanic, s)
	panic(s)
}

// Stack Stack
func (log *Logger) Stack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, LogMaxBuf)
	// 得到当前堆栈信息
	n := runtime.Stack(buf, true)
	s += string(buf[:n])
	s += "\n"
	_ = log.OutPut(LogError, s)
}

// Track Track
func (log *Logger) Track(logTip string, v ...int) {
	b := bytes.NewBufferString(logTip)
	b.WriteString("\n")
	l := len(v)
	depth := log.calldDepth - 1
	max := 1
	if l == 1 {
		max = v[0]
	} else if l > 1 {
		depth = depth + v[1]
		max = v[0]
	}
	
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
		b.WriteString(fmt.Sprintf("    %v:%d %v\n", file, line, name))
	}
	
	_ = log.OutPut(LogDebug, b.String())
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

// SetLogFile Setting log file output
func (log *Logger) SetLogFile(fileDir string, fileName string) {
	var file *os.File
	
	_ = mkdirLog(fileDir)
	
	fullPath := fileDir + "/" + fileName
	if log.checkFileExist(fullPath) {
		file, _ = os.OpenFile(fullPath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		file, _ = os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	}
	log.DisableConsoleColor()
	log.mu.Lock()
	defer log.mu.Unlock()
	
	log.closeFile()
	log.file = file
	log.out = file
}

func (log *Logger) SetSaveLogFile(fileDir string, fileName string) {
	log.SetLogFile(fileDir, fileName)
	log.out = io.MultiWriter(log.file, os.Stdout)
}

func (log *Logger) closeFile() {
	if log.file != nil {
		_ = log.file.Close()
		log.file = nil
		log.out = os.Stderr
	}
}

// SetLogLevel Setting log display level
func (log *Logger) SetLogLevel(level int) {
	log.level = level
}

// GetLogLevel Get log display level
func (log *Logger) GetLogLevel() int {
	return log.level
}

func (log *Logger) checkFileExist(filename string) bool {
	return zls.FileExist(filename)
}

func mkdirLog(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0775); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func itoa(buf *bytes.Buffer, i int, wid int) {
	u := uint(i)
	if u == 0 && wid <= 1 {
		buf.WriteByte('0')
		return
	}
	
	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	
	// avoid slicing b to avoid an allocation.
	for bp < len(b) {
		buf.WriteByte(b[bp])
		bp++
	}
}
