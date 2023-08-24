/*
Package zreflect provides reflection tools

	package main

	import (
		"github.com/sohaha/zlsgo/zreflect"
	)

	type Demo struct {
		Name string
	}

	func main() {
		typ := zreflect.TypeOf(Demo{})

		println(typ.NumMethod())
	}
*/
package zreflect
