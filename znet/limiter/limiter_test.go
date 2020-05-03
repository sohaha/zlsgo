package limiter

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestProcess(T *testing.T) {
	t := zlsgo.NewTest(T)
	var managerClients uint64 = 0
	for i := 0; i < 1000; i++ {
		go process(&managerClients, 997,
			func(managerClients uint64) {
				time.Sleep(10 * time.Millisecond)
			},
			func(current uint64) {
				t.Log("溢出", current)
			})
	}
	time.Sleep(1 * time.Second)
}

func BenchmarkProcess(b *testing.B) {
	var managerClients uint64 = 0
	for i := 0; i < b.N; i++ {
		go process(&managerClients, 90,
			func(managerClients uint64) {
			},
			func(current uint64) {
			})
	}
}
