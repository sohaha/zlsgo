/*
 * @Author: seekwe
 * @Date:   2019-05-30 12:43:26
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-30 12:46:39
 */

package zhttp

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

func Do(method, url string, v ...interface{}) (*Res, error) {
	return std.Do(method, url, v...)
}
