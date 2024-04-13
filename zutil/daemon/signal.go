package daemon

import (
	"sync"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	singleSignal sync.Once
	single       = make(chan bool)
	singleNum    = zutil.NewInt32(0)
)

func singelDo() {
	singleSignal.Do(func() {
		go func() {
			kill := KillSignal()
			for {
				if int(singleNum.Sub(1)) < 0 {
					singleNum.Add(1)
					break
				}
				single <- kill
			}
			close(single)
		}()
	})
}

func SingleKillSignal() <-chan bool {
	singleNum.Add(1)
	singelDo()

	return single
}

func ReSingleKillSignal() {
	if singleNum.Load() > 0 {
		return
	}

	single = make(chan bool)
	singleSignal = sync.Once{}

	singelDo()
}
