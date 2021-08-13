package zfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/sohaha/zlsgo"
)

func TestFile(tt *testing.T) {
	t := NewTest(tt)

	tt.Log("c:" + RealPath("/"+"d"))

	baseDir := "c:\\test"
	fileName := "isPic"
	tt.Log(filepath.Join(baseDir, fileName))
	tt.Log(RealPath(baseDir + "/" + "d" + "/" + "fileName" + ".jpg"))
	tt.Log(RealPath(strings.Join([]string{"a", "b", "c", "dd\\ddd", ".jpg"}, "/")))

	_ = WriteFile("./tmp.tmp", []byte(""))
	defer os.RemoveAll("./tmp.tmp")

	filePath := "./tmp.tmp"
	tIsFile := FileExist(filePath)
	t.Equal(true, tIsFile)

	notPath := "zlsgo.php"
	status, _ := PathExist(notPath)
	t.Equal(0, status)

	size := FileSize(filePath)
	tt.Log(size)
	t.Equal("0 B" == size, true)

	RealPath("")

	tt.Log(filepath.Glob(RealPath("/Users/seekwe/Code/Go/zlsgo\\*\\file.go")))

	dirPath := RealPathMkdir("../zfile", true)
	tt.Log(dirPath)
	tIsDir := DirExist(dirPath)
	t.Equal(true, tIsDir)

	dirPath = SafePath(dirPath, RealPath(".."))
	tt.Log(dirPath, SafePath(dirPath))
	t.Equal("zfile/", dirPath)

	tmpPath := TmpPath("")
	t.EqualTrue(tmpPath != "")
	tt.Log(tmpPath, TmpPath("666"))

	path := RealPathMkdir("../tmp")
	path2 := RealPathMkdir(path + "/ooo")
	tt.Log(path, path2)
	t.Equal(true, Rmdir(path, true))
	t.Equal(true, Rmdir(path))
	ePath := ProgramPath(true)
	ProjectPath = ePath
	path = RealPathMkdir("../ppppp")
	testPath := ePath + "../ppppp"
	tt.Log(path, testPath)
	t.EqualTrue(DirExist(path))
	t.EqualTrue(DirExist(testPath))
	ok := Rmdir(testPath)
	Rmdir(path)
	tt.Log(path, testPath, ok)
}

func TestPut(t *testing.T) {
	var err error
	tt := NewTest(t)
	defer os.Remove("./text.txt")
	err = PutOffset("./text.txt", []byte(time.Now().String()+"\n"), 0)
	tt.EqualNil(err)
	err = PutAppend("./text.txt", []byte(time.Now().String()+"\n"))
	tt.EqualNil(err)
	_ = os.Remove("./text.txt")
	err = PutAppend("./put/text.txt", []byte(time.Now().String()+"\n"))
	tt.EqualNil(err)
	_ = os.Remove("./put/text.txt")
	err = PutAppend("./put2/text.txt", []byte(time.Now().String()+"\n"))
	tt.EqualNil(err)
	_ = os.Remove("./put2/text.txt")
	err = PutOffset("./text.txt", []byte("\n(ok)\n"), 5)
	tt.EqualNil(err)
}
