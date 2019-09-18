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
	"github.com/sohaha/zlsgo/zutil"
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
		buffer := zutil.GetBuff()
		blw := &bodyWriter{body: buffer, ResponseWriter: c.Writer}
		finish := make(chan struct{}, 1)
		ctx, cancel := context.WithTimeout(c.Request.Context(), waitingTime)
		c.Writer = blw
		c.Request = c.Request.WithContext(ctx)
		go func() {
			c.Next()
			finish <- struct{}{}
		}()
		select {
		case <-ctx.Done():
			c.Writer = blw.ResponseWriter
			cancel()
			if len(custom) > 0 {
				custom[0](c)
				zutil.PutBuff(buffer)
				c.Abort()
			} else {
				c.Abort(http.StatusGatewayTimeout)
			}
		case <-finish:
			_, _ = blw.ResponseWriter.Write(buffer.Bytes())
			zutil.PutBuff(buffer)
		}
		cancel()
	}
}
