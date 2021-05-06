package limiter

import (
	"fmt"
	"sort"

	"github.com/sohaha/zlsgo/znet"
)

// Remaining Remaining visits
func (r *Rule) Remaining(key interface{}) []int {
	arr := make([]int, 0, len(r.rules))
	for i := range r.rules {
		arr = append(arr, r.rules[i].remainingVisits(key))
	}
	return arr
}

// RemainingVisitsByIP Remaining Visits IP
func (r *Rule) RemainingVisitsByIP(ip string) []int {
	ipUint, _ := znet.IPString2Long(ip)
	if ipUint == 0 {
		return []int{}
	}
	return r.Remaining(ipUint)
}

// GetOnline Get all current online users
func (r *Rule) GetOnline() []string {
	var insertIgnoreString = func(s []string, v string) []string {
		for _, val := range s {
			if val == v {
				return s
			}
		}
		s = append(s, v)
		return s
	}
	var users []string
	for i := range r.rules {
		f := func(k, v interface{}) bool {
			var user string
			switch v := k.(type) {
			case uint:
				user, _ = znet.Long2IPString(v)
			default:
				user = fmt.Sprint(k)
			}
			users = insertIgnoreString(users, user)
			return true
		}
		r.rules[i].usedRecordsIndex.Range(f)
	}
	sort.Strings(users)
	return users
}
