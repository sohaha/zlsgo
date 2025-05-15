// Package zjson provides fast and flexible JSON manipulation functions.
// It offers path-based operations for getting, setting, and modifying JSON data
// without the need for intermediate unmarshaling and marshaling.
package zjson

import (
	jsongo "encoding/json"
	"errors"
	"strconv"
	"unsafe"

	"github.com/sohaha/zlsgo/zstring"
)

// Error definitions for common JSON operations
var (
	// ErrNoChange is returned when an operation doesn't modify the JSON
	ErrNoChange              = errors.New("no change")
	// ErrPathEmpty is returned when an empty path is provided
	ErrPathEmpty             = errors.New("path cannot be empty")
	// ErrInvalidJSON is returned when the input is not valid JSON
	ErrInvalidJSON           = errors.New("invalid json")
	// ErrNotAllowedWildcard is returned when a wildcard is used in a path where not allowed
	ErrNotAllowedWildcard    = errors.New("wildcard characters not allowed in path")
	// ErrNotAllowedArrayAccess is returned when array access is used in a path where not allowed
	ErrNotAllowedArrayAccess = errors.New("array access character not allowed in path")
	// ErrTypeError is returned when the JSON value is not of the expected type
	ErrTypeError             = errors.New("json must be an object or array")
)

// MatchKeys returns a new Res containing only the key-value pairs where the key
// matches one of the provided keys.
func (r *Res) MatchKeys(keys []string) *Res {
	return r.Filter(func(key, value *Res) bool {
		for i := range keys {
			if key.String() == keys[i] {
				return true
			}
		}
		return false
	})
}

// Filter returns a new Res containing only the key-value pairs that satisfy
// the provided filter function.
func (r *Res) Filter(fn func(key, value *Res) bool) *Res {
	j := "{}"
	r.ForEach(func(key, value *Res) bool {
		if fn(key, value) {
			j, _ = Set(j, key.String(), value.Value())
		}
		return true
	})
	return Parse(j)
}

// stringHeader represents the header of a string for unsafe pointer operations.
type stringHeader struct {
	data unsafe.Pointer
	len  int
}

// fillIndex calculates the index of the value within the original JSON string.
func fillIndex(json string, c *parseContext) {
	if len(c.value.raw) > 0 && !c.calcd {
		jhdr := *(*stringHeader)(unsafe.Pointer(&json))
		rhdr := *(*stringHeader)(unsafe.Pointer(&(c.value.raw)))
		c.value.index = int(uintptr(rhdr.data) - uintptr(jhdr.data))
		if c.value.index < 0 || c.value.index >= len(json) {
			c.value.index = 0
		}
	}
}

// set performs the core JSON modification operation, handling various types of modifications.
// It supports setting values, deleting values, and optimistic path resolution.
func set(s, path, raw string, stringify, del, optimistic, place bool) ([]byte, error) {
	if path == "" {
		if !Valid(raw) {
			return nil, ErrPathEmpty
		}
		return zstring.String2Bytes(raw), nil
	}
	if !del && optimistic && isOptimisticPath(path) {
		res := Get(s, path)
		if res.Exists() && res.index > 0 {
			sz := len(s) - len(res.raw) + len(raw)
			if stringify {
				sz += 2
			}
			if place && sz <= len(s) {
				if !stringify || !mustMarshalString(raw) {
					jbytes := []byte(s)
					if stringify {
						jbytes[res.index] = '"'
						copy(jbytes[res.index+1:], zstring.String2Bytes(raw))
						jbytes[res.index+1+len(raw)] = '"'
						copy(jbytes[res.index+1+len(raw)+1:],
							jbytes[res.index+len(res.raw):])
					} else {
						copy(jbytes[res.index:], zstring.String2Bytes(raw))
						copy(jbytes[res.index+len(raw):],
							jbytes[res.index+len(res.raw):])
					}
					return jbytes[:sz], nil
				}
				return nil, nil
			}
			buf := make([]byte, 0, sz)
			buf = append(buf, s[:res.index]...)
			if stringify {
				buf = appendStringify(buf, raw)
			} else {
				buf = append(buf, raw...)
			}
			buf = append(buf, s[res.index+len(res.raw):]...)
			return buf, nil
		}
	}
	paths := make([]pathResult, 0, 4)
	r, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	paths = append(paths, r)
	for r.more {
		if r, err = parsePath(r.path); err != nil {
			return nil, err
		}
		paths = append(paths, r)
	}

	njson, err := appendRawPaths(nil, s, paths, raw, stringify, del)
	if err != nil {
		return nil, err
	}
	return njson, nil
}

// SetOptions sets a JSON value at the specified path with custom options.
// It returns the modified JSON string and any error encountered.
func SetOptions(json, path string, value interface{},
	opts *Options) (string, error) {
	if opts != nil && opts.ReplaceInPlace {
		nopts := *opts
		opts = &nopts
		opts.ReplaceInPlace = false
	}
	if json == "" {
		json = "{}"
	}
	jsonb := zstring.String2Bytes(json)
	res, err := SetBytesOptions(jsonb, path, value, opts)
	return zstring.Bytes2String(res), err
}

// SetBytesOptions sets a JSON value at the specified path with custom options.
// It works directly with byte slices for better performance.
func SetBytesOptions(json []byte, path string, value interface{},
	opts *Options) ([]byte, error) {
	var optimistic, inplace bool
	if opts != nil {
		optimistic = opts.Optimistic
		inplace = opts.ReplaceInPlace
	}
	jstr := zstring.Bytes2String(json)
	var res []byte
	var err error
	switch v := value.(type) {
	default:
		b, merr := jsongo.Marshal(value)
		if merr != nil {
			return nil, merr
		}
		raw := zstring.Bytes2String(b)
		res, err = set(jstr, path, raw, false, false, optimistic, inplace)
	case dtype:
		res, err = set(jstr, path, "", false, true, optimistic, inplace)
	case string:
		res, err = set(jstr, path, v, true, false, optimistic, inplace)
	case []byte:
		raw := zstring.Bytes2String(v)
		res, err = set(jstr, path, raw, true, false, optimistic, inplace)
	case bool:
		if v {
			res, err = set(jstr, path, "true", false, false, optimistic, inplace)
		} else {
			res, err = set(jstr, path, "false", false, false, optimistic, inplace)
		}
	case int:
		res, err = set(jstr, path, strconv.FormatInt(int64(v), 10),
			false, false, optimistic, inplace)
	case int8:
		res, err = set(jstr, path, strconv.FormatInt(int64(v), 10),
			false, false, optimistic, inplace)
	case int16:
		res, err = set(jstr, path, strconv.FormatInt(int64(v), 10),
			false, false, optimistic, inplace)
	case int32:
		res, err = set(jstr, path, strconv.FormatInt(int64(v), 10),
			false, false, optimistic, inplace)
	case int64:
		res, err = set(jstr, path, strconv.FormatInt(int64(v), 10),
			false, false, optimistic, inplace)
	case uint:
		res, err = set(jstr, path, strconv.FormatUint(uint64(v), 10),
			false, false, optimistic, inplace)
	case uint8:
		res, err = set(jstr, path, strconv.FormatUint(uint64(v), 10),
			false, false, optimistic, inplace)
	case uint16:
		res, err = set(jstr, path, strconv.FormatUint(uint64(v), 10),
			false, false, optimistic, inplace)
	case uint32:
		res, err = set(jstr, path, strconv.FormatUint(uint64(v), 10),
			false, false, optimistic, inplace)
	case uint64:
		res, err = set(jstr, path, strconv.FormatUint(v, 10),
			false, false, optimistic, inplace)
	case float32:
		res, err = set(jstr, path, strconv.FormatFloat(float64(v), 'f', -1, 64),
			false, false, optimistic, inplace)
	case float64:
		res, err = set(jstr, path, strconv.FormatFloat(v, 'f', -1, 64),
			false, false, optimistic, inplace)
	}
	if err == ErrNoChange {
		return json, nil
	}
	return res, err
}

// SetRawBytesOptions sets a raw JSON value at the specified path in a JSON byte slice with custom options.
// It accepts raw JSON bytes for both the target JSON and the value to be set.
func SetRawBytesOptions(json []byte, path string, value []byte,
	opts *Options) ([]byte, error) {
	jstr := zstring.Bytes2String(json)
	vstr := zstring.Bytes2String(value)
	var optimistic, inplace bool
	if opts != nil {
		optimistic = opts.Optimistic
		inplace = opts.ReplaceInPlace
	}
	res, err := set(jstr, path, vstr, false, false, optimistic, inplace)
	if err == ErrNoChange {
		return json, nil
	}
	return res, err
}

func safeInt(f float64) (n int, ok bool) {
	if f < -9007199254740991 || f > 9007199254740991 {
		return 0, false
	}
	return int(f), true
}

func squash(json string) string {
	ss, _ := switchJson(json, 0, false)
	return ss
}

func parseSquash(json string, i int) (string, int) {
	return switchJson(json, i, true)
}

// switchJson processes a JSON string starting at position i, handling nested structures.
// It returns the processed JSON string and the new position.
func switchJson(json string, i int, isParse bool) (string, int) {
	depth := 1
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] >= '"' && json[i] <= '}' {
			switch json[i] {
			case '"':
				i++
				s2 := i
				for ; i < len(json); i++ {
					if json[i] > '\\' {
						continue
					}
					if json[i] == '"' {
						if json[i-1] == '\\' {
							n := 0
							for j := i - 2; j > s2-1; j-- {
								if json[j] != '\\' {
									break
								}
								n++
							}
							if n%2 == 0 {
								continue
							}
						}
						break
					}
				}
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 0 {
					if isParse {
						i++
						return json[s:i], i
					} else {
						return json[:i+1], i
					}
				}
			}
		}
	}
	if isParse {
		return json[s:], i
	} else {
		return json, i
	}
}
