package zlsgo

import (
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

// Unequal Unequal
func (u *TestUtil) Unequal(expected, actual interface{}) {
	if reflect.DeepEqual(expected, actual) {
		u.T.Errorf("Did not expect %v (type %v) - Got %v (type %v)", expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

func (u *TestUtil) Log(v ...interface{}) {
	u.T.Log(v...)
}

// PrintMyName PrintMyName
func (u *TestUtil) PrintMyName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	file, line := f.FileLine(pc)
	return file + ":" + strconv.Itoa(line)
}
