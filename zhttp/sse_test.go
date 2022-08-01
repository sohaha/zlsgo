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
e:
	for {
		select {
		case <-s.Done():
			break e
		case ev := <-s.Event():
			t.Logf("id:%s msg:%s [%s]\n", ev.ID, string(ev.Data), ev.Event)
			i++
		}
	}
	tt.Equal(2, i)
}
