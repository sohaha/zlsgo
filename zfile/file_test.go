package zfile

import (
	"fmt"
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
	defer func() {
		_ = Remove(tmpFile)
	}()

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
	tt.Equal("0 B", size)

	RealPath("")

	t.Log(filepath.Glob(RealPath("/Users/seekwe/Code/Go/zlsgo\\*\\file.go")))

	dirPath := RealPathMkdir("../zfile", true)
	t.Log(dirPath)
	tIsDir := DirExist(dirPath)
	tt.Equal(true, tIsDir)

	dirPath = SafePath(dirPath, RealPath(".."))
	t.Log(dirPath, SafePath(dirPath))
	tt.Equal("/zfile", dirPath)

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

func TestProgramPath(t *testing.T) {
	tt := NewTest(t)
	ePath := ProgramPath(true)
	tt.Log(ePath)
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

func TestGetMimeType(t *testing.T) {
	tt := NewTest(t)

	m := GetMimeType("test.jpg", nil)
	tt.Log(m)
	tt.Equal("image/jpeg", m)

	m = GetMimeType("test.html", nil)
	tt.Log(m)
	tt.Equal("text/html", strings.Split(m, ";")[0])

	m = GetMimeType("test", []byte("<html></html>"))
	tt.Log(m)
	tt.Equal("text/html", strings.Split(m, ";")[0])
}

func TestPermissionDenied(t *testing.T) {
	tt := NewTest(t)
	dir := TmpPath() + "/permission_denied"

	os.Mkdir(dir, 0o000)

	tt.EqualTrue(!HasReadWritePermission(dir))
	tt.EqualTrue(!HasReadWritePermission(dir + "/ddd"))

	tt.EqualTrue(HasReadWritePermission("./"))

	tt.EqualTrue(HasReadWritePermission("./ddd2"))

	tt.EqualTrue(!HasPermission(dir, 0o400))
	tt.EqualTrue(!HasPermission(dir+"/ddd2", 0o400))
	tt.EqualTrue(HasPermission("./ddd2", 0o400))
	tt.EqualTrue(!HasPermission("./ddd2", 0o400, true))
	tt.Log(HasPermission("/", 0o664))

	os.Chmod(dir, 0o777)
	tt.EqualTrue(Rmdir(dir))
}

func TestGetDirSize(t *testing.T) {
	tt := NewTest(t)

	tmp := RealPathMkdir("./tmp-size")
	defer Rmdir(tmp)

	for i := 0; i < 20; i++ {
		WriteFile(filepath.Join(tmp, fmt.Sprintf("file-%d.txt", i)), []byte(strings.Repeat("a", (i+1)*8*KB)))
	}

	size, total, err := StatDir(tmp)
	tt.NoError(err)
	tt.EqualTrue(size > 1*MB)
	tt.EqualTrue(total == 20)
	tt.Log(SizeFormat(int64(size)), size)

	size, total, err = StatDir(tmp, DirStatOptions{MaxSize: 1 * MB, MaxTotal: 3})
	tt.NoError(err)
	tt.EqualTrue(size < 1*MB)
	tt.EqualTrue(total == 3)
	tt.Log(SizeFormat(int64(size)), size)
}

func TestExecutablePath(t *testing.T) {
	tt := NewTest(t)
	tt.Log(ExecutablePath())
}
