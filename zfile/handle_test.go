package zfile

import (
	"os"
	"testing"

	. "github.com/sohaha/zlsgo"
)

func TestCopy(tt *testing.T) {
	t := NewTest(tt)
	dest := RealPathMkdir("../tmp", true)
	defer Rmdir(dest)
	err := CopyFile("../doc.go", dest+"tmp.tmp")
	t.Equal(nil, err)
	err = CopyDir("../znet", dest, func(srcFilePath, destFilePath string) bool {
		return srcFilePath == "../znet/timeout/timeout.go"
	})
	t.Equal(nil, err)
}

func TestRW(t *testing.T) {
	var err error
	var text []byte
	tt := NewTest(t)
	str := []byte("666")

	_ = WriteFile("./text.txt", str)
	text, err = ReadFile("./text.txt")
	tt.EqualNil(err)
	tt.Equal(str, text)
	t.Log(string(text))

	_ = WriteFile("./text.txt", str, true)
	text, err = ReadFile("./text.txt")
	tt.EqualNil(err)
	t.Log(string(text))
	tt.Equal([]byte("666666"), text)

	_ = WriteFile("./text.txt", str)
	text, err = ReadFile("./text.txt")
	tt.EqualNil(err)
	t.Log(string(text))
	tt.Equal(str, text)
	_ = os.Remove("./text.txt")
}

func TestReadLineFile(t *testing.T) {
	_ = WriteFile("./TestReadLineFile.txt", []byte("111\n2222\nTestReadLineFile\n88"))
	defer os.Remove("./TestReadLineFile.txt")
	tt := NewTest(t)
	file := "./TestReadLineFile.txt"
	err := ReadLineFile(file, func(line int, data []byte) error{
		t.Log(line, string(data))
		return nil
	})
	tt.EqualNil(err)
}
