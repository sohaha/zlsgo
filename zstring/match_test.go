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
	u = "/api/hello/kiki/k999"
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
	t.EqualExit(false, Match(z, "h*{大}当家"))
	t.EqualExit(false, Match(z, "h*大当家"))
	t.EqualExit(true, Match(z, "h*{小}当家"))
	t.EqualExit(true, Match(z, "h*小当家"))
	t.EqualExit(true, Match(z, "h*{小,大}当家"))
	t.EqualExit(false, Match(z, "h*{中,大}当家"))
	t.EqualExit(true, Match(z, "h*{小当家,大当家}"))
	t.EqualExit(false, Match(z, "h*{不当家,大当家}"))
	t.EqualExit(true, Match(z, "hell{o 中华小当,大当}家"))
	t.EqualExit(true, Match(z, "h{ll,ello} 中华小当家"))
	t.EqualExit(false, Match(z, "h{ll,ell} 中华小当家"))

	t.EqualExit(true, Match("超级马里{奥", "超级马里{奥"))
	t.EqualExit(false, Match("超级马里奥{", "超级马里{奥"))
	t.EqualExit(true, Match("超级马里奥{}", "超级马里奥{}"))
	t.EqualExit(true, Match("超级马里奥{1,2.3}!", "超级马里奥{1,2.3}!"))

	t.EqualExit(true, Match(u, "/api/hello/kiki/k999"))
	t.EqualExit(false, Match("/api/hi2/kiki/k999", "/api/{hello,hi}/kiki/k999"))
	t.EqualExit(true, Match("/api/hi/kiki/k999", "/api/{hello,hi}/kiki/k999"))
	t.EqualExit(true, Match(u, "/api/{hello,hi}/kiki/k999"))
	t.EqualExit(true, Match(u, "/api/*/kiki/*"))
	t.EqualExit(false, Match("/api/kiki/k999", "/api/*/kiki/*"))
	t.EqualExit(false, Match("/api/kiki/", "/api/*/kiki/*"))
	t.EqualExit(true, Match(u, "/api/*"))
	t.EqualExit(true, Match(u, "/api/**"))
	t.EqualExit(true, Match(u, "/api/*/*/k999"))
	t.EqualExit(true, Match(u, "/api/*/k999"))
	t.EqualExit(true, Match(u, "/api*k999"))
	t.EqualExit(true, Match(u, "/a*9"))
	t.EqualExit(false, Match(u, "a*9"))
	t.EqualExit(false, Match(u, "/a*8"))
	t.EqualExit(true, Match(u, "/api/hello/kiki/{k999,k1}"))
}

func TestMatch_fold(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualExit(false, Match(s, "hello P*"))
	tt.EqualExit(true, Match(s, "hello P*", true))

	tt.EqualExit(false, Match("你好呀 A!", "你好呀 a!"))
	tt.EqualExit(true, Match("你好呀 A!", "你好呀 a!", true))
}

func TestIsPattern(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Equal(true, IsPattern("hello ?ickup"))
	t.Equal(true, IsPattern("hello*"))
	t.Equal(false, IsPattern("hello pickup"))
}

func BenchmarkMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Match(z, "hello ki*")
	}
}

func BenchmarkMatch_fold(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Match(z, "hello ki*", true)
	}
}

func BenchmarkMatch1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Match(z, "he?lo ki*")
	}
}

func BenchmarkRegexMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RegexMatch(`hello ki*`, z)
	}
}

func BenchmarkContains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strings.Contains(z, "hello ki")
	}
}
