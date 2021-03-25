package zcli

import (
	"testing"
)

func TestService(t *testing.T) {
	s, err := LaunchService("test", "", func() {
		t.Log("TestService")
	})
	t.Log(s, err)
}
