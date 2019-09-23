package zcfg

import (
	"errors"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"io/ioutil"
	"net/url"
	"path/filepath"
)

var res = zjson.Parse("{}")

func GetJSON(name ...string) zjson.Res {
	if len(name) > 0 {
		return res.Get(name[0])
	}
	return res
}

func SetJSON(name string, value interface{}) (err error) {
	str, err := zjson.Set(res.String(), name, value)
	if err == nil {
		res = zjson.Parse(str)
	}
	return
}

func SaveJSON(file, json string) (err error) {
	path, err := filepath.Abs(file)
	if err != nil {
		return
	}
	return ioutil.WriteFile(path, []byte(json), 0644)
}

func UnmarshalJSON(v interface{}) error {
	return res.Unmarshal(v)
}

func JSONCfg(cfgPath string) (zjson.Res, error) {
	var err error
	var json []byte
	if u, err := url.Parse(cfgPath); err == nil && u.Host != "" {
		json, err = GetRemoteCfgContent(cfgPath)
	} else {
		json, err = GetCfgContent(cfgPath)
	}
	if !zjson.ValidBytes(json) {
		err = errors.New("not valid json")
	} else {
		res = zjson.ParseBytes(json)
	}
	return res, err
}

func GetCfgContent(path string) (content []byte, err error) {
	file, err := filepath.Abs(path)
	if err != nil {
		return
	}
	content, err = ioutil.ReadFile(file)
	return
}

func GetRemoteCfgContent(url string) (content []byte, err error) {
	res, err := zhttp.Get(url)
	if err != nil {
		return
	}
	content = res.Bytes()
	return
}
