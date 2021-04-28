package zjson

import (
	"github.com/sohaha/zlsgo/zstring"
	"strconv"
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

	strBytes, err = SetBytesOptions(strBytes, "setBytes3.op", "new set", &Options{Optimistic: true, ReplaceInPlace: true})

	t.Log(string(strBytes), err)
	_, _ = DeleteBytes(strBytes, "setRawBytes")

	var j = struct {
		Name string `json:"n"`
	}{"isName"}
	jj, err := Marshal(j)
	t.Log(string(jj), err)
	t.Log(Stringify(j))
}

func TestSetSt(tt *testing.T) {
	var j = struct {
		Name string `json:"n"`
	}{"isName"}
	t := zlsgo.NewTest(tt)
	json, _ := Set("{}", "code", 200)
	json, _ = Set(json, "code2", uint(200))
	json, _ = Set(json, "int8", int8(8))
	json, _ = Set(json, "int32", int32(200))
	json, _ = Set(json, "int64", int64(200))
	json, _ = Set(json, "uint8", uint8(8))
	json, _ = Set(json, "uint32", uint32(200))
	json, _ = Set(json, "uint64", uint64(200))
	json, _ = SetOptions(json, "code2", 200.01, &Options{
		Optimistic: true,
	})
	tt.Log(json)
	json, _ = Set(json, "data", j)
	tt.Log(json)
	t.Equal("isName", Get(json, "data.n").String())
}

func BenchmarkSet(b *testing.B) {
	s := zstring.Rand(100)
	json := "{}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Set(json, strconv.Itoa(i), s)
	}
}

func BenchmarkSetBytes(b *testing.B) {
	s := []byte(zstring.Rand(100))
	json := []byte("{}")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetBytes(json, strconv.Itoa(i), s)
	}
}
