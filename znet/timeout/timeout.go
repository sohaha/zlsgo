package timeout

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sohaha/zlsgo/znet"
)

func New(waitingTime time.Duration, custom ...znet.HandlerFunc) znet.HandlerFunc {
	return func(c *znet.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), waitingTime)
		defer cancel()
		finish := make(chan error, 1)
		go func(c *znet.Context) {
			defer func() {
				if err := recover(); err != nil {
					errMsg, ok := err.(error)
					if !ok {
						errMsg = errors.New(fmt.Sprint(err))
					}
					finish <- errMsg
				} else {
					finish <- nil
				}
			}()
			c.Next()
		}(c)
		select {
		case err := <-finish:
			if err != nil {
				panic(err)
			}
		case <-ctx.Done():
			c.Abort(http.StatusGatewayTimeout)
			if len(custom) > 0 {
				custom[0](c)
			}
		}
	}
}
