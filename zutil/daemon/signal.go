package daemon

import (
	"sync"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	singleSignal sync.Once
	single       = make(chan bool)
	singleNum    = zutil.NewUint32(0)
)

func SingleKillSignal() <-chan bool {
	singleNum.Add(1)
	singleSignal.Do(func() {
		go func() {
			kill := KillSignal()
			for i := singleNum.Load(); i > 0; i-- {
				single <- kill
			}
			close(single)
		}()
	})

	return single
}
