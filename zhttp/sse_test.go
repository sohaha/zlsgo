package zhttp

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestSSE(t *testing.T) {
	tt := zlsgo.NewTest(t)

	time.Sleep(time.Second)

	s := SSE("http://127.0.0.1:18181/sse", NoRedirect(true))
	i := 0
	c, err := s.OnMessage(func(ev *SSEEvent) {
		t.Logf("id:%s msg:%s [%s] %s\n", ev.ID, string(ev.Data), ev.Event, ev.Undefined)
		i++
	})

	if err != nil {
		t.Error(err)
		return
	}
	<-c
	tt.Equal(2, i)
}
