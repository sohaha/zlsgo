//go:build !windows
// +build !windows

package zfile

import (
	"os"
)

// MoveFile Move File
func MoveFile(source string, dest string, force ...bool) error {
	source = RealPath(source)
	dest = RealPath(dest)
	if len(force) > 0 && force[0] {
		if exist, _ := PathExist(dest); exist != 0 && source != dest {
			_ = os.RemoveAll(dest)
		}
	}
	return os.Rename(source, dest)
}
