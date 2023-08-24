package zreflect

import (
	"time"
)

type (
	Child2 struct {
		P DemoChildSt
	}

	DemoChildSt struct {
		ChildName int
	}

	DemoSt struct {
		Date2  time.Time
		Child3 *DemoChildSt
		Remark string `json:"remark"`
		note   string
		Name   string `json:"username"`
		Slice  [][]string
		Hobby  []string
		Child  struct {
			Title       string `json:"child_user_title"`
			DemoChild2  Child2 `json:"demo_child_2"`
			IsChildName bool
		} `json:"child"`
		Year   float64
		Child2 Child2
		child4 DemoChildSt
		Age    uint
		Lovely bool
	}
	TestSt struct {
		Name string
		I    int `z:"iii"`
		Note int `json:"note,omitempty"`
	}
)

func (d DemoSt) Text() string {
	return d.Name + ":" + d.Remark
}
