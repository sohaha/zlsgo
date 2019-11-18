package zjson

import (
	jsong "encoding/json"
	"errors"
	"github.com/sohaha/zlsgo/zstring"
	"strconv"
)

type StSetOptions struct {
	Optimistic     bool
	ReplaceInPlace bool
}

type pathResult struct {
	part  string
	gpart string
	path  string
	force bool
	more  bool
}

func Stringify(value interface{}) (json string) {
	if jsonByte, err := jsong.Marshal(value); err != nil {
		json = zstring.Bytes2String(jsonByte)
	} else {
		json = "{}"
	}

	return
}

func parsePath(path string) (pathResult, error) {
	var r pathResult
	if len(path) > 0 && path[0] == ':' {
		r.force = true
		path = path[1:]
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '.' {
			r.part = path[:i]
			r.gpart = path[:i]
			r.path = path[i+1:]
			r.more = true
			return r, nil
		}
		if path[i] == '*' || path[i] == '?' {
			return r, ErrNotAllowedWildcard
		} else if path[i] == '#' {
			return r, ErrNotAllowedArrayAccess
		}
		if path[i] == '\\' {
			epart := []byte(path[:i])
			gpart := []byte(path[:i+1])
			i++
			if i < len(path) {
				epart = append(epart, path[i])
				gpart = append(gpart, path[i])
				i++
				for ; i < len(path); i++ {
					if path[i] == '\\' {
						gpart = append(gpart, '\\')
						i++
						if i < len(path) {
							epart = append(epart, path[i])
							gpart = append(gpart, path[i])
						}
						continue
					} else if path[i] == '.' {
						r.part = zstring.Bytes2String(epart)
						r.gpart = zstring.Bytes2String(gpart)
						r.path = path[i+1:]
						r.more = true
						return r, nil
					} else if path[i] == '*' || path[i] == '?' {
						return r, ErrNotAllowedWildcard
					} else if path[i] == '#' {
						return r, ErrNotAllowedArrayAccess
					}
					epart = append(epart, path[i])
					gpart = append(gpart, path[i])
				}
			}
			r.part = zstring.Bytes2String(epart)
			r.gpart = zstring.Bytes2String(gpart)
			return r, nil
		}
	}
	r.part = path
	r.gpart = path
	return r, nil
}

func mustMarshalString(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < ' ' || s[i] > 0x7f || s[i] == '"' || s[i] == '\\' {
			return true
		}
	}
	return false
}

func appendStringify(buf []byte, s string) []byte {
	if mustMarshalString(s) {
		b, _ := jsong.Marshal(s)
		return append(buf, b...)
	}
	buf = append(buf, '"')
	buf = append(buf, s...)
	buf = append(buf, '"')
	return buf
}

func appendBuild(buf []byte, array bool, paths []pathResult, raw string,
	stringify bool) []byte {
	if !array {
		buf = appendStringify(buf, paths[0].part)
		buf = append(buf, ':')
	}
	if len(paths) > 1 {
		n, numeric := atoui(paths[1])
		if numeric || (!paths[1].force && paths[1].part == "-1") {
			buf = append(buf, '[')
			buf = appendRepeat(buf, "null,", n)
			buf = appendBuild(buf, true, paths[1:], raw, stringify)
			buf = append(buf, ']')
		} else {
			buf = append(buf, '{')
			buf = appendBuild(buf, false, paths[1:], raw, stringify)
			buf = append(buf, '}')
		}
	} else {
		if stringify {
			buf = appendStringify(buf, raw)
		} else {
			buf = append(buf, raw...)
		}
	}
	return buf
}

func atoui(r pathResult) (n int, ok bool) {
	if r.force {
		return 0, false
	}
	for i := 0; i < len(r.part); i++ {
		if r.part[i] < '0' || r.part[i] > '9' {
			return 0, false
		}
		n = n*10 + int(r.part[i]-'0')
	}
	return n, true
}

func appendRepeat(buf []byte, s string, n int) []byte {
	for i := 0; i < n; i++ {
		buf = append(buf, s...)
	}
	return buf
}

func deleteTailItem(buf []byte) ([]byte, bool) {
loop:
	for i := len(buf) - 1; i >= 0; i-- {
		switch buf[i] {
		case '[':
			return buf, true
		case ',':
			return buf[:i], false
		case ':':
			i--
			for ; i >= 0; i-- {
				if buf[i] == '"' {
					i--
					for ; i >= 0; i-- {
						if buf[i] == '"' {
							i--
							if i >= 0 && buf[i] == '\\' {
								i--
								continue
							}
							for ; i >= 0; i-- {
								switch buf[i] {
								case '{':
									return buf[:i+1], true
								case ',':
									return buf[:i], false
								}
							}
						}
					}
					break
				}
			}
			break loop
		}
	}
	return buf, false
}

func appendRawPaths(buf []byte, jstr string, paths []pathResult, raw string,
	stringify, del bool) ([]byte, error) {
	var err error
	var res Res
	var found bool
	if del {
		if paths[0].part == "-1" && !paths[0].force {
			res = Get(jstr, "#")
			if res.Int() > 0 {
				res = Get(jstr, strconv.FormatInt(int64(res.Int()-1), 10))
				found = true
			}
		}
	}
	if !found {
		res = Get(jstr, paths[0].gpart)
	}
	if res.Index > 0 {
		if len(paths) > 1 {
			buf = append(buf, jstr[:res.Index]...)
			buf, err = appendRawPaths(buf, res.Raw, paths[1:], raw,
				stringify, del)
			if err != nil {
				return nil, err
			}
			buf = append(buf, jstr[res.Index+len(res.Raw):]...)
			return buf, nil
		}
		buf = append(buf, jstr[:res.Index]...)
		var exidx int
		if del {
			var delNextComma bool
			buf, delNextComma = deleteTailItem(buf)
			if delNextComma {
				i, j := res.Index+len(res.Raw), 0
				for ; i < len(jstr); i, j = i+1, j+1 {
					if jstr[i] <= ' ' {
						continue
					}
					if jstr[i] == ',' {
						exidx = j + 1
					}
					break
				}
			}
		} else {
			if stringify {
				buf = appendStringify(buf, raw)
			} else {
				buf = append(buf, raw...)
			}
		}
		buf = append(buf, jstr[res.Index+len(res.Raw)+exidx:]...)
		return buf, nil
	}
	if del {
		return nil, ErrNoChange
	}
	n, numeric := atoui(paths[0])
	isempty := true
	for i := 0; i < len(jstr); i++ {
		if jstr[i] > ' ' {
			isempty = false
			break
		}
	}
	if isempty {
		if numeric {
			jstr = "[]"
		} else {
			jstr = "{}"
		}
	}
	jsres := Parse(jstr)
	if jsres.Type != JSON {
		if numeric {
			jstr = "[]"
		} else {
			jstr = "{}"
		}
		jsres = Parse(jstr)
	}
	var comma bool
	for i := 1; i < len(jsres.Raw); i++ {
		if jsres.Raw[i] <= ' ' {
			continue
		}
		if jsres.Raw[i] == '}' || jsres.Raw[i] == ']' {
			break
		}
		comma = true
		break
	}
	switch jsres.Raw[0] {
	case '{':
		buf = append(buf, '{')
		buf = appendBuild(buf, false, paths, raw, stringify)
		if comma {
			buf = append(buf, ',')
		}
		buf = append(buf, jsres.Raw[1:]...)
		return buf, nil
	case '[':
		var appendit bool
		if !numeric {
			if paths[0].part == "-1" && !paths[0].force {
				appendit = true
			} else {
				return nil, errors.New("cannot set array element for non-numeric key '" + paths[0].part + "'")
			}
		}
		if appendit {
			njson := trim(jsres.Raw)
			if njson[len(njson)-1] == ']' {
				njson = njson[:len(njson)-1]
			}
			buf = append(buf, njson...)
			if comma {
				buf = append(buf, ',')
			}

			buf = appendBuild(buf, true, paths, raw, stringify)
			buf = append(buf, ']')
			return buf, nil
		}
		buf = append(buf, '[')
		ress := jsres.Array()
		for i := 0; i < len(ress); i++ {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = append(buf, ress[i].Raw...)
		}
		if len(ress) == 0 {
			buf = appendRepeat(buf, "null,", n-len(ress))
		} else {
			buf = appendRepeat(buf, ",null", n-len(ress))
			if comma {
				buf = append(buf, ',')
			}
		}
		buf = appendBuild(buf, true, paths, raw, stringify)
		buf = append(buf, ']')
		return buf, nil
	default:
		return nil, ErrTypeError
	}
}

func isOptimisticPath(path string) bool {
	for i := 0; i < len(path); i++ {
		if path[i] < '.' || path[i] > 'z' {
			return false
		}
		if path[i] > '9' && path[i] < 'A' {
			return false
		}
		if path[i] > 'z' {
			return false
		}
	}
	return true
}

func Marshal(json interface{}) ([]byte, error) {
	return SetBytes([]byte{}, "", json)
}

func Set(json, path string, value interface{}) (string, error) {
	return SetOptions(json, path, value, nil)
}

func SetBytes(json []byte, path string, value interface{}) ([]byte, error) {
	return SetBytesOptions(json, path, value, nil)
}

func SetRaw(json, path, value string) (string, error) {
	return SetRawOptions(json, path, value, nil)
}

func SetRawOptions(json, path, value string, opts *StSetOptions) (string, error) {
	var optimistic bool
	if opts != nil {
		optimistic = opts.Optimistic
	}
	res, err := set(json, path, value, false, false, optimistic, false)
	if err == ErrNoChange {
		return json, nil
	}
	return zstring.Bytes2String(res), err
}

func SetRawBytes(json []byte, path string, value []byte) ([]byte, error) {
	return SetRawBytesOptions(json, path, value, nil)
}

type dtype struct{}

func Delete(json, path string) (string, error) {
	return Set(json, path, dtype{})
}

func DeleteBytes(json []byte, path string) ([]byte, error) {
	return SetBytes(json, path, dtype{})
}
