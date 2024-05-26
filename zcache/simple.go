package zcache

import (
	"time"

	"github.com/sohaha/zlsgo/ztype"
)

var simple = NewFast()

func Set(key string, val interface{}, expiration ...time.Duration) {
	simple.Set(key, val, expiration...)
}

func Delete(key string) {
	simple.Delete(key)
}

func Get(key string) (interface{}, bool) {
	return simple.Get(key)
}

func GetAny(key string) (ztype.Type, bool) {
	return simple.GetAny(key)
}

func ProvideGet(key string, provide func() (interface{}, bool), expiration ...time.Duration) (interface{}, bool) {
	return simple.ProvideGet(key, provide, expiration...)
}
