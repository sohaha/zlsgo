package zfile

import (
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/sohaha/zlsgo"
)

func TestFile(T *testing.T) {
	t := NewTest(T)

	filePath := "../doc.go"
	tIsFile := FileExist(filePath)
	t.Equal(true, tIsFile)

	notPath := "zlsgo.php"
	status, _ := PathExist(notPath)
	t.Equal(0, status)

	size := FileSize("../doc.go")
	t.Equal("0 B" != size, true)

	size = FileSize("../_doc.go")
	t.Equal("0 B" == size, true)

	dirPath := RealPathMkdir("../zfile")
	tIsDir := DirExist(dirPath)
	t.Equal(true, tIsDir)

	dirPath = SafePath("../zfile/ok")
	t.Equal("ok", dirPath)

	path := RealPathMkdir("../tmp")
	path2 := RealPathMkdir(path + "/ooo")
	T.Log(path, path2)
	t.Equal(true, Rmdir(path, true))
	t.Equal(true, Rmdir(path))
	ePath := ProgramPath(true)
	ProjectPath = ePath
	path = RealPathMkdir("../ppppp")
	testPath := ePath + "../ppppp"
	T.Log(path, testPath)
	t.EqualTrue(DirExist(path))
	t.EqualTrue(DirExist(testPath))
	ok := Rmdir(testPath)

	T.Log(path, testPath, ok)
	var g sync.WaitGroup
	g.Add(1)
	// g.Wait()

}

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

func TestPut(t *testing.T) {
	var err error
	tt := NewTest(t)
	defer os.Remove("./text.txt")
	err = PutOffset("./text.txt", []byte(time.Now().String()+"\n"), 0)
	tt.EqualNil(err)
	err = PutAppend("./text.txt", []byte(time.Now().String()+"\n"))
	tt.EqualNil(err)
	os.Remove("./text.txt")
	err = PutAppend("./text.txt", []byte(time.Now().String()+"\n"))
	tt.EqualNil(err)
	err = PutOffset("./text.txt", []byte("\n(ok)\n"), 5)
	tt.EqualNil(err)
}
func TestRW(t *testing.T) {

	var err error
	var text []byte
	tt := NewTest(t)
	str := []byte("666")

	WriteFile("./text.txt", str)
	text, err = ReadFile("./text.txt")
	tt.EqualNil(err)
	tt.Equal(str, text)
	t.Log(string(text))

	WriteFile("./text.txt", str, true)
	text, err = ReadFile("./text.txt")
	tt.EqualNil(err)
	t.Log(string(text))
	tt.Equal([]byte("666666"), text)

	WriteFile("./text.txt", str)
	text, err = ReadFile("./text.txt")
	tt.EqualNil(err)
	t.Log(string(text))
	tt.Equal(str, text)
	os.Remove("./text.txt")
}
