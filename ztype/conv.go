package ztype

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/ztime"
)

// Conver provides configuration for type conversion operations
type Conver struct {
	// MatchName defines the function used to match map keys to struct field names
	// Default is case-insensitive matching using strings.EqualFold
	MatchName func(mapKey, fieldName string) bool

	// ConvHook is an optional hook that can be used to customize the conversion process
	// If it returns false as the second return value, the default conversion is skipped
	ConvHook func(name string, i reflect.Value, o reflect.Type) (reflect.Value, bool)

	// TagName specifies the struct tag name to use for field mapping
	// Default is "z"
	TagName string

	// IgnoreTagName if true, ignores struct tags during conversion
	IgnoreTagName bool

	// ZeroFields if true, zero values will be written to the destination
	ZeroFields bool

	// Squash if true, embedded structs are "squashed" (fields are at the same level as parent)
	Squash bool

	// Deep if true, performs a deep copy of nested structures
	Deep bool

	// Merge if true, merges maps and slices instead of replacing them
	Merge bool
}

var conv = Conver{TagName: tagName, Squash: true, MatchName: strings.EqualFold}

// To converts input value to the output type specified by out parameter.
// The out parameter must be a pointer to the target type.
// Optional configuration functions can be provided to customize the conversion behavior.
func To(input, out interface{}, opt ...func(*Conver)) error {
	return ValueConv(input, zreflect.ValueOf(out), opt...)
}

// ValueConv converts input value to the output reflect.Value.
// The out parameter must be a pointer value that can be addressed.
// Optional configuration functions can be provided to customize the conversion behavior.
func ValueConv(input interface{}, out reflect.Value, opt ...func(*Conver)) error {
	o := conv
	for _, f := range opt {
		f(&o)
	}
	if out.Kind() != reflect.Ptr {
		return errors.New("out must be a pointer")
	}
	if !out.Elem().CanAddr() {
		return errors.New("out must be addressable (a pointer)")
	}
	return o.to("", input, out, true)
}

func (d *Conver) to(name string, input interface{}, outVal reflect.Value, deep bool) error {
	var inputVal reflect.Value
	if input != nil {
		inputVal = zreflect.ValueOf(input)
		if inputVal.Kind() == reflect.Ptr && inputVal.IsNil() {
			input = nil
		}
	}

	t := outVal.Type()
	if input == nil {
		if d.ZeroFields {
			outVal.Set(reflect.Zero(t))
		}
		return nil
	}

	if !inputVal.IsValid() {
		outVal.Set(reflect.Zero(t))
		return nil
	}

	var err error
	outputKind := zreflect.GetAbbrKind(outVal)

	if d.ConvHook != nil {
		i, next := d.ConvHook(name, inputVal, t)
		if !next {
			outVal.Set(i)
			return nil
		}
		input = i.Interface()
	}
	switch outputKind {
	case reflect.Bool:
		outVal.SetBool(ToBool(input))
	case reflect.Interface:
		err = d.basic(name, input, outVal)
	case reflect.String:
		outVal.SetString(ToString(input))
	case reflect.Int:
		outVal.SetInt(ToInt64(input))
	case reflect.Uint:
		outVal.SetUint(ToUint64(input))
	case reflect.Float64:
		outVal.SetFloat(ToFloat64(input))
	case reflect.Struct:
		err = d.toStruct(name, input, outVal)
	case reflect.Map:
		err = d.toMap(name, input, outVal, deep)
	case reflect.Ptr:
		err = d.toPtr(name, input, outVal)
	case reflect.Slice:
		err = d.toSlice(name, input, outVal)
	case reflect.Array:
		err = d.toArray(name, input, outVal)
	case reflect.Func:
		err = d.toFunc(name, input, outVal)
	default:
		return errors.New("unsupported type: " + outputKind.String())
	}

	return err
}

func (d *Conver) basic(name string, data interface{}, val reflect.Value) error {
	if val.IsValid() && val.Elem().IsValid() {
		elem, copied := val.Elem(), false
		if !elem.CanAddr() {
			copied = true
			nVal := reflect.New(elem.Type())
			nVal.Elem().Set(elem)
			elem = nVal
		}

		if err := d.to(name, data, elem, true); err != nil || !copied {
			return err
		}

		val.Set(elem.Elem())
		return nil
	}

	dataVal := zreflect.ValueOf(data)
	if dataVal.Kind() == reflect.Ptr && dataVal.Type().Elem() == val.Type() {
		dataVal = reflect.Indirect(dataVal)
	}

	if !dataVal.IsValid() {
		dataVal = reflect.Zero(val.Type())
	}

	dataValType := dataVal.Type()
	if !dataValType.AssignableTo(val.Type()) {
		return fmt.Errorf("expected type '%s', got '%s'", val.Type(), dataValType)
	}

	val.Set(dataVal)
	return nil
}

func (d *Conver) toStruct(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(zreflect.ValueOf(data))
	if dataVal.Type() == val.Type() {
		val.Set(dataVal)
		return nil
	}

	switch dataVal.Kind() {
	case reflect.Map:
		return d.toStructFromMap(name, dataVal, val)
	case reflect.Struct:
		mapType := reflect.TypeOf((map[string]interface{})(nil))
		mval := reflect.MakeMap(mapType)
		addrVal := reflect.New(mval.Type())
		reflect.Indirect(addrVal).Set(mval)
		err := d.toMapFromStruct(name, dataVal, reflect.Indirect(addrVal), mval)
		if err != nil {
			return err
		}

		return d.toStructFromMap(name, reflect.Indirect(addrVal), val)

	default:
		vTyp := val.Type()
		vTypStr := vTyp.String()
		if (isTime(vTypStr) || vTyp.ConvertibleTo(timeType)) && dataVal.Kind() == reflect.String {
			t, err := ztime.Parse(data.(string))
			if err == nil {
				if vTypStr == "ztime.LocalTime" {
					val.Set(zreflect.ValueOf(ztime.LocalTime{Time: t}))
					return nil
				}

				val.Set(zreflect.ValueOf(t).Convert(vTyp))
				return nil
			}
		}
		return fmt.Errorf("expected a map, got '%s'", dataVal.Kind())
	}
}

// structFieldInfo struct field info
type structFieldInfo struct {
	val      reflect.Value
	field    reflect.StructField
	isRemain bool
}

// validateMapKeyType validate map key type
func validateMapKeyType(_ string, mapType reflect.Type) error {
	if kind := mapType.Key().Kind(); kind != reflect.String && kind != reflect.Interface {
		return fmt.Errorf("needs a map with string keys, has '%s' keys", mapType.Key().Kind())
	}
	return nil
}

// collectMapKeys collect map keys
func collectMapKeys(dataVal reflect.Value) (map[reflect.Value]struct{}, map[interface{}]struct{}) {
	mapLen := dataVal.Len()
	keys := make(map[reflect.Value]struct{}, mapLen)
	unusedKeys := make(map[interface{}]struct{}, mapLen)

	mapKeys := dataVal.MapKeys()
	for i := 0; i < len(mapKeys); i++ {
		key := mapKeys[i]
		keys[key] = struct{}{}
		unusedKeys[key.Interface()] = struct{}{}
	}

	return keys, unusedKeys
}

// processFieldTag process field tag
func (d *Conver) processFieldTag(fieldType reflect.StructField, fieldVal reflect.Value) (squash, remain bool) {
	squash = d.Squash && fieldVal.Kind() == reflect.Struct && fieldType.Anonymous
	if d.IgnoreTagName {
		return squash, false
	}
	_, opt := zreflect.GetStructTag(fieldType, d.TagName, tagNameLesser)
	if opt == "" {
		return squash, false
	}

	const (
		tagSquash = "squash"
		tagRemain = "remain"
	)
	if !strings.Contains(opt, ",") {
		switch opt {
		case tagSquash:
			return true, false
		case tagRemain:
			return false, true
		}
		return squash, false
	}

	for _, tag := range strings.Split(opt, ",") {
		switch tag {
		case tagSquash:
			squash = true
		case tagRemain:
			remain = true
		}
	}

	return squash, remain
}

// collectStructFields collect struct fields
func (d *Conver) collectStructFields(val reflect.Value) ([]structFieldInfo, *structFieldInfo, error) {
	queue := make([]reflect.Value, 0, 4)
	queue = append(queue, val)

	fields := make([]structFieldInfo, 0, val.NumField()*2)
	var remainField *structFieldInfo

	for len(queue) > 0 {
		structVal := queue[0]
		queue = queue[1:]
		structType := structVal.Type()

		for i := 0; i < structType.NumField(); i++ {
			fieldType := structType.Field(i)
			fieldVal := structVal.Field(i)

			if fieldVal.Kind() == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct {
				fieldVal = fieldVal.Elem()
			}

			squash, remain := d.processFieldTag(fieldType, fieldVal)

			if squash {
				if fieldVal.Kind() != reflect.Struct {
					return nil, nil, fmt.Errorf("cannot squash non-struct type '%s'", fieldVal.Type())
				}
				queue = append(queue, fieldVal)
				continue
			}

			field := structFieldInfo{val: fieldVal, field: fieldType, isRemain: remain}

			if remain {
				remainField = &field
			} else {
				fields = append(fields, field)
			}
		}
	}

	return fields, remainField, nil
}

// getFieldName get field name
func (d *Conver) getFieldName(field reflect.StructField) string {
	if d.IgnoreTagName {
		return field.Name
	}

	name, _ := zreflect.GetStructTag(field, d.TagName, tagNameLesser)
	return name
}

// findMapValue find map value
func (d *Conver) findMapValue(fieldName string, dataVal reflect.Value, dataValKeys map[reflect.Value]struct{}) (reflect.Value, reflect.Value, bool) {
	rawMapKey := zreflect.ValueOf(fieldName)
	if rawMapVal := dataVal.MapIndex(rawMapKey); rawMapVal.IsValid() {
		return rawMapKey, rawMapVal, true
	}
	for dataValKey := range dataValKeys {
		mK, ok := dataValKey.Interface().(string)
		if !ok || !d.MatchName(mK, fieldName) {
			continue
		}

		if rawMapVal := dataVal.MapIndex(dataValKey); rawMapVal.IsValid() {
			return dataValKey, rawMapVal, true
		}
	}

	return reflect.Value{}, reflect.Value{}, false
}

// processRemainField process remain field
func (d *Conver) processRemainField(remainField *structFieldInfo, dataVal reflect.Value, unusedKeys map[interface{}]struct{}, name string) error {
	if remainField == nil || len(unusedKeys) == 0 {
		return nil
	}

	remain := make(map[interface{}]interface{}, len(unusedKeys))
	for key := range unusedKeys {
		remain[key] = dataVal.MapIndex(zreflect.ValueOf(key)).Interface()
	}

	return d.toMap(name, remain, remainField.val, true)
}

// toStructFromMap converts a map value to a struct value
func (d *Conver) toStructFromMap(name string, dataVal, val reflect.Value) error {
	if err := validateMapKeyType(name, dataVal.Type()); err != nil {
		return fmt.Errorf("invalid map key type: %w", err)
	}

	dataValKeys, dataValKeysUnused := collectMapKeys(dataVal)
	targetValKeysUnused := make(map[interface{}]struct{}, len(dataValKeys))

	fields, remainField, err := d.collectStructFields(val)
	if err != nil {
		return err
	}
	for _, f := range fields {
		fieldName := d.getFieldName(f.field)
		fieldValue := f.val

		rawMapKey, rawMapVal, found := d.findMapValue(fieldName, dataVal, dataValKeys)
		if !found {
			targetValKeysUnused[fieldName] = struct{}{}
			continue
		}

		if !fieldValue.IsValid() {
			return errors.New("field is not valid")
		}

		if !fieldValue.CanSet() {
			continue
		}

		delete(dataValKeysUnused, rawMapKey.Interface())

		if name != "" {
			fieldName = name + "." + fieldName
		}

		if err := d.to(fieldName, rawMapVal.Interface(), fieldValue, false); err != nil {
			return err
		}
	}

	return d.processRemainField(remainField, dataVal, dataValKeysUnused, name)
}

func (d *Conver) toMap(name string, data interface{}, val reflect.Value, deep bool) error {
	valType := val.Type()
	valKeyType := valType.Key()
	valElemType := valType.Elem()
	valMap := val

	if valMap.IsNil() || d.ZeroFields || !deep {
		mapType := reflect.MapOf(valKeyType, valElemType)
		valMap = reflect.MakeMap(mapType)
	}

	dataVal := reflect.Indirect(zreflect.ValueOf(data))
	switch dataVal.Kind() {
	case reflect.Map:
		return d.toMapFromMap(name, dataVal, val, valMap)
	case reflect.Struct:
		return d.toMapFromStruct(name, dataVal, val, valMap)
	case reflect.Array, reflect.Slice:
		return d.toMapFromSlice(name, dataVal, val, valMap)
	default:
		return fmt.Errorf("'%s' expected a map, got '%s'", name, dataVal.Kind())
	}
}

func (d *Conver) toMapFromSlice(name string, dataVal reflect.Value, val reflect.Value, valMap reflect.Value) error {
	if dataVal.Len() == 0 {
		val.Set(valMap)
		return nil
	}

	for i := 0; i < dataVal.Len(); i++ {
		err := d.to(
			name+"["+strconv.Itoa(i)+"]",
			dataVal.Index(i).Interface(), val, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Conver) toMapFromMap(name string, dataVal reflect.Value, val reflect.Value, valMap reflect.Value) error {
	valType := val.Type()
	valKeyType := valType.Key()
	valElemType := valType.Elem()

	if dataVal.Len() == 0 {
		if dataVal.IsNil() {
			if !val.IsNil() {
				val.Set(dataVal)
			}
		} else {
			val.Set(valMap)
		}

		return nil
	}

	for _, k := range dataVal.MapKeys() {
		fieldName := name + "[" + k.String() + "]"
		currentKey := reflect.Indirect(reflect.New(valKeyType))
		if err := d.to(fieldName, k.Interface(), currentKey, true); err != nil {
			return err
		}

		v := dataVal.MapIndex(k).Interface()
		currentVal := reflect.Indirect(reflect.New(valElemType))
		if err := d.to(fieldName, v, currentVal, true); err != nil {
			return err
		}

		valMap.SetMapIndex(currentKey, currentVal)
	}

	val.Set(valMap)

	return nil
}

func (d *Conver) toMapFromStruct(name string, dataVal reflect.Value, val reflect.Value, valMap reflect.Value) error {
	typ := dataVal.Type()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}

		v := dataVal.Field(i)
		vTyp := v.Type()
		if !vTyp.AssignableTo(valMap.Type().Elem()) {
			return fmt.Errorf("cannot assign type '%s' to map value field of type '%s'", vTyp, valMap.Type().Elem())
		}

		if vTyp.Kind() == reflect.Ptr && !v.IsZero() {
			vTyp = v.Elem().Type()
		}

		var (
			keyName string
			opt     string
		)

		if d.IgnoreTagName {
			keyName = f.Name
		} else {
			keyName, opt = zreflect.GetStructTag(f, d.TagName, tagNameLesser)
			if keyName == "" {
				continue
			}
		}

		squash := d.Squash && v.Kind() == reflect.Struct && f.Anonymous

		if !(v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct) {
			nv := v.Elem()
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				var keyName string
				if d.IgnoreTagName {
					keyName = f.Name
				} else {
					keyName, _ = zreflect.GetStructTag(f, d.TagName, tagNameLesser)
				}
				if keyName != "" {
					v = nv
					break
				}
			}
		}

		if opt != "" {
			if strings.Contains(opt, "omitempty") && !zreflect.Nonzero(v) {
				continue
			}

			squash = squash || strings.Contains(opt, "squash")
			if squash {
				if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
					v = v.Elem()
				}
				if v.Kind() != reflect.Struct {
					return fmt.Errorf("cannot squash non-struct type '%s'", vTyp)
				}
			}
		}

		switch v.Kind() {
		case reflect.Struct:
			if isTime(vTyp.String()) {
				valMap.SetMapIndex(zreflect.ValueOf(keyName), v)
			} else {
				x := reflect.New(vTyp)
				x.Elem().Set(v)
				vType := valMap.Type()
				vKeyType := vType.Key()
				vElemType := vType.Elem()
				mType := reflect.MapOf(vKeyType, vElemType)
				vMap := reflect.MakeMap(mType)
				addrVal := reflect.New(vMap.Type())
				reflect.Indirect(addrVal).Set(vMap)

				err := d.to(keyName, x.Interface(), reflect.Indirect(addrVal), false)
				if err != nil {
					return err
				}

				vMap = reflect.Indirect(addrVal)
				if squash {
					for _, k := range vMap.MapKeys() {
						valMap.SetMapIndex(k, vMap.MapIndex(k))
					}
				} else {
					valMap.SetMapIndex(zreflect.ValueOf(keyName), vMap)
				}
			}
		default:
			valMap.SetMapIndex(zreflect.ValueOf(keyName), v)
		}
	}

	if val.CanAddr() {
		val.Set(valMap)
	}

	return nil
}

func (d *Conver) toPtr(name string, data interface{}, val reflect.Value) error {
	isNil := data == nil
	if !isNil {
		switch v := reflect.Indirect(zreflect.ValueOf(data)); v.Kind() {
		case reflect.Chan,
			reflect.Func,
			reflect.Interface,
			reflect.Map,
			reflect.Ptr,
			reflect.Slice:
			isNil = v.IsNil()
		}
	}

	if isNil {
		if !val.IsNil() && val.CanSet() {
			nilValue := reflect.New(val.Type()).Elem()
			val.Set(nilValue)
		}

		return nil
	}

	valType := val.Type()
	valElemType := valType.Elem()
	if val.CanSet() {
		realVal := val
		if realVal.IsNil() || d.ZeroFields {
			realVal = reflect.New(valElemType)
		}

		if err := d.to(name, data, reflect.Indirect(realVal), true); err != nil {
			return err
		}

		val.Set(realVal)
	} else {
		if err := d.to(name, data, reflect.Indirect(val), true); err != nil {
			return err
		}
	}
	return nil
}

func (d *Conver) toSlice(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(zreflect.ValueOf(data))
	dataValKind := dataVal.Kind()
	valType := val.Type()
	valElemType := valType.Elem()
	sliceType := reflect.SliceOf(valElemType)

	if dataValKind != reflect.Array && dataValKind != reflect.Slice {
		switch {
		case dataValKind == reflect.Slice, dataValKind == reflect.Array:
			break
		case dataValKind == reflect.Map:
			if dataVal.Len() == 0 {
				val.Set(reflect.MakeSlice(sliceType, 0, 0))
				return nil
			}
			return d.toSlice(name, []interface{}{data}, val)
		case dataValKind == reflect.String && valElemType.Kind() == reflect.Uint8:
			return d.toSlice(name, []byte(dataVal.String()), val)
		default:
			return d.toSlice(name, []interface{}{data}, val)
		}
	}

	if dataValKind != reflect.Array && dataVal.IsNil() {
		return nil
	}

	valSlice := val
	if valSlice.IsNil() || d.ZeroFields {
		valSlice = reflect.MakeSlice(sliceType, dataVal.Len(), dataVal.Len())
	} else if valSlice.Len() > dataVal.Len() {
		valSlice = valSlice.Slice(0, dataVal.Len())
	}

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		for valSlice.Len() <= i {
			valSlice = reflect.Append(valSlice, reflect.Zero(valElemType))
		}
		currentField := valSlice.Index(i)
		fieldName := name + "[" + strconv.Itoa(i) + "]"
		if err := d.to(fieldName, currentData, currentField, true); err != nil {
			return err
		}
	}

	val.Set(valSlice)

	return nil
}

func (d *Conver) toArray(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(zreflect.ValueOf(data))
	dataValKind, valType := dataVal.Kind(), val.Type()
	valElemType := valType.Elem()
	arrayType, valArray := reflect.ArrayOf(valType.Len(), valElemType), val
	if valArray.Interface() == reflect.Zero(valArray.Type()).Interface() || d.ZeroFields {
		if dataValKind != reflect.Array && dataValKind != reflect.Slice {
			switch {
			case dataValKind == reflect.Map:
				if dataVal.Len() == 0 {
					val.Set(reflect.Zero(arrayType))
					return nil
				}
			default:
				return d.toArray(name, []interface{}{data}, val)
			}
		}
		if dataVal.Len() > arrayType.Len() {
			return fmt.Errorf(
				"'%s': expected source data to have length less or equal to %d, got %d", name, arrayType.Len(), dataVal.Len())
		}

		valArray = reflect.New(arrayType).Elem()
	}

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		currentField := valArray.Index(i)

		fieldName := name + "[" + strconv.Itoa(i) + "]"
		if err := d.to(fieldName, currentData, currentField, true); err != nil {
			return err
		}
	}

	val.Set(valArray)

	return nil
}

func (d *Conver) toFunc(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(zreflect.ValueOf(data))
	if val.Type() != dataVal.Type() {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}
	val.Set(dataVal)
	return nil
}

func isTime(vTyp string) bool {
	switch vTyp {
	case "time.Time", "ztime.LocalTime":
		return true
	}

	return false
}
