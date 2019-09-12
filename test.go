package zlsgo

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

// TestUtil Test aid
type TestUtil struct {
	T *testing.T
}

// NewTest testing object
func NewTest(t *testing.T) *TestUtil {
	return &TestUtil{t}
}

// Equal Equal
func (u *TestUtil) Equal(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		u.T.Errorf("%s 期待:%v (type %v) - 结果:%v (type %v)", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

func (u *TestUtil) EqualExit(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		u.T.Fatalf("%s 期待:%v (type %v) - 结果:%v (type %v)", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

// Log log
func (u *TestUtil) Log(v ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	file, line := f.FileLine(pc)
	base, _ := filepath.Abs(".")
	path, _ := filepath.Rel(base, file)
	tip := []interface{}{"  " + path + ":" + strconv.Itoa(line)}
	va := append(tip, v...)
	fmt.Println(va...)
}

// Fatal Fatal
func (u *TestUtil) Fatal(v ...interface{}) {
	u.T.Fatal(v...)
}

// PrintMyName PrintMyName
func (u *TestUtil) PrintMyName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	file, line := f.FileLine(pc)
	return file + ":" + strconv.Itoa(line)
}
