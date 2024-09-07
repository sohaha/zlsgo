package zfile_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
)

func Test_Byte(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(1024, zfile.KB)
	tt.Equal(1048576, zfile.MB)
	tt.Equal(1073741824, zfile.GB)
	tt.Equal(1099511627776, zfile.TB)

	tt.Equal("2.0 KB", zfile.SizeFormat(2*zfile.KB))
	tt.Equal("2.0 MB", zfile.SizeFormat(2*zfile.MB))
	tt.Equal("2.0 GB", zfile.SizeFormat(2*zfile.GB))
	tt.Equal("2.0 TB", zfile.SizeFormat(2*zfile.TB))
}
