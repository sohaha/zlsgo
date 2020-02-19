package zvalid

import (
	"fmt"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
	"net"
	"strconv"
	"strings"
	"unicode"
)

// IsBool booleanValue
func (v *Engine) IsBool(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if _, err := strconv.ParseBool(v.value); err != nil {
			v.err = v.setError("必须是布尔值", customError...)
		}
		return v
	})
}

// IsLower lowerCase letters
func (v *Engine) IsLower(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.IsLower(rv) {
				v.err = v.setError("必须是小写字母", customError...)
				return v
			}
		}
		return v
	})
}

// IsUpper uppercase letter
func (v *Engine) IsUpper(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.IsUpper(rv) {
				v.err = v.setError("必须是大写字母", customError...)
				return v
			}
		}
		return v
	})
}

// IsLetter uppercase and lowercase letters
func (v *Engine) IsLetter(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.IsLetter(rv) {
				v.err = v.setError("必须是字母", customError...)
				return v
			}
		}
		return v
	})
}

// IsNumber number
func (v *Engine) IsNumber(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if _, err := strconv.Atoi(v.value); err != nil {
			v.err = v.setError("必须是数字", customError...)
		}
		return v
	})
}

// IsLowerOrDigit lowercase letters or numbers
func (v *Engine) IsLowerOrDigit(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.IsLower(rv) && !unicode.IsDigit(rv) {
				v.err = v.setError("必须是小写字母或数字", customError...)
				return v
			}
		}
		return v
	})
}

// IsUpperOrDigit uppercase letters or numbers
func (v *Engine) IsUpperOrDigit(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.IsUpper(rv) && !unicode.IsDigit(rv) {
				v.err = v.setError("必须是大写字母或数字", customError...)
				return v
			}
		}
		return v
	})
}

// IsLetterOrDigit letters or numbers
func (v *Engine) IsLetterOrDigit(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.IsLetter(rv) && !unicode.IsDigit(rv) {
				v.err = v.setError("必须是字母或数字", customError...)
				return v
			}
		}
		return v
	})
}

// IsChinese chinese character
func (v *Engine) IsChinese(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for _, rv := range v.value {
			if !unicode.Is(unicode.Scripts["Han"], rv) {
				v.err = v.setError("必须是汉字", customError...)
				return v
			}
		}
		return v
	})
}

// IsMobile chinese mobile
func (v *Engine) IsMobile(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if zstring.RegexMatch(`^1[\d]{10}$`, v.value) {
			v.err = v.setError("必须是手机号码", customError...)
			return v
		}

		return v
	})
}

// IsMail email address
func (v *Engine) IsMail(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		emailSlice := strings.Split(v.value, "@")
		if len(emailSlice) != 2 {
			v.err = v.setError("格式不正确", customError...)
			return v
		}
		if emailSlice[0] == "" || emailSlice[1] == "" {
			v.err = v.setError("格式不正确", customError...)
			return v
		}

		for k, rv := range emailSlice[0] {
			if k == 0 && !unicode.IsLetter(rv) && !unicode.IsDigit(rv) {
				v.err = v.setError("格式不正确", customError...)
				return v
			} else if !unicode.IsLetter(rv) && !unicode.IsDigit(rv) && rv != '@' && rv != '.' && rv != '_' && rv != '-' {
				v.err = v.setError("格式不正确", customError...)
				return v
			}
		}

		domainSlice := strings.Split(emailSlice[1], ".")
		if len(domainSlice) < 2 {
			v.err = v.setError("格式不正确", customError...)
			return v
		}
		domainSliceLen := len(domainSlice)
		for i := 0; i < domainSliceLen; i++ {
			for k, rv := range domainSlice[i] {
				// nolint
				if i != domainSliceLen-1 && k == 0 && !unicode.IsLetter(rv) && !unicode.IsDigit(rv) {
					v.err = v.setError("格式不正确", customError...)
					return v
				} else if !unicode.IsLetter(rv) && !unicode.IsDigit(rv) && rv != '.' && rv != '_' && rv != '-' {
					v.err = v.setError("格式不正确", customError...)
					return v
				} else if i == domainSliceLen-1 && !unicode.IsLetter(rv) {
					v.err = v.setError("格式不正确", customError...)
					return v
				}
			}
		}

		return v
	})
}

// isURL links
func (v *Engine) IsURL(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if len(v.value) < 10 || !(strings.HasPrefix(v.value, "https://") || strings.HasPrefix(v.value, "http://")) {
			v.err = v.setError("格式不正确", customError...)
			return v
		}
		return v
	})
}

// IsIP ipv4 v6 address
func (v *Engine) IsIP(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if net.ParseIP(v.value) == nil {
			v.err = v.setError("格式不正确", customError...)
			return v
		}

		return v
	})
}

// IsJSON valid json format
func (v *Engine) IsJSON(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if zjson.Valid(v.value) {
			v.err = v.setError("格式不正确", customError...)
			return v
		}
		return v
	})
}

// IsChineseIDNumber mainland china id number
func (v *Engine) IsChineseIDNumber(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		var idV int
		if v.value[17:] == "X" {
			idV = 88
		} else {
			var err error
			if idV, err = strconv.Atoi(v.value[17:]); err != nil {
				v.err = v.setError("格式不正确", customError...)
				return v
			}
		}

		var verify int
		id := v.value[:17]
		arr := make([]int, 17)
		for i := 0; i < 17; i++ {
			arr[i], v.err = strconv.Atoi(string(id[i]))
			if v.err != nil {
				return v
			}
		}
		wi := [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
		var res int
		for i := 0; i < 17; i++ {
			res += arr[i] * wi[i]
		}
		verify = res % 11

		var temp int
		a18 := [11]int{1, 0, 88 /* 'X' */, 9, 8, 7, 6, 5, 4, 3, 2}
		for i := 0; i < 11; i++ {
			if i == verify {
				temp = a18[i]
				break
			}
		}
		if temp != idV {
			v.err = v.setError("格式不正确", customError...)
			return v
		}

		return v
	})
}

// MinLength minimum length
func (v *Engine) MinLength(min int, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if len(v.value) < min {
			v.err = v.setError("长度不能小于"+strconv.Itoa(min)+"个字符", customError...)
		}
		return v
	})
}

// MinUTF8Length utf8 encoding minimum length
func (v *Engine) MinUTF8Length(min int, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if zstring.Len(v.value) < min {
			v.err = v.setError("长度不能小于"+strconv.Itoa(min)+"个字符", customError...)
		}
		return v
	})
}

// MaxLength the maximum length
func (v *Engine) MaxLength(max int, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if !ignore(v) && len(v.value) > max {
			v.err = v.setError("长度不能大于"+strconv.Itoa(max)+"个字符", customError...)
		}
		return v
	})
}

// MaxUTF8Length utf8 encoding maximum length
func (v *Engine) MaxUTF8Length(max int, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if !ignore(v) && zstring.Len(v.value) > max {
			v.err = v.setError("长度不能大于"+strconv.Itoa(max)+"个字符", customError...)
		}
		return v
	})
}

// MinInt minimum integer value
func (v *Engine) MinInt(min int, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if v.err != nil {
			return v
		}
		if v.value == "" {
			v.err = v.setError("检查失败，不能小于"+strconv.Itoa(min), customError...)
			return v
		}
		if v.i == 0 {
			i, err := strconv.Atoi(v.value)
			if err != nil {
				v.err = v.setError("检查失败，不能小于"+strconv.Itoa(min), customError...)
				return v
			}
			v.i = i
		}
		if v.i < min {
			v.err = v.setError("不能小于"+strconv.Itoa(min), customError...)
			return v
		}
		return v
	})
}

// MinInt maximum integer value
func (v *Engine) MaxInt(max int, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if v.i == 0 {
			var err error
			v.i, err = strconv.Atoi(v.value)
			if err != nil {
				v.err = v.setError("检查失败，不能大于"+strconv.Itoa(max), customError...)
				return v
			}
		}
		if v.i > max {
			v.err = v.setError("不能大于"+strconv.Itoa(max), customError...)
			return v
		}
		return v
	})
}

// MinFloat minimum floating point value
func (v *Engine) MinFloat(min float64, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if v.f == 0 {
			var err error
			v.f, err = strconv.ParseFloat(v.value, 64)
			if err != nil {
				v.err = v.setError("检查失败，不能小于"+fmt.Sprint(min), customError...)
				return v
			}
		}

		if v.f < min {
			v.err = v.setError("不能小于"+fmt.Sprint(min), customError...)
			return v
		}
		return v
	})
}

// MinFloat maximum floating point value
func (v *Engine) MaxFloat(max float64, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if v.f == 0 {
			var err error
			v.f, err = strconv.ParseFloat(v.value, 64)
			if err != nil {
				v.err = v.setError("检查失败，不能大于"+fmt.Sprint(max), customError...)
				return v
			}
		}
		if v.f > max {
			v.err = v.setError("不能大于"+fmt.Sprint(max), customError...)
			return v
		}
		return v
	})
}

// EnumString allow only values ​​in []string
func (v *Engine) EnumString(slice []string, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		for k := range slice {
			if slice[k] == v.value {
				return v
			}
		}
		v.err = v.setError("不在允许的范围", customError...)
		return v
	})
}

// EnumInt allow only values ​​in []int
func (v *Engine) EnumInt(i []int, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		value, err := strconv.Atoi(v.value)
		if err != nil {
			v.err = v.setError(err.Error(), customError...)
			return v
		}
		v.i = value
		for k := range i {
			if value == i[k] {
				return v
			}
		}
		v.err = v.setError("不在允许的范围", customError...)
		return v
	})
}

// EnumFloat64 allow only values ​​in []float64
func (v *Engine) EnumFloat64(f []float64, customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if ignore(v) {
			return v
		}
		if v.f == 0 {
			var err error
			v.f, err = strconv.ParseFloat(v.value, 64)
			if err != nil {
				v.err = v.setError(err.Error(), customError...)
				return v
			}
		}
		for k := range f {
			if v.f == f[k] {
				return v
			}
		}
		v.err = v.setError("不在允许的范围", customError...)
		return v
	})
}
