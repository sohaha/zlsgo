package zjson

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestSet(t *testing.T) {
	var demoData = "{}"
	var err error
	var str string
	tt := zlsgo.NewTest(t)

	str, err = Set(demoData, "set", "new set")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "set.b", true)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "arr1", []string{"one"})
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw("", "arr2", "[1,2,3]")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set("", "arr2", "[1,2,3]")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "arr3.:3", "[1,2,3]")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "obj", `{"name":"pp"}`)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "set\\.ff", 1.66)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "newSet", `"haha"`)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	_, err = SetRaw(str, "", `haha`)
	tt.EqualExit(true, err != nil)
	t.Log(err)

	newSet := Get(str, "newSet").String()
	t.Log(newSet)
	newStr, err := Delete(str, "newSet")
	tt.EqualExit(false, Get(newStr, "newSet").Exists())
	t.Log(newStr, str, err)

	strBytes := []byte("{}")

	strBytes, err = SetBytes(strBytes, "setBytes", "new set")
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetRawBytes(strBytes, "setRawBytes", []byte("Raw"))
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetRawBytes(strBytes, "set\\.other", []byte("Raw"))
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetRawBytes(strBytes, "set.other", []byte("Raw"))
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	_, _ = SetBytes(strBytes, "setBytes2.s", "new set")

	strBytes, err = SetBytesOptions(strBytes, "setBytes3.op", "new set", &StSetOptions{Optimistic: true, ReplaceInPlace: true})

	t.Log(string(strBytes), err)
	_, _ = DeleteBytes(strBytes, "setRawBytes")

	var j = struct {
		Name string `json:"n"`
	}{"isName"}
	jj, err := Marshal(j)
	t.Log(string(jj), err)
	t.Log(Stringify(j))
}
