package zstring

import (
	"math/rand"
	"time"
)

// RandInt random numbers in the specified range
func RandInt(min int, max int) int {
	rand.Seed(rand.Int63n(time.Now().UnixNano()))
	return min + rand.Intn(max-min)
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

func UUID(workerid ...int64) int64 {
	var wid int64 = 0
	if len(workerid) > 0 {
		wid = workerid[0]
	}
	w, _ := NewIdWorker(wid)
	id, _ := w.Id()
	return id
}
