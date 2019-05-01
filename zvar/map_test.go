package zvar_test

import (
	. "github.com/sohaha/zlsgo/ztest"
	. "github.com/sohaha/zlsgo/zvar"
	"testing"
)

func TestMap(t *testing.T) {
	m := make(map[interface{}]interface{})
	m["T"] = "test"
	tMapKeyExists := MapKeyExists("T", m)
	Equal(t, true, tMapKeyExists)
}
