package gtest

import (
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

// Equal 对等
func Equal(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s 期待:%v (type %v) - 结果:%v (type %v)", PrintMyName(t), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

// Unequal 不对等
func Unequal(t *testing.T, expected, actual interface{}) {
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

// PrintMyName 获取测试文件名
func PrintMyName(t *testing.T) string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	file, line := f.FileLine(pc)
	return file + ":" + strconv.Itoa(line)
}
