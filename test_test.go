package zlsgo

import "testing"

func TestNewTest(T *testing.T) {
	t := NewTest(T)
	t.Equal(1, 1)
	t.EqualExit(1, 1)
	t.EqualTrue(true)
	t.EqualNil(nil)
	t.Log("ok")
}
