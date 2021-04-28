package zcli

import "testing"

func TestDaemon(t *testing.T) {
	quit, err := Daemon()
	t.Log(quit, err)
}
