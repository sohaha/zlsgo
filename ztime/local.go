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

func (t *LocalTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime{Time: value}
		return nil
	}

	return fmt.Errorf("expected time.Time, got %T", v)
}
