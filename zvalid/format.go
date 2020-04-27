package zvalid

import (
	"strings"

	"github.com/sohaha/zlsgo/zstring"
)

// Trim remove leading and trailing spaces
func (v Engine) Trim() Engine {
	return pushQueue(v, func(v Engine) Engine {
		if notEmpty(v) {
			v.value = strings.TrimSpace(v.value)
		}

		return v
	})
}

// RemoveSpace remove all spaces
func (v Engine) RemoveSpace() Engine {
	return pushQueue(v, func(v Engine) Engine {
		if notEmpty(v) {
			v.value = strings.Replace(v.value, " ", "", -1)
		}
		return v
	})
}

// ReplaceAll replace
func (v Engine) Replace(old, new string, n int) Engine {
	return pushQueue(v, func(v Engine) Engine {
		if notEmpty(v) {
			v.value = strings.Replace(v.value, old, new, n)
		}
		return v
	})
}

// ReplaceAll replace all
func (v Engine) ReplaceAll(old, new string) Engine {
	return v.Replace(old, new, -1)
}

// XssClean clean html tag
func (v Engine) XssClean() Engine {
	return pushQueue(v, func(v Engine) Engine {
		if notEmpty(v) {
			v.value = zstring.XssClean(v.value)
		}
		return v
	})
}

// SnakeCaseToCamelCase snakeCase To CamelCase: hello_world => helloWorld
func (v Engine) SnakeCaseToCamelCase(ucfirst bool, delimiter ...string) Engine {
	return pushQueue(v, func(v Engine) Engine {
		if notEmpty(v) {
			v.value = zstring.SnakeCaseToCamelCase(v.value, ucfirst, delimiter...)
		}
		return v
	})
}

// CamelCaseToSnakeCase camelCase To SnakeCase helloWorld/HelloWorld => hello_world
func (v Engine) CamelCaseToSnakeCase(delimiter ...string) Engine {
	return pushQueue(v, func(v Engine) Engine {
		if notEmpty(v) {
			v.value = zstring.CamelCaseToSnakeCase(v.value, delimiter...)
		}
		return v
	})
}
