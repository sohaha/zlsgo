package zvalid

import (
	"github.com/sohaha/zlsgo/zstring"
	"strings"
)

// Trim remove leading and trailing spaces
func (v *Engine) Trim() *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if notEmpty(v) {
			v.value = strings.TrimSpace(v.value)
		}

		return v
	})
}

// RemoveSpace remove all spaces
func (v *Engine) RemoveSpace() *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if notEmpty(v) {
			v.value = strings.ReplaceAll(v.value, " ", "")
		}
		return v
	})
}

// ReplaceAll replace all
func (v *Engine) ReplaceAll(old, new string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if notEmpty(v) {
			v.value = strings.ReplaceAll(v.value, old, new)
		}
		return v
	})
}

// XssClean clean html tag
func (v *Engine) XssClean() *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if notEmpty(v) {
			v.value = zstring.XssClean(v.value)
		}
		return v
	})
}

// SnakeCaseToCamelCase snakeCase To CamelCase: hello_world => helloWorld
func (v *Engine) SnakeCaseToCamelCase(ucfirst bool, delimiter ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if notEmpty(v) {
			v.value = zstring.SnakeCaseToCamelCase(v.value, ucfirst, delimiter...)
		}
		return v
	})
}

// CamelCaseToSnakeCase camelCase To SnakeCase helloWorld/HelloWorld => hello_world
func (v *Engine) CamelCaseToSnakeCase(str string, delimiter ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if notEmpty(v) {
			v.value = zstring.CamelCaseToSnakeCase(v.value, delimiter...)
		}
		return v
	})
}
