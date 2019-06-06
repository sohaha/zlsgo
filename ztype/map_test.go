package ztype

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
)

func TestMap(t *testing.T) {

	T := zls.NewTest(t)
	m := make(map[interface{}]interface{})
	m["T"] = "test"
	tMapKeyExists := MapKeyExists("T", m)
	T.Equal(true, tMapKeyExists)
}
