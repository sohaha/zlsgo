// Package zjson json data read and write operations
package zjson

import (
	jsongo "encoding/json"
	"errors"
	"strconv"
	"unsafe"

	"github.com/sohaha/zlsgo/zstring"
)

var (
	ErrNoChange              = errors.New("no change")
	ErrPathEmpty             = errors.New("path cannot be empty")
	ErrInvalidJSON           = errors.New("invalid json")
	ErrNotAllowedWildcard    = errors.New("wildcard characters not allowed in path")
	ErrNotAllowedArrayAccess = errors.New("array access character not allowed in path")
	ErrTypeError             = errors.New("json must be an object or array")
)

type stringHeader struct {
	data unsafe.Pointer
	len  int
}

func fillIndex(json string, c *parseContext) {
	if len(c.value.Raw) > 0 && !c.calcd {
		jhdr := *(*stringHeader)(unsafe.Pointer(&json))
		rhdr := *(*stringHeader)(unsafe.Pointer(&(c.value.Raw)))
		c.value.Index = int(uintptr(rhdr.data) - uintptr(jhdr.data))
		if c.value.Index < 0 || c.value.Index >= len(json) {
			c.value.Index = 0
		}
	}
}

func set(s, path, raw string, stringify, del, optimistic, place bool) ([]byte, error) {
	if path == "" {
		if !Valid(raw) {
			return nil, ErrPathEmpty
		}
		return zstring.String2Bytes(raw), nil
	}
	if !del && optimistic && isOptimisticPath(path) {
		res := Get(s, path)
		if res.Exists() && res.Index > 0 {
			sz := len(s) - len(res.Raw) + len(raw)
			if stringify {
				sz += 2
			}
			if place && sz <= len(s) {
				if !stringify || !mustMarshalString(raw) {
					jbytes := []byte(s)
					if stringify {
						jbytes[res.Index] = '"'
						copy(jbytes[res.Index+1:], zstring.String2Bytes(raw))
						jbytes[res.Index+1+len(raw)] = '"'
						copy(jbytes[res.Index+1+len(raw)+1:],
							jbytes[res.Index+len(res.Raw):])
					} else {
						copy(jbytes[res.Index:], zstring.String2Bytes(raw))
						copy(jbytes[res.Index+len(raw):],
							jbytes[res.Index+len(res.Raw):])
					}
					return jbytes[:sz], nil
				}
				return nil, nil
			}
			buf := make([]byte, 0, sz)
			buf = append(buf, s[:res.Index]...)
			if stringify {
				buf = appendStringify(buf, raw)
			} else {
				buf = append(buf, raw...)
			}
			buf = append(buf, s[res.Index+len(res.Raw):]...)
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
		res, err = set(jstr, path, strconv.FormatUint(uint64(v), 10),
			false, false, optimistic, inplace)
	case float32:
		res, err = set(jstr, path, strconv.FormatFloat(float64(v), 'f', -1, 64),
			false, false, optimistic, inplace)
	case float64:
		res, err = set(jstr, path, strconv.FormatFloat(float64(v), 'f', -1, 64),
			false, false, optimistic, inplace)
	}
	if err == ErrNoChange {
		return json, nil
	}
	return res, err
}

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
