package zjson

import (
	"github.com/sohaha/zlsgo"
	"testing"
)

func TestSet(T *testing.T) {
	var demoData = "{}"
	var err error
	var str string
	t := zlsgo.NewTest(T)

	str, err = Set(demoData, "set", "new set")
	t.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "set.b", true)
	t.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "arr1", []string{"one"})
	t.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "arr2", "[1,2,3]")
	t.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "obj", `{"name":"pp"}`)
	t.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "set\\.ff", 1.66)
	t.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "newSet", `"haha"`)
	t.EqualExit(true, err == nil)
	t.Log(str, err)

	_, err = SetRaw(str, "", `haha`)
	t.EqualExit(true, err != nil)
	t.Log(err)

	newSet := Get(str, "newSet").String()
	t.Log(newSet)
	newStr, err := Delete(str, "newSet")
	t.EqualExit(false, Get(newStr, "newSet").Exists())
	t.Log(newStr, str, err)

	strBytes := []byte("{}")

	strBytes, err = SetBytes(strBytes, "setBytes", "new set")
	t.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetRawBytes(strBytes, "setRawBytes", []byte("Raw"))
	t.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetBytes(strBytes, "setBytes2.s", "new set")
	strBytes, err = SetBytesOptions(strBytes, "setBytes3.op", "new set", &StSetOptions{Optimistic: true, ReplaceInPlace: true})

	t.Log(string(strBytes), err)
	_, _ = DeleteBytes(strBytes, "setRawBytes")

	var j = struct {
		Name string `json:"n"`
	}{"isName"}
	jj, err := SetMarshal("", "", j)
	t.Log(string(jj), err)
}
