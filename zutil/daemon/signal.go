package daemon

import (
	"sync"
)

var (
	singleSignal sync.Once
	single            = make(chan bool)
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
				single <- kill
			}
			close(single)
			singleLock.Unlock()
		}()
	})
}

func SingleKillSignal() <-chan bool {
	singleLock.Lock()
	defer singleLock.Unlock()

	singleNum++
	singelDo()

	return single
}

func ReSingleKillSignal() {
	singleLock.Lock()
	defer singleLock.Unlock()

	if singleNum > 0 {
		return
	}

	single = make(chan bool)
	singleSignal = sync.Once{}

	singelDo()
}
