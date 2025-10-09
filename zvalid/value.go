package zvalid

import (
	"strconv"
	"strings"

	"github.com/sohaha/zlsgo/ztype"
)

// Ok not err
func (v Engine) Ok() bool {
	return v.Error() == nil
}

// Error or whether the verification fails
func (v Engine) Error() error {
	if v.result {
		return v.err
	}
	return v.valid().err
}

// Value get the final value
func (v Engine) Value() (value string) {
	return v.valid().value
}

// String to string
func (v Engine) String() (string, error) {
	v.valid()
	return v.value, v.err
}

// Bool to bool
func (v Engine) Bool() (bool, error) {
	v.valid()
	if ignore(&v) {
		return false, v.err
	}
	value, err := strconv.ParseBool(v.value)
	if err != nil {
		return false, err
	}
	return value, nil
}

// Int convert to int
func (v Engine) Int() (int, error) {
	v.valid()
	if ignore(&v) {
		return 0, v.err
	}
	if v.valueInt != 0 {
		return v.valueInt, nil
	}
	value, err := strconv.Atoi(v.value)
	if err != nil {
		return 0, err
	}
	v.valueInt = value
	return value, nil
}

// Float64 convert to float64
func (v Engine) Float64() (float64, error) {
	v.valid()
	if ignore(&v) {
		return 0, v.err
	}
	if v.valueFloat != 0 {
		return v.valueFloat, nil
	}
	value, err := strconv.ParseFloat(v.value, 64)
	if err != nil {
		return 0, err
	}
	v.valueFloat = value
	return value, nil
}

// Split converted to [] string
func (v Engine) Split(sep string) ([]string, error) {
	v.valid()
	if ignore(&v) {
		return []string{}, v.err
	}
	value := strings.Split(v.value, sep)
	if len(value) == 0 {
		return []string{}, v.err
	}
	return value, nil
}

// Valid executes the validation queue and returns the final result.
// Note: All queue operations will be executed even after validation errors occur,
// ensuring that operations like Default() can work properly.
// This method includes defensive checks to prevent panic from nil queue functions.
func (v *Engine) valid() *Engine {
	if v.result {
		return v
	}
	v.result = true
	if v.err == nil && !v.setRawValue {
		v.err = ErrNoValidationValueSet
		return v
	}

	if len(v.queue) == 0 {
		return v
	}

	for _, q := range v.queue {
		if q == nil {
			continue // Skip nil queue functions to prevent panic
		}
		nv := q(v)
		if nv == nil {
			continue // Skip nil results to prevent nil pointer dereference
		}
		v.value = nv.value
		v.err = nv.err
		v.defaultValue = nv.defaultValue
		// Continue executing queue items even after error to allow Default() to work
	}
	return v
}

// SetAlias set alias
func (v Engine) SetAlias(name string) Engine {
	v.name = name
	return v
}

// Verifi validate specified data
func (v Engine) Verifi(value string, name ...string) Engine {
	v.value = value
	v.setRawValue = true
	if len(name) > 0 {
		v.name = name[0]
	}
	return v
}

// VerifiAny validate specified data
func (v Engine) VerifiAny(value interface{}, name ...string) Engine {
	var s string
	switch vv := value.(type) {
	case string:
		s = vv
	default:
		s = ztype.ToString(vv)
		// v.err = setError(&v, "unsupported type")
	}
	return v.Verifi(s, name...)
}
