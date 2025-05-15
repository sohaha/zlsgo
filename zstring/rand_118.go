//go:build go1.18
// +build go1.18

package zstring

import (
	_ "unsafe"
)

// fastrand is linked directly to the Go runtime's internal fast random number generator.
// This provides better performance than crypto/rand for non-cryptographic purposes.
//
//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

// RandUint32 returns a pseudorandom uint32 value using the Go runtime's internal
// random number generator, which is much faster than crypto/rand for non-security-critical uses.
func RandUint32() uint32 {
	return fastrand()
}
