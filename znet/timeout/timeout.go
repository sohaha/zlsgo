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

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{}, 1)
		panicErr := make(chan interface{}, 1)

		go func() {
			defer func() {
				if err := recover(); err != nil {
					select {
					case panicErr <- err:
					default:
					}
					return
				}
				select {
				case done <- struct{}{}:
				default:
				}
			}()
			c.Next()
		}()

		select {
		case <-done:
			return
		case err := <-panicErr:
			panic(err)
		case <-ctx.Done():
			if len(custom) > 0 {
				custom[0](c)
				c.Abort()
			} else {
				c.Abort(http.StatusGatewayTimeout)
			}
			return
		}
	}
}
