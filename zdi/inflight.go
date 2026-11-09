package zdi

import (
	"fmt"
	"runtime"
)

// inflightCall is one in-progress provider construction.
//
// singleflight cannot be used here: it is not reentrant, so a provider that
// resolves through the public API back to its own key would wait for itself
// forever. Tracking the owning goroutine turns that hang into an error.
type inflightCall struct {
	owner uint64
	done  chan struct{}
	err   error
}

// goid returns the current goroutine id.
//
// This is only used on the construction path (a provider runs at most once per
// container), never on the cached read path, so the runtime.Stack cost is not
// on any hot path.
func goid() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	s := buf[:n]
	const prefix = "goroutine "
	if len(s) < len(prefix) {
		return 0
	}
	s = s[len(prefix):]
	var id uint64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		id = id*10 + uint64(s[i]-'0')
	}
	return id
}

// beginBuild returns the in-flight construction for key.
//
// leader is true when the caller must run the provider. When the current
// goroutine already owns the construction, an error is returned instead of
// blocking on itself.
func (inj *injector) beginBuild(key string) (call *inflightCall, leader bool, err error) {
	id := goid()

	inj.mu.Lock()
	defer inj.mu.Unlock()

	if inflight, ok := inj.inflight[key]; ok {
		if inflight.owner == id {
			return nil, false, fmt.Errorf("zdi: circular dependency detected while resolving %s", key)
		}
		return inflight, false, nil
	}

	inflight := &inflightCall{owner: id, done: make(chan struct{})}
	inj.inflight[key] = inflight
	return inflight, true, nil
}

// endBuild publishes the construction result to every waiter and releases key.
func (inj *injector) endBuild(key string, call *inflightCall, err error) {
	inj.mu.Lock()
	call.err = err
	delete(inj.inflight, key)
	inj.mu.Unlock()
	close(call.done)
}

// runBuild executes fn as the leader of key, guaranteeing the key is released
// even if fn panics, so waiters can never block forever.
func (inj *injector) runBuild(key string, call *inflightCall, fn func() error) (err error) {
	defer func() {
		inj.endBuild(key, call, err)
	}()
	return fn()
}
