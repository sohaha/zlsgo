package zstring

import (
	"errors"
	"sync"
	"time"
)

// The Snowflake algorithm is inspired by Twitter's famous snowflake implementation.
// Reference: https://github.com/twitter/snowflake/releases/tag/snowflake-2010
// ID format: timestamp(ms)42 bits | worker id(10 bits) | sequence(12 bits)

const (
	// sEpoch is the Snowflake epoch timestamp (milliseconds since UNIX epoch)
	sEpoch = 1474802888000
	// sWorkerIDBits is the number of bits allocated for worker ID
	sWorkerIDBits = 10
	// sWorkerIDShift is the bit shift for worker ID in the ID
	sWorkerIDShift = 12
	// sTimeStampShift is the bit shift for timestamp in the ID
	sTimeStampShift = 22
	// sequenceMask is the mask for sequence number (12 bits)
	sequenceMask = 0xfff
	// sMaxWorker is the maximum worker ID value
	sMaxWorker = 0x3ff
)

// IDWorker represents a Snowflake ID generator instance.
// Each worker generates unique IDs based on its worker ID.
type IDWorker struct {
	workerID      int64
	lastTimeStamp int64
	sequence      int64
	maxWorkerID   int64
	sync.RWMutex
}

// NewIDWorker creates a new Snowflake ID generator with the given worker ID.
// Returns an error if the worker ID is invalid (outside the allowed range).
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

// getMaxWorkerID calculates the maximum worker ID based on the allocated bits.
func getMaxWorkerID() int64 {
	return -1 ^ -1<<sWorkerIDBits
}

// timeGen returns the current timestamp in milliseconds.
func (iw *IDWorker) timeGen() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}

// timeReGen ensures the timestamp is greater than the last timestamp.
// It spins until a newer timestamp is obtained.
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

// ID generates the next unique ID.
// Returns the generated ID and any error that occurred during generation.
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

// ParseID extracts the components from a Snowflake ID.
// Returns the timestamp as a time.Time, raw timestamp value, worker ID, and sequence number.
func ParseID(id int64) (t time.Time, ts int64, workerId int64, seq int64) {
	seq = id & sequenceMask
	workerId = (id >> sWorkerIDShift) & sMaxWorker
	ts = (id >> sTimeStampShift) + sEpoch
	t = time.Unix(ts/1000, (ts%1000)*1000000)
	return
}
