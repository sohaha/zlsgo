package zjson

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztype"
)

type (
	Type int
	Res  struct {
		raw   string
		str   string
		typ   Type
		num   float64
		index int
	}
	fieldMaps struct {
		m  map[string]map[string]int
		mu sync.RWMutex
	}
)

const (
	Null Type = iota
	False
	Number
	String
	True
	JSON
)

func (t Type) String() string {
	switch t {
	case Null:
		return "Null"
	case False:
		return "False"
	case Number:
		return "Number"
	case String:
		return "String"
	case True:
		return "True"
	case JSON:
		return "JSON"
	default:
		return ""
	}
}

func (r *Res) Raw() string {
	return r.raw
}

func (r *Res) Bytes() []byte {
	return zstring.String2Bytes(r.String())
}

func (r *Res) String(def ...string) string {
	switch r.typ {
	case False:
		return "false"
	case Number:
		if len(r.raw) == 0 {
			return strconv.FormatFloat(r.num, 'f', -1, 64)
		}
		var i int
		if r.raw[0] == '-' {
			i++
		}
		for ; i < len(r.raw); i++ {
			if r.raw[i] < '0' || r.raw[i] > '9' {
				return strconv.FormatFloat(r.num, 'f', -1, 64)
			}
		}
		return r.raw
	case String:
		return r.str
	case JSON:
		return r.raw
	case True:
		return "true"
	default:
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
}

func (r *Res) Bool(def ...bool) bool {
	switch r.typ {
	case True:
		return true
	case String:
		b, _ := strconv.ParseBool(strings.ToLower(r.str))
		return b
	case Number:
		return r.num != 0
	default:
		if len(def) > 0 {
			return def[0]
		}
		return false
	}
}

func (r *Res) Int(def ...int) int {
	switch r.typ {
	case True:
		return 1
	case String:
		n, _ := parseInt(r.str)
		return n
	case Number:
		i, ok := safeInt(r.num)
		if ok {
			return i
		}
		// now try to parse the raw string
		i, ok = parseInt(r.raw)
		if ok {
			return i
		}
		return int(r.num)
	default:
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
}

func (r *Res) Int8(def ...int8) int8 {
	var i int
	if len(def) > 0 {
		i = int(def[0])
	}
	return ztype.ToInt8(r.Int(i))
}

func (r *Res) Int16(def ...int16) int16 {
	var i int
	if len(def) > 0 {
		i = int(def[0])
	}
	return ztype.ToInt16(r.Int(i))
}

func (r *Res) Int32(def ...int32) int32 {
	var i int
	if len(def) > 0 {
		i = int(def[0])
	}
	return ztype.ToInt32(r.Int(i))
}

func (r *Res) Int64(def ...int64) int64 {
	var i int
	if len(def) > 0 {
		i = int(def[0])
	}
	return ztype.ToInt64(r.Int(i))
}

func (r *Res) Uint(def ...uint) uint {
	switch r.typ {
	default:
		if len(def) > 0 {
			return def[0]
		}
		return 0
	case True:
		return 1
	case String:
		n, _ := parseUint(r.str)
		return n
	case Number:
		i, ok := safeInt(r.num)
		if ok && i >= 0 {
			return uint(i)
		}
		u, ok := parseUint(r.raw)
		if ok {
			return u
		}
		return uint(r.num)
	}
}

func (r *Res) Uint8(def ...uint8) uint8 {
	var i uint
	if len(def) > 0 {
		i = uint(def[0])
	}
	return ztype.ToUint8(r.Uint(i))
}

func (r *Res) Uint16(def ...uint16) uint16 {
	var i uint
	if len(def) > 0 {
		i = uint(def[0])
	}
	return ztype.ToUint16(r.Uint(i))
}

func (r *Res) Uint32(def ...uint32) uint32 {
	var i uint
	if len(def) > 0 {
		i = uint(def[0])
	}
	return ztype.ToUint32(r.Uint(i))
}

func (r *Res) Uint64(def ...uint64) uint64 {
	var i uint
	if len(def) > 0 {
		i = uint(def[0])
	}
	return ztype.ToUint64(r.Uint(i))
}

func (r *Res) Float64(def ...float64) float64 {
	switch r.typ {
	default:
		if len(def) > 0 {
			return def[0]
		}
		return 0
	case True:
		return 1
	case String:
		n, _ := strconv.ParseFloat(r.str, 64)
		return n
	case Number:
		return r.num
	}
}

func (r *Res) Float(def ...float64) float64 {
	return r.Float64(def...)
}

func (r *Res) Float32(def ...float32) float32 {
	var i float64
	if len(def) > 0 {
		i = float64(def[0])
	}
	return ztype.ToFloat32(r.Float64(i))
}

func (r *Res) Unmarshal(v interface{}) error {
	return Unmarshal(r.raw, v)
}

func (r *Res) Time(format ...string) (t time.Time) {
	t, _ = ztime.Parse(r.String(), format...)
	return t
}

func (r *Res) Array() []*Res {
	if r.typ == Null {
		return []*Res{}
	}
	if r.typ != JSON {
		return []*Res{r}
	}
	rr := r.arrayOrMap('[', false)
	return rr.a
}

func (r *Res) Slice() ztype.SliceType {
	if !r.IsArray() {
		return ztype.SliceType{}
	}

	return ztype.ToSlice(r.Value())
}

func (r *Res) SliceString() []string {
	return r.Slice().String()
}

func (r *Res) SliceInt() []int {
	return r.Slice().Int()
}

func (r *Res) Maps() ztype.Maps {
	if !r.IsArray() {
		return ztype.Maps{}
	}

	return ztype.ToMaps(r.Value())
}

func (r *Res) IsObject() bool {
	return r.firstCharacter() == '{'
}

func (r *Res) IsArray() bool {
	return r.firstCharacter() == '['
}

func (r *Res) firstCharacter() uint8 {
	if r.typ == JSON && len(r.raw) > 0 {
		return r.raw[0]
	}
	return 0
}

func (r *Res) ForEach(fn func(key, value *Res) bool) {
	if !r.Exists() || r.typ == Null {
		return
	}
	if r.typ != JSON {
		fn(&Res{}, r)
		return
	}
	var (
		keys bool
		i    int
	)
	key, value := Res{}, &Res{}
	j := r.raw
	for ; i < len(j); i++ {
		if j[i] == '{' {
			i++
			key.typ = String
			keys = true
			break
		} else if j[i] == '[' {
			i++
			key.typ = Number
			key.num = -1
			break
		}
		if j[i] > ' ' {
			return
		}
	}
	var (
		str  string
		vesc bool
		ok   bool
	)

	for ; i < len(j); i++ {
		if key.typ == Number {
			key.num = key.num + 1
		} else if keys {
			if j[i] != '"' {
				continue
			}
			s := i
			i, str, vesc, ok = parseString(j, i+1)
			if !ok {
				return
			}
			if vesc {
				key.str = unescape(str[1 : len(str)-1])
			} else {
				key.str = str[1 : len(str)-1]
			}

			key.raw = str
			key.index = s
		}
		for ; i < len(j); i++ {
			if j[i] <= ' ' || j[i] == ',' || j[i] == ':' {
				continue
			}
			break
		}
		s := i
		i, value, ok = parseAny(j, i, true)
		if !ok {
			return
		}
		value.index = s
		if !fn(&key, value) {
			return
		}
	}
}

func (r *Res) MapRes() map[string]*Res {
	if r.typ != JSON {
		return map[string]*Res{}
	}
	rr := r.arrayOrMap('{', false)
	return rr.o
}

func (r *Res) Map() ztype.Map {
	if !r.IsObject() {
		return map[string]interface{}{}
	}
	v, _ := r.Value().(map[string]interface{})
	return v
}

func (r *Res) MapKeys(exclude ...string) (keys []string) {
	m := r.MapRes()
	keys = make([]string, 0, len(m))
lo:
	for k := range m {
		for i := range exclude {
			if k == exclude[i] {
				continue lo
			}
		}
		keys = append(keys, k)
	}
	return
}

func (r *Res) Get(path string) *Res {
	return Get(r.raw, path)
}

func (r *Res) Set(path string, value interface{}) (err error) {
	r.raw, err = Set(r.raw, path, value)
	return
}

func (r *Res) Delete(path string) (err error) {
	r.raw, err = Delete(r.raw, path)
	return
}

type arrayOrMapResult struct {
	o  map[string]*Res
	oi map[string]interface{}
	a  []*Res
	ai []interface{}
	vc byte
}

func (r *Res) arrayOrMap(vc byte, valueize bool) (ar arrayOrMapResult) {
	var (
		count int
		key   Res
		i     int
		j     = r.raw
	)

	if vc == 0 {
		for ; i < len(j); i++ {
			if j[i] == '{' || j[i] == '[' {
				ar.vc = j[i]
				i++
				break
			}
			if j[i] > ' ' {
				goto end
			}
		}
	} else {
		for ; i < len(j); i++ {
			if j[i] == vc {
				i++
				break
			}
			if j[i] > ' ' {
				goto end
			}
		}
		ar.vc = vc
	}
	if ar.vc == '{' {
		if valueize {
			ar.oi = make(map[string]interface{})
		} else {
			ar.o = make(map[string]*Res)
		}
	} else {
		if valueize {
			ar.ai = make([]interface{}, 0)
		} else {
			ar.a = make([]*Res, 0)
		}
	}
	for ; i < len(j); i++ {
		if j[i] <= ' ' {
			continue
		}
		if j[i] == ']' || j[i] == '}' {
			break
		}

		var value Res
		switch j[i] {
		default:
			if (j[i] >= '0' && j[i] <= '9') || j[i] == '-' {
				value.typ = Number
				value.raw, value.num = tonum(j[i:])
				value.str = ""
			} else {
				continue
			}
		case '{', '[':
			value.typ = JSON
			value.raw = squash(j[i:])
			value.str, value.num = "", 0
		case 'n':
			value.typ = Null
			value.raw = tolit(j[i:])
			value.str, value.num = "", 0
		case 'r':
			value.typ = True
			value.raw = tolit(j[i:])
			value.str, value.num = "", 0
		case 'f':
			value.typ = False
			value.raw = tolit(j[i:])
			value.str, value.num = "", 0
		case '"':
			value.typ = String
			value.raw, value.str = tostr(j[i:])
			value.num = 0
		}
		i += len(value.raw) - 1

		if ar.vc == '{' {
			if count%2 == 0 {
				key = value
			} else {
				if valueize {
					if _, ok := ar.oi[key.str]; !ok {
						ar.oi[key.str] = value.Value()
					}
				} else {
					if _, ok := ar.o[key.str]; !ok {
						ar.o[key.str] = &value
					}
				}
			}
			count++
		} else {
			if valueize {
				ar.ai = append(ar.ai, value.Value())
			} else {
				ar.a = append(ar.a, &value)
			}
		}
	}
end:
	return
}

func Parse(json string) *Res {
	value := &Res{}
	for i := 0; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			value.typ = JSON
			value.raw = json[i:]
			break
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		default:
			if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' {
				value.typ = Number
				value.raw, value.num = tonum(json[i:])
			} else {
				return &Res{}
			}
		case 'n':
			value.typ = Null
			value.raw = tolit(json[i:])
		case 't':
			value.typ = True
			value.raw = tolit(json[i:])
		case 'f':
			value.typ = False
			value.raw = tolit(json[i:])
		case '"':
			value.typ = String
			value.raw, value.str = tostr(json[i:])
		}
		break
	}
	return value
}

func ParseBytes(json []byte) *Res {
	return Parse(zstring.Bytes2String(json))
}

func tonum(json string) (raw string, num float64) {
	for i := 1; i < len(json); i++ {
		if json[i] <= '-' {
			if json[i] <= ' ' || json[i] == ',' {
				raw = json[:i]
				num, _ = strconv.ParseFloat(raw, 64)
				return
			}
			continue
		}
		if json[i] < ']' {
			continue
		}
		if json[i] == 'e' || json[i] == 'E' {
			continue
		}
		raw = json[:i]
		num, _ = strconv.ParseFloat(raw, 64)
		return
	}
	raw = json
	num, _ = strconv.ParseFloat(raw, 64)
	return
}

func tolit(json string) (raw string) {
	for i := 1; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return json[:i]
		}
	}
	return json
}

func tostr(json string) (raw string, str string) {
	for i := 1; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return json[:i+1], json[1:i]
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
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
			var ret string
			if i+1 < len(json) {
				ret = json[:i+1]
			} else {
				ret = json[:i]
			}
			return ret, unescape(json[1:i])
		}
	}
	return json, json[1:]
}

func (r *Res) Exists() bool {
	return r.typ != Null || len(r.raw) != 0
}

func (r *Res) Value() interface{} {
	if r.typ == String {
		return r.str
	}
	switch r.typ {
	default:
		return nil
	case False:
		return false
	case Number:
		return r.num
	case JSON:
		r := r.arrayOrMap(0, true)
		if r.vc == '{' {
			return r.oi
		} else if r.vc == '[' {
			return r.ai
		}
		return nil
	case True:
		return true
	}
}

func parseString(json string, i int) (int, string, bool, bool) {
	s := i
	for ; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return i + 1, json[s-1 : i+1], false, true
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if json[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					return i + 1, json[s-1 : i+1], true, true
				}
			}
			break
		}
	}
	return i, json[s-1:], false, false
}

func parseNumber(json string, i int) (int, string) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ']' ||
			json[i] == '}' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

func parseLiteral(json string, i int) (int, string) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

type arrayPathResult struct {
	part    string
	path    string
	pipe    string
	alogkey string
	query   struct {
		path  string
		op    string
		value string
		on    bool
		all   bool
	}
	piped  bool
	more   bool
	alogok bool
	arrch  bool
}

// parseArrayPath parses a path that points to an array element or query.
// It extracts array index, query conditions, and pipe operations from the path.
func parseArrayPath(path string) (r arrayPathResult) {
	for i := 0; i < len(path); i++ {
		if path[i] == '|' {
			r.part = path[:i]
			r.pipe = path[i+1:]
			r.piped = true
			return
		}
		if path[i] == '.' {
			r.part = path[:i]
			r.path = path[i+1:]
			r.more = true
			return
		}
		if path[i] == '#' {
			r.arrch = true
			if i == 0 && len(path) > 1 {
				if path[1] == '.' {
					r.alogok = true
					r.alogkey = path[2:]
					r.path = path[:1]
				} else if path[1] == '[' || path[1] == '(' {
					r.query.on = true
					if true {
						qpath, op, value, _, fi, ok := parseQuery(path[i:])
						if !ok {
							break
						}
						r.query.path = qpath
						r.query.op = op
						r.query.value = value
						i = fi - 1
						if i+1 < len(path) && path[i+1] == '#' {
							r.query.all = true
						}
					} else {
						var end byte
						if path[1] == '[' {
							end = ']'
						} else {
							end = ')'
						}
						i += 2
						for ; i < len(path); i++ {
							if path[i] > ' ' {
								break
							}
						}
						s := i
						for ; i < len(path); i++ {
							if path[i] <= ' ' ||
								path[i] == '!' ||
								path[i] == '=' ||
								path[i] == '<' ||
								path[i] == '>' ||
								path[i] == '%' ||
								path[i] == end {
								break
							}
						}
						r.query.path = path[s:i]
						for ; i < len(path); i++ {
							if path[i] > ' ' {
								break
							}
						}
						if i < len(path) {
							s = i
							if path[i] == '!' {
								if i < len(path)-1 && (path[i+1] == '=' ||
									path[i+1] == '%') {
									i++
								}
							} else if path[i] == '<' || path[i] == '>' {
								if i < len(path)-1 && path[i+1] == '=' {
									i++
								}
							} else if path[i] == '=' {
								if i < len(path)-1 && path[i+1] == '=' {
									s++
									i++
								}
							}
							i++
							r.query.op = path[s:i]
							for ; i < len(path); i++ {
								if path[i] > ' ' {
									break
								}
							}
							s = i
							for ; i < len(path); i++ {
								if path[i] == '"' {
									i++
									s2 := i
									for ; i < len(path); i++ {
										if path[i] > '\\' {
											continue
										}
										if path[i] == '"' {
											if path[i-1] == '\\' {
												n := 0
												for j := i - 2; j > s2-1; j-- {
													if path[j] != '\\' {
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
								} else if path[i] == end {
									if i+1 < len(path) && path[i+1] == '#' {
										r.query.all = true
									}
									break
								}
							}
							if i > len(path) {
								i = len(path)
							}
							v := path[s:i]
							for len(v) > 0 && v[len(v)-1] <= ' ' {
								v = v[:len(v)-1]
							}
							r.query.value = v
						}
					}
				}
			}
			continue
		}
	}
	r.part = path
	r.path = ""
	return
}

func parseQuery(query string) (
	path, op, value, remain string, i int, ok bool,
) {
	if len(query) < 2 || query[0] != '#' ||
		(query[1] != '(' && query[1] != '[') {
		return "", "", "", "", i, false
	}
	i = 2
	j := 0
	depth := 1
	for ; i < len(query); i++ {
		if depth == 1 && j == 0 {
			switch query[i] {
			case '!', '=', '<', '>', '%':
				j = i
				continue
			}
		}
		if query[i] == '\\' {
			i++
		} else if query[i] == '[' || query[i] == '(' {
			depth++
		} else if query[i] == ']' || query[i] == ')' {
			depth--
			if depth == 0 {
				break
			}
		} else if query[i] == '"' {
			i++
			for ; i < len(query); i++ {
				if query[i] == '\\' {
					i++
				} else if query[i] == '"' {
					break
				}
			}
		}
	}
	if depth > 0 {
		return "", "", "", "", i, false
	}
	if j > 0 {
		path = zstring.TrimSpace(query[2:j])
		value = zstring.TrimSpace(query[j:i])
		remain = query[i+1:]
		var opsz int
		switch {
		case len(value) == 1:
			opsz = 1
		case value[0] == '!' && value[1] == '=':
			opsz = 2
		case value[0] == '!' && value[1] == '%':
			opsz = 2
		case value[0] == '<' && value[1] == '=':
			opsz = 2
		case value[0] == '>' && value[1] == '=':
			opsz = 2
		case value[0] == '=' && value[1] == '=':
			value = value[1:]
			opsz = 1
		case value[0] == '<':
			opsz = 1
		case value[0] == '>':
			opsz = 1
		case value[0] == '=':
			opsz = 1
		case value[0] == '%':
			opsz = 1
		}
		op = value[:opsz]
		value = zstring.TrimSpace(value[opsz:])
	} else {
		path = zstring.TrimSpace(query[2:i])
		remain = query[i+1:]
	}
	return path, op, value, remain, i + 1, true
}

type objectPathResult struct {
	part  string
	path  string
	pipe  string
	piped bool
	wild  bool
	more  bool
}

func parseObjectPath(path string) (r objectPathResult) {
	for i := 0; i < len(path); i++ {
		if path[i] == '|' {
			r.part = path[:i]
			r.pipe = path[i+1:]
			r.piped = true
			return
		}
		if path[i] == '.' {
			r.part = path[:i]
			if ModifiersState() &&
				i < len(path)-1 &&
				(path[i+1] == '@' ||
					path[i+1] == '[' || path[i+1] == '{') {
				r.pipe = path[i+1:]
				r.piped = true
			} else {
				r.path = path[i+1:]
				r.more = true
			}
			return
		}
		if path[i] == '*' || path[i] == '?' {
			r.wild = true
			continue
		}
		if path[i] == '\\' {
			epart := []byte(path[:i])
			i++
			if i < len(path) {
				epart = append(epart, path[i])
				i++
				for ; i < len(path); i++ {
					if path[i] == '\\' {
						i++
						if i < len(path) {
							epart = append(epart, path[i])
						}
						continue
					} else if path[i] == '.' {
						r.part = zstring.Bytes2String(epart)
						if ModifiersState() &&
							i < len(path)-1 && path[i+1] == '@' {
							r.pipe = path[i+1:]
							r.piped = true
						} else {
							r.path = path[i+1:]
							r.more = true
						}
						r.more = true
						return
					} else if path[i] == '|' {
						r.part = zstring.Bytes2String(epart)
						r.pipe = path[i+1:]
						r.piped = true
						return
					} else if path[i] == '*' || path[i] == '?' {
						r.wild = true
					}
					epart = append(epart, path[i])
				}
			}
			r.part = zstring.Bytes2String(epart)
			return
		}
	}
	r.part = path
	return
}

func parseObject(c *parseContext, i int, path string) (int, bool) {
	var pmatch, kesc, vesc, ok, hit bool
	var key, val string
	rp := parseObjectPath(path)
	if !rp.more && rp.piped {
		c.pipe = rp.pipe
		c.piped = true
	}
	for i < len(c.json) {
		for ; i < len(c.json); i++ {
			if c.json[i] == '"' {
				i++
				s := i
				for ; i < len(c.json); i++ {
					if c.json[i] > '\\' {
						continue
					}
					if c.json[i] == '"' {
						i, key, kesc, ok = i+1, c.json[s:i], false, true
						goto parseKeyStringDone
					}
					if c.json[i] == '\\' {
						i++
						for ; i < len(c.json); i++ {
							if c.json[i] > '\\' {
								continue
							}
							if c.json[i] == '"' {
								if c.json[i-1] == '\\' {
									n := 0
									for j := i - 2; j > 0; j-- {
										if c.json[j] != '\\' {
											break
										}
										n++
									}
									if n%2 == 0 {
										continue
									}
								}
								i, key, kesc, ok = i+1, c.json[s:i], true, true
								goto parseKeyStringDone
							}
						}
						break
					}
				}
				key, kesc, ok = c.json[s:], false, false
			parseKeyStringDone:
				break
			}
			if c.json[i] == '}' {
				return i + 1, false
			}
		}
		if !ok {
			return i, false
		}
		if rp.wild {
			if kesc {
				pmatch = zstring.Match(unescape(key), rp.part)
			} else {
				pmatch = zstring.Match(key, rp.part)
			}
		} else {
			if kesc {
				pmatch = rp.part == unescape(key)
			} else {
				pmatch = rp.part == key
			}
		}
		hit = pmatch && !rp.more
		for ; i < len(c.json); i++ {
			switch c.json[i] {
			default:
				continue
			case '"':
				i++
				i, val, vesc, ok = parseString(c.json, i)
				if !ok {
					return i, false
				}
				if hit {
					if vesc {
						c.value.str = unescape(val[1 : len(val)-1])
					} else {
						c.value.str = val[1 : len(val)-1]
					}
					c.value.raw = val
					c.value.typ = String
					return i, true
				}
			case '{':
				if pmatch && !hit {
					i, hit = parseObject(c, i+1, rp.path)
					if hit {
						return i, true
					}
				} else {
					val, i = parseSquash(c.json, i)
					if hit {
						c.value.raw = val
						c.value.typ = JSON
						return i, true
					}
				}
			case '[':
				if pmatch && !hit {
					i, hit = parseArray(c, i+1, rp.path)
					if hit {
						return i, true
					}
				} else {
					val, i = parseSquash(c.json, i)
					if hit {
						c.value.raw = val
						c.value.typ = JSON
						return i, true
					}
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumber(c.json, i)
				if hit {
					c.value.raw = val
					c.value.typ = Number
					c.value.num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteral(c.json, i)
				if hit {
					c.value.raw = val
					switch vc {
					case 't':
						c.value.typ = True
					case 'f':
						c.value.typ = False
					}
					return i, true
				}
			}
			break
		}
	}
	return i, false
}

// queryMatches checks if a JSON value matches the query conditions in the array path.
// It supports various comparison operators like '=', '!=', '>', '<', etc.
func queryMatches(rp *arrayPathResult, value *Res) bool {
	rpv := rp.query.value
	if len(rpv) > 2 && rpv[0] == '"' && rpv[len(rpv)-1] == '"' {
		rpv = rpv[1 : len(rpv)-1]
	}
	if !value.Exists() {
		return false
	}
	if rp.query.op == "" {
		return true
	}
	switch value.typ {
	case String:
		switch rp.query.op {
		case "=":
			return value.str == rpv
		case "!=":
			return value.str != rpv
		case "<":
			return value.str < rpv
		case "<=":
			return value.str <= rpv
		case ">":
			return value.str > rpv
		case ">=":
			return value.str >= rpv
		case "%":
			return zstring.Match(value.str, rpv)
		case "!%":
			return !zstring.Match(value.str, rpv)
		}
	case Number:
		rpvn, _ := strconv.ParseFloat(rpv, 64)
		switch rp.query.op {
		case "=":
			return value.num == rpvn
		case "!=":
			return value.num != rpvn
		case "<":
			return value.num < rpvn
		case "<=":
			return value.num <= rpvn
		case ">":
			return value.num > rpvn
		case ">=":
			return value.num >= rpvn
		}
	case True:
		switch rp.query.op {
		case "=":
			return rpv == "true"
		case "!=":
			return rpv != "true"
		case ">":
			return rpv == "false"
		case ">=":
			return true
		}
	case False:
		switch rp.query.op {
		case "=":
			return rpv == "false"
		case "!=":
			return rpv != "false"
		case "<":
			return rpv == "true"
		case "<=":
			return true
		}
	}
	return false
}

// parseArray parses a JSON array at the given position and processes it according to the path.
// It handles array indexing, array queries, and piped operations on array elements.
func parseArray(c *parseContext, i int, path string) (int, bool) {
	var pmatch, vesc, ok, hit bool
	var val string
	var h int
	var alog []int
	var partidx int
	var multires []byte
	rp := parseArrayPath(path)
	if !rp.arrch {
		n, ok := parseUint(rp.part)
		if !ok {
			partidx = -1
		} else {
			partidx = int(n)
		}
	}
	if !rp.more && rp.piped {
		c.pipe = rp.pipe
		c.piped = true
	}

	// procQuery processes a query against a JSON value and determines if it matches.
	// It also handles collecting all matches for "all" queries.
	procQuery := func(qval *Res) bool {
		if rp.query.all && len(multires) == 0 {
			multires = append(multires, '[')
		}
		var res *Res
		if qval.typ == JSON {
			res = qval.Get(rp.query.path)
		} else {
			if rp.query.path != "" {
				return false
			}
			res = qval
		}
		if queryMatches(&rp, res) {
			if rp.more {
				left, right, ok := splitPossiblePipe(rp.path)
				if ok {
					rp.path = left
					c.pipe = right
					c.piped = true
				}
				res = qval.Get(rp.path)
			} else {
				res = qval
			}
			if rp.query.all {
				raw := res.raw
				if len(raw) == 0 {
					raw = res.String()
				}
				if raw != "" {
					if len(multires) > 1 {
						multires = append(multires, ',')
					}
					multires = append(multires, raw...)
				}
			} else {
				c.value = res
				return true
			}
		}
		return false
	}

	for i < len(c.json)+1 {
		if !rp.arrch {
			pmatch = partidx == h
			hit = pmatch && !rp.more
		}
		h++
		if rp.alogok {
			alog = append(alog, i)
		}
		for ; ; i++ {
			var ch byte
			if i > len(c.json) {
				break
			} else if i == len(c.json) {
				ch = ']'
			} else {
				ch = c.json[i]
			}
			switch ch {
			default:
				continue
			case '"':
				i++
				i, val, vesc, ok = parseString(c.json, i)
				if !ok {
					return i, false
				}
				if rp.query.on {
					qval := &Res{
						raw: val,
						typ: String,
					}
					if vesc {
						qval.str = unescape(val[1 : len(val)-1])
					} else {
						qval.str = val[1 : len(val)-1]
					}
					if procQuery(qval) {
						return i, true
					}
				} else if hit {
					if rp.alogok {
						break
					}
					if vesc {
						c.value.str = unescape(val[1 : len(val)-1])
					} else {
						c.value.str = val[1 : len(val)-1]
					}
					c.value.raw = val
					c.value.typ = String
					return i, true
				}
			case '{':
				if pmatch && !hit {
					i, hit = parseObject(c, i+1, rp.path)
					if hit {
						if rp.alogok {
							break
						}
						return i, true
					}
				} else {
					val, i = parseSquash(c.json, i)
					if rp.query.on {
						if procQuery(&Res{raw: val, typ: JSON}) {
							return i, true
						}
					} else if hit {
						if rp.alogok {
							break
						}
						c.value.raw = val
						c.value.typ = JSON
						return i, true
					}
				}
			case '[':
				if pmatch && !hit {
					i, hit = parseArray(c, i+1, rp.path)
					if hit {
						if rp.alogok {
							break
						}
						return i, true
					}
				} else {
					val, i = parseSquash(c.json, i)
					if rp.query.on {
						if procQuery(&Res{raw: val, typ: JSON}) {
							return i, true
						}
					} else if hit {
						if rp.alogok {
							break
						}
						c.value.raw = val
						c.value.typ = JSON
						return i, true
					}
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumber(c.json, i)
				if rp.query.on {
					qval := &Res{
						raw: val,
						typ: Number,
					}
					qval.num, _ = strconv.ParseFloat(val, 64)
					if procQuery(qval) {
						return i, true
					}
				} else if hit {
					if rp.alogok {
						break
					}
					c.value.raw = val
					c.value.typ = Number
					c.value.num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteral(c.json, i)
				if rp.query.on {
					qval := &Res{
						raw: val,
					}
					switch vc {
					case 't':
						qval.typ = True
					case 'f':
						qval.typ = False
					case 'n':
						qval.typ = Null
					}
					if procQuery(qval) {
						return i, true
					}
				} else if hit {
					if rp.alogok {
						break
					}
					c.value.raw = val
					switch vc {
					case 't':
						c.value.typ = True
					case 'f':
						c.value.typ = False
					}
					return i, true
				}
			case ']':
				if rp.arrch && rp.part == "#" {
					if rp.alogok {
						left, right, ok := splitPossiblePipe(rp.alogkey)
						if ok {
							rp.alogkey = left
							c.pipe = right
							c.piped = true
						}
						jsons := make([]byte, 0, 64)
						jsons = append(jsons, '[')
						for j, k := 0, 0; j < len(alog); j++ {
							_, res, ok := parseAny(c.json, alog[j], true)
							if ok {
								res := res.Get(rp.alogkey)
								if res.Exists() {
									if k > 0 {
										jsons = append(jsons, ',')
									}
									raw := res.raw
									if len(raw) == 0 {
										raw = res.String()
									}
									jsons = append(jsons, zstring.String2Bytes(raw)...)
									k++
								}
							}
						}
						jsons = append(jsons, ']')
						c.value.typ = JSON
						c.value.raw = zstring.Bytes2String(jsons)
						return i + 1, true
					}
					if rp.alogok {
						break
					}

					c.value.typ = Number
					c.value.num = float64(h - 1)
					c.value.raw = strconv.Itoa(h - 1)
					c.calcd = true
					return i + 1, true
				}
				if len(multires) > 0 && !c.value.Exists() {
					c.value = &Res{
						raw: zstring.Bytes2String(append(multires, ']')),
						typ: JSON,
					}
				}
				return i + 1, false
			}
			break
		}
	}
	return i, false
}

// splitPossiblePipe splits a path string at a pipe character ('|').
// Returns the left part, right part, and a boolean indicating if a pipe was found.
func splitPossiblePipe(path string) (left, right string, ok bool) {
	var possible bool
	for i := 0; i < len(path); i++ {
		if path[i] == '|' {
			possible = true
			break
		}
	}
	if !possible {
		return
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '\\' {
			i++
		} else if path[i] == '.' {
			if i == len(path)-1 {
				return
			}
			if path[i+1] == '#' {
				i += 2
				if i == len(path) {
					return
				}
				if path[i] == '[' || path[i] == '(' {
					var start, end byte
					if path[i] == '[' {
						start, end = '[', ']'
					} else {
						start, end = '(', ')'
					}
					i++
					depth := 1
					for ; i < len(path); i++ {
						if path[i] == '\\' {
							i++
						} else if path[i] == start {
							depth++
						} else if path[i] == end {
							depth--
							if depth == 0 {
								break
							}
						} else if path[i] == '"' {
							i++
							for ; i < len(path); i++ {
								if path[i] == '\\' {
									i++
								} else if path[i] == '"' {
									break
								}
							}
						}
					}
				}
			}
		} else if path[i] == '|' {
			return path[:i], path[i+1:], true
		}
	}
	return
}

func ForEachLine(json string, fn func(line *Res) bool) {
	var res *Res
	var i int
	for {
		i, res, _ = parseAny(json, i, true)
		if !res.Exists() {
			break
		}
		if !fn(res) {
			return
		}
	}
}

type subSelector struct {
	name string
	path string
}

func parseSubSelectors(path string) (sels []subSelector, out string, ok bool) {
	depth := 1
	colon := 0
	start := 1
	i := 1
	pushSel := func() {
		var sel subSelector
		if colon == 0 {
			sel.path = path[start:i]
		} else {
			sel.name = path[start:colon]
			sel.path = path[colon+1 : i]
		}
		sels = append(sels, sel)
		colon = 0
		start = i + 1
	}
	for ; i < len(path); i++ {
		switch path[i] {
		case '\\':
			i++
		case ':':
			if depth == 1 {
				colon = i
			}
		case ',':
			if depth == 1 {
				pushSel()
			}
		case '"':
			i++
		loop:
			for ; i < len(path); i++ {
				switch path[i] {
				case '\\':
					i++
				case '"':
					break loop
				}
			}
		case '[', '(', '{':
			depth++
		case ']', ')', '}':
			depth--
			if depth == 0 {
				pushSel()
				path = path[i+1:]
				return sels, path, true
			}
		}
	}
	return
}

func nameOfLast(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '|' || path[i] == '.' {
			if i > 0 && path[i-1] == '\\' {
				continue
			}

			return path[i+1:]
		}
	}
	return path
}

func isSimpleName(component string) bool {
	for i := 0; i < len(component); i++ {
		if component[i] < ' ' {
			return false
		}
		switch component[i] {
		case '[', ']', '{', '}', '(', ')', '#', '|':
			return false
		}
	}
	return true
}

func appendJSONString(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		if s[i] < ' ' || s[i] == '\\' || s[i] == '"' || s[i] > 126 {
			d, _ := json.Marshal(s)
			return append(dst, zstring.Bytes2String(d)...)
		}
	}
	dst = append(dst, '"')
	dst = append(dst, s...)
	dst = append(dst, '"')
	return dst
}

type parseContext struct {
	json  string
	value *Res
	pipe  string
	piped bool
	calcd bool
	lines bool
}

func ModifiersState() bool {
	return openModifiers
}

func SetModifiersState(b bool) {
	openModifiers = b
}

func Get(json, path string) *Res {
	if len(path) > 1 {
		if ModifiersState() && path[0] == '@' {
			var ok bool
			var npath string
			var rjson string
			npath, rjson, ok = execModifier(json, path)
			if ok {
				path = npath
				if len(path) > 0 && (path[0] == '|' || path[0] == '.') {
					res := Get(rjson, path[1:])
					res.index = 0
					return res
				}
				return Parse(rjson)
			}
		}
		if path[0] == '[' || path[0] == '{' {
			kind := path[0]
			var ok bool
			var subs []subSelector
			subs, path, ok = parseSubSelectors(path)
			if ok && len(path) == 0 || (path[0] == '|' || path[0] == '.') {
				var b []byte
				b = append(b, kind)
				var i int
				for _, sub := range subs {
					res := Get(json, sub.path)
					if res.Exists() {
						if i > 0 {
							b = append(b, ',')
						}
						if kind == '{' {
							if len(sub.name) > 0 {
								if sub.name[0] == '"' && Valid(sub.name) {
									b = append(b, sub.name...)
								} else {
									b = appendJSONString(b, sub.name)
								}
							} else {
								last := nameOfLast(sub.path)
								if isSimpleName(last) {
									b = appendJSONString(b, last)
								} else {
									b = appendJSONString(b, "_")
								}
							}
							b = append(b, ':')
						}
						var raw string
						if len(res.raw) == 0 {
							raw = res.String()
							if len(raw) == 0 {
								raw = "null"
							}
						} else {
							raw = res.raw
						}
						b = append(b, raw...)
						i++
					}
				}
				b = append(b, kind+2)
				res := &Res{}
				res.raw = zstring.Bytes2String(b)
				res.typ = JSON
				if len(path) > 0 {
					res = res.Get(path[1:])
				}
				res.index = 0
				return res
			}
		}
	}

	var i int
	c := &parseContext{json: json, value: &Res{}}
	if len(path) >= 2 && path[0] == '.' && path[1] == '.' {
		c.lines = true
		parseArray(c, 0, path[2:])
	} else {
		for ; i < len(c.json); i++ {
			if c.json[i] == '{' {
				i++
				parseObject(c, i, path)
				break
			}
			if c.json[i] == '[' {
				i++
				parseArray(c, i, path)
				break
			}
		}
	}
	if c.piped {
		res := c.value.Get(c.pipe)
		res.index = 0
		return res
	}
	fillIndex(json, c)
	return c.value
}

func GetBytes(json []byte, path string) *Res {
	return Get(zstring.Bytes2String(json), path)
}

func runeit(json string) rune {
	n, _ := strconv.ParseUint(json[:4], 16, 64)
	return rune(n)
}

func unescape(json string) string {
	str := make([]byte, 0, len(json))
	for i := 0; i < len(json); i++ {
		switch {
		case json[i] < ' ':
			return zstring.Bytes2String(str)
		case json[i] == '\\':
			i++
			if i >= len(json) {
				return zstring.Bytes2String(str)
			}
			switch json[i] {
			default:
				return zstring.Bytes2String(str)
			case '\\':
				str = append(str, '\\')
			case '/':
				str = append(str, '/')
			case 'b':
				str = append(str, '\b')
			case 'f':
				str = append(str, '\f')
			case 'n':
				str = append(str, '\n')
			case 'r':
				str = append(str, '\r')
			case 't':
				str = append(str, '\t')
			case '"':
				str = append(str, '"')
			case 'u':
				if i+5 > len(json) {
					return zstring.Bytes2String(str)
				}
				r := runeit(json[i+1:])
				i += 5
				if utf16.IsSurrogate(r) && len(json[i:]) >= 6 && json[i] == '\\' &&
					json[i+1] == 'u' {
					r = utf16.DecodeRune(r, runeit(json[i+2:]))
					i += 6
				}

				str = append(str, 0, 0, 0, 0, 0, 0, 0, 0)
				n := utf8.EncodeRune(str[len(str)-8:], r)
				str = str[:len(str)-8+n]
				i--
			}
		default:
			str = append(str, json[i])
		}
	}
	return zstring.Bytes2String(str)
}

func parseAny(json string, i int, hit bool) (int, *Res, bool) {
	res := &Res{}
	var val string
	for ; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			val, i = parseSquash(json, i)
			if hit {
				res.raw = val
				res.typ = JSON
			}
			return i, res, true
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		case '"':
			i++
			var vesc bool
			var ok bool
			i, val, vesc, ok = parseString(json, i)
			if !ok {
				return i, res, false
			}
			if hit {
				res.typ = String
				res.raw = val
				if vesc {
					res.str = unescape(val[1 : len(val)-1])
				} else {
					res.str = val[1 : len(val)-1]
				}
			}
			return i, res, true
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i, val = parseNumber(json, i)
			if hit {
				res.raw = val
				res.typ = Number
				res.num, _ = strconv.ParseFloat(val, 64)
			}
			return i, res, true
		case 't', 'f', 'n':
			vc := json[i]
			i, val = parseLiteral(json, i)
			if hit {
				res.raw = val
				switch vc {
				case 't':
					res.typ = True
				case 'f':
					res.typ = False
				}
				return i, res, true
			}
		}
	}
	return i, res, false
}

func GetMultiple(json string, path ...string) []*Res {
	res := make([]*Res, len(path))
	for i, path := range path {
		res[i] = Get(json, path)
	}
	return res
}

func GetMultipleBytes(json []byte, path ...string) []*Res {
	res := make([]*Res, len(path))
	for i, path := range path {
		res[i] = GetBytes(json, path)
	}
	return res
}

func assign(jsval *Res, val reflect.Value, fmap *fieldMaps) {
	if jsval.typ == Null {
		return
	}
	// TODO Dev
	t := val.Type()
	switch val.Kind() {
	default:
	case reflect.Ptr:
		if !val.IsNil() {
			elem := val.Elem()
			assign(jsval, elem, fmap)
		} else {
			newval := reflect.New(t.Elem())
			assign(jsval, newval.Elem(), fmap)
			val.Set(newval)
		}
	case reflect.Struct:
		fmap.mu.RLock()
		name := t.String()
		sf := fmap.m[name]
		fmap.mu.RUnlock()
		if sf == nil {
			fmap.mu.Lock()
			sf = make(map[string]int)
			for i := 0; i < t.NumField(); i++ {
				tag, _ := zreflect.GetStructTag(t.Field(i), "json")
				sf[tag] = i
			}
			fmap.m[name] = sf
			fmap.mu.Unlock()
		}
		jsval.ForEach(func(key, value *Res) bool {
			if idx, ok := sf[key.str]; ok {
				f := val.Field(idx)
				if f.CanSet() {
					assign(value, f, fmap)
				}
			}
			return true
		})
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 &&
			jsval.typ == String {
			data, _ := base64.StdEncoding.DecodeString(jsval.String())
			val.Set(zreflect.ValueOf(data))
		} else {
			jsvals := jsval.Array()
			l := len(jsvals)
			slice := reflect.MakeSlice(t, l, l)
			for i := 0; i < l; i++ {
				assign(jsvals[i], slice.Index(i), fmap)
			}
			val.Set(slice)
		}
	case reflect.Array:
		i, n := 0, val.Len()
		jsval.ForEach(func(_, value *Res) bool {
			if i == n {
				return false
			}
			assign(value, val.Index(i), fmap)
			i++
			return true
		})
	case reflect.Map:
		key := t.Key()
		s := key.Kind() == reflect.String
		if s {
			kind := t.Elem().Kind()
			switch kind {
			case reflect.Interface:
				val.Set(zreflect.ValueOf(jsval.Value()))
			default:
				v := reflect.MakeMap(t)
				jsval.ForEach(func(key, value *Res) bool {
					newval := reflect.New(t.Elem())
					elem := newval.Elem()
					assign(value, elem, fmap)
					v.SetMapIndex(zreflect.ValueOf(key.Value()), elem)
					return true
				})
				val.Set(v)
			}
		}
	case reflect.Interface:
		val.Set(zreflect.ValueOf(jsval.Value()))
	case reflect.Bool:
		val.SetBool(jsval.Bool())
	case reflect.Float32, reflect.Float64:
		val.SetFloat(jsval.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		val.SetInt(int64(jsval.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		val.SetUint(uint64(jsval.Uint()))
	case reflect.String:
		val.SetString(jsval.String())
	}
	if len(t.PkgPath()) > 0 {
		v := val.Addr()
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(json.Unmarshaler); ok {
				_ = u.UnmarshalJSON([]byte(jsval.raw))
			}
		}
	}
}

func Unmarshal(json, v interface{}) error {
	var r *Res
	switch v := json.(type) {
	case string:
		r = Parse(v)
	case []byte:
		r = ParseBytes(v)
	case Res:
		r = &v
	}
	if v := zreflect.ValueOf(v); v.Kind() == reflect.Ptr {
		if r.String() == "" {
			return errors.New("invalid json")
		}
		assign(r, v, &fieldMaps{m: make(map[string]map[string]int)})
		return nil
	}
	return errors.New("assignment must be a pointer")
}

func validPayload(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			i, ok = validany(data, i)
			if !ok {
				return i, false
			}
			for ; i < len(data); i++ {
				switch data[i] {
				default:
					return i, false
				case ' ', '\t', '\n', '\r':
					continue
				}
			}
			return i, true
		case ' ', '\t', '\n', '\r':
			continue
		}
	}
	return i, false
}

func validany(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return validobject(data, i+1)
		case '[':
			return validarray(data, i+1)
		case '"':
			return validstring(data, i+1)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return validnumber(data, i+1)
		case 't':
			return validtrue(data, i+1)
		case 'f':
			return validfalse(data, i+1)
		case 'n':
			return validnull(data, i+1)
		default:
			return i, false
		}
	}
	return i, false
}

func validobject(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case '}':
			return i + 1, true
		case '"':
		key:
			if i, ok = validstring(data, i+1); !ok {
				return i, false
			}
			if i, ok = validcolon(data, i); !ok {
				return i, false
			}
			if i, ok = validany(data, i); !ok {
				return i, false
			}
			if i, ok = validcomma(data, i, '}'); !ok {
				return i, false
			}
			if data[i] == '}' {
				return i + 1, true
			}
			i++
			for ; i < len(data); i++ {
				switch data[i] {
				default:
					return i, false
				case ' ', '\t', '\n', '\r':
					continue
				case '"':
					goto key
				}
			}
			return i, false
		}
	}
	return i, false
}

func validcolon(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case ':':
			return i + 1, true
		}
	}
	return i, false
}

func validcomma(data []byte, i int, end byte) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case ',':
			return i, true
		case end:
			return i, true
		}
	}
	return i, false
}

func validarray(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			for ; i < len(data); i++ {
				if i, ok = validany(data, i); !ok {
					return i, false
				}
				if i, ok = validcomma(data, i, ']'); !ok {
					return i, false
				}
				if data[i] == ']' {
					return i + 1, true
				}
			}
		case ' ', '\t', '\n', '\r':
			continue
		case ']':
			return i + 1, true
		}
	}
	return i, false
}

func validstring(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		if data[i] < ' ' {
			return i, false
		} else if data[i] == '\\' {
			i++
			if i == len(data) {
				return i, false
			}
			switch data[i] {
			default:
				return i, false
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				for j := 0; j < 4; j++ {
					i++
					if i >= len(data) {
						return i, false
					}
					if !((data[i] >= '0' && data[i] <= '9') ||
						(data[i] >= 'a' && data[i] <= 'f') ||
						(data[i] >= 'A' && data[i] <= 'F')) {
						return i, false
					}
				}
			}
		} else if data[i] == '"' {
			return i + 1, true
		}
	}
	return i, false
}

func validnumber(data []byte, i int) (outi int, ok bool) {
	i--
	if data[i] == '-' {
		i++
	}
	if i == len(data) {
		return i, false
	}
	if data[i] == '0' {
		i++
	} else {
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	if i == len(data) {
		return i, true
	}
	if data[i] == '.' {
		i++
		if i == len(data) {
			return i, false
		}
		if data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	if i == len(data) {
		return i, true
	}
	if data[i] == 'e' || data[i] == 'E' {
		i++
		if i == len(data) {
			return i, false
		}
		if data[i] == '+' || data[i] == '-' {
			i++
		}
		if i == len(data) {
			return i, false
		}
		if data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	return i, true
}

func validtrue(data []byte, i int) (outi int, ok bool) {
	if i+3 <= len(data) && data[i] == 'r' && data[i+1] == 'u' &&
		data[i+2] == 'e' {
		return i + 3, true
	}
	return i, false
}

func validfalse(data []byte, i int) (outi int, ok bool) {
	if i+4 <= len(data) && data[i] == 'a' && data[i+1] == 'l' &&
		data[i+2] == 's' && data[i+3] == 'e' {
		return i + 4, true
	}
	return i, false
}

func validnull(data []byte, i int) (outi int, ok bool) {
	if i+3 <= len(data) && data[i] == 'u' && data[i+1] == 'l' &&
		data[i+2] == 'l' {
		return i + 3, true
	}
	return i, false
}

func Valid(json string) (ok bool) {
	_, ok = validPayload(zstring.String2Bytes(json), 0)
	return
}

func ValidBytes(json []byte) bool {
	_, ok := validPayload(json, 0)
	return ok
}

// parseUint parses a string into an unsigned integer.
// Returns the parsed value and a boolean indicating success.
func parseUint(s string) (n uint, ok bool) {
	var i int
	if i == len(s) {
		return 0, false
	}
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + uint(s[i]-'0')
		} else {
			return 0, false
		}
	}
	return n, true
}

// parseInt parses a string into a signed integer.
// Returns the parsed value and a boolean indicating success.
func parseInt(s string) (n int, ok bool) {
	var i int
	var sign bool
	if len(s) > 0 && s[0] == '-' {
		sign = true
		i++
	}
	if i == len(s) {
		return 0, false
	}
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int(s[i]-'0')
		} else {
			return 0, false
		}
	}
	if sign {
		return n * -1, true
	}
	return n, true
}

func execModifier(json, path string) (pathOut, res string, ok bool) {
	name := path[1:]
	var hasArgs bool
	for i := 1; i < len(path); i++ {
		if path[i] == ':' {
			pathOut = path[i+1:]
			name = path[1:i]
			hasArgs = len(pathOut) > 0
			break
		}
		if path[i] == '|' {
			pathOut = path[i:]
			name = path[1:i]
			break
		}
		if path[i] == '.' {
			pathOut = path[i:]
			name = path[1:i]
			break
		}
	}
	if fn, ok := modifiers[name]; ok {
		var args string
		if hasArgs {
			var parsedArgs bool
			switch pathOut[0] {
			case '{', '[', '"':
				res := Parse(pathOut)
				if res.Exists() {
					args, _ = parseSquash(pathOut, 0)
					pathOut = pathOut[len(args):]
					parsedArgs = true
				}
			}
			if !parsedArgs {
				idx := strings.IndexByte(pathOut, '|')
				if idx == -1 {
					args = pathOut
					pathOut = ""
				} else {
					args = pathOut[:idx]
					pathOut = pathOut[idx:]
				}
			}
		}
		return pathOut, fn(json, args), true
	}
	return pathOut, res, false
}

var (
	openModifiers = false
	modifiers     = map[string]func(json, arg string) string{
		"format":  modifierPretty,
		"ugly":    modifierUgly,
		"reverse": modifierReverse,
	}
)

// AddModifier registers a new modifier function with the given name.
// Modifiers can be used in paths to transform JSON values.
func AddModifier(name string, fn func(json, arg string) string) {
	modifiers[name] = fn
}

// ModifierExists checks if a modifier with the given name exists.
func ModifierExists(name string) bool {
	_, ok := modifiers[name]
	return ok
}

func modifierPretty(json, arg string) string {
	if len(arg) > 0 {
		opts := *DefOptions
		Parse(arg).ForEach(func(key, value *Res) bool {
			switch key.String() {
			case "sortKeys":
				opts.SortKeys = value.Bool()
			case "indent":
				opts.Indent = value.String()
			case "prefix":
				opts.Prefix = value.String()
			case "width":
				opts.Width = value.Int()
			}
			return true
		})
		return zstring.Bytes2String(FormatOptions(zstring.String2Bytes(json), &opts))
	}
	return zstring.Bytes2String(Format(zstring.String2Bytes(json)))
}

func modifierUgly(json, _ string) string {
	return zstring.Bytes2String(Ugly(zstring.String2Bytes(json)))
}

func modifierReverse(json, _ string) string {
	res := Parse(json)
	if res.IsArray() {
		var values []*Res
		res.ForEach(func(_, value *Res) bool {
			values = append(values, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '[')
		for i, j := len(values)-1, 0; i >= 0; i, j = i-1, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, values[i].raw...)
		}
		out = append(out, ']')
		return zstring.Bytes2String(out)
	}
	if res.IsObject() {
		var keyValues []*Res
		res.ForEach(func(key, value *Res) bool {
			keyValues = append(keyValues, key, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '{')
		for i, j := len(keyValues)-2, 0; i >= 0; i, j = i-2, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, keyValues[i+0].raw...)
			out = append(out, ':')
			out = append(out, keyValues[i+1].raw...)
		}
		out = append(out, '}')
		return zstring.Bytes2String(out)
	}
	return json
}
