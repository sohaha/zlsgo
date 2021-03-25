package zfile

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestGz(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := WriteFile("./tmp_gz/log.txt", []byte("ok\n"))
	tt.EqualNil(err)
	err = WriteFile("./tmp_gz/tmp2/log.txt", []byte("ok\n"))
	tt.EqualNil(err)
	gz := "dd.tar.gz"
	err = GzCompress(".", gz)
	tt.EqualNil(err)
	Rmdir("./tmp_gz")
	err = GzDeCompress(gz, "tmp2")
	tt.EqualNil(err)
	err = GzDeCompress(gz+"1", "tmp2")
	tt.Equal(true, err != nil)

	ok := Rmdir("tmp2")
	tt.EqualTrue(ok)
	Rmdir(gz)
}

func TestZip(t *testing.T) {
	tt := zlsgo.NewTest(t)
	zip := "tmp.zip"
	err := WriteFile("./tmp/log.txt", []byte("ok\n"))
	tt.EqualNil(err)
	err = WriteFile("./tmp/tmp2/log.txt", []byte("ok\n"))
	tt.EqualNil(err)
	err = ZipCompress("./", zip)
	tt.EqualNil(err)
	tt.EqualNil(ZipDeCompress(zip, "zip"))
	tt.EqualTrue(FileExist("./zip/tmp/log.txt"))
	Rmdir(zip)
	Rmdir("zip")
	Rmdir("./tmp")
}
