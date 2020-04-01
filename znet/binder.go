package znet

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Error struct {
	Field   string
	Message string
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Field, err.Message)
}

// Binder handles binding of parameter maps to Go data structures.
type Binder struct {
	Values map[string][]string
	Files  map[string][]*multipart.FileHeader
}

// Request returns a binder initialized with the request's form and query string
// data (including multipart forms).
func Request(req *http.Request) Binder {
	_ = req.ParseMultipartForm(10 * 1024 * 1024)
	if req.MultipartForm != nil {
		return Binder{req.MultipartForm.Value, req.MultipartForm.File}
	}
	return Values(req.Form)
}

// Values returns a binder for the given query parameter map
func Values(params map[string][]string) Binder {
	return Binder{params, nil}
}

// Map returns a binder for the given parameter map
func Map(m map[string]string) Binder {
	var p = make(map[string][]string, len(m))
	for k, v := range m {
		p[k] = []string{v}
	}
	return Values(p)
}

// All unpacks the entire set of form data into the given struct.
func (b Binder) All(dst interface{}) (err error) {
	return b.Field(dst, "")
}

// Field binds the given destination to a field of the given name from one or
// more values in this binder.  The destination must be a pointer.
// Returns an error of type bind.Error upon any sort of failure.
func (b Binder) Field(dst interface{}, name string) (err error) {
	return b.field(reflect.ValueOf(dst), name)
}

func (b Binder) field(dstval reflect.Value, name string) (err error) {
	if !dstval.IsValid() {
		return Error{name, "destination is not valid"}
	}

	var typ = dstval.Type()
	if typ.Kind() != reflect.Ptr {
		return Error{name, fmt.Sprintf("destination must be a pointer, got %v", typ)}
	}
	var elemtyp = typ.Elem()
	if dstval.IsNil() {
		if !dstval.CanSet() {
			return Error{name, fmt.Sprintf("destination is nil and non-addressable: %v", typ)}
		}
		dstval.Set(reflect.New(elemtyp))
	}
	if fn, found := binderForType(elemtyp); found {
		return fn(b, name, dstval.Elem())
	}
	return Error{name, fmt.Sprintf("no binder found for type %v", typ)}
}

func binderForType(typ reflect.Type) (Func, bool) {
	binder, ok := TypeBinders[typ]
	if !ok {
		binder, ok = KindBinders[typ.Kind()]
		if !ok {
			return nil, false
		}
	}
	return binder, true
}

// Func is a binding function that is responsible for extracting and converting
// the relevant parameters from the binder and writing the result to the given
// destination.
type Func func(b Binder, name string, dst reflect.Value) error

var (
	// KindBinders is a lookup from the kind of a type to the bind.Func that binds
	// it. It is less specific than the TypeBinders and used as a fallback.
	KindBinders map[reflect.Kind]Func

	// TypeBinders is a lookup from a specific type to the bind.Func that binds it.
	// Applications may add custom binders to this map to override the default behavior.
	TypeBinders map[reflect.Type]Func

	// TimeFormats are the time layout strings used to attempt to parse data into a time.Time.
	// They are attempted in order.
	TimeFormats = []string{"2006-01-02 15:04", "2006-01-02"}
)

func init() {
	var (
		intBinder   = valueBinder(bindInt)
		uintBinder  = valueBinder(bindUint)
		floatBinder = valueBinder(bindFloat)
	)

	KindBinders = map[reflect.Kind]Func{
		reflect.Int:     intBinder,
		reflect.Int8:    intBinder,
		reflect.Int16:   intBinder,
		reflect.Int32:   intBinder,
		reflect.Int64:   intBinder,
		reflect.Uint:    uintBinder,
		reflect.Uint8:   uintBinder,
		reflect.Uint16:  uintBinder,
		reflect.Uint32:  uintBinder,
		reflect.Uint64:  uintBinder,
		reflect.Float32: floatBinder,
		reflect.Float64: floatBinder,
		reflect.String:  valueBinder(bindString),
		reflect.Bool:    valueBinder(bindBool),
		reflect.Slice:   bindSlice,
		reflect.Struct:  bindStruct,
		reflect.Ptr:     bindPointer,
	}

	TypeBinders = map[reflect.Type]Func{
		reflect.TypeOf(time.Time{}):                  valueBinder(bindTime),
		reflect.TypeOf(&os.File{}):                   bindFile,
		reflect.TypeOf([]byte{}):                     bindByteArray,
		reflect.TypeOf([]io.Reader{}).Elem():         bindReadSeeker,
		reflect.TypeOf([]io.ReadSeeker{}).Elem():     bindReadSeeker,
		reflect.TypeOf((*multipart.FileHeader)(nil)): bindFileHeader,
	}
}

// An adapter for easily making one-key-value binders.
func valueBinder(f func(value string, dst reflect.Value) error) Func {
	return func(b Binder, name string, dst reflect.Value) error {
		vals, ok := b.Values[name]
		if !ok || len(vals) == 0 {
			return Error{name, "no value found"}
		}
		var err = f(vals[0], dst)
		if err != nil {
			return Error{name, err.Error()}
		}
		return nil
	}
}

func bindValue(val string, dst reflect.Value) error {
	return Binder{
		Values: map[string][]string{"": {val}},
	}.field(dst, "")
}

func bindFileValue(val *multipart.FileHeader, dst reflect.Value) error {
	return Binder{
		Files: map[string][]*multipart.FileHeader{"": {val}},
	}.field(dst, "")
}

func bindInt(val string, dst reflect.Value) error {
	intValue, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	dst.SetInt(intValue)
	return nil
}

func bindUint(val string, dst reflect.Value) error {
	uintValue, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return err
	}
	dst.SetUint(uintValue)
	return nil
}

func bindFloat(val string, dst reflect.Value) error {
	floatValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	dst.SetFloat(floatValue)
	return nil
}

func bindString(val string, dst reflect.Value) error {
	dst.SetString(val)
	return nil
}

// Booleans support a couple different value formats:
// "true" and "false"
// "on" and "" (a checkbox)
// "1" and "0" (why not)
func bindBool(val string, dst reflect.Value) error {
	v := strings.TrimSpace(strings.ToLower(val))
	switch v {
	case "true", "on", "1":
		dst.SetBool(true)
	case "false", "", "0":
		dst.SetBool(false)
	default:
		return errors.New("unrecognized boolean value")
	}
	return nil
}

func bindPointer(b Binder, name string, dst reflect.Value) error {
	return b.field(dst, name)
}

func bindTime(val string, dst reflect.Value) error {
	var err error
	var r time.Time
	for _, f := range TimeFormats {
		if r, err = time.Parse(f, val); err == nil {
			dst.Set(reflect.ValueOf(r))
			return nil
		}
	}
	return err
}

// sliceValue keep track of the index for individual keyvalues.
type sliceValue struct {
	index int           // Index extracted from brackets.  If -1, no index was provided.
	value reflect.Value // the bound value for this slice element.
}

// This function creates a slice of the given type, Binds each of the individual
// elements, and then sets them to their appropriate location in the slice.
// If elements are provided without an explicit index, they are added (in
// unspecified order) to the end of the slice.
func bindSlice(binder Binder, name string, dst reflect.Value) error {
	// Collect an array of slice elements with their indexes (and the max index).
	var (
		maxIndex    = -1
		numNoIndex  = 0
		sliceValues = []sliceValue{}
		elemType    = dst.Type().Elem()
	)

	// Factor out the common slice logic (between form values and files).
	processElement := func(key string, vals []string, files []*multipart.FileHeader) error {
		if key != name && !strings.HasPrefix(key, name+"[") {
			return nil
		}

		// Extract the index, and the index where a sub-key starts. (e.g. field[0].subkey)
		index := -1
		leftBracket, rightBracket := len(name), strings.Index(key[len(name):], "]")+len(name)
		if rightBracket > leftBracket+1 {
			index, _ = strconv.Atoi(key[leftBracket+1 : rightBracket])
		}
		subKeyIndex := rightBracket + 1

		// Handle the indexed case.
		if index > -1 {
			if index > maxIndex {
				maxIndex = index
			}
			var element = sliceValue{index, reflect.New(elemType)}
			sliceValues = append(sliceValues, element)
			return binder.field(element.value, key[:subKeyIndex])
		}

		// It's an un-indexed element.  (e.g. element[])
		numNoIndex += len(vals) + len(files)
		for _, val := range vals {
			// Unindexed values can only be direct-bound.
			var element = sliceValue{-1, reflect.New(elemType)}
			sliceValues = append(sliceValues, element)
			if err := bindValue(val, element.value); err != nil {
				return err
			}
		}

		for _, fileHeader := range files {
			var element = sliceValue{-1, reflect.New(elemType)}
			sliceValues = append(sliceValues, element)
			if err := bindFileValue(fileHeader, element.value); err != nil {
				return err
			}
		}
		return nil
	}

	for key, vals := range binder.Values {
		if err := processElement(key, vals, nil); err != nil {
			return err
		}
	}
	for key, fileHeaders := range binder.Files {
		if err := processElement(key, nil, fileHeaders); err != nil {
			return err
		}
	}

	resultArray := reflect.MakeSlice(reflect.SliceOf(elemType), maxIndex+1, maxIndex+1+numNoIndex)
	for _, sv := range sliceValues {
		if sv.index != -1 {
			resultArray.Index(sv.index).Set(sv.value.Elem())
		} else {
			resultArray = reflect.Append(resultArray, sv.value.Elem())
		}
	}

	dst.Set(resultArray)
	return nil
}

// Break on dots and brackets.
// e.g. bar => "bar", bar.baz => "bar", bar[0] => "bar"
func nextKey(key string) string {
	fieldLen := strings.IndexAny(key, ".[")
	if fieldLen == -1 {
		return key
	}
	return key[:fieldLen]
}

func bindStruct(binder Binder, name string, dst reflect.Value) error {
	fieldValues := make(map[string]struct{})
	for key := range binder.Values {
		if name != "" && !strings.HasPrefix(key, name+".") {
			continue
		}

		// Get the name of the struct property.
		// Strip off the prefix. e.g. foo.bar.baz => bar.baz
		suffix := key
		if name != "" {
			suffix = key[len(name)+1:]
		}
		fieldName := nextKey(suffix)
		fieldLen := len(fieldName)

		if _, ok := fieldValues[fieldName]; !ok {
			// Time to bind this field.  Get it and make sure we can set it.
			fieldValue := dst.FieldByName(fieldName)
			if !fieldValue.IsValid() {
				return Error{name, "field not found: " + fieldName}
			}
			if !fieldValue.CanSet() {
				return Error{name, "field not settable: " + fieldName}
			}
			subName := key[:fieldLen]
			if name != "" {
				subName = key[:len(name)+1+fieldLen]
			}
			err := binder.field(fieldValue.Addr(), subName)
			if err != nil {
				return Error{key[:len(name)+1+fieldLen], err.(Error).Message}
			}
			fieldValues[fieldName] = struct{}{}
		}
	}

	return nil
}

// getMultipartFile returns an upload of the given name, or nil.
func getMultipartFile(binder Binder, name string) (multipart.File, error) {
	for _, fileHeader := range binder.Files[name] {
		return fileHeader.Open()
	}
	return nil, Error{name, "not found"}
}

func bindFile(binder Binder, name string, dst reflect.Value) error {
	reader, err := getMultipartFile(binder, name)
	if err != nil {
		return err
	}

	// If it's already stored in a temp file, just return that.
	if osFile, ok := reader.(*os.File); ok {
		dst.Set(reflect.ValueOf(osFile))
		return nil
	}

	// Otherwise, have to store it.
	tmpFile, err := ioutil.TempFile("", "binder")
	if err != nil {
		return Error{name, err.Error()}
	}

	_, err = io.Copy(tmpFile, reader)
	if err != nil {
		return Error{name, err.Error()}
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return Error{name, err.Error()}
	}

	dst.Set(reflect.ValueOf(tmpFile))
	return nil
}

func bindByteArray(binder Binder, name string, dst reflect.Value) error {
	reader, err := getMultipartFile(binder, name)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	dst.Set(reflect.ValueOf(b))
	return nil
}

func bindReadSeeker(binder Binder, name string, dst reflect.Value) error {
	reader, err := getMultipartFile(binder, name)
	if err != nil {
		return err
	}
	dst.Set(reflect.ValueOf(reader.(io.ReadSeeker)))
	return nil
}

func bindFileHeader(binder Binder, name string, dst reflect.Value) error {
	fileHeader, ok := binder.Files[name]
	if !ok {
		return Error{name, "not found"}
	}
	dst.Set(reflect.ValueOf(fileHeader[0]))
	return nil
}
