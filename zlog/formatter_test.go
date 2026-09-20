package zlog

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func newBufLogger(buf *bytes.Buffer, flag int) *Logger {
	return NewZLog(buf, "", flag, LogDump, false, 3)
}

func TestTextFormatterDefaultUnchanged(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.Info("hello")
	out := buf.String()
	if out != "[INFO]  hello\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestEntryTextFields(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.WithFields(Fields{"b": 2, "a": 1}).Info("hi")
	out := buf.String()
	if out != "[INFO]  hi a=1 b=2\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestEntryFieldQuoting(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.WithField("msg", "a b").Info("x")
	out := buf.String()
	if !strings.Contains(out, `msg="a b"`) {
		t.Fatalf("expected quoted field, got %q", out)
	}
}

func TestTextFieldTokenEncoding(t *testing.T) {
	for _, tc := range []struct{ input, encoded string }{
		{"simple", "simple"},
		{"", `""`},
		{"line\nbreak", `"line\nbreak"`},
		{"two words", `"two words"`},
		{"a=b", `"a=b"`},
		{"a\rb", `"a\rb"`},
		{"a\tb", `"a\tb"`},
		{"a\x00b", `"a\x00b"`},
		{"a\x7fb", `"a\x7fb"`},
		{"a\\b", `"a\\b"`},
		{"a\"b", `"a\"b"`},
		{"a\u2028b", `"a\u2028b"`},
	} {
		t.Run(tc.encoded, func(t *testing.T) {
			var buf bytes.Buffer
			l := newBufLogger(&buf, 0)
			l.WithField(tc.input, tc.input).Info("message")
			want := "message " + tc.encoded + "=" + tc.encoded + "\n"
			if buf.String() != want {
				t.Fatalf("got %q, want %q", buf.String(), want)
			}
			if strings.Count(buf.String(), "\n") != 1 || strings.Contains(buf.String(), "\r") {
				t.Fatalf("multiline record: %q", buf.String())
			}
		})
	}
}

func TestEntryImmutableAndFieldsCopied(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)

	src := Fields{"a": 1}
	e1 := l.WithFields(src)
	src["a"] = 99 // mutating caller map must not affect entry

	e2 := e1.WithField("b", 2)
	e1.Info("one")
	e2.Info("two")

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %v", lines)
	}
	if !strings.Contains(lines[0], "a=1") || strings.Contains(lines[0], "b=") {
		t.Fatalf("parent entry polluted: %q", lines[0])
	}
	if !strings.Contains(lines[1], "a=1") || !strings.Contains(lines[1], "b=2") {
		t.Fatalf("child entry missing fields: %q", lines[1])
	}
}

func TestEntryLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.SetLogLevel(LogError)
	e := l.WithField("k", "v")
	e.Debug("skip")
	e.Info("skip")
	e.Error("keep")
	out := buf.String()
	if strings.Contains(out, "skip") || !strings.Contains(out, "keep") {
		t.Fatalf("level filter broken: %q", out)
	}
}

func TestEntryGenericLog(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	e := l.WithField("k", "v")
	e.Log(LogWarn, "warn via Log")
	e.Logf(LogError, "err %d", 42)
	out := buf.String()
	if !strings.Contains(out, "[WARN]  warn via Log k=v") || !strings.Contains(out, "[ERROR] err 42 k=v") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel|BitShortFile)
	l.SetFormatter(JSONFormatter{DisableTimestamp: true})
	l.WithFields(Fields{"user": "bob", "n": 3}).Info("hi")

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	if m["level"] != "info" || m["msg"] != "hi" || m["user"] != "bob" || m["n"] != float64(3) {
		t.Fatalf("bad record: %v", m)
	}
	if caller, ok := m["caller"].(string); !ok || !strings.Contains(caller, "_test.go:") {
		t.Fatalf("bad caller: %v", m["caller"])
	}
	if _, ok := m["time"]; ok {
		t.Fatal("timestamp should be disabled")
	}
}

func TestJSONFormatterTimestampAndFieldMap(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.SetFormatter(JSONFormatter{
		TimestampFormat: "2006-01-02",
		FieldMap:        map[string]string{"msg": "message"},
	})
	l.Info("hi")

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if m["message"] != "hi" {
		t.Fatalf("FieldMap not applied: %v", m)
	}
	if ts, ok := m["time"].(string); !ok || len(ts) != len("2006-01-02") {
		t.Fatalf("bad timestamp: %v", m["time"])
	}
}

func TestJSONFormatterUnmarshalableField(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.SetFormatter(JSONFormatter{DisableTimestamp: true})
	l.WithField("fn", func() {}).Info("hi")

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	if _, ok := m["fn"].(string); !ok {
		t.Fatalf("unmarshalable field should degrade to string: %v", m["fn"])
	}
}

func TestJSONFormatterPrintlnNoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.SetFormatter(JSONFormatter{DisableTimestamp: true})
	l.Println("plain")

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	if _, ok := m["level"]; ok {
		t.Fatalf("Print record should not carry level: %v", m)
	}
	if m["msg"] != "plain" {
		t.Fatalf("bad msg: %v", m)
	}
}

func TestEntryWithErrorAndData(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	err := &testError{"boom"}
	e := l.WithError(err).WithField("id", 7)
	d := e.Data()
	if d["error"] != err || d["id"] != 7 {
		t.Fatalf("bad Data(): %v", d)
	}
	d["evil"] = true // must be a copy
	e.Info("x")
	if strings.Contains(buf.String(), "evil") {
		t.Fatal("Data() leaked internal map")
	}
	if !strings.Contains(buf.String(), "error=boom") || !strings.Contains(buf.String(), "id=7") {
		t.Fatalf("missing fields: %q", buf.String())
	}
}

func TestSetFormatterNilRestoresText(t *testing.T) {
	var buf bytes.Buffer
	l := newBufLogger(&buf, BitLevel)
	l.SetFormatter(JSONFormatter{DisableTimestamp: true})
	l.Info("json")
	l.SetFormatter(nil)
	l.Info("text")
	out := buf.String()
	if !strings.Contains(out, `"msg":"json"`) || !strings.Contains(out, "[INFO]  text") {
		t.Fatalf("formatter switch broken: %q", out)
	}
}

func TestConcurrentStructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	l := NewZLog(&lockedWriter{buf: &buf}, "", BitLevel, LogDump, false, 3)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e := l.WithField("worker", i)
			for j := 0; j < 50; j++ {
				e.Info("tick")
				l.SetLogLevel(LogDump)
				l.SetFormatter(nil)
			}
		}(i)
	}
	wg.Wait()
}

type lockedWriter struct {
	buf *bytes.Buffer
	mu  sync.Mutex
}

func (w *lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

type testError struct{ s string }

func (e *testError) Error() string { return e.s }
