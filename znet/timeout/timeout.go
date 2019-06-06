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

	"github.com/sohaha/zlsgo/zls"
	"github.com/sohaha/zlsgo/znet"
)

type bodyWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func New(t time.Duration) znet.HandlerFunc {
	return func(c *znet.Context) {
		if c.Request.URL.Path != "/ip" {
			c.Next()
			return
		}
		buffer := zls.GetBuff()
		blw := &bodyWriter{body: buffer, ResponseWriter: c.Writer}
		finish := make(chan struct{}, 1)
		ctx, cancel := context.WithTimeout(c.Request.Context(), t)
		c.Writer = blw
		c.Request = c.Request.WithContext(ctx)
		go func() {
			c.Next()
			finish <- struct{}{}
		}()
		select {
		case <-ctx.Done():
			cancel()
			c.Abort(http.StatusGatewayTimeout)
			// It can't be released manually here.
			// zls.PutBuff(buffer)
		case <-finish:
			blw.ResponseWriter.Write(buffer.Bytes())
			zls.PutBuff(buffer)
		}
	}
}
