//go:build !windows
// +build !windows

package zutil

func IsDoubleClickStartUp() bool {
	return false
}

func GetParentProcessName() (string, error) {
	return "", nil
}
