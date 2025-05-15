package zstring

import (
	"regexp"
	"sync"
	"time"
)

// regexMapStruct holds a compiled regular expression with its last access time
// and a mutex for concurrent access control.
type regexMapStruct struct {
	Value *regexp.Regexp
	Time  int64
	sync.RWMutex
}

var (
	l                 sync.RWMutex                                // Global mutex for regexCache access
	regexCache                     = map[string]*regexMapStruct{} // Cache of compiled regular expressions
	regexCacheTimeout uint         = 1800                         // Cache timeout in seconds (30 minutes)
)

// init starts a background goroutine that periodically cleans up
// expired regular expression cache entries every 10 minutes.
func init() {
	go func() {
		ticker := time.NewTicker(600 * time.Second)
		for range ticker.C {
			clearRegexpCompile()
		}
	}()
}

// RegexMatch checks if a string matches a regular expression pattern.
// Returns true if the pattern matches the string, false otherwise.
func RegexMatch(pattern string, str string) bool {
	if r, err := getRegexpCompile(pattern); err == nil {
		return r.Match(String2Bytes(str))
	}
	return false
}

// RegexExtract extracts the first matching substring and any capture groups
// defined in the regular expression pattern.
func RegexExtract(pattern string, str string) ([]string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		return r.FindStringSubmatch(str), nil
	}
	return nil, err
}

// RegexExtractAll extracts all matching substrings and their capture groups.
// An optional count parameter limits the number of matches returned.
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

// RegexFind returns the positions (start and end indices) of all matches.
// The count parameter limits the number of matches returned.
func RegexFind(pattern string, str string, count int) [][]int {
	if r, err := getRegexpCompile(pattern); err == nil {
		return r.FindAllIndex(String2Bytes(str), count)
	}
	return [][]int{}
}

// RegexReplace replaces all matches of the pattern with the replacement string.
// Returns the modified string and any error that occurred.
func RegexReplace(pattern string, str, repl string) (string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		str = r.ReplaceAllString(str, repl)
	}
	return str, err
}

// RegexReplaceFunc replaces all matches of the pattern using a replacement function.
// The function receives each match and returns the replacement string.
func RegexReplaceFunc(pattern string, str string, repl func(string) string) (string, error) {
	r, err := getRegexpCompile(pattern)
	if err == nil {
		str = r.ReplaceAllStringFunc(str, repl)
	}
	return str, err
}

// RegexSplit splits the string at each occurrence of the pattern.
// Returns the resulting string slices and any error that occurred.
func RegexSplit(pattern string, str string) ([]string, error) {
	r, err := getRegexpCompile(pattern)
	var result []string
	if err == nil {
		result = r.Split(str, -1)
	}
	return result, err
}

// clearRegexpCompile removes expired entries from the regular expression cache.
// An entry is considered expired if it hasn't been accessed within the timeout period.
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

// getRegexpCompile retrieves a compiled regular expression from the cache or compiles it if not found.
// The compiled expression is cached for future use to improve performance.
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
