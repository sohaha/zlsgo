package limiter

import (
	"net/http"
	"sort"
	"time"

	"github.com/sohaha/zlsgo/znet"
)

// Rule user access control strategy
type Rule struct {
	rules []*singleRule
}

// New New limiter
func New(allowed uint64, overflow ...func(c *znet.Context)) znet.HandlerFunc {
	r := NewRule()
	f := func(c *znet.Context) {
		c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
	}
	if len(overflow) > 0 {
		f = overflow[0]
	}
	r.AddRule(time.Second, int(allowed))
	return func(c *znet.Context) {
		if !r.AllowVisitByIP(c.GetClientIP()) {
			f(c)
			return
		}
		c.Next()
	}
}

// NewRule Custom limiter rule
func NewRule() *Rule {
	return &Rule{}
}

// AddRule increase user access control strategy
// If less than 1s, please use golang.org/x/time/rate
func (r *Rule) AddRule(exp time.Duration, allowed int, estimated ...int) {
	r.rules = append(r.rules, newRule(exp, allowed, estimated...))
	sort.Slice(r.rules, func(i int, j int) bool {
		return r.rules[i].defaultExpiration < r.rules[j].defaultExpiration
	})
}

// AllowVisit Is access allowed
func (r *Rule) AllowVisit(key interface{}) bool {
	if len(r.rules) == 0 {
		return true
	}
	for i := range r.rules {
		if !r.rules[i].allowVisit(key) {
			return false
		}
	}
	return true
}

// AllowVisitByIP AllowVisit IP
func (r *Rule) AllowVisitByIP(ip string) bool {
	i, _ := znet.IPToLong(ip)
	if i == 0 {
		return false
	}
	return r.AllowVisit(i)
}
