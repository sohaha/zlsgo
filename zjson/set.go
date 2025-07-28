package zjson

import (
	"bytes"
	jsong "encoding/json"
	"errors"
	"strconv"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	// Options provides configuration for JSON operations.
	Options struct {
		// Optimistic enables optimistic path processing.
		Optimistic bool
		// ReplaceInPlace modifies the JSON in place without reallocation when possible.
		ReplaceInPlace bool
	}
	dtype struct{}
	// pathResult represents the parsed components of a JSON path.
	pathResult struct {
		part  string // The current path segment
		gpart string // The escaped path segment
		path  string // The remaining path
		force bool   // Force creation of missing elements
		more  bool   // Indicates if there are more path segments
	}
)

// Stringify converts any Go value to its JSON string representation.
func Stringify(value interface{}) (json string) {
	if jsonByte, err := jsong.Marshal(value); err == nil {
		json = zstring.Bytes2String(jsonByte)
	} else {
		json = "{}"
	}

	return
}

// parsePath parses a dot notation path into a pathResult structure.
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

// mustMarshalString determines if a string needs to be JSON escaped.
func mustMarshalString(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < ' ' || s[i] > 0x7f || s[i] == '"' || s[i] == '\\' {
			return true
		}
	}
	return false
}

// appendStringify appends a JSON string representation to a byte buffer.
func appendStringify(buf *bytes.Buffer, s string) {
	if mustMarshalString(s) {
		b, _ := jsong.Marshal(s)
		buf.Write(b)
		return
	}

	buf.WriteByte('"')
	buf.WriteString(s)
	buf.WriteByte('"')
}

// appendBuild recursively builds a JSON structure based on the provided paths.
func appendBuild(buf *bytes.Buffer, array bool, paths []pathResult, raw string,
	stringify bool,
) *bytes.Buffer {
	if !array {
		appendStringify(buf, paths[0].part)
		buf.WriteByte(':')
	}
	if len(paths) > 1 {
		n, numeric := atoui(paths[1])
		if numeric || (!paths[1].force && paths[1].part == "-1") {
			buf.WriteByte('[')
			appendRepeat(buf, "null,", n)
			appendBuild(buf, true, paths[1:], raw, stringify)
			buf.WriteByte(']')
		} else {
			buf.WriteByte('{')
			appendBuild(buf, false, paths[1:], raw, stringify)
			buf.WriteByte('}')
		}
	} else {
		if stringify {
			appendStringify(buf, raw)
		} else {
			buf.WriteString(raw)
		}
	}
	return buf
}

// atoui converts a path segment to an unsigned integer if possible.
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

// appendRepeat appends a string to a buffer n times.
func appendRepeat(buf *bytes.Buffer, s string, n int) {
	for i := 0; i < n; i++ {
		buf.WriteString(s)
	}
}

// deleteTailItem removes the last item from a JSON array or object.
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

// appendRawPaths appends or deletes a value at the specified path in the JSON.
func appendRawPaths(buf *bytes.Buffer, jstr string, paths []pathResult, raw string,
	stringify, del bool,
) error {
	var err error
	var res *Res
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
	if res.index > 0 {
		if len(paths) > 1 {
			buf.WriteString(jstr[:res.index])
			err = appendRawPaths(buf, res.raw, paths[1:], raw,
				stringify, del)
			if err != nil {
				return err
			}
			buf.WriteString(jstr[res.index+len(res.raw):])
			return nil
		}
		buf.WriteString(jstr[:res.index])
		var exidx int
		if del {
			bufNew, delNextComma := deleteTailItem(buf.Bytes())
			buf.Reset()
			buf.Write(bufNew)
			if delNextComma {
				i, j := res.index+len(res.raw), 0
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
				appendStringify(buf, raw)
			} else {
				buf.WriteString(raw)
			}
		}
		buf.WriteString(jstr[res.index+len(res.raw)+exidx:])
		return nil
	}
	if del {
		return ErrNoChange
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
	if jsres.typ != JSON {
		if numeric {
			jstr = "[]"
		} else {
			jstr = "{}"
		}
		jsres = Parse(jstr)
	}
	var comma bool
	for i := 1; i < len(jsres.raw); i++ {
		if jsres.raw[i] <= ' ' {
			continue
		}
		if jsres.raw[i] == '}' || jsres.raw[i] == ']' {
			break
		}
		comma = true
		break
	}
	switch jsres.raw[0] {
	case '{':
		buf.WriteString("{")
		appendBuild(buf, false, paths, raw, stringify)
		if comma {
			buf.WriteString(",")
		}
		buf.WriteString(jsres.raw[1:])
		return nil
	case '[':
		var appendit bool
		if !numeric {
			if paths[0].part == "-1" && !paths[0].force {
				appendit = true
			} else {
				return errors.New("cannot set array element for non-numeric key '" + paths[0].part + "'")
			}
		}
		if appendit {
			njson := zstring.TrimSpace(jsres.raw)
			if njson[len(njson)-1] == ']' {
				njson = njson[:len(njson)-1]
			}
			buf.WriteString(njson)
			if comma {
				buf.WriteString(",")
			}

			appendBuild(buf, true, paths, raw, stringify)
			buf.WriteString("]")
			return nil
		}
		buf.WriteString("[")
		ress := jsres.Array()
		for i := 0; i < len(ress); i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(ress[i].raw)
		}
		if len(ress) == 0 {
			buf.WriteString("null,")
		} else {
			appendRepeat(buf, ",null", n-len(ress))
			if comma {
				buf.WriteString(",")
			}
		}
		appendBuild(buf, true, paths, raw, stringify)
		buf.WriteString("]")
		return nil
	default:
		return ErrTypeError
	}
}

// isOptimisticPath determines if a path can be processed optimistically.
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

// Marshal converts a Go value to a JSON byte slice.
func Marshal(json interface{}) ([]byte, error) {
	return jsong.Marshal(json)
}

// Set sets a value at the specified path in a JSON string.
func Set(json, path string, value interface{}) (string, error) {
	return SetOptions(json, path, value, nil)
}

// SetBytes sets a value at the specified path in a JSON byte slice.
func SetBytes(json []byte, path string, value interface{}) ([]byte, error) {
	return SetBytesOptions(json, path, value, nil)
}

// SetRaw sets a raw JSON value at the specified path in a JSON string.
func SetRaw(json, path, value string) (string, error) {
	return SetRawOptions(json, path, value, nil)
}

// SetRawOptions sets a raw JSON value at the specified path with custom options.
func SetRawOptions(json, path, value string, opts *Options) (string, error) {
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

// SetRawBytes sets a raw JSON value at the specified path in a JSON byte slice.
func SetRawBytes(json []byte, path string, value []byte) ([]byte, error) {
	return SetRawBytesOptions(json, path, value, nil)
}

// Delete removes a value at the specified path from a JSON string.
func Delete(json, path string) (string, error) {
	return Set(json, path, dtype{})
}

// DeleteBytes removes a value at the specified path from a JSON byte slice.
func DeleteBytes(json []byte, path string) ([]byte, error) {
	return SetBytes(json, path, dtype{})
}
