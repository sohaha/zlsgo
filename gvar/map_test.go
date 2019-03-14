package gvar

import (
	"testing"
	. "github.com/sohaha/zlsgo/gtest"
)

func TestMap(t *testing.T) {
	m := make(map[interface{}]interface{})
	m["T"] = "test"
	tMapKeyExists := MapKeyExists("T", m)
	Equal(t, true, tMapKeyExists)
}
