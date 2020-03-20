package zstring

import (
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"github.com/sohaha/zlsgo"
)

func TestBuffer(T *testing.T) {
	t := zlsgo.NewTest(T)
	l := "0"
	l += "1"
	b := Buffer()
	b.WriteString("0")
	b.WriteString("1")
	t.Equal(l, b.String())
}

func TestLen(T *testing.T) {
	t := zlsgo.NewTest(T)
	s := "我是中国人"
	t.Equal(5, Len(s))
	t.Log(Len(s), len(s))
}

func TestSubstr(T *testing.T) {
	t := zlsgo.NewTest(T)
	s := "0123"
	t.Equal(Substr(s, 1), "123")
	t.Equal(Substr(s, 2, 1), "2")
	t.Equal("A,是我呀", Substr("你好A,是我呀", 2))
	t.Equal("是我呀", Substr("你好A,是我呀", -3))
	t.Equal("A,是", Substr("你好A,是我呀", 2, -2))
}

func TestPad(T *testing.T) {
	t := zlsgo.NewTest(T)
	l := "我的这里一共8字"
	t.Equal(8, Len(l))

	s := "我的长度是二十,不够右边补零"
	t.Equal("我的长度是二十,不够右边补零000000", Pad(s, 20, "0", PadRight))

	s2 := "我的长度是二十,不够左边补零"
	t.Equal("000000我的长度是二十,不够左边补零", Pad(s2, 20, "0", PadLeft))

	s3 := "我的长度很长不需要填充"
	t.Equal("我的长度很长不需要填充", Pad(s3, 5, "我的长度很长不需要填充", PadRight))

	t.Equal("长度", Substr(s3, 2, 2))

	s4 := "我的长度是二十,不够两边补零"
	t.Equal("000我的长度是二十,不够两边补零000", Pad(s4, 20, "0", PadSides))
}

func TestFirst(T *testing.T) {
	t := zlsgo.NewTest(T)
	str := "my Name"
	str = Ucfirst(str)
	t.Equal(true, IsUcfirst(str))
	t.Equal(false, IsLcfirst(str))
	t.Equal("My Name", str)
	str = Lcfirst(str)
	t.Equal(true, IsLcfirst(str))
	t.Equal(false, IsUcfirst(str))
}

func TestSnakeCaseCamelCase(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Equal("", SnakeCaseToCamelCase("", true))
	t.Equal("HelloWorld", SnakeCaseToCamelCase("hello_world", true))
	t.Equal("helloWorld", SnakeCaseToCamelCase("hello_world", false))
	t.Equal("helloWorld", SnakeCaseToCamelCase("hello-world", false, "-"))

	t.Equal("", CamelCaseToSnakeCase(""))
	t.Equal("hello_world", CamelCaseToSnakeCase("HelloWorld"))
	t.Equal("hello_world", CamelCaseToSnakeCase("helloWorld"))
	t.Equal("hello-world", CamelCaseToSnakeCase("helloWorld", "-"))
}

func TestXss(T *testing.T) {
	t := zlsgo.NewTest(T)
	htmls := [][]string{
		{"", ""},
		{"Hello, World!", "Hello, World!"},
		{"foo&amp;bar", "foo&amp;bar"},
		{`Hello <a href="www.example.com/">World</a>!`, "Hello World!"},
		{"Foo <textarea>Bar</textarea> Baz", "Foo Bar Baz"},
		{"Foo <!-- Bar --> Baz", "Foo Baz"},
		{"<", "<"},
		{"foo < bar", "foo < bar"},
		{`Foo<script type="text/javascript">alert(1337)</script>Bar`, "FooBar"},
		{`Foo<div title="1>2">Bar`, "Foo2\">Bar"},
		{`I <3 Ponies!`, `I <3 Ponies!`},
		{`<script>foo()</script>`, ``},
		{`<script>foo()</script>`, ``},
	}
	for _, v := range htmls {
		t.Equal(v[1], XssClean(v[0]))
	}
}

func BenchmarkStr(b *testing.B) {
	s := ""
	for i := 0; i < b.N; i++ {
		s += "1"
	}
}

func BenchmarkStrBuffer(b *testing.B) {
	s := Buffer()
	for i := 0; i < b.N; i++ {
		s.WriteString("1")
	}
	_ = s.String()
}

func TestTo(T *testing.T) {
	t := zlsgo.NewTest(T)
	s := "我是中国人"
	b := String2Bytes(s)
	b = TrimBOM(b)
	s2 := Bytes2String(b)
	t.Equal(s, s2)
}

func BenchmarkTo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "我是中国人"
		b := String2Bytes(s)
		_ = Bytes2String(b)
	}
}

func BenchmarkTo2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "我是中国人"
		b := []byte(s)
		_ = string(b)
	}
}

func BenchmarkTo3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "我是中国人"
		b := []byte(s)
		_ = string(*(*[]byte)(unsafe.Pointer(&b)))
	}
}

func getRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func BenchmarkBuffer1(b *testing.B) {
	bb := Buffer()
	str := getRandomString(99999)
	for i := 0; i < b.N; i++ {
		bb.WriteString(str)
	}
}

func BenchmarkBuffer2(b *testing.B) {
	bb := Buffer(99999 * b.N)
	str := getRandomString(99999)
	for i := 0; i < b.N; i++ {
		bb.WriteString(str)
	}
}
