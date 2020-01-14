package zcfg

import (
	"github.com/sohaha/zlsgo/zhttp"
	"io/ioutil"
	"path/filepath"
)

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
