package ztime

import (
	"fmt"
	"time"
)

func Diff(t1, t2 string, format ...string) (time.Duration, error) {
	t1t, err := Parse(t1, format...)
	if err != nil {
		return time.Duration(0), err
	}

	t2t, err := Parse(t2, format...)
	if err != nil {
		return time.Duration(0), err
	}

	return t2t.Sub(t1t), nil
}

// FindRange parses a slice of time strings and returns the earliest and latest time.
func FindRange(times []string, format ...string) (time.Time, time.Time, error) {
	if len(times) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("empty time slice")
	}

	if len(times) == 1 {
		t, err := Parse(times[0], format...)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return t, t, nil
	}

	var minTime, maxTime time.Time
	var minUnix, maxUnix int64
	var minNano, maxNano int
	var initialized bool

	for i := 0; i < len(times); i++ {
		t, err := Parse(times[i], format...)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		if !initialized {
			minTime = t
			maxTime = t
			minUnix = t.Unix()
			maxUnix = minUnix
			minNano = t.Nanosecond()
			maxNano = minNano
			initialized = true
			continue
		}

		tUnix := t.Unix()
		tNano := t.Nanosecond()

		if tUnix < minUnix || (tUnix == minUnix && tNano < minNano) {
			minTime = t
			minUnix = tUnix
			minNano = tNano
		}

		if tUnix > maxUnix || (tUnix == maxUnix && tNano > maxNano) {
			maxTime = t
			maxUnix = tUnix
			maxNano = tNano
		}
	}

	return minTime, maxTime, nil
}
