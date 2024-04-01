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
		Writer: os.Stderr,
	}

	l1 := zlog.NewZLog(w, "Custom1", zlog.BitDefault|zlog.BitLongFile, zlog.LogDump, true, 2)
	l1.Info("Test")

	// or

	l2 := zlog.New("Custom2")
	l2.ResetWriter(w)
	l2.Info("Test")
}
