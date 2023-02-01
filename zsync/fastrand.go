//go:build go1.18
// +build go1.18

package zsync

import (
	_ "unsafe"
)

//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32
