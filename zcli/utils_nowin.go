// +build !windows

package zcli

func IsDoubleClickStartUp() bool {
	return false
}

func GetParentProcessName() (string, error) {
	return "", nil
}
