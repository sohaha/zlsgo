package zjson

import (
	jsongo "encoding/json"
	"errors"
	"github.com/sohaha/zlsgo/zstring"
	"reflect"
	"strconv"
	"unsafe"
)

var (
	ErrNoChange              = errors.New("no change")
	ErrPathEmpty             = errors.New("path cannot be empty")
	ErrInvalidJSON           = errors.New("invalid json")
	ErrNotAllowedWildcard    = errors.New("wildcard characters not allowed in path")
	ErrNotAllowedArrayAccess = errors.New("array access character not allowed in path")
	ErrTypeError             = errors.New("json must be an object or array")
)

func getBytes(json []byte, path string) Res {
	var result Res
	if json != nil {
		result = Get(zstring.Bytes2String(json), path)
		rawhi := *(*reflect.StringHeader)(unsafe.Pointer(&result.Raw))
		strhi := *(*reflect.StringHeader)(unsafe.Pointer(&result.Str))
		rawh := reflect.SliceHeader{Data: rawhi.Data, Len: rawhi.Len}
		strh := reflect.SliceHeader{Data: strhi.Data, Len: strhi.Len}
		if strh.Data == 0 {
			if rawh.Data == 0 {
				result.Raw = ""
			} else {
				result.Raw = sliceHeaderToString(&rawh)
			}
			result.Str = ""
		} else if rawh.Data == 0 {
			result.Raw = ""
			result.Str = sliceHeaderToString(&strh)
		} else if strh.Data >= rawh.Data &&
			int(strh.Data)+strh.Len <= int(rawh.Data)+rawh.Len {
			start := int(strh.Data - rawh.Data)
			result.Raw = sliceHeaderToString(&rawh)
			result.Str = result.Raw[start : start+strh.Len]
		} else {
			result.Raw = sliceHeaderToString(&rawh)
			result.Str = sliceHeaderToString(&strh)
		}
	}
	return result
}

func fillIndex(json string, c *parseContext) {
	if len(c.value.Raw) > 0 && !c.calcd {
		jhdr := *(*reflect.StringHeader)(unsafe.Pointer(&json))
		rhdr := *(*reflect.StringHeader)(unsafe.Pointer(&(c.value.Raw)))
		c.value.Index = int(rhdr.Data - jhdr.Data)
		if c.value.Index < 0 || c.value.Index >= len(json) {
			c.value.Index = 0
		}
	}
}

func sliceHeaderToString(s *reflect.SliceHeader) string {
	return string(*(*[]byte)(unsafe.Pointer(s)))
}

func trim(s string) string {
	for len(s) > 0 {
		if s[0] <= ' ' {
			s = s[1:]
			continue
		}
		break
	}
	for len(s) > 0 {
		if s[len(s)-1] <= ' ' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}

func set(jstr, path, raw string, stringify, del, optimistic, inplace bool) ([]byte, error) {
	if path == "" {
		if !Valid(raw) {
			return nil, ErrPathEmpty
		}
		return zstring.String2Bytes(&raw), nil
	}
	if !del && optimistic && isOptimisticPath(path) {
		res := Get(jstr, path)
		if res.Exists() && res.Index > 0 {
			sz := len(jstr) - len(res.Raw) + len(raw)
			if stringify {
				sz += 2
			}
			if inplace && sz <= len(jstr) {
				if !stringify || !mustMarshalString(raw) {
					jbytes := []byte(jstr)
					if stringify {
						jbytes[res.Index] = '"'
						copy(jbytes[res.Index+1:], []byte(raw))
						jbytes[res.Index+1+len(raw)] = '"'
						copy(jbytes[res.Index+1+len(raw)+1:],
							jbytes[res.Index+len(res.Raw):])
					} else {
						copy(jbytes[res.Index:], []byte(raw))
						copy(jbytes[res.Index+len(raw):],
							jbytes[res.Index+len(res.Raw):])
					}
					return jbytes[:sz], nil
				}
				return nil, nil
			}
			buf := make([]byte, 0, sz)
			buf = append(buf, jstr[:res.Index]...)
			if stringify {
				buf = appendStringify(buf, raw)
			} else {
				buf = append(buf, raw...)
			}
			buf = append(buf, jstr[res.Index+len(res.Raw):]...)
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

	njson, err := appendRawPaths(nil, jstr, paths, raw, stringify, del)
	if err != nil {
		return nil, err
	}
	return njson, nil
}

func SetOptions(json, path string, value interface{},
	opts *StSetOptions) (string, error) {
	if opts != nil && opts.ReplaceInPlace {
		nopts := *opts
		opts = &nopts
		opts.ReplaceInPlace = false
	}
	jsonb := []byte(json)
	res, err := SetBytesOptions(jsonb, path, value, opts)
	return string(res), err
}

func SetBytesOptions(json []byte, path string, value interface{},
	opts *StSetOptions) ([]byte, error) {
	var optimistic, inplace bool
	if opts != nil {
		optimistic = opts.Optimistic
		inplace = opts.ReplaceInPlace
	}
	jstr := string(json)
	var res []byte
	var err error
	switch v := value.(type) {
	default:
		b, merr := jsongo.Marshal(value)
		if merr != nil {
			return nil, merr
		}
		raw := string(b)
		res, err = set(jstr, path, raw, false, false, optimistic, inplace)
	case dtype:
		res, err = set(jstr, path, "", false, true, optimistic, inplace)
	case string:
		res, err = set(jstr, path, v, true, false, optimistic, inplace)
	case []byte:
		raw := string(v)
		res, err = set(jstr, path, raw, true, false, optimistic, inplace)
	case bool:
		if v {
			res, err = set(jstr, path, "true", false, false, optimistic, inplace)
		} else {
			res, err = set(jstr, path, "false", false, false, optimistic, inplace)
		}
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
	opts *StSetOptions) ([]byte, error) {
	jstr := string(json)
	vstr := string(value)
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
