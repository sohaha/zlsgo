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
)

type (
	Type int
	Res  struct {
		Type  Type
		Raw   string
		Str   string
		Num   float64
		Index int
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

func (r Res) String() string {
	switch r.Type {
	case False:
		return "false"
	case Number:
		if len(r.Raw) == 0 {
			return strconv.FormatFloat(r.Num, 'f', -1, 64)
		}
		var i int
		if r.Raw[0] == '-' {
			i++
		}
		for ; i < len(r.Raw); i++ {
			if r.Raw[i] < '0' || r.Raw[i] > '9' {
				return strconv.FormatFloat(r.Num, 'f', -1, 64)
			}
		}
		return r.Raw
	case String:
		return r.Str
	case JSON:
		return r.Raw
	case True:
		return "true"
	default:
		return ""
	}
}

func (r Res) Bool() bool {
	switch r.Type {
	case True:
		return true
	case String:
		b, _ := strconv.ParseBool(strings.ToLower(r.Str))
		return b
	case Number:
		return r.Num != 0
	default:
		return false
	}
}

func (r Res) Int() int {
	switch r.Type {
	case True:
		return 1
	case String:
		n, _ := parseInt(r.Str)
		return n
	case Number:
		i, ok := safeInt(r.Num)
		if ok {
			return i
		}
		// now try to parse the raw string
		i, ok = parseInt(r.Raw)
		if ok {
			return i
		}
		return int(r.Num)
	default:
		return 0
	}
}

func (r Res) Uint() uint {
	switch r.Type {
	default:
		return 0
	case True:
		return 1
	case String:
		n, _ := parseUint(r.Str)
		return n
	case Number:
		i, ok := safeInt(r.Num)
		if ok && i >= 0 {
			return uint(i)
		}
		u, ok := parseUint(r.Raw)
		if ok {
			return u
		}
		return uint(r.Num)
	}
}

func (r Res) Float() float64 {
	switch r.Type {
	default:
		return 0
	case True:
		return 1
	case String:
		n, _ := strconv.ParseFloat(r.Str, 64)
		return n
	case Number:
		return r.Num
	}
}

func (r Res) Unmarshal(v interface{}) error {
	return Unmarshal(r.Raw, v)
}

func (r Res) Time(format ...string) time.Time {
	f := "2006-01-02 15:04:05"
	if len(format) > 0 {
		f = format[0]
	}
	loc, _ := time.LoadLocation("Local")
	res, _ := time.ParseInLocation(f, r.String(), loc)
	return res
}

func (r Res) Array() []Res {
	if r.Type == Null {
		return []Res{}
	}
	if r.Type != JSON {
		return []Res{r}
	}
	rr := r.arrayOrMap('[', false)
	return rr.a
}

func (r Res) IsObject() bool {
	return r.Type == JSON && len(r.Raw) > 0 && r.Raw[0] == '{'
}

func (r Res) IsArray() bool {
	return r.Type == JSON && len(r.Raw) > 0 && r.Raw[0] == '['
}

func (r Res) ForEach(iterator func(key, value Res) bool) {
	if !r.Exists() {
		return
	}
	if r.Type != JSON {
		iterator(Res{}, r)
		return
	}
	j := r.Raw
	var keys bool
	var i int
	var key, value Res
	for ; i < len(j); i++ {
		if j[i] == '{' {
			i++
			key.Type = String
			keys = true
			break
		} else if j[i] == '[' {
			i++
			break
		}
		if j[i] > ' ' {
			return
		}
	}
	var str string
	var vesc bool
	var ok bool
	for ; i < len(j); i++ {
		if keys {
			if j[i] != '"' {
				continue
			}
			s := i
			i, str, vesc, ok = parseString(j, i+1)
			if !ok {
				return
			}
			if vesc {
				key.Str = unescape(str[1 : len(str)-1])
			} else {
				key.Str = str[1 : len(str)-1]
			}
			key.Raw = str
			key.Index = s
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
		value.Index = s
		if !iterator(key, value) {
			return
		}
	}
}

func (r Res) Map() map[string]Res {
	if r.Type != JSON {
		return map[string]Res{}
	}
	rr := r.arrayOrMap('{', false)
	return rr.o
}

func (r Res) MapKeys(exclude ...string) (keys []string) {
	m := r.Map()
	keys = make([]string, 0, len(m))
	for k := range m {
		skip := false
		for i := range exclude {
			if k == exclude[i] {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		keys = append(keys, k)
	}
	return
}

func (r Res) Get(path string) Res {
	return Get(r.Raw, path)
}

type arrayOrMapResult struct {
	a  []Res
	ai []interface{}
	o  map[string]Res
	oi map[string]interface{}
	vc byte
}

func (r Res) arrayOrMap(vc byte, valueize bool) (ar arrayOrMapResult) {
	var (
		value Res
		count int
		key   Res
		i     int
		j     = r.Raw
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
			ar.o = make(map[string]Res)
		}
	} else {
		if valueize {
			ar.ai = make([]interface{}, 0)
		} else {
			ar.a = make([]Res, 0)
		}
	}
	for ; i < len(j); i++ {
		if j[i] <= ' ' {
			continue
		}
		if j[i] == ']' || j[i] == '}' {
			break
		}
		switch j[i] {
		default:
			if (j[i] >= '0' && j[i] <= '9') || j[i] == '-' {
				value.Type = Number
				value.Raw, value.Num = tonum(j[i:])
				value.Str = ""
			} else {
				continue
			}
		case '{', '[':
			value.Type = JSON
			value.Raw = squash(j[i:])
			value.Str, value.Num = "", 0
		case 'n':
			value.Type = Null
			value.Raw = tolit(j[i:])
			value.Str, value.Num = "", 0
		case 'r':
			value.Type = True
			value.Raw = tolit(j[i:])
			value.Str, value.Num = "", 0
		case 'f':
			value.Type = False
			value.Raw = tolit(j[i:])
			value.Str, value.Num = "", 0
		case '"':
			value.Type = String
			value.Raw, value.Str = tostr(j[i:])
			value.Num = 0
		}
		i += len(value.Raw) - 1

		if ar.vc == '{' {
			if count%2 == 0 {
				key = value
			} else {
				if valueize {
					if _, ok := ar.oi[key.Str]; !ok {
						ar.oi[key.Str] = value.Value()
					}
				} else {
					if _, ok := ar.o[key.Str]; !ok {
						ar.o[key.Str] = value
					}
				}
			}
			count++
		} else {
			if valueize {
				ar.ai = append(ar.ai, value.Value())
			} else {
				ar.a = append(ar.a, value)
			}
		}
	}
end:
	return
}

func Parse(json string) Res {
	var value Res
	for i := 0; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			value.Type = JSON
			value.Raw = json[i:]
			break
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		default:
			if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' {
				value.Type = Number
				value.Raw, value.Num = tonum(json[i:])
			} else {
				return Res{}
			}
		case 'n':
			value.Type = Null
			value.Raw = tolit(json[i:])
		case 't':
			value.Type = True
			value.Raw = tolit(json[i:])
		case 'f':
			value.Type = False
			value.Raw = tolit(json[i:])
		case '"':
			value.Type = String
			value.Raw, value.Str = tostr(json[i:])
		}
		break
	}
	return value
}

func ParseBytes(json []byte) Res {
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

func (r Res) Exists() bool {
	return r.Type != Null || len(r.Raw) != 0
}

func (r Res) Value() interface{} {
	if r.Type == String {
		return r.Str
	}
	switch r.Type {
	default:
		return nil
	case False:
		return false
	case Number:
		return r.Num
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
	var s = i
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
	var s = i
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
	var s = i
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
	piped   bool
	more    bool
	alogok  bool
	arrch   bool
	alogkey string
	query   struct {
		on    bool
		path  string
		op    string
		value string
		all   bool
	}
}

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
				var s = i
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
						c.value.Str = unescape(val[1 : len(val)-1])
					} else {
						c.value.Str = val[1 : len(val)-1]
					}
					c.value.Raw = val
					c.value.Type = String
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
						c.value.Raw = val
						c.value.Type = JSON
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
						c.value.Raw = val
						c.value.Type = JSON
						return i, true
					}
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumber(c.json, i)
				if hit {
					c.value.Raw = val
					c.value.Type = Number
					c.value.Num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteral(c.json, i)
				if hit {
					c.value.Raw = val
					switch vc {
					case 't':
						c.value.Type = True
					case 'f':
						c.value.Type = False
					}
					return i, true
				}
			}
			break
		}
	}
	return i, false
}
func queryMatches(rp *arrayPathResult, value Res) bool {
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
	switch value.Type {
	case String:
		switch rp.query.op {
		case "=":
			return value.Str == rpv
		case "!=":
			return value.Str != rpv
		case "<":
			return value.Str < rpv
		case "<=":
			return value.Str <= rpv
		case ">":
			return value.Str > rpv
		case ">=":
			return value.Str >= rpv
		case "%":
			return zstring.Match(value.Str, rpv)
		case "!%":
			return !zstring.Match(value.Str, rpv)
		}
	case Number:
		rpvn, _ := strconv.ParseFloat(rpv, 64)
		switch rp.query.op {
		case "=":
			return value.Num == rpvn
		case "!=":
			return value.Num != rpvn
		case "<":
			return value.Num < rpvn
		case "<=":
			return value.Num <= rpvn
		case ">":
			return value.Num > rpvn
		case ">=":
			return value.Num >= rpvn
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

	procQuery := func(qval Res) bool {
		if rp.query.all && len(multires) == 0 {
			multires = append(multires, '[')
		}
		var res Res
		if qval.Type == JSON {
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
				raw := res.Raw
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
					var qval Res
					if vesc {
						qval.Str = unescape(val[1 : len(val)-1])
					} else {
						qval.Str = val[1 : len(val)-1]
					}
					qval.Raw = val
					qval.Type = String
					if procQuery(qval) {
						return i, true
					}
				} else if hit {
					if rp.alogok {
						break
					}
					if vesc {
						c.value.Str = unescape(val[1 : len(val)-1])
					} else {
						c.value.Str = val[1 : len(val)-1]
					}
					c.value.Raw = val
					c.value.Type = String
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
						if procQuery(Res{Raw: val, Type: JSON}) {
							return i, true
						}
					} else if hit {
						if rp.alogok {
							break
						}
						c.value.Raw = val
						c.value.Type = JSON
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
						if procQuery(Res{Raw: val, Type: JSON}) {
							return i, true
						}
					} else if hit {
						if rp.alogok {
							break
						}
						c.value.Raw = val
						c.value.Type = JSON
						return i, true
					}
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumber(c.json, i)
				if rp.query.on {
					var qval Res
					qval.Raw = val
					qval.Type = Number
					qval.Num, _ = strconv.ParseFloat(val, 64)
					if procQuery(qval) {
						return i, true
					}
				} else if hit {
					if rp.alogok {
						break
					}
					c.value.Raw = val
					c.value.Type = Number
					c.value.Num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteral(c.json, i)
				if rp.query.on {
					var qval Res
					qval.Raw = val
					switch vc {
					case 't':
						qval.Type = True
					case 'f':
						qval.Type = False
					}
					if procQuery(qval) {
						return i, true
					}
				} else if hit {
					if rp.alogok {
						break
					}
					c.value.Raw = val
					switch vc {
					case 't':
						c.value.Type = True
					case 'f':
						c.value.Type = False
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
						var jsons = make([]byte, 0, 64)
						jsons = append(jsons, '[')
						for j, k := 0, 0; j < len(alog); j++ {
							_, res, ok := parseAny(c.json, alog[j], true)
							if ok {
								res := res.Get(rp.alogkey)
								if res.Exists() {
									if k > 0 {
										jsons = append(jsons, ',')
									}
									raw := res.Raw
									if len(raw) == 0 {
										raw = res.String()
									}
									jsons = append(jsons, zstring.String2Bytes(raw)...)
									k++
								}
							}
						}
						jsons = append(jsons, ']')
						c.value.Type = JSON
						c.value.Raw = zstring.Bytes2String(jsons)
						return i + 1, true
					}
					if rp.alogok {
						break
					}

					c.value.Type = Number
					c.value.Num = float64(h - 1)
					c.value.Raw = strconv.Itoa(h - 1)
					c.calcd = true
					return i + 1, true
				}
				if len(multires) > 0 && !c.value.Exists() {
					c.value = Res{
						Raw:  zstring.Bytes2String(append(multires, ']')),
						Type: JSON,
					}
				}
				return i + 1, false
			}
			break
		}
	}
	return i, false
}

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

func ForEachLine(json string, iterator func(line Res) bool) {
	var res Res
	var i int
	for {
		i, res, _ = parseAny(json, i, true)
		if !res.Exists() {
			break
		}
		if !iterator(res) {
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
	value Res
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

func Get(json, path string) Res {
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
					res.Index = 0
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
						if len(res.Raw) == 0 {
							raw = res.String()
							if len(raw) == 0 {
								raw = "null"
							}
						} else {
							raw = res.Raw
						}
						b = append(b, raw...)
						i++
					}
				}
				b = append(b, kind+2)
				var res Res
				res.Raw = zstring.Bytes2String(b)
				res.Type = JSON
				if len(path) > 0 {
					res = res.Get(path[1:])
				}
				res.Index = 0
				return res
			}

		}
	}

	var i int
	var c = &parseContext{json: json}
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
		res.Index = 0
		return res
	}
	fillIndex(json, c)
	return c.value
}

func GetBytes(json []byte, path string) Res {
	return Get(zstring.Bytes2String(json), path)
}

func runeit(json string) rune {
	n, _ := strconv.ParseUint(json[:4], 16, 64)
	return rune(n)
}

func unescape(json string) string {
	var str = make([]byte, 0, len(json))
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

func parseAny(json string, i int, hit bool) (int, Res, bool) {
	var res Res
	var val string
	for ; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			val, i = parseSquash(json, i)
			if hit {
				res.Raw = val
				res.Type = JSON
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
				res.Type = String
				res.Raw = val
				if vesc {
					res.Str = unescape(val[1 : len(val)-1])
				} else {
					res.Str = val[1 : len(val)-1]
				}
			}
			return i, res, true
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i, val = parseNumber(json, i)
			if hit {
				res.Raw = val
				res.Type = Number
				res.Num, _ = strconv.ParseFloat(val, 64)
			}
			return i, res, true
		case 't', 'f', 'n':
			vc := json[i]
			i, val = parseLiteral(json, i)
			if hit {
				res.Raw = val
				switch vc {
				case 't':
					res.Type = True
				case 'f':
					res.Type = False
				}
				return i, res, true
			}
		}
	}
	return i, res, false
}

func GetMultiple(json string, path ...string) []Res {
	res := make([]Res, len(path))
	for i, path := range path {
		res[i] = Get(json, path)
	}
	return res
}

func GetMultipleBytes(json []byte, path ...string) []Res {
	res := make([]Res, len(path))
	for i, path := range path {
		res[i] = GetBytes(json, path)
	}
	return res
}

var (
	fieldsmu  sync.RWMutex
	fieldsMap sync.Map
	fields    = make(map[string]map[string]int)
)

func assign(jsval Res, val reflect.Value) {
	if jsval.Type == Null {
		return
	}
	t := val.Type()
	switch val.Kind() {
	default:
	case reflect.Ptr:
		if !val.IsNil() {
			elem := val.Elem()
			assign(jsval, elem)
		} else {
			newval := reflect.New(t.Elem())
			assign(jsval, newval.Elem())
			val.Set(newval)
		}
	case reflect.Struct:
		fieldsmu.RLock()
		name := t.String()
		sf := fields[name]
		fieldsmu.RUnlock()
		if sf == nil {
			fieldsmu.Lock()
			sf = make(map[string]int)
			for i := 0; i < t.NumField(); i++ {
				sf[zreflect.GetStructTag(t.Field(i))] = i
			}
			// fieldsMap.Store(t.String(), sf)
			fields[name] = sf
			fieldsmu.Unlock()
		}
		jsval.ForEach(func(key, value Res) bool {
			if idx, ok := sf[key.Str]; ok {
				f := val.Field(idx)
				if f.CanSet() {
					assign(value, f)
				}
			}
			return true
		})
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 &&
			jsval.Type == String {
			data, _ := base64.StdEncoding.DecodeString(jsval.String())
			val.Set(reflect.ValueOf(data))
		} else {
			jsvals := jsval.Array()
			l := len(jsvals)
			slice := reflect.MakeSlice(t, l, l)
			for i := 0; i < l; i++ {
				assign(jsvals[i], slice.Index(i))
			}
			val.Set(slice)
		}
	case reflect.Array:
		i, n := 0, val.Len()
		jsval.ForEach(func(_, value Res) bool {
			if i == n {
				return false
			}
			assign(value, val.Index(i))
			i++
			return true
		})
	case reflect.Map:
		if t.Key().Kind() == reflect.String &&
			t.Elem().Kind() == reflect.Interface {
			val.Set(reflect.ValueOf(jsval.Value()))
		}
	case reflect.Interface:
		val.Set(reflect.ValueOf(jsval.Value()))
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
				_ = u.UnmarshalJSON([]byte(jsval.Raw))
			}
		}
	}
}

func Unmarshal(json, v interface{}) error {

	var data []byte
	switch v := json.(type) {
	case string:
		data = zstring.String2Bytes(v)
	case []byte:
		data = v
	}
	if v := reflect.ValueOf(v); v.Kind() == reflect.Ptr {
		r := ParseBytes(data)
		if r.String() == "" {
			return errors.New("invalid json")
		}
		assign(r, v)
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

func AddModifier(name string, fn func(json, arg string) string) {
	modifiers[name] = fn
}

func ModifierExists(name string) bool {
	_, ok := modifiers[name]
	return ok
}

func modifierPretty(json, arg string) string {
	if len(arg) > 0 {
		opts := *DefOptions
		Parse(arg).ForEach(func(key, value Res) bool {
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
		var values []Res
		res.ForEach(func(_, value Res) bool {
			values = append(values, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '[')
		for i, j := len(values)-1, 0; i >= 0; i, j = i-1, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, values[i].Raw...)
		}
		out = append(out, ']')
		return zstring.Bytes2String(out)
	}
	if res.IsObject() {
		var keyValues []Res
		res.ForEach(func(key, value Res) bool {
			keyValues = append(keyValues, key, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '{')
		for i, j := len(keyValues)-2, 0; i >= 0; i, j = i-2, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, keyValues[i+0].Raw...)
			out = append(out, ':')
			out = append(out, keyValues[i+1].Raw...)
		}
		out = append(out, '}')
		return zstring.Bytes2String(out)
	}
	return json
}

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
