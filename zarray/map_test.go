//go:build go1.18
// +build go1.18

package zarray

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestKeys(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(3, len(Keys(map[int]int{1: 1, 2: 2, 3: 3})))
	tt.Equal(3, len(Keys(map[int]interface{}{1: 1, 2: "2", 3: 3})))
}

func TestValues(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(3, len(Values(map[int]int{1: 1, 2: 2, 3: 3})))
	tt.Equal(3, len(Values(map[int]interface{}{1: 1, 2: "2", 3: 3})))
}
