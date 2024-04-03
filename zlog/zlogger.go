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
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
)

// Log header information tag bit, using bitmap mode
const (
	BitDate         int                                 = 1 << iota // Date marker  2019/01/23
	BitTime                                                         // Time Label Bit  01:23:12
	BitMicroSeconds                                                 // Microsecond label bit 01:23:12.111222
	BitLongFile                                                     // Full file name, example: /home/go/src/github.com/sohaha/zlsgo/doc.go
	BitShortFile                                                    // Final File name   doc.go
	BitLevel                                                        // Current log level
	BitStdFlag      = BitDate | BitTime                             // Standard header log format
	BitDefault      = BitLevel | BitShortFile | BitTime             // Default log header format
	// LogMaxBuf LogMaxBuf
	LogMaxBuf = 1024 * 1024
)

// log level
const (
	LogFatal = iota
	LogPanic
	LogTrack
	LogError
	LogWarn
	LogTips
	LogSuccess
	LogInfo
	LogDebug
	LogDump
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
	// Logger logger struct
	Logger struct {
		out        io.Writer
		file       *zfile.MemoryFile
		prefix     string
		fileDir    string
		fileName   string
		ignoreLogs []string
		// buf           bytes.Buffer
		calldDepth    int
		level         int
		flag          int
		mu            sync.RWMutex
		color         bool
		fileAndStdout bool
	}
	formatter struct {
		v     reflect.Value
		force bool
		quote bool
	}
	visit struct {
		typ reflect.Type
		v   uintptr
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
	if log.flag&(BitDate|BitTime|BitMicroSeconds|BitLevel) != 0 {

		if log.flag&BitDate != 0 {
			buf.WriteString(ztime.FormatTime(t, "Y/m/d "))
		}

		if log.flag&(BitTime|BitMicroSeconds) != 0 {
			buf.WriteString(ztime.FormatTime(t, "H:i:s"))
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

// outPut Output log
func (log *Logger) outPut(level int, s string, isWrap bool, fn func() func(), prefixText ...string) error {

	if log.ignoreLogs != nil && len(s) > 0 {
		p := s
		if isWrap {
			p = p[:len(p)-1]
		}
		for _, v := range log.ignoreLogs {
			if zstring.Match(p, v) {
				return nil
			}
		}
	}

	var after func()
	if fn != nil {
		log.mu.Lock()
		after = fn()
	}

	defer func() {
		if after != nil {
			after()
			log.mu.Unlock()
		}
	}()

	if len(prefixText) > 0 {
		s = prefixText[0] + s
	}

	buf := zutil.GetBuff(len(s) + 34)
	defer zutil.PutBuff(buf)

	now := ztime.Time()
	if level != LogNot {
		file, line := log.fileLocation()
		log.formatHeader(buf, now, file, line, level)
	}

	if log.prefix != "" {
		buf.WriteString(log.prefix)
	}

	buf.WriteString(s)
	if isWrap && len(s) > 0 && s[len(s)-1] != '\n' {
		buf.WriteByte('\n')
	}
	_, err := log.out.Write(buf.Bytes())
	return err
}

// Printf formats according to a format specifier and writes to standard output
func (log *Logger) Printf(format string, v ...interface{}) {
	_ = log.outPut(LogNot, fmt.Sprintf(format, v...), false, nil)
}

// Println Println
func (log *Logger) Println(v ...interface{}) {
	_ = log.outPut(LogNot, sprintStr(v...)+"\n", true, nil)
}

// Debugf Debugf
func (log *Logger) Debugf(format string, v ...interface{}) {
	if log.level < LogDebug {
		return
	}
	_ = log.outPut(LogDebug, fmt.Sprintf(format, v...), false, nil)
}

// Debug Debug
func (log *Logger) Debug(v ...interface{}) {
	if log.level < LogDebug {
		return
	}
	_ = log.outPut(LogDebug, sprintStr(v...)+"\n", true, nil)
}

// Dump pretty print format
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

	_ = log.outPut(LogDump, fmt.Sprintln(args...), true, nil)
}

// Successf output Success
func (log *Logger) Successf(format string, v ...interface{}) {
	if log.level < LogSuccess {
		return
	}
	_ = log.outPut(LogSuccess, fmt.Sprintf(format, v...), false, nil)
}

// Success output Success
func (log *Logger) Success(v ...interface{}) {
	if log.level < LogSuccess {
		return
	}
	_ = log.outPut(LogSuccess, sprintStr(v...)+"\n", true, nil)
}

// Infof output Info
func (log *Logger) Infof(format string, v ...interface{}) {
	if log.level < LogInfo {
		return
	}
	_ = log.outPut(LogInfo, fmt.Sprintf(format, v...), false, nil)
}

// Info output Info
func (log *Logger) Info(v ...interface{}) {
	if log.level < LogInfo {
		return
	}
	_ = log.outPut(LogInfo, sprintStr(v...)+"\n", true, nil)
}

// Tipsf output Tips
func (log *Logger) Tipsf(format string, v ...interface{}) {
	if log.level < LogTips {
		return
	}
	_ = log.outPut(LogTips, fmt.Sprintf(format, v...), false, nil)
}

// Tips output Tips
func (log *Logger) Tips(v ...interface{}) {
	if log.level < LogTips {
		return
	}
	_ = log.outPut(LogTips, sprintStr(v...)+"\n", true, nil)
}

// Warnf output Warn
func (log *Logger) Warnf(format string, v ...interface{}) {
	if log.level < LogWarn {
		return
	}
	_ = log.outPut(LogWarn, fmt.Sprintf(format, v...), false, nil)
}

// Warn output Warn
func (log *Logger) Warn(v ...interface{}) {
	if log.level < LogWarn {
		return
	}
	_ = log.outPut(LogWarn, sprintStr(v...)+"\n", true, nil)
}

func sprintStr(v ...interface{}) string {
	if len(v) == 1 {
		return ztype.ToString(v[0])
	}
	return fmt.Sprint(v...)
}

// Errorf output Error
func (log *Logger) Errorf(format string, v ...interface{}) {
	if log.level < LogError {
		return
	}
	_ = log.outPut(LogError, fmt.Sprintf(format, v...), false, nil)
}

// Error output Error
func (log *Logger) Error(v ...interface{}) {
	if log.level < LogError {
		return
	}
	_ = log.outPut(LogError, sprintStr(v...)+"\n", true, nil)
}

// Fatalf output Fatal
func (log *Logger) Fatalf(format string, v ...interface{}) {
	if log.level < LogFatal {
		return
	}
	_ = log.outPut(LogFatal, fmt.Sprintf(format, v...), false, nil)
	osExit(1)
}

// Fatal output Fatal
func (log *Logger) Fatal(v ...interface{}) {
	if log.level < LogFatal {
		return
	}
	_ = log.outPut(LogFatal, sprintStr(v...)+"\n", true, nil)
	osExit(1)
}

// Panicf output Panic
func (log *Logger) Panicf(format string, v ...interface{}) {
	if log.level < LogPanic {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = log.outPut(LogPanic, fmt.Sprintf(format, s), false, nil)
	panic(s)
}

// Panic output panic
func (log *Logger) Panic(v ...interface{}) {
	if log.level < LogPanic {
		return
	}
	s := sprintStr(v...) + "\n"
	_ = log.outPut(LogPanic, s, true, nil)
	panic(s)
}

// Stack output Stack
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
	_ = log.outPut(LogTrack, s, true, nil)
}

// Track output Track
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
	_ = log.outPut(LogTrack, text, true, nil)
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

func (log *Logger) GetPrefix() string {
	return log.prefix
}

// SetLogLevel Setting log display level
func (log *Logger) SetLogLevel(level int) {
	log.level = level
}

// GetLogLevel Get log display level
func (log *Logger) GetLogLevel() int {
	return log.level
}

func (log *Logger) Write(b []byte) (n int, err error) {
	_ = log.outPut(LogWarn, zstring.Bytes2String(b), false, nil)
	return len(b), nil
}

func (log *Logger) SetIgnoreLog(logs ...string) {
	log.ignoreLogs = append(log.ignoreLogs, logs...)
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

func (log *Logger) ResetWriter(w io.Writer) {
	log.out = w
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

func (log *Logger) fileLocation() (file string, line int) {
	if log.flag&(BitShortFile|BitLongFile) != 0 {
		var ok bool
		_, file, line, ok = runtime.Caller(log.calldDepth + 1)
		if !ok {
			file = "unknown-file"
			line = 0
		}
	}
	return
}
