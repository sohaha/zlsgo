package zlsgo

import "testing"

func TestNewTest(t *testing.T) {
	tt := NewTest(t)

	tt.Equal(1, 1)
	tt.EqualExit(1, 1)
	tt.EqualTrue(true)
	tt.EqualNil(nil)
	tt.NoError(nil)
	tt.Log("ok")
}
