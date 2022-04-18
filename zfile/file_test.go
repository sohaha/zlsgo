package zfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/sohaha/zlsgo"
)

func TestFile(t *testing.T) {
	tt := NewTest(t)

	t.Log("c:" + RealPath("/"+"d"))

	baseDir := "c:\\test"
	fileName := "isPic"
	t.Log(filepath.Join(baseDir, fileName))
	t.Log(RealPath(baseDir + "/" + "d" + "/" + "fileName" + ".jpg"))
	t.Log(RealPath(strings.Join([]string{"a", "b", "c", "dd\\ddd", ".jpg"}, "/")))

	tmpFile := RealPath("./tmp.tmp")
	_ = WriteFile(tmpFile, []byte(""))
	defer Remove(tmpFile)

	filePath := "./tmp.tmp"
	tIsFile := FileExist(filePath)
	tt.Equal(true, tIsFile)

	err := MoveFile(filePath, filePath+".new")
	tt.EqualNil(err)
	tt.EqualTrue(!FileExist(filePath))
	tt.EqualTrue(FileExist(filePath + ".new"))
	_ = WriteFile("./tmp.tmp", []byte(""))
	err = MoveFile(filePath, filePath+".new", true)
	tt.EqualNil(err)
	err = MoveFile(filePath+".new", filePath+".new", true)
	tt.EqualNil(err)
	_ = MoveFile(filePath+".new", filePath)

	notPath := "zlsgo.php"
	status, _ := PathExist(notPath)
	tt.Equal(0, status)

	size := FileSize(filePath)
	t.Log(size)
	tt.Equal("0 B" == size, true)

	RealPath("")

	t.Log(filepath.Glob(RealPath("/Users/seekwe/Code/Go/zlsgo\\*\\file.go")))

	dirPath := RealPathMkdir("../zfile", true)
	t.Log(dirPath)
	tIsDir := DirExist(dirPath)
	tt.Equal(true, tIsDir)

	dirPath = SafePath(dirPath, RealPath(".."))
	t.Log(dirPath, SafePath(dirPath))
	tt.Equal("zfile", dirPath)

	tmpPath := TmpPath("")
	tt.EqualTrue(tmpPath != "")
	t.Log(tmpPath, TmpPath("666"))

	path := RealPathMkdir("../tmp")
	path2 := RealPathMkdir(path + "/ooo")
	t.Log(path, path2)
	tt.Equal(true, Rmdir(path, true))
	tt.Equal(true, Rmdir(path))
	ePath := ProgramPath(true)
	ProjectPath = ePath
	path = RealPathMkdir("../ppppp")
	testPath := ePath + "../ppppp"
	t.Log(path, testPath)
	tt.EqualTrue(DirExist(path))
	tt.EqualTrue(DirExist(testPath))
	ok := Rmdir(testPath)
	Rmdir(path)
	t.Log(path, testPath, ok)
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
