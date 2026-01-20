package realip

import (
	"testing"
	"time"
)

func TestGetCloudflare(t *testing.T) {
	t.Log(GetCloudflare())
	t.Log(GetCloudflare(time.Microsecond))
}

func TestGetCloudflareDefaultForNonPositiveTimeout(t *testing.T) {
	cidrs := GetCloudflare(0)
	if len(cidrs) == 0 {
		t.Fatal("expected default cidrs")
	}
	if !hasCIDR(cidrs, "173.245.48.0/20") {
		t.Fatalf("missing expected ipv4 cidr: %v", cidrs)
	}
	if !hasCIDR(cidrs, "2400:cb00::/32") {
		t.Fatalf("missing expected ipv6 cidr: %v", cidrs)
	}

	cidrs = GetCloudflare(-1)
	if len(cidrs) == 0 {
		t.Fatal("expected default cidrs for negative timeout")
	}
}

func hasCIDR(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
