package zstring

import (
	"errors"
	"sync"
	"time"
)

// The algorithm is inspired by Twitter's famous snowflake
// its link is: https://github.com/twitter/snowflake/releases/tag/snowflake-2010
// timestamp(ms)42  | worker id(10) | sequence(12)

const (
	SEpoch          = 1474802888000
	SWorkerIdBits   = 10
	SWorkerIdShift  = 12
	STimeStampShift = 22
	SequenceMask    = 0xfff
	SMaxWorker      = 0x3ff
)

// IdWorker Struct
type IdWorker struct {
	workerId      int64
	lastTimeStamp int64
	sequence      int64
	maxWorkerId   int64
	sync.RWMutex
}

// NewIdWorker Generate NewIdWorker with Given workerid
func NewIdWorker(workerid int64) (iw *IdWorker, err error) {
	iw = new(IdWorker)

	iw.maxWorkerId = getMaxWorkerId()

	if workerid > iw.maxWorkerId || workerid < 0 {
		return nil, errors.New("worker not fit")
	}
	iw.workerId = workerid
	iw.lastTimeStamp = -1
	iw.sequence = 0
	// iw.lock = new(sync.RWMutex)
	return iw, nil
}

func getMaxWorkerId() int64 {
	return -1 ^ -1<<SWorkerIdBits
}

func (iw *IdWorker) timeGen() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}

func (iw *IdWorker) timeReGen(last int64) int64 {
	ts := iw.timeGen()
	println(ts, last, last-ts)
	for {
		if ts < last {
			ts = iw.timeGen()
		} else {
			break
		}
	}
	return ts
}

// Id Generate next id
func (iw *IdWorker) Id() (ts int64, err error) {
	iw.Lock()
	defer iw.Unlock()
	ts = iw.timeGen()
	if ts == iw.lastTimeStamp {
		iw.sequence = (iw.sequence + 1) & SequenceMask
		if iw.sequence == 0 {
			ts = iw.timeReGen(ts)
		}
	} else {
		iw.sequence = 0
	}

	if ts < iw.lastTimeStamp {
		err = errors.New("clock moved backwards, refuse gen id")
		return 0, err
	}
	iw.lastTimeStamp = ts
	ts = (ts-SEpoch)<<STimeStampShift | iw.workerId<<SWorkerIdShift | iw.sequence
	return ts, nil
}

// ParseId reverse uid to timestamp, workid, seq
func ParseId(id int64) (t time.Time, ts int64, workerId int64, seq int64) {
	seq = id & SequenceMask
	workerId = (id >> SWorkerIdShift) & SMaxWorker
	ts = (id >> STimeStampShift) + SEpoch
	t = time.Unix(ts/1000, (ts%1000)*1000000)
	return
}
