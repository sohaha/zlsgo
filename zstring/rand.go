package zstring

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"strconv"
	"time"
)

// RandUint32Max returns pseudorandom uint32 in the range [0..max)
func RandUint32Max(max uint32) uint32 {
	x := RandUint32()
	return uint32((uint64(x) * uint64(max)) >> 32)
}

// RandInt random numbers in the specified range
func RandInt(min int, max int) int {
	if max < min {
		max = min
	}
	return min + int(RandUint32Max(uint32(max+1-min)))
}

// Rand random string of specified length, the second parameter limit can only appear the specified character
func Rand(n int, tpl ...string) string {
	var s string
	b := make([]byte, n)
	if len(tpl) > 0 {
		s = tpl[0]
	} else {
		s = letterBytes
	}
	l := len(s) - 1
	for i := n - 1; i >= 0; i-- {
		idx := RandInt(0, l)
		b[i] = s[idx]
	}
	return Bytes2String(b)
}

// UniqueID unique id minimum 6 digits
func UniqueID(n int) string {
	if n < 6 {
		n = 6
	}
	k := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return Rand(n-2) + strconv.Itoa(time.Now().Nanosecond()/10000000)
	}
	return hex.EncodeToString(k)
}

var idWorkers, _ = NewIDWorker(0)

func UUID() int64 {
	id, _ := idWorkers.ID()
	return id
}
