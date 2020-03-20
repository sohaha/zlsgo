package zstring

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestMd5(t *testing.T) {
	tt := zlsgo.NewTest(t)
	str := "zlsgo"
	tt.Equal("e058a5f5dd76183d00d902c61c250fe3", Md5(str))
	t.Log(Md5File("./md5.go"))
	t.Log(Md5File("./md5.go.bak"))
}
