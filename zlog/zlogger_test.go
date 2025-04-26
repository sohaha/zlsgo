package zlog_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sohaha/zlsgo/zlog"
)

type Report struct {
	Writer io.Writer
}

func (w *Report) Write(p []byte) (n int, err error) {
	// Here you can initiate an HTTP request to report an error
	fmt.Println("Report: ", string(p))

	return w.Writer.Write(p)
}

func TestCustomWriter(t *testing.T) {
	w := &Report{
		Writer: os.Stdout,
	}

	l1 := zlog.NewZLog(w, "[Custom1] ", zlog.BitLevel, zlog.LogDump, true, 3)
	l1.Info("Test")

	// or

	l2 := zlog.New("[Custom2] ")
	l2.Info("Test")
	l2.Writer().Set(w)
	l2.Info("Test 2")

	// or
	l3 := zlog.New("[Custom3] ")
	l3.Info("Test")
	l3.Writer().Reset(l2)
	l3.Info("Test")
}
