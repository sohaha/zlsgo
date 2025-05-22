package znet

import (
	"net/http"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztime"
)

func Test_isModified(t *testing.T) {
	tt := zlsgo.NewTest(t)

	now := time.Now()
	c := &Context{
		Request: &http.Request{
			Header: http.Header{
				"If-Modified-Since": []string{ztime.In(now).Format("Mon, 02 Jan 2006 15:04:05 GMT")},
			},
		},
		header: map[string][]string{},
	}
	m := isModified(c, now)
	tt.Equal(false, m, true)

	m = isModified(c, now.Add(-time.Second))
	tt.Equal(true, m, true)

	c.Request.Header.Del("If-Modified-Since")
	m = isModified(c, now)
	tt.Equal(true, m, true)
}
