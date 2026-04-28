package zstring

import (
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

// regexMapStruct holds a compiled regular expression with its last access time.
type regexMapStruct struct {
	Value *regexp.Regexp
	Time  int64
}

var (
	// regexCache uses sync.Map for lock-free concurrent access
	// This provides better performance under high concurrency compared to mutex-protected map
	regexCache        sync.Map
	regexCacheTimeout uint = 1800 // Cache timeout in seconds (30 minutes)
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
// This function is called periodically by a background goroutine.
func clearRegexpCompile() {
	now := time.Now().Unix()
	regexCache.Range(func(key, value interface{}) bool {
		data := value.(*regexMapStruct)
		if uint(now-atomic.LoadInt64(&data.Time)) > regexCacheTimeout {
			regexCache.Delete(key)
		}
		return true
	})
}

// getRegexpCompile retrieves a compiled regular expression from the cache or compiles it if not found.
// The compiled expression is cached for future use to improve performance.
// Uses sync.Map for lock-free concurrent access, providing better performance under high concurrency.
func getRegexpCompile(pattern string) (r *regexp.Regexp, err error) {
	// Fast path: try to load from cache without locking
	if v, ok := regexCache.Load(pattern); ok {
		data := v.(*regexMapStruct)
		atomic.StoreInt64(&data.Time, time.Now().Unix())
		return data.Value, nil
	}

	// Slow path: compile and store
	r, err = regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// Store in cache for future use
	regexCache.Store(pattern, &regexMapStruct{
		Value: r,
		Time:  time.Now().Unix(),
	})

	return r, nil
}
