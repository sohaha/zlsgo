package zstring

import (
	"regexp"
	"sync"
	"time"
)

type regexMapStruct struct {
	Value *regexp.Regexp
	Time  int64
	sync.RWMutex
}

var (
	l                 sync.RWMutex
	regexCache             = map[string]*regexMapStruct{}
	regexCacheTimeout uint = 1800
)

func init() {
	go func() {
		ticker := time.NewTicker(600 * time.Second)
		for range ticker.C {
			clearRegexpCompile()
		}
	}()
}

// RegexMatch check for match
func RegexMatch(pattern string, str string) bool {
	if r, err := getRegexpCompile(pattern); err == nil {
		return r.Match(String2Bytes(str))
	}
	return false
}

// RegexExtract extract matching text
func RegexExtract(pattern string, str string) ([]string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		return r.FindStringSubmatch(str), nil
	}
	return nil, err
}

// RegexExtractAll extract matching all text
func RegexExtractAll(pattern string, str string, count ...int) ([][]string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		n := -1
		if len(count) > 0 {
			n = count[0]
		}
		return r.FindAllStringSubmatch(str, n), nil
	}
	return nil, err
}

// RegexFind return matching position
func RegexFind(pattern string, str string, count int) [][]int {
	if r, err := getRegexpCompile(pattern); err == nil {
		return r.FindAllIndex(String2Bytes(str), count)
	}
	return [][]int{}
}

// RegexReplace replacing matches of the Regexp
func RegexReplace(pattern string, str, repl string) (string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		str = r.ReplaceAllString(str, repl)
	}
	return str, err
}

// RegexReplaceFunc replacing matches of the Regexp
func RegexReplaceFunc(pattern string, str string, repl func(string) string) (string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		str = r.ReplaceAllStringFunc(str, repl)
	}
	return str, err
}

// RegexSplit split the string
func RegexSplit(pattern string, str string) ([]string, error) {
	r, err := getRegexpCompile(pattern)
	var result []string
	if err == nil {
		result = r.Split(str, -1)
	}
	return result, err
}

func clearRegexpCompile() {
	newRegexCache := map[string]*regexMapStruct{}
	l.Lock()
	defer l.Unlock()
	if len(regexCache) == 0 {
		return
	}
	now := time.Now().Unix()
	for k := range regexCache {
		if uint(now-regexCache[k].Time) <= regexCacheTimeout {
			newRegexCache[k] = &regexMapStruct{Value: regexCache[k].Value, Time: now}
		}
	}
	regexCache = newRegexCache
}

func getRegexpCompile(pattern string) (r *regexp.Regexp, err error) {
	l.RLock()
	var data *regexMapStruct
	var ok bool
	data, ok = regexCache[pattern]
	l.RUnlock()
	if ok {
		r = data.Value
		return
	}
	r, err = regexp.Compile(pattern)
	if err != nil {
		return
	}
	l.Lock()
	regexCache[pattern] = &regexMapStruct{Value: r, Time: time.Now().Unix()}
	l.Unlock()
	return
}
