package zjson

import (
	"bytes"
	"sort"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	Map             map[string]string
	StFormatOptions struct {
		Prefix   string
		Indent   string
		Width    int
		SortKeys bool
	}
	pair struct {
		ks, kd int
		vs, vd int
	}
	byKey struct {
		json   []byte
		pairs  []pair
		sorted bool
	}
)

var (
	DefOptions = &StFormatOptions{Width: 80, Prefix: "", Indent: "  ", SortKeys: false}
	Matches    = []Map{
		{"start": "//", "end": "\n"},
		{"start": "/*", "end": "*/"},
	}
)

func Format(json []byte) []byte { return FormatOptions(json, nil) }

func FormatOptions(json []byte, opts *StFormatOptions) []byte {
	if opts == nil {
		opts = DefOptions
	}
	buf := make([]byte, 0, len(json))
	if len(opts.Prefix) != 0 {
		buf = append(buf, opts.Prefix...)
	}
	buf, _, _, _ = appendAny(buf, json, 0, true,
		opts.Width, opts.Prefix, opts.Indent, opts.SortKeys,
		0, 0, -1)
	if len(buf) > 0 {
		buf = append(buf, '\n')
	}
	return buf
}

func Ugly(json []byte) []byte {
	jsonStr, err := Discard(zstring.Bytes2String(json))
	if err == nil {
		json = zstring.String2Bytes(jsonStr)
	}
	buf := make([]byte, 0, len(json))
	return ugly(buf, json)
}

func ugly(dst, src []byte) []byte {
	dst = dst[:0]
	for i := 0; i < len(src); i++ {
		if src[i] > ' ' {
			dst = append(dst, src[i])
			if src[i] == '"' {
				for i = i + 1; i < len(src); i++ {
					dst = append(dst, src[i])
					if src[i] == '"' {
						j := i - 1
						for ; ; j-- {
							if src[j] != '\\' {
								break
							}
						}
						if (j-i)%2 != 0 {
							break
						}
					}
				}
			}
		}
	}
	return dst
}

func appendAny(buf, json []byte, i int, pretty bool, width int, prefix, indent string, sortkeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == '"' {
			return appendString(buf, json, i, nl)
		}
		if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' {
			return appendNumber(buf, json, i, nl)
		}
		if json[i] == '{' {
			return appendObject(buf, json, i, '{', '}', pretty, width, prefix, indent, sortkeys, tabs, nl, max)
		}
		if json[i] == '[' {
			return appendObject(buf, json, i, '[', ']', pretty, width, prefix, indent, sortkeys, tabs, nl, max)
		}
		switch json[i] {
		case 't':
			return append(buf, 't', 'r', 'u', 'e'), i + 4, nl, true
		case 'f':
			return append(buf, 'f', 'a', 'l', 's', 'e'), i + 5, nl, true
		case 'n':
			return append(buf, 'n', 'u', 'l', 'l'), i + 4, nl, true
		}
	}
	return buf, i, nl, true
}

func (arr *byKey) Len() int {
	return len(arr.pairs)
}

func (arr *byKey) Less(i, j int) bool {
	key1 := arr.json[arr.pairs[i].ks+1 : arr.pairs[i].kd-1]
	key2 := arr.json[arr.pairs[j].ks+1 : arr.pairs[j].kd-1]
	return zstring.Bytes2String(key1) < zstring.Bytes2String(key2)
}

func (arr *byKey) Swap(i, j int) {
	arr.pairs[i], arr.pairs[j] = arr.pairs[j], arr.pairs[i]
	arr.sorted = true
}

func appendObject(buf, json []byte, i int, open, close byte, pretty bool, width int, prefix, indent string, sortkeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	var ok bool
	if width > 0 {
		if pretty && open == '[' && max == -1 {
			max := width - (len(buf) - nl)
			if max > 3 {
				s1, s2 := len(buf), i
				buf, i, _, ok = appendObject(buf, json, i, '[', ']', false, width, prefix, "", sortkeys, 0, 0, max)
				if ok && len(buf)-s1 <= max {
					return buf, i, nl, true
				}
				buf = buf[:s1]
				i = s2
			}
		} else if max != -1 && open == '{' {
			return buf, i, nl, false
		}
	}
	buf = append(buf, open)
	i++
	var pairs []pair
	if open == '{' && sortkeys {
		pairs = make([]pair, 0, 8)
	}
	var n int
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == close {
			if pretty {
				if open == '{' && sortkeys {
					buf = sortPairs(json, buf, pairs)
				}
				if n > 0 {
					nl = len(buf)
					buf = append(buf, '\n')
				}
				if buf[len(buf)-1] != open {
					buf = appendTabs(buf, prefix, indent, tabs)
				}
			}
			buf = append(buf, close)
			return buf, i + 1, nl, open != '{'
		}
		if open == '[' || json[i] == '"' {
			if n > 0 {
				buf = append(buf, ',')
				if width != -1 && open == '[' {
					buf = append(buf, ' ')
				}
			}
			var p pair
			if pretty {
				nl = len(buf)
				buf = append(buf, '\n')
				if open == '{' && sortkeys {
					p.ks = i
					p.vs = len(buf)
				}
				buf = appendTabs(buf, prefix, indent, tabs+1)
			}
			if open == '{' {
				buf, i, nl, _ = appendString(buf, json, i, nl)
				if sortkeys {
					p.kd = i
				}
				buf = append(buf, ':')
				if pretty {
					buf = append(buf, ' ')
				}
			}
			buf, i, nl, ok = appendAny(buf, json, i, pretty, width, prefix, indent, sortkeys, tabs+1, nl, max)
			if max != -1 && !ok {
				return buf, i, nl, false
			}
			if pretty && open == '{' && sortkeys {
				p.vd = len(buf)
				if p.ks > p.kd || p.vs > p.vd {
					sortkeys = false
				} else {
					pairs = append(pairs, p)
				}
			}
			i--
			n++
		}
	}
	return buf, i, nl, open != '{'
}

func sortPairs(json, buf []byte, pairs []pair) []byte {
	if len(pairs) == 0 {
		return buf
	}
	vstart := pairs[0].vs
	vend := pairs[len(pairs)-1].vd
	arr := byKey{sorted: false, json: json, pairs: pairs}
	sort.Sort(&arr)
	if !arr.sorted {
		return buf
	}
	nbuf := make([]byte, 0, vend-vstart)
	for i, p := range pairs {
		nbuf = append(nbuf, buf[p.vs:p.vd]...)
		if i < len(pairs)-1 {
			nbuf = append(nbuf, ',')
			nbuf = append(nbuf, '\n')
		}
	}
	return append(buf[:vstart], nbuf...)
}

func appendString(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] == '"' {
			var sc int
			for j := i - 1; j > s; j-- {
				if json[j] == '\\' {
					sc++
				} else {
					break
				}
			}
			if sc%2 == 1 {
				continue
			}
			i++
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

func appendNumber(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ':' || json[i] == ']' || json[i] == '}' {
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

func appendTabs(buf []byte, prefix, indent string, tabs int) []byte {
	if len(prefix) != 0 {
		buf = append(buf, prefix...)
	}
	if len(indent) == 2 && indent[0] == ' ' && indent[1] == ' ' {
		for i := 0; i < tabs; i++ {
			buf = append(buf, ' ', ' ')
		}
	} else {
		for i := 0; i < tabs; i++ {
			buf = append(buf, indent...)
		}
	}
	return buf
}

func Discard(json string) (string, error) {
	var (
		buffer    bytes.Buffer
		flag      int
		v         rune
		protected bool
	)
	runes := []rune(json)
	flag = -1
	for i := 0; i < len(runes); {
		v = runes[i]
		if flag == -1 {
			for f, v := range Matches {
				l := match(&runes, i, v["start"])
				if l != 0 {
					flag = f
					i += l
					break
				}
			}
			if flag == -1 {
				if protected {
					buffer.WriteRune(v)
					if v == '"' {
						protected = true
					}
				} else {
					r := filter(v)
					if r != 0 {
						buffer.WriteRune(v)
					}
				}
			} else {
				continue
			}
		} else {
			l := match(&runes, i, Matches[flag]["end"])
			if l != 0 {
				flag = -1
				i += l
				continue
			}
		}
		i++
	}
	return buffer.String(), nil
}

func filter(v rune) rune {
	switch v {
	case ' ':
	case '\n':
	case '\t':
	default:
		return v
	}
	return 0
}

func match(runes *[]rune, i int, dst string) int {
	dstLen := len([]rune(dst))
	if len(*runes)-i >= dstLen && string((*runes)[i:i+dstLen]) == dst {
		return dstLen
	}
	return 0
}
