package ztype

import (
	"fmt"
	"time"
)

func (m Map) GetToString(key string, def ...string) (val string) {
	v := m.Get(key)
	if v.Exists() {
		val = v.String()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToBytes(key string, def ...[]byte) (val []byte) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Bytes()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToBool(key string, def ...bool) (val bool) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Bool()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToInt(key string, def ...int) (val int) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Int()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToInt8(key string, def ...int8) (val int8) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Int8()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToInt16(key string, def ...int16) (val int16) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Int16()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToInt32(key string, def ...int32) (val int32) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Int32()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToInt64(key string, def ...int64) (val int64) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Int64()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToUint(key string, def ...uint) (val uint) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Uint()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToUint8(key string, def ...uint8) (val uint8) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Uint8()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToUint16(key string, def ...uint16) (val uint16) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Uint16()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToUint32(key string, def ...uint32) (val uint32) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Uint32()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToUint64(key string, def ...uint64) (val uint64) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Uint64()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToFloat32(key string, def ...float32) (val float32) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Float32()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToFloat64(key string, def ...float64) (val float64) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Float64()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToTime(key string, def ...time.Time) (val time.Time) {
	var ok bool
	var err error
	v := m.Get(key)
	if v.Exists() {
		val, err = v.Time()
		ok = err == nil
		fmt.Println(err)
		_ = m.Set(key, val)
	}

	if !ok && len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToMap(key string, def ...Map) (val Map) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Map()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToSlice(key string, def ...SliceType) (val SliceType) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Slice()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}

func (m Map) GetToSliceValue(key string, def ...[]interface{}) (val []interface{}) {
	v := m.Get(key)
	if v.Exists() {
		val = v.Slice().Value()
		_ = m.Set(key, val)
	} else if len(def) > 0 {
		val = def[0]
		_ = m.Set(key, val)
	}

	return
}
