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
	sEpoch = 1474802888000
	// sWorkerIDBits Num of WorkerId Bits
	sWorkerIDBits   = 10
	sWorkerIDShift  = 12
	sTimeStampShift = 22
	// sequenceMask equal as getSequenceMask()
	sequenceMask = 0xfff
	sMaxWorker   = 0x3ff
)

// IDWorker Struct
type IDWorker struct {
	workerID      int64
	lastTimeStamp int64
	sequence      int64
	maxWorkerID   int64
	sync.RWMutex
}

// NewIDWorker Generate NewIDWorker with Given workerid
func NewIDWorker(workerid int64) (iw *IDWorker, err error) {
	iw = new(IDWorker)

	iw.maxWorkerID = getMaxWorkerID()

	if workerid > iw.maxWorkerID || workerid < 0 {
		return nil, errors.New("worker not fit")
	}
	iw.workerID = workerid
	iw.lastTimeStamp = -1
	iw.sequence = 0
	return iw, nil
}

func getMaxWorkerID() int64 {
	return -1 ^ -1<<sWorkerIDBits
}

func (iw *IDWorker) timeGen() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}

func (iw *IDWorker) timeReGen(last int64) int64 {
	ts := iw.timeGen()
	for {
		if ts < last {
			ts = iw.timeGen()
		} else {
			break
		}
	}
	return ts
}

// ID Generate next id
func (iw *IDWorker) ID() (ts int64, err error) {
	iw.Lock()
	defer iw.Unlock()
	ts = iw.timeGen()
	if ts == iw.lastTimeStamp {
		iw.sequence = (iw.sequence + 1) & sequenceMask
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
	ts = (ts-sEpoch)<<sTimeStampShift | iw.workerID<<sWorkerIDShift | iw.sequence
	return ts, nil
}

// ParseID reverse uid to timestamp, workid, seq
func ParseID(id int64) (t time.Time, ts int64, workerId int64, seq int64) {
	seq = id & sequenceMask
	workerId = (id >> sWorkerIDShift) & sMaxWorker
	ts = (id >> sTimeStampShift) + sEpoch
	t = time.Unix(ts/1000, (ts%1000)*1000000)
	return
}
