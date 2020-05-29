package compress

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
)

func TestGz(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := zfile.PutAppend("./tmp/log.txt", []byte("ok\n"))
	tt.EqualNil(err)
	err = zfile.PutAppend("./tmp/tmp2/log.txt", []byte("ok\n"))
	tt.EqualNil(err)
	gz := "dd.tar.gz"
	err = Compress(".", gz)
	tt.EqualNil(err)
	err = DeCompress(gz, "tmp2")
	tt.EqualNil(err)
	err = DeCompress(gz+"1", "tmp2")
	tt.Equal(true, err != nil)

	zfile.Rmdir("tmp")
	zfile.Rmdir("tmp2")
	zfile.Rmdir(gz)
}
