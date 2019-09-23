package zcfg

import (
	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"testing"
)

func TestJSONCfg(T *testing.T) {
	t := zlsgo.NewTest(T)
	remote := "https://raw.githubusercontent.com/sohaha/ZlsPHP/master/composer.json"
	json, err := JSONCfg(remote)
	t.EqualExit(json, GetJSON())
	t.EqualExit(nil, err)
	name := GetJSON("name").String()
	t.EqualExit("zls/zls", name)
	_ = SetJSON("test", "ok")
	t.EqualExit("ok", GetJSON("test").String())
	_ = SaveJSON("test.json", GetJSON().String())

	json, _ = JSONCfg("test.json")
	t.EqualExit(name, json.Get("name").String())
	zfile.Rmdir("test.json")
}
