package daemon

import (
	"sync"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	// singleSignal ensures the signal handler is initialized only once
	singleSignal sync.Once
	// single is a channel for broadcasting kill signals to multiple subscribers
	single = zutil.NewChan[bool]()
	// singleNum tracks the number of active subscribers to the kill signal
	singleNum uint = 0
	// singleLock protects access to the shared signal handling state
	singleLock sync.Mutex
)

// singelDo initializes the signal handler goroutine if it hasn't been initialized yet.
// The goroutine waits for a kill signal and broadcasts it to all subscribers.
// This is an internal function used by SingleKillSignal and ReSingleKillSignal.
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

// SingleKillSignal returns a channel that will receive a value when the process
// receives a termination signal (such as SIGTERM or SIGINT).
// Multiple goroutines can call this function to receive the same signal.
// The channel will receive true if the signal was caught, false otherwise.
//
// Returns:
//   - <-chan bool: A channel that will receive a value when a kill signal is received
func SingleKillSignal() <-chan bool {
	singleLock.Lock()
	defer singleLock.Unlock()

	singleNum++
	singelDo()

	return single.Out()
}

// ReSingleKillSignal resets the signal handling system if there are no active subscribers.
// This allows the signal handling to be reused after all previous subscribers have been notified.
// If there are still active subscribers, this function does nothing.
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
