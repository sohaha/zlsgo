/*
 * @Author: seekwe
 * @Date:   2019-05-29 15:08:41
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 17:08:04
 */

package timeout

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/sohaha/zlsgo/znet"
)

type (
	bodyWriter struct {
		http.ResponseWriter
		body *bytes.Buffer
	}
)

func (w bodyWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func New(waitingTime time.Duration, custom ...znet.HandlerFunc) znet.HandlerFunc {
	return func(c *znet.Context) {
		ctx, _ := context.WithTimeout(c.Request.Context(), waitingTime)
		finish := make(chan struct{}, 1)
		go func(c *znet.Context) {
			c.Next()
			finish <- struct{}{}
		}(c)
		select {
		case <-finish:
		case <-ctx.Done():
			c.Abort(http.StatusGatewayTimeout)
			if len(custom) > 0 {
				custom[0](c)
			}
		}
	}
}
