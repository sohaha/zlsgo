package zvalid

import (
	"container/list"
	"strconv"
	"strings"
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
	return v.Result().err
}

// Value get the final value
func (v Engine) Value() (value string) {
	v = v.Result()
	return v.value
}

// String to string
func (v Engine) String() (string, error) {
	v = v.Result()
	return v.value, v.err
}

// Bool to bool
func (v Engine) Bool() (bool, error) {
	v = v.Result()
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
func (v Engine) Int() (int, error) {
	v = v.Result()
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
func (v Engine) Float64() (float64, error) {
	v = v.Result()
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
func (v Engine) Split(sep string) ([]string, error) {
	v = v.Result()
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
func (v Engine) Result() Engine {
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
			}
			queues.Remove(queue)
		}
	}
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
