package realip

import (
	"testing"
	"time"
)

func TestGetCloudflare(t *testing.T) {
	t.Log(GetCloudflare())
	t.Log(GetCloudflare(time.Microsecond))
}
