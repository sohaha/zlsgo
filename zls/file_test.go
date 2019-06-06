/*
 * @Author: seekwe
 * @Date:   2019-05-09 13:08:23
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-25 14:15:18
 */

package zls

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestFile(t *testing.T) {
	T := zlsgo.NewTest(t)
	dirPath := "."
	tIsDir := DirExist(dirPath)
	T.Equal(true, tIsDir)

	filePath := "../doc.go"
	tIsFile := FileExist(filePath)
	T.Equal(true, tIsFile)

	notPath := "zlsgo.php"
	status, _ := PathExist(notPath)
	T.Equal(0, status)

	size := FileSize("../doc.go")
	T.Equal("0 B" != size, true)

	size = FileSize("../_doc.go")
	T.Equal("0 B" == size, true)
}
