package zcfg

import (
	"github.com/sohaha/zlsgo"
	"testing"
)

func TestIni(T *testing.T) {
	t := zlsgo.NewTest(T)
	res, _ := Ini("./conf/conf.ini")
	t.Log(res)
}
