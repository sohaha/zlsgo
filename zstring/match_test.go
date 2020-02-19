package zstring

import (
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
)

var (
	z = "hello 中华小当家"
	s = "hello pickup"
	h = "hello kiki"
)

func TestMatch(T *testing.T) {
	t := zlsgo.NewTest(T)

	t.EqualExit(true, Match(s, s))
	t.EqualExit(false, Match(s, h))

	t.EqualExit(true, Match(s, "hello p*"))
	t.EqualExit(false, Match(s, "hello k*"))

	t.EqualExit(true, Match(s, "hello p?ckup"))
	t.EqualExit(true, Match(s, "he?lo?p?ck*p"))
	t.EqualExit(false, Match(s, "hello ?iki"))

	t.EqualExit(true, Match(s, "hello*?ckup"))
	t.EqualExit(false, Match(s, "hello*?iki"))

	t.EqualExit(true, Match(s, "*"))
	t.EqualExit(true, Match(z, "*"))

	t.EqualExit(true, Match(z, "*当家"))
	t.EqualExit(false, Match(z, "h?o*当家"))
}

func TestIsPattern(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Equal(true, IsPattern("hello ?ickup"))
	t.Equal(true, IsPattern("hello*"))
	t.Equal(false, IsPattern("hello pickup"))
}

func BenchmarkMatch1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Match(z, "hello *")
	}
}

func BenchmarkMatch2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RegexMatch(`hello`, z)
	}
}

func BenchmarkMatch3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strings.Contains(z, "hello ")
	}
}
