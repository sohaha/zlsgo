package zstring

import (
	"strings"
	"unicode/utf8"
)

func Match(str, pattern string, equalFold ...bool) bool {
	if pattern == "*" {
		return true
	}
	var f bool
	if len(equalFold) > 0 {
		f = equalFold[0]
	}
	return deepMatch(str, pattern, f)
}

func deepMatch(str, pattern string, fold bool) bool {
	// label:
	for len(pattern) > 0 {
		if pattern[0] > 0x7f {
			return deepMatchRune(str, pattern, fold)
		}

		switch pattern[0] {
		default:
			if len(str) == 0 {
				return false
			}
			if str[0] > 0x7f {
				return deepMatchRune(str, pattern, fold)
			}

			if !equal(rune(str[0]), rune(pattern[0]), fold) {
				return false
			}
		case '{':
			i, l := 1, len(pattern)
			for ; i < l; i++ {
				if pattern[i] == '}' {
					break
				}
			}
			if i > 2 {
				for _, p := range strings.Split(pattern[1:i], ",") {
					if len(pattern) > i {
						p = p + pattern[i+1:]
					}
					if deepMatch(str, p, fold) {
						return true
					}
				}
				return false
			}
		case '?':
			if len(str) == 0 {
				return false
			}
		case '*':
			return deepMatch(str, pattern[1:], fold) ||
				(len(str) > 0 && deepMatch(str[1:], pattern, fold))
		}
		str = str[1:]
		pattern = pattern[1:]
	}
	return len(str) == 0 && len(pattern) == 0
}

func x7f(str string) (r rune, p int) {
	if len(str) <= 0 {
		return utf8.RuneError, 0
	}
	var s uint8 = str[0]
	if s > 0x7f {
		r, p = utf8.DecodeRuneInString(str)
	} else {
		r, p = rune(s), 1
	}
	return
}

func equal(tr, sr rune, fold bool) bool {
	if tr == sr {
		return true
	}

	if !fold {
		return false
	}

	if tr < sr {
		tr, sr = sr, tr
	}

	return 'A' <= sr && sr <= 'Z' && tr == sr+'a'-'A'
}

func deepMatchRune(str, pattern string, fold bool) bool {
	var sr, pr rune
	var srsz, prsz int

	sr, srsz = x7f(str)
	pr, prsz = x7f(pattern)

	for pr != utf8.RuneError {
		switch pr {
		default:
			if srsz == utf8.RuneError || !equal(sr, pr, fold) {
				return false
			}
		case '?':
			if srsz == utf8.RuneError {
				return false
			}
		case '*':
			return deepMatchRune(str, pattern[prsz:], fold) ||
				(srsz > 0 && deepMatchRune(str[srsz:], pattern, fold))
		}

		pattern = pattern[prsz:]
		str = str[srsz:]

		sr, srsz = x7f(str)
		pr, prsz = x7f(pattern)
	}

	return srsz == 0 && prsz == 0
}

func IsPattern(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] == '*' || str[i] == '?' {
			return true
		}
	}
	return false
}
