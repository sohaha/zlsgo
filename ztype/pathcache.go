package ztype

import (
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zcache/fast"
)

// pathToken represents a parsed token in a path
type pathToken struct {
	kind  int    // Token type: 0=field access, 1=array index
	key   string // Field name or index string
	index int    // Array index (used when kind=1)
}

// pathResult represents the result of a parsed path
type pathResult struct {
	tokens []pathToken
	simple bool // Whether this is a simple path (no escape characters)
}

// pathCache stores compiled path results, using sync.Map to avoid circular dependencies
var pathCache = fast.NewFast(func(o *fast.Options) {
	o.Cap = 1 << 10
	o.Bucket = 4
	o.Expiration = time.Second * 60 * 60
})

// compilePath compiles a path string into pathResult
func compilePath(path string) *pathResult {
	if path == "" {
		return &pathResult{tokens: nil, simple: true}
	}

	cached, ok := pathCache.ProvideGet(path, func() (interface{}, bool) {
		result := &pathResult{
			tokens: make([]pathToken, 0, 4),
			simple: true,
		}

		start := 0
		for i := 0; i < len(path); i++ {
			switch path[i] {
			case '\\':
				result.simple = false
				i++
			case '.':
				if i > start {
					key := path[start:i]
					if !result.simple {
						key = unescapePathKey(key)
					}
					result.tokens = append(result.tokens, pathToken{
						kind: 0,
						key:  key,
					})
				}
				start = i + 1
			}
		}

		if start < len(path) {
			key := path[start:]
			if !result.simple {
				key = unescapePathKey(key)
			}

			result.tokens = append(result.tokens, pathToken{
				kind: 0,
				key:  key,
			})
		}

		return result, true
	})
	if ok {
		return cached.(*pathResult)
	}

	return &pathResult{
		tokens: make([]pathToken, 0, 4),
		simple: true,
	}
}

// unescapePathKey handles escape characters in path keys
func unescapePathKey(key string) string {
	if key == "" {
		return key
	}

	result := make([]byte, 0, len(key))
	for i := 0; i < len(key); i++ {
		if key[i] == '\\' && i+1 < len(key) {
			result = append(result, key[i+1])
			i++
		} else {
			result = append(result, key[i])
		}
	}
	return string(result)
}

// executeCompiledPath executes compiled path lookup
func executeCompiledPath(result *pathResult, v interface{}) (interface{}, bool) {
	if len(result.tokens) == 0 {
		return v, true
	}

	current := v
	for _, token := range result.tokens {
		var ok bool
		current, ok = executePathToken(token, current)
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// executePathToken executes a single path token
func executePathToken(token pathToken, v interface{}) (interface{}, bool) {
	if v == nil {
		return nil, false
	}

	switch token.kind {
	case 0:
		return executeFieldAccess(token.key, v)
	case 1:
		return executeArrayAccess(token.index, v)
	default:
		return nil, false
	}
}

// executeFieldAccess executes field access
func executeFieldAccess(key string, v interface{}) (interface{}, bool) {
	switch val := v.(type) {
	case Map:
		result, exist := val[key]
		return result, exist
	case map[string]interface{}:
		result, exist := val[key]
		return result, exist
	case map[string]string:
		result, exist := val[key]
		return result, exist
	case map[string]int:
		result, exist := val[key]
		return result, exist
	default:
		if idx, err := strconv.Atoi(key); err == nil {
			return executeArrayAccess(idx, v)
		}
		mapVal := ToMap(v)
		result, exist := mapVal[key]
		return result, exist
	}
}

// executeArrayAccess executes array access
func executeArrayAccess(index int, v interface{}) (interface{}, bool) {
	switch val := v.(type) {
	case []Map:
		if len(val) > index && index >= 0 {
			return val[index], true
		}
	case []interface{}:
		if len(val) > index && index >= 0 {
			return val[index], true
		}
	case []string:
		if len(val) > index && index >= 0 {
			return val[index], true
		}
	default:
		aval := ToSlice(v).Value()
		if len(aval) > index && index >= 0 {
			return aval[index], true
		}
	}
	return nil, false
}
