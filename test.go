package zlsgo

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestUtil Test aid
type TestUtil struct {
	t testing.TB
}

// NewTest testing object
func NewTest(t testing.TB) *TestUtil {
	return &TestUtil{
		t: t,
	}
}

// GetCallerInfo GetCallerInfo
func (u *TestUtil) GetCallerInfo() string {
	var info string

	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		basename := file
		if !strings.HasSuffix(basename, "_test.go") {
			continue
		}

		funcName := runtime.FuncForPC(pc).Name()
		index := strings.LastIndex(funcName, ".Test")
		if index == -1 {
			index = strings.LastIndex(funcName, ".Benchmark")
			if index == -1 {
				continue
			}
		}
		funcName = funcName[index+1:]

		if index := strings.IndexByte(funcName, '.'); index > -1 {
			// funcName = funcName[:index]
			// info = funcName + "(" + basename + ":" + strconv.Itoa(line) + ")"
			info = basename + ":" + strconv.Itoa(line)
			continue
		}

		info = basename + ":" + strconv.Itoa(line)
		break
	}

	if info == "" {
		info = "<Unable to get information>"
	}
	return info
}

// Equal Equal
func (u *TestUtil) Equal(expected, actual interface{}) bool {
	if !reflect.DeepEqual(expected, actual) {
		u.t.Helper()
		fmt.Printf("        %s 期待:%v (type %v) - 结果:%v (type %v)\n", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
		u.t.Fail()
		return false
	}
	return true
}

// EqualTrue EqualTrue
func (u *TestUtil) EqualTrue(actual interface{}) {
	u.Equal(true, actual)
}

// EqualNil EqualNil
func (u *TestUtil) EqualNil(actual interface{}) {
	u.Equal(nil, actual)
}

// NoError NoError
func (u *TestUtil) NoError(err error, exit ...bool) bool {
	if err == nil {
		return true
	}

	fmt.Printf("    %s Error: %s\n", u.PrintMyName(), err)

	if len(exit) > 0 && exit[0] {
		u.t.Fatal()
	}
	return false
}

// EqualExit EqualExit
func (u *TestUtil) EqualExit(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		fmt.Printf("        %s 期待:%v (type %v) - 结果:%v (type %v)\n", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
		u.t.Fatal()
	}
}

// Log log
func (u *TestUtil) Log(v ...interface{}) {
	u.t.Helper()
	u.t.Log(v...)
}

// Logf Logf
func (u *TestUtil) Logf(format string, args ...interface{}) {
	u.t.Helper()
	u.t.Logf(format, args...)
}

// Fatal Fatal
func (u *TestUtil) Fatal(v ...interface{}) {
	u.t.Helper()
	u.t.Fatal(v...)
}

// PrintMyName PrintMyName
func (u *TestUtil) PrintMyName() string {
	return u.GetCallerInfo()
}

func (u *TestUtil) Run(name string, f func(tt *TestUtil)) {
	u.t.Helper()
	u.t.(*testing.T).Run(name, func(t *testing.T) {
		f(NewTest(t))
	})
}

func (u *TestUtil) T() *testing.T {
	return u.t.(*testing.T)
}
