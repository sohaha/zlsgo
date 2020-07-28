package zstring

import (
	"math/rand"
	"time"
)

// RandInt random numbers in the specified range
func RandInt(min int, max int) int {
	if max < min {
		max = min
	}
	rand.Seed(rand.Int63n(time.Now().UnixNano()))
	return min + rand.Intn(max+1-min)
}

// RandString random string of specified length, the second parameter limit can only appear the specified character
func Rand(n int, tpl ...string) string {
	var src = rand.NewSource(time.Now().UnixNano())
	var s string
	b := make([]byte, n)
	if len(tpl) > 0 {
		s = tpl[0]
	} else {
		s = letterBytes
	}
	for i, cache, remain := n-1, src.Int63(), 10; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), 10
		}
		b[i] = s[int(cache&letterIdxMask)%len(s)]
		i--
		cache >>= 6
		remain--
	}
	return Bytes2String(b)
}

var idWorkers, _ = NewIDWorker(0)

func UUID() int64 {
	id, _ := idWorkers.ID()
	return id
}
