package fast_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo/zcache/fast"
)

func BenchmarkFastCacheSet(b *testing.B) {
	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = time.Hour
		o.AutoCleaner = true
		o.Bucket = 4
		o.Cap = 1024
	})
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + string(rune('A'+i%26))
			cache.Set(key, "value_"+string(rune('A'+i%26)))
			i++
		}
	})
}

func BenchmarkFastCacheGet(b *testing.B) {
	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = time.Hour
		o.AutoCleaner = true
		o.Bucket = 4
		o.Cap = 1024
	})
	defer cache.Close()

	for i := 0; i < 100; i++ {
		key := "key_" + string(rune('A'+i%26))
		cache.Set(key, "value_"+string(rune('A'+i%26)))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + string(rune('A'+i%26))
			cache.Get(key)
			i++
		}
	})
}

func BenchmarkFastCacheSetGet(b *testing.B) {
	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = time.Hour
		o.AutoCleaner = true
		o.Bucket = 4
		o.Cap = 1024
	})
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + string(rune('A'+i%26))
			if i%2 == 0 {
				cache.Set(key, "value_"+string(rune('A'+i%26)))
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}

func BenchmarkFastCacheWithExpiration(b *testing.B) {
	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 100 * time.Millisecond
		o.AutoCleaner = true
		o.Bucket = 4
		o.Cap = 1024
	})
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + string(rune('A'+i%26))
			cache.Set(key, "value_"+string(rune('A'+i%26)), 50*time.Millisecond)
			i++
		}
	})
}

func BenchmarkFastCacheNoAutoCleaner(b *testing.B) {
	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = time.Hour
		o.AutoCleaner = false
		o.Bucket = 4
		o.Cap = 1024
	})
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + string(rune('A'+i%26))
			if i%2 == 0 {
				cache.Set(key, "value_"+string(rune('A'+i%26)))
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}
