package zvalid

import (
	"strconv"
	"strings"
)

// Ok not err
func (v *Engine) Ok() bool {
	return v.Error() == nil
}

// Error or whether the verification fails
func (v *Engine) Error() (err error) {
	if v.err != nil {
		return v.err
	}
	_, err = v.Result()
	return
}

// Value get the final value
func (v *Engine) Value() (value string) {
	value, _ = v.Result()
	return
}

// String to string
func (v *Engine) String() (string, error) {
	return v.Result()
}

// Bool to bool
func (v *Engine) Bool() (bool, error) {
	_, _ = v.Result()
	if ignore(v) {
		return false, v.err
	}
	value, err := strconv.ParseBool(v.value)
	if err != nil {
		return false, err
	}
	return value, nil
}

// Int convert to int
func (v *Engine) Int() (int, error) {
	_, _ = v.Result()
	if ignore(v) {
		return 0, v.err
	}
	if v.i != 0 {
		return v.i, nil
	}
	value, err := strconv.Atoi(v.value)
	if err != nil {
		return 0, err
	}
	v.i = value
	return value, nil
}

// Float64 convert to float64
func (v *Engine) Float64() (float64, error) {
	_, _ = v.Result()
	if ignore(v) {
		return 0, v.err
	}
	if v.f != 0 {
		return v.f, nil
	}
	value, err := strconv.ParseFloat(v.value, 64)
	if err != nil {
		return 0, err
	}
	v.f = value
	return value, nil
}

// Split converted to [] string
func (v *Engine) Split(sep string) ([]string, error) {
	_, _ = v.Result()
	if ignore(v) {
		return []string{}, v.err
	}
	value := strings.Split(v.value, sep)
	if len(value) == 0 {
		return []string{}, v.err
	}
	return value, nil
}

// Result get the final value, or an notEmpty string if an error occurs
func (v *Engine) Result() (string, error) {
	if v.err == nil && !v.setRawValue {
		v.err = ErrNoValidationValueSet
	}
	l := v.queue.Len()
	if l > 0 {
		for i := 0; i < l; i++ {
			queue := v.queue.Front()
			if q, ok := queue.Value.(queueT); ok {
				q(v)
			}
			v.queue.Remove(queue)
		}
	}

	if v.err != nil {
		return "", v.err
	}
	return v.value, nil
}

// Verifi validate specified data
func (v *Engine) Verifi(value string, name ...string) *Engine {
	vc := clone(v)
	vc.value = value
	vc.setRawValue = true
	if len(name) > 0 {
		vc.name = name[0]
	}
	return vc
}
