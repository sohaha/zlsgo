package limiter

import (
	"sync"
	"time"

	"github.com/sohaha/zlsgo/znet"
)

type singleRule struct {
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	allowed           int
	estimated         int
	records           []*circleQueue
	notRecordsIndex   map[int]struct{}
	usedRecordsIndex  sync.Map
	locker            *sync.Mutex
}

// newRule Initialize an access control policy
func newRule(defaultExpiration time.Duration, allowed int, estimated ...int) *singleRule {
	if allowed <= 0 {
		allowed = 1
	}
	userEstimated := 0
	if len(estimated) > 0 {
		userEstimated = estimated[0]
	}
	if userEstimated <= 0 {
		userEstimated = allowed
	}
	cleanupInterval := defaultExpiration / 100
	if cleanupInterval < time.Second*1 {
		cleanupInterval = time.Second * 1
	}
	if cleanupInterval > time.Second*60 {
		cleanupInterval = time.Second * 60
	}
	vc := createRule(defaultExpiration, cleanupInterval, allowed, userEstimated)
	go vc.deleteExpired()
	return vc
}

func createRule(defaultExpiration, cleanupInterval time.Duration, allowed, userEstimated int) *singleRule {
	var vc singleRule
	var locker sync.Mutex
	vc.defaultExpiration = defaultExpiration
	vc.cleanupInterval = cleanupInterval
	vc.allowed = allowed
	vc.estimated = userEstimated
	vc.notRecordsIndex = make(map[int]struct{})
	vc.locker = &locker
	vc.records = make([]*circleQueue, vc.estimated)
	for i := range vc.records {
		vc.records[i] = newCircleQueue(vc.allowed)
		vc.notRecordsIndex[i] = struct{}{}
	}
	return &vc

}

// allowVisit Whether access is allowed or not. If access is allowed, an access record is added to the access record
func (r *singleRule) allowVisit(key interface{}) bool {
	return r.add(key) == nil
}

// remainingVisits Remaining visits
func (r *singleRule) remainingVisits(key interface{}) int {
	if index, exist := r.usedRecordsIndex.Load(key); exist {
		r.records[index.(int)].deleteExpired()
		return r.records[index.(int)].unUsedSize()
	}
	return r.allowed
}

// remainingVisitsIP Remaining access times of an IP
func (r *singleRule) remainingVisitsIP(ip string) int {
	i, _ := znet.IPToLong(ip)
	if i == 0 {
		return 0
	}
	return r.remainingVisits(i)
}

// add Add an access record
func (r *singleRule) add(key interface{}) (err error) {
	r.locker.Lock()
	defer r.locker.Unlock()

	if index, exist := r.usedRecordsIndex.Load(key); exist {
		r.records[index.(int)].deleteExpired()
		return r.records[index.(int)].push(time.Now().Add(r.defaultExpiration).UnixNano())
	}

	if len(r.notRecordsIndex) > 0 {
		for index := range r.notRecordsIndex {
			delete(r.notRecordsIndex, index)
			r.usedRecordsIndex.Store(key, index)
			return r.records[index].push(time.Now().Add(r.defaultExpiration).UnixNano())
		}
	}
	queue := newCircleQueue(r.allowed)
	r.records = append(r.records, queue)
	index := len(r.records) - 1
	r.usedRecordsIndex.Store(key, index)
	return r.records[index].push(time.Now().Add(r.defaultExpiration).UnixNano())
}

// deleteExpired Delete expired data
func (r *singleRule) deleteExpired() {
	finished := true
	for range time.Tick(r.cleanupInterval) {
		if finished {
			finished = false
			r.deleteExpiredOnce()
			r.recovery()
			finished = true
		}
	}
}

// deleteExpiredOnce Delete expired data once in a specific time interval
func (r *singleRule) deleteExpiredOnce() {
	r.usedRecordsIndex.Range(func(k, v interface{}) bool {
		r.locker.Lock()
		index := v.(int)
		if index < len(r.records) && index >= 0 {
			r.records[index].deleteExpired()
			if r.records[index].usedSize() == 0 {
				r.usedRecordsIndex.Delete(k)
				r.notRecordsIndex[index] = struct{}{}
			}
		} else {
			r.usedRecordsIndex.Delete(k)
		}
		r.locker.Unlock()
		return true
	})
}

func (r *singleRule) recovery() {
	r.locker.Lock()
	defer r.locker.Unlock()
	if r.needRecovery() {
		curLen := len(r.records)
		unUsedLen := len(r.notRecordsIndex)
		usedLen := curLen - unUsedLen
		var newLen int
		if usedLen < r.estimated {
			newLen = r.estimated
		} else {
			newLen = usedLen * 2
		}
		visitorRecordsNew := make([]*circleQueue, newLen)
		for i := range visitorRecordsNew {
			visitorRecordsNew[i] = newCircleQueue(r.allowed)
		}
		r.notRecordsIndex = make(map[int]struct{})
		indexNew := 0
		r.usedRecordsIndex.Range(func(k, v interface{}) bool {
			indexOld := v.(int)
			visitorRecordsNew[indexNew] = r.records[indexOld]
			indexNew++
			return true
		})
		r.records = visitorRecordsNew
		for index := range r.records {
			if index >= indexNew {
				r.notRecordsIndex[index] = struct{}{}
			}
		}
	}
}

func (r *singleRule) needRecovery() bool {
	curLen := len(r.records)
	unUsedLen := len(r.notRecordsIndex)
	usedLen := curLen - unUsedLen
	if curLen < 2*r.estimated {
		return false
	}
	if usedLen*2 < unUsedLen {
		return true
	}
	return false
}
