package zvalid

import (
	"strings"
	"unicode"
)

// HasLetter must contain letters not case sensitive
func (v *Engine) HasLetter(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}

		for _, rv := range v.value {
			if unicode.IsLetter(rv) {
				return v
			}
		}

		v.err = v.setError("必须包含字母", customError...)
		return v
	})
}

// HasLower must contain lowercase letters
func (v *Engine) HasLower(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}

		for _, rv := range v.value {
			if unicode.IsLower(rv) {
				return v
			}
		}

		v.err = v.setError("必须包含小写字母", customError...)
		return v
	})
}

// HasUpper must contain uppercase letters
func (v *Engine) HasUpper(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if unicode.IsUpper(rv) {
				return v
			}
		}

		v.err = v.setError("必须包含大写字母", customError...)
		return v
	})
}

// HasNumber must contain numbers
func (v *Engine) HasNumber(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if unicode.IsDigit(rv) {
				return v
			}
		}

		v.err = v.setError("必须包含数字", customError...)
		return v
	})
}

// HasSymbol must contain symbols
func (v *Engine) HasSymbol(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.IsDigit(rv) && !unicode.IsLetter(rv) && !unicode.Is(unicode.Han, rv) {
				return v
			}
		}

		v.err = v.setError("必须包含符号", customError...)
		return v
	})
}

// HasString must contain a specific string
func (v *Engine) HasString(sub string, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if !ignore(v) && !strings.Contains(v.value, sub) {
			v.err = v.setError("必须包含特定的字符串", customError...)
			return v
		}
		return v
	})
}

// HasPrefix must contain the specified prefix string
func (v *Engine) HasPrefix(sub string, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if !ignore(v) && !strings.HasPrefix(v.value, sub) {
			v.err = v.setError("不允许的值", customError...)
			return v
		}
		return v
	})
}

// HasSuffix contains the specified suffix string
func (v *Engine) HasSuffix(sub string, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if !ignore(v) && !strings.HasSuffix(v.value, sub) {
			v.err = v.setError("不允许的值", customError...)
			return v
		}
		return v
	})
}

// Password Universal password (any visible character, length between 6 ~ 20)
func (v *Engine) Password(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		v.MinLength(6).MaxLength(20)
		if v.err != nil && len(customError) > 0 {
			v.err = v.setError(customError[0])
		}
		return v
	})
}

// StrongPassword Strong equal strength password (length is 6 ~ 20, must include uppercase and lowercase letters, numbers and special characters)
func (v *Engine) StrongPassword(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		v.MinLength(6).MaxLength(20).HasSymbol().HasNumber().HasLetter().HasLower()
		if v.err != nil && len(customError) > 0 {
			v.err = v.setError(customError[0])
		}
		return v
	})
}
