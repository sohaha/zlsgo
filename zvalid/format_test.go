package zvalid

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestFormat(t *testing.T) {
	tt := zlsgo.NewTest(t)
	str := Text(" is test ").Trim().Value()
	tt.Equal("is test", str)

	str = Text(" is test ").RemoveSpace().Value()
	tt.Equal("istest", str)

	str = Text("is test is").Replace("is", "yes", 1).Value()
	tt.Equal("yes test is", str)

	str = Text("is test is").ReplaceAll("is", "yes").Value()
	tt.Equal("yes test yes", str)

	str = Text("is <script> alert(666); </script> js").XssClean().Value()
	tt.Equal("is js", str)

	str = Text("hello_world").SnakeCaseToCamelCase(false).Value()
	tt.Equal("helloWorld", str)

	str = Text("hello_world").SnakeCaseToCamelCase(true).Value()
	tt.Equal("HelloWorld", str)

	str = Text("hello-world").SnakeCaseToCamelCase(true, "-").Value()
	tt.Equal("HelloWorld", str)

	str = Text("HelloWorld").CamelCaseToSnakeCase().Value()
	tt.Equal("hello_world", str)

	str = Text("helloWorld").CamelCaseToSnakeCase().Value()
	tt.Equal("hello_world", str)

	str = Text("helloWorld").CamelCaseToSnakeCase("-").Value()
	tt.Equal("hello-world", str)
}
