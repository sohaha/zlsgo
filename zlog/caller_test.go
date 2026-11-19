package zlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// TestEntryCallerLineStable pins the stack depth used by outPutFields. If the
// skip count or the call structure changes, the reported file:line shifts by a
// frame and this test fails.
func TestEntryCallerLineStable(t *testing.T) {
	var buf bytes.Buffer
	l := NewZLog(&buf, "", BitLevel|BitShortFile, LogDump, false, 3)

	_, _, wantLine, _ := runtime.Caller(0)
	l.WithField("k", "v").Info("caller") // wantLine+1

	want := fmt.Sprintf("caller_test.go:%d", wantLine+1)
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("caller = %q, want it to contain %q", buf.String(), want)
	}
}

// TestJSONEntryCallerLineStable covers the same depth through JSONFormatter.
func TestJSONEntryCallerLineStable(t *testing.T) {
	var buf bytes.Buffer
	l := NewZLog(&buf, "", BitLevel|BitShortFile, LogDump, false, 3)
	l.SetFormatter(JSONFormatter{DisableTimestamp: true})

	_, _, wantLine, _ := runtime.Caller(0)
	l.WithField("k", "v").Info("caller") // wantLine+1

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	want := fmt.Sprintf("caller_test.go:%d", wantLine+1)
	caller, _ := m["caller"].(string)
	if !strings.Contains(caller, want) {
		t.Fatalf("caller = %v, want it to contain %q", m["caller"], want)
	}
}
