package zlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Fields is a set of key-value pairs attached to a log record.
// It is used by Entry and Formatter to provide structured logging.
type Fields map[string]interface{}

// LevelNames provides clean level names (no brackets/padding) aligned with
// the Levels order, suitable for structured output.
var LevelNames = []string{
	"FATAL",
	"PANIC",
	"TRACK",
	"ERROR",
	"WARNING",
	"TIPS",
	"SUCCESS",
	"INFO",
	"DEBUG",
	"DUMP",
}

func levelName(level int) string {
	if level >= 0 && level < len(LevelNames) {
		return LevelNames[level]
	}
	return ""
}

// Record is a single log event passed to a Formatter for serialization.
type Record struct {
	// Time is the moment the record was created.
	Time time.Time
	// Level is the zlog level constant; LogNot (-1) for the Print family.
	Level int
	// LevelText is the clean level name (e.g. "INFO"); empty when Level is out of range.
	LevelText string
	// Message is the raw log message.
	Message string
	// Fields are the structured key-value pairs bound via WithFields; may be nil.
	Fields Fields
	// File and Line are populated only when the logger flag requests file info.
	File string
	Line int
	// Prefix is the logger prefix.
	Prefix string
	// Tag is an optional extra text rendered between prefix and message.
	Tag string
	// Flag is the Bit* header flag bitmap in effect (rendering hint for text formatters).
	Flag int
	// Color indicates whether ANSI color output is enabled (rendering hint).
	Color bool
}

// Formatter serializes a Record into buf. Implementations must be safe for
// concurrent use. Set one on a Logger via SetFormatter; nil selects TextFormatter.
type Formatter interface {
	Format(r *Record, buf *bytes.Buffer)
}

var defaultFormatter = TextFormatter{}

// TextFormatter renders records in the classic zlog text format.
// It is the default formatter and reproduces the historical output exactly;
// structured fields (if any) are appended as sorted key=value pairs.
type TextFormatter struct{}

// Format implements Formatter.
func (TextFormatter) Format(r *Record, buf *bytes.Buffer) {
	formatHeader(buf, r)
	buf.WriteString(r.Prefix)
	buf.WriteString(r.Tag)
	msg := r.Message
	trailing := ""
	if len(r.Fields) > 0 && strings.HasSuffix(msg, "\n") {
		// Render fields before the trailing newline so each record stays one line.
		msg = msg[:len(msg)-1]
		trailing = "\n"
	}
	buf.WriteString(msg)
	appendTextFields(buf, r.Fields)
	buf.WriteString(trailing)
}

func appendTextFields(buf *bytes.Buffer, fields Fields) {
	if len(fields) == 0 {
		return
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteByte(' ')
		buf.WriteString(k)
		buf.WriteByte('=')
		v := fmt.Sprint(fields[k])
		if strings.IndexAny(v, " \t\n\"") >= 0 {
			fmt.Fprintf(buf, "%q", v)
		} else {
			buf.WriteString(v)
		}
	}
}

// JSONFormatter renders records as single-line JSON objects.
//
// Default keys: "time", "level", "msg", "caller", "prefix", "tag".
// FieldMap may rename the built-in keys (e.g. {"msg": "message"});
// unmapped keys keep their defaults. Structured Fields are merged at top
// level and take precedence over built-in keys on collision.
// Field values that fail JSON encoding are degraded to fmt strings.
type JSONFormatter struct {
	// TimestampFormat is the time.Format layout for the "time" key.
	// Defaults to time.RFC3339Nano.
	TimestampFormat string
	// DisableTimestamp omits the "time" key entirely.
	DisableTimestamp bool
	// DisableCaller omits the "caller" key even when the logger flag requests file info.
	DisableCaller bool
	// Pretty enables indented output for debugging.
	Pretty bool
	// FieldMap renames built-in keys: time, level, msg, caller, prefix, tag.
	FieldMap map[string]string
}

func jsonLevelText(level int, fallback string) string {
	if level == LogWarn {
		return "warning"
	}
	return strings.ToLower(fallback)
}

func (f JSONFormatter) fieldKey(name string) string {
	if f.FieldMap != nil {
		if k, ok := f.FieldMap[name]; ok {
			return k
		}
	}
	return name
}

// Format implements Formatter.
func (f JSONFormatter) Format(r *Record, buf *bytes.Buffer) {
	data := make(map[string]interface{}, len(r.Fields)+6)

	if !f.DisableTimestamp {
		tf := f.TimestampFormat
		if tf == "" {
			tf = time.RFC3339Nano
		}
		data[f.fieldKey("time")] = r.Time.Format(tf)
	}

	if r.LevelText != "" {
		data[f.fieldKey("level")] = jsonLevelText(r.Level, r.LevelText)
	}

	data[f.fieldKey("msg")] = strings.TrimSuffix(r.Message, "\n")

	if r.File != "" && !f.DisableCaller {
		data[f.fieldKey("caller")] = fmt.Sprintf("%s:%d", r.File, r.Line)
	}

	if r.Prefix != "" {
		data[f.fieldKey("prefix")] = r.Prefix
	}

	if r.Tag != "" {
		data[f.fieldKey("tag")] = r.Tag
	}

	for k, v := range r.Fields {
		if b, err := json.Marshal(v); err == nil {
			data[k] = json.RawMessage(b)
		} else {
			data[k] = fmt.Sprintf("%v", v)
		}
	}

	var (
		b   []byte
		err error
	)
	if f.Pretty {
		b, err = json.MarshalIndent(data, "", "  ")
	} else {
		b, err = json.Marshal(data)
	}
	if err != nil {
		fmt.Fprintf(buf, `{"level":"error","msg":%q}`, "zlog json encode failed: "+err.Error())
		return
	}
	buf.Write(b)
}
