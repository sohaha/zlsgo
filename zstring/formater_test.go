package zstring

import (
	"github.com/sohaha/zlsgo"
	"testing"
)

func TestSnakeCaseCamelCase(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Equal("",SnakeCaseToCamelCase("",true))
	t.Equal("HelloWorld",SnakeCaseToCamelCase("hello_world",true))
	t.Equal("helloWorld",SnakeCaseToCamelCase("hello_world",false))
	t.Equal("helloWorld",SnakeCaseToCamelCase("hello-world",false,"-"))

	t.Equal("",CamelCaseToSnakeCase(""))
	t.Equal("hello_world",CamelCaseToSnakeCase("HelloWorld"))
	t.Equal("hello_world",CamelCaseToSnakeCase("helloWorld"))
	t.Equal("hello-world",CamelCaseToSnakeCase("helloWorld","-"))
}
