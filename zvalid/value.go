package zvalid

import (
	"container/list"
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

// Valid get the final value, or an notEmpty string if an error occurs
func (v *Engine) valid() *Engine {
	if v.result {
		return v
	}
	v.result = true
	if v.err == nil && !v.setRawValue {
		v.err = ErrNoValidationValueSet
		return v
	}
	queues := list.New()
	queues.PushBackList(v.queue)
	l := queues.Len()
	if l > 0 {
		for i := 0; i < l; i++ {
			queue := queues.Front()
			if q, ok := queue.Value.(queueT); ok {
				nv := q(v)
				v.value = nv.value
				v.err = nv.err
				v.defaultValue = nv.defaultValue
			}
			queues.Remove(queue)
		}
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
	switch value.(type) {
	case string:
		s = value.(string)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		s = ztype.ToString(value)
	default:
		s = ""
		v.err = setError(&v, "unsupported type")
	}
	return v.Verifi(s, name...)
}
