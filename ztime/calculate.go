package ztime

import "time"

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
