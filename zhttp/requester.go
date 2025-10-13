package zhttp

import (
	"time"
)

func DisableChunked(enable ...bool) {
	std.DisableChunked(enable...)
}

func Get(url string, v ...interface{}) (*Res, error) {
	return std.Get(url, v...)
}

func Post(url string, v ...interface{}) (*Res, error) {
	return std.Post(url, v...)
}

func Put(url string, v ...interface{}) (*Res, error) {
	return std.Put(url, v...)
}

func Head(url string, v ...interface{}) (*Res, error) {
	return std.Head(url, v...)
}

func Options(url string, v ...interface{}) (*Res, error) {
	return std.Options(url, v...)
}

func Delete(url string, v ...interface{}) (*Res, error) {
	return std.Delete(url, v...)
}

func Patch(url string, v ...interface{}) (*Res, error) {
	return std.Patch(url, v...)
}

func Connect(url string, v ...interface{}) (*Res, error) {
	return std.Connect(url, v...)
}

func Trace(url string, v ...interface{}) (*Res, error) {
	return std.Trace(url, v...)
}

func Do(method, rawurl string, v ...interface{}) (resp *Res, err error) {
	return std.Do(method, rawurl, v...)
}

func DoRetry(attempt int, sleep time.Duration, fn func() (*Res, error)) (*Res, error) {
	return std.DoRetry(attempt, sleep, fn)
}
