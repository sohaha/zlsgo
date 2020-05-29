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
