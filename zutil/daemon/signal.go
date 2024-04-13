package daemon

import (
	"sync"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	singleSignal sync.Once
	single            = zutil.NewChan[bool]()
	singleNum    uint = 0
	singleLock   sync.Mutex
)

func singelDo() {
	singleSignal.Do(func() {
		go func() {
			kill := KillSignal()
			singleLock.Lock()
			for {
				if singleNum == 0 {
					break
				}

				singleNum--
				single.In() <- kill
			}
			single.Close()
			singleLock.Unlock()
		}()
	})
}

func SingleKillSignal() <-chan bool {
	singleLock.Lock()
	defer singleLock.Unlock()

	singleNum++
	singelDo()

	return single.Out()
}

func ReSingleKillSignal() {
	singleLock.Lock()
	defer singleLock.Unlock()

	if singleNum > 0 {
		return
	}

	single = zutil.NewChan[bool]()
	singleSignal = sync.Once{}

	singelDo()
}
