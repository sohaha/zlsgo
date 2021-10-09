//go:build !windows
// +build !windows

package zlog

// IsSupportColor IsSupportColor
func IsSupportColor() bool {
	return supportColor
}
