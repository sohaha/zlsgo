package realip

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zsync"
)

func GetCloudflare() []string {
	const (
		cfIPv4CIDRsEndpoint = "https://www.cloudflare.com/ips-v4"
		cfIPv6CIDRsEndpoint = "https://www.cloudflare.com/ips-v6"
	)

	var (
		wg      zsync.WaitGroup
		mu      = zsync.NewRBMutex()
		cfCIDRs = make([]string, 0)
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	for _, v := range []string{cfIPv4CIDRsEndpoint, cfIPv6CIDRsEndpoint} {
		endpoint := v
		wg.Go(func() {
			resp, err := zhttp.Get(endpoint, ctx)
			if err == nil && resp.StatusCode() == http.StatusOK {
				mu.Lock()
				cfCIDRs = append(cfCIDRs, strings.Split(resp.String(), "\n")...)
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	return cfCIDRs
}
