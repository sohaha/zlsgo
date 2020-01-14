package zcfg

import (
	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"testing"
)

func TestJSONCfg(T *testing.T) {
	t := zlsgo.NewTest(T)
	remote := "https://unpkg.com/zls-cli@0.6.0/package.json"
	json, err := JSON(remote)
	if err != nil {
		t.Log(err)
	}
	t.EqualExit(json, GetJSON())
	t.EqualExit(nil, err)
	name := GetJSON("name").String()
	t.EqualExit("zls-cli", name)
	_ = SetJSON("test", "ok")
	t.EqualExit("ok", GetJSON("test").String())
	_ = SaveJSON("test.json", GetJSON().String())

	json, _ = JSON("test.json")
	t.EqualExit(name, json.Get("name").String())
	zfile.Rmdir("test.json")
}
