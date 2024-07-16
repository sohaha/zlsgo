package ztime

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type LocalTime struct {
	time.Time
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + inlay.FormatTime(t.Time) + `"`), nil
}

func (t LocalTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}

	return t.Time, nil
}

func (t LocalTime) String() string {
	return inlay.FormatTime(t.Time)
}

func (t LocalTime) Format(layout string) string {
	return inlay.FormatTime(t.Time, layout)
}

func (t *LocalTime) Scan(v interface{}) error {
	switch vv := v.(type) {
	case time.Time:
		*t = LocalTime{Time: vv}
		return nil
	case LocalTime:
		*t = vv
		return nil
	default:
		return fmt.Errorf("expected time.Time, got %T", v)
	}
}
