package zstring

import (
	"regexp"
	"sync"
	"time"
)

type regexMapStruct struct {
	Value *regexp.Regexp
	Time  int64
}

var (
	regexMu         = sync.RWMutex{}
	regexMap        = make(map[string]*regexp.Regexp)
	regexs          sync.Map
	clearRegexCache = time.Now().Unix()
)

func getRegexpCompile(pattern string) (*regexp.Regexp, error) {
	if r := getRegexCache(pattern); r != nil {
		return r, nil
	}
	r, err := regexp.Compile(pattern)
	if err == nil {
		setRegexCache(pattern, r)
		return r, nil
	}
	return nil, err
}

func getRegexCache(pattern string) (regex *regexp.Regexp) {
	v, ok := regexs.Load(pattern)
	now := time.Now().Unix()
	if ((now / 1 % 10) >= 6) && ((now - clearRegexCache) > 1800) {
		clearRegexCache = now
		regexs.Range(func(k, v interface{}) bool {
			reg := v.(*regexMapStruct)
			if (now - reg.Time) >= 1800 {
				regexs.Delete(k)
			}
			return true
		})
	}

	if !ok {
		return
	}
	reg := v.(*regexMapStruct)
	return reg.Value
}

func setRegexCache(pattern string, regex *regexp.Regexp) {
	regexs.Store(pattern, &regexMapStruct{
		Value: regex,
		Time:  time.Now().Unix(),
	})
}

// IsMatch check for match
func IsMatch(pattern string, str string) bool {
	if r, err := getRegexpCompile(pattern); err == nil {
		return r.Match(String2Bytes(str))
	}
	return false
}

// Extract extract matching text
func Extract(pattern string, str string) ([]string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		return r.FindStringSubmatch(str), nil
	}
	return nil, err
}
