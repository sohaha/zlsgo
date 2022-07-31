package timeout

import (
	"context"
	"net/http"
	"time"

	"github.com/sohaha/zlsgo/znet"
)

func New(waitingTime time.Duration, custom ...znet.HandlerFunc) znet.HandlerFunc {
	return func(c *znet.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), waitingTime)
		defer cancel()
		done := make(chan struct{}, 1)
		go func() {
			c.Next()
			done <- struct{}{}
		}()
		for {
			select {
			case _, _ = <-done:
				return
			case <-ctx.Done():
				if len(custom) > 0 {
					custom[0](c)
				} else {
					c.Abort(http.StatusGatewayTimeout)
				}
				return
			}
		}
	}
}
