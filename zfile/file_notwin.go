//go:build !windows
// +build !windows

package zfile

import (
	"os"
)

// MoveFile moves a file from source to destination path.
// If force is true and destination exists, it will be removed before moving.
// On non-Windows systems, this uses os.Rename which is atomic if both paths
// are on the same filesystem.
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
