//go:build go1.18
// +build go1.18

package ztime

import (
	"fmt"
	"time"
)

func Diff[T time.Time | string](t1, t2 T, format ...string) (time.Duration, error) {
	t1t, err := parseGenericTime(t1, format...)
	if err != nil {
		return time.Duration(0), err
	}

	t2t, err := parseGenericTime(t2, format...)
	if err != nil {
		return time.Duration(0), err
	}

	return t2t.Sub(t1t), nil
}

// FindRange parses a slice of time strings and returns the earliest and latest time.
func FindRange[T time.Time | string](times []T, format ...string) (time.Time, time.Time, error) {
	if len(times) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("empty time slice")
	}

	if len(times) == 1 {
		t, err := parseGenericTime(times[0], format...)
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
		t, err := parseGenericTime(times[i], format...)
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

func parseGenericTime[T time.Time | string](t T, format ...string) (time.Time, error) {
	switch v := any(t).(type) {
	case time.Time:
		return v, nil
	case string:
		return Parse(v, format...)
	default:
		return time.Time{}, fmt.Errorf("unsupported type: %T", t)
	}
}

// Sequence generates a sequence of time strings between start and end times based on the given format.
func Sequence[T time.Time | string](start, end T, stepFn func(time.Time) time.Time, format ...string) ([]string, error) {
	startTime, err := parseGenericTime(start, format...)
	if err != nil {
		return nil, err
	}

	endTime, err := parseGenericTime(end, format...)
	if err != nil {
		return nil, err
	}

	if startTime.After(endTime) {
		return nil, fmt.Errorf("start time cannot be after end time")
	}

	tpl := TimeTpl
	if len(format) > 0 {
		tpl = FormatTlp(format[0])
	}

	result := make([]string, 0)
	current := startTime

	if stepFn == nil {
		stepFn = func(t time.Time) time.Time {
			return t.Add(time.Hour * 24)
		}
	}

	for !current.After(endTime) {
		result = append(result, inlay.In(current).Format(tpl))
		current = stepFn(current)
	}

	if len(result) > 0 {
		lastTime, _ := Parse(result[len(result)-1], format...)
		if !lastTime.Equal(endTime) && !endTime.Before(lastTime) {
			result = append(result, inlay.In(endTime).Format(tpl))
		}
	}

	return result, nil
}
