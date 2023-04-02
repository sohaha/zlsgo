package zhttp

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestSSE(t *testing.T) {
	tt := zlsgo.NewTest(t)

	time.Sleep(time.Second)

	s := SSE("http://127.0.0.1:18181/sse")
	i := 0

	s.OnMessage(func(ev *SSEEvent, err error) {
		if err != nil {
			t.Error(err)

			return
		}
		t.Logf("id:%s msg:%s [%s] %s\n", ev.ID, string(ev.Data), ev.Event, ev.Undefined)
		i++
	})
	tt.Equal(2, i)
}
