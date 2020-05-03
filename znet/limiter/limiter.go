package limiter

import (
	"sync/atomic"

	"github.com/sohaha/zlsgo/znet"
)

func New(maxClients uint64, overflowFn func(c *znet.Context)) znet.HandlerFunc {
	var managerClients uint64 = 0
	return func(c *znet.Context) {
		process(&managerClients, maxClients, func(_ uint64) {
			c.Next()
		}, func(_ uint64) {
			overflowFn(c)
		})
	}
}

func process(managerClients *uint64, max uint64, fn func(managerClients uint64), overflowFn func(current uint64)) {
	s := atomic.LoadUint64(managerClients)
	if s >= max {
		overflowFn(s)
		return
	}
	atomic.AddUint64(managerClients, +1)
	fn(s)
	atomic.AddUint64(managerClients, ^uint64(0))
}
