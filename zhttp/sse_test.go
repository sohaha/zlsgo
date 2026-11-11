package zhttp

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestSSE(t *testing.T) {
	tt := zlsgo.NewTest(t)

	time.Sleep(time.Second)

	s, err := SSE(testServerURL+"/sse", NoRedirect(true))
	tt.NoError(err, true)
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
