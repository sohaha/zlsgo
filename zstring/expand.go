package zstring

// Expand replaces ${var} or $var in the string based on the mapping function.
// It's similar to shell variable expansion, supporting both ${var} and $var syntax.
func Expand(s string, process func(key string) string) string {
	var buf []byte
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getShellName(s[j+1:])
			if name != "" {
				buf = append(buf, process(name)...)
			} else if w == 0 {
				buf = append(buf, s[j])
			}
			j += w
			i = j + 1
		}
	}

	if buf == nil {
		return s
	}

	return Bytes2String(buf) + s[i:]
}

// getShellName extracts a shell variable name from a string starting with a variable reference.
// It returns the variable name and the number of bytes consumed from the input string.
func getShellName(s string) (string, int) {
	switch {
	case s[0] == '{':
		if len(s) > 2 && isShellSpecialVar(s[1]) && s[2] == '}' {
			return s[1:2], 3
		}
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				if i == 1 {
					return "", 2
				}
				return s[1:i], i + 1
			}
		}
		return "", 1
	case isShellSpecialVar(s[0]):
		return s[0:1], 1
	}
	var i int
	for i = 0; i < len(s) && isAlphaNum(s[i]); i++ {
	}
	return s[:i], i
}

// isShellSpecialVar checks if a character is a special shell variable character.
// Special variables include *, #, $, @, !, ?, -, and digits 0-9.
func isShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// isAlphaNum checks if a character is alphanumeric or underscore.
// These characters are valid in variable names.
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}
