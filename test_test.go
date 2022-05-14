package zlsgo_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNewTest(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(1, 1)
	tt.EqualExit(1, 1)
	tt.EqualTrue(true)
	tt.EqualNil(nil)
	tt.NoError(nil)
	tt.Log("ok")
}
