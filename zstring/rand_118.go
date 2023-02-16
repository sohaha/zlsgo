//go:build go1.18
// +build go1.18

package zstring

import (
	_ "unsafe"
)

//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

func RandUint32() uint32 {
	return fastrand()
}
