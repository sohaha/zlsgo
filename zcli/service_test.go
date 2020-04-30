package zcli

import (
	"testing"
)

func TestService(t *testing.T) {
	_ = LaunchServiceRun("test", "", func() {
	})
}
