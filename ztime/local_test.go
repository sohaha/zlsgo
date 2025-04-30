package ztime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztime"
)

type demo struct {
	BirthdayLocal ztime.LocalTime
	Birthday      time.Time
	Name          string
}

func TestLocalTime(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ztime.SetTimeZone(0)

	now := time.Now()
	lt := ztime.LocalTime{now}
	tt.Equal(now.Unix(), lt.Unix())

	j, err := json.Marshal(lt)
	tt.NoError(err)
	tt.Log(string(j))

	nj, err := json.Marshal(now)
	tt.NoError(err)
	tt.Log(string(nj))

	data := demo{Name: "anna", BirthdayLocal: lt, Birthday: now}
	dj, err := json.Marshal(data)
	tt.NoError(err)
	tt.Log(string(dj))

	v, err := lt.Value()
	tt.NoError(err)
	tt.Log(v)

	nt, _ := ztime.Parse("2021-01-01 00:00:00")
	err = lt.Scan(nt)
	tt.NoError(err)

	j2, err := lt.MarshalJSON()
	tt.NoError(err)
	tt.Log(string(j2))
	tt.EqualTrue(string(j2) != string(nj))

	lt3 := ztime.LocalTime{}
	lt3.Scan(data.Birthday)
	tt.Log(lt3.String())

	lt3.Scan(data.BirthdayLocal)
	tt.Log(lt3.String())
}
