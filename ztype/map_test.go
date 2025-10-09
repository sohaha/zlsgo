package ztype

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
)

// Test validator implementations for testing Map validation methods
type testValidator struct {
	rules []string
}

type testValidatorResult struct {
	err error
}

func (r testValidatorResult) Error() error {
	return r.err
}

func newTestValidator(rules ...string) *testValidator {
	return &testValidator{rules: rules}
}

func (v *testValidator) VerifyAny(value interface{}, name ...string) ValidatorResult {
	str := ToString(value)

	for _, rule := range v.rules {
		switch rule {
		case "required":
			if str == "" {
				return testValidatorResult{err: errors.New("field is required")}
			}
		case "email":
			if !strings.Contains(str, "@") || !strings.Contains(str, ".") {
				return testValidatorResult{err: errors.New("invalid email format")}
			}
		case "int":
			if _, err := strconv.Atoi(str); err != nil {
				return testValidatorResult{err: errors.New("must be an integer")}
			}
		case "min18":
			if val, err := strconv.Atoi(str); err != nil || val < 18 {
				return testValidatorResult{err: errors.New("must be at least 18")}
			}
		case "min2":
			if len(str) < 2 {
				return testValidatorResult{err: errors.New("minimum length is 2")}
			}
		case "mobile":
			if len(str) != 11 || !strings.HasPrefix(str, "1") {
				return testValidatorResult{err: errors.New("invalid mobile format")}
			}
		}
	}

	return testValidatorResult{err: nil}
}

// nullValidator returns nil result to test error handling
type nullValidator struct{}

func (nv *nullValidator) VerifyAny(value interface{}, name ...string) ValidatorResult {
	return nil
}

func TestMap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	m := make(map[interface{}]interface{})
	m["T"] = "test"
	tMapKeyExists := MapKeyExists("T", m)
	tt.Equal(true, tMapKeyExists)

	mm := ToMap(m)
	tt.Log(mm.Keys())
	tt.EqualTrue(!mm.Valid("val"))
	tt.EqualTrue(mm.Valid("T"))
}

func TestMapCopy(t *testing.T) {
	tt := zlsgo.NewTest(t)

	z := Map{"a": 1}
	m := Map{"1": 1, "z": z}
	m2 := m.DeepCopy()
	m3 := m

	tt.Equal(m, m2)
	tt.Equal(m, m3)

	m["1"] = 2
	z["a"] = 2

	tt.EqualTrue(m.Get("z.a").String() != m2.Get("z.a").String())
	t.Log(m, m2, m3)
	tt.EqualTrue(m.Get("z.a").String() == m3.Get("z.a").String())
}

func TestMapNil(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var m Map
	tt.Equal(true, m.IsEmpty())
	err := m.Delete("no")
	t.Log(err)

	err = m.Set("val", "99")
	t.Log(err)
	tt.EqualTrue(err != nil)

	m2 := &Map{}
	tt.Equal(true, m2.IsEmpty())
	err = m.Delete("no")
	t.Log(err)

	err = m2.Set("val", "99")
	tt.NoError(err)
	tt.Equal("99", m2.Get("val").String())
}

type other struct {
	Sex int
}

type (
	Str string
	Obj struct {
		Name Str `json:"name"`
	}
	u struct {
		Other  *other
		Name   Str                      `json:"name"`
		Region struct{ Country string } `json:"reg"`
		Objs   []Obj
		Key    int
		Status bool
	}
)

var user = &u{
	Name: "n666",
	Key:  9,
	Objs: []Obj{
		{"n1"},
		{"n2"},
		{"n3"},
		{"n4"},
		{"n5"},
	},
	Status: true,
	Region: struct {
		Country string
	}{"中国"},
}

func TestToMap(T *testing.T) {
	t := zlsgo.NewTest(T)
	type u struct {
		Name   string
		Key    int
		Status bool
	}
	user := &u{
		Name:   "666",
		Key:    9,
		Status: true,
	}
	userMap := Map{
		"Name":   user.Name,
		"Status": user.Status,
		"Key":    user.Key,
	}
	toUserMap := ToMap(user)
	t.Log(user)
	t.Log(userMap)
	t.Log(toUserMap)
	t.Equal(userMap, toUserMap)

	t.Equal(1, ToMap(map[interface{}]interface{}{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]string{"name": "1"}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]int{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]uint{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]float64{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]bool{"name": true}).Get("name").Int())
	t.Equal(1, ToMap(map[string]interface{}{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]string{"name": "1"}).Get("name").Int())
	t.Equal(1, ToMap(map[string]int{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]uint{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]float64{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]bool{"name": true}).Get("name").Int())
}

func TestToMaps(T *testing.T) {
	t := zlsgo.NewTest(T)
	type u struct {
		Other  *other
		Name   string
		Key    int
		Status bool
	}
	rawData := make([]u, 2)
	rawData[0] = u{
		Name:   "666",
		Key:    9,
		Status: true,
		Other: &other{
			Sex: 18,
		},
	}
	rawData[1] = u{
		Name:   "666",
		Key:    9,
		Status: true,
	}
	toSliceMapString := ToMaps(rawData)
	t.Log(toSliceMapString)
	t.Equal(18, toSliceMapString[0].Get("Other").Get("Sex").Int())

	data := make([]map[string]interface{}, 2)
	data[0] = map[string]interface{}{"name": "hi"}
	data[1] = map[string]interface{}{"name": "golang"}
	toSliceMapString = ToMaps(data)
	t.Equal("hi", toSliceMapString.Index(0).Get("name").String())

	data2 := Maps{{"name": "hi"}, {"name": "golang"}, {"name": "!"}}
	toSliceMapString = ToMaps(data2)
	t.Equal("hi", toSliceMapString.Index(0).Get("name").String())
	t.EqualTrue(!data2.IsEmpty())

	t.Equal("hi", data2.First().Get("name").String())
	t.Equal("!", data2.Last().Get("name").String())
}

func TestConvContainTime(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type JSONTime time.Time

	type S struct {
		Date1 JSONTime `z:"date"`
		Date2 time.Time
		Name  string
	}

	now := time.Now()
	v := map[string]interface{}{
		"date":  now,
		"Date2": now,
		"Name":  "123",
	}

	var s S
	isTime := zreflect.TypeOf(time.Time{})
	err := To(v, &s, func(conver *Conver) {
		conver.ConvHook = func(name string, i reflect.Value, o reflect.Type) (reflect.Value, bool) {
			t := i.Type()
			if t == isTime && t.ConvertibleTo(o) {
				return i.Convert(o), true
			}
			return i, true
		}
	})
	tt.NoError(err)

	tt.Equal(now.Unix(), time.Time(s.Date1).Unix())
	tt.Equal(now.Unix(), s.Date2.Unix())
}

func BenchmarkName(b *testing.B) {
	b.Run("toMapString", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = toMapString(user)
		}
	})

	b.Run("ToMap", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ToMap(user)
		}
	})

	b.Run("toMapString", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = toMapString(user)
			}
		})
	})

	b.Run("ToMap", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ToMap(user)
			}
		})
	})
}

func TestMapValidate(t *testing.T) {
	tt := zlsgo.NewTest(t)

	m := Map{
		"email":  "test@example.com",
		"age":    "25",
		"name":   "张三",
		"phone":  "13812345678",
		"empty":  "",
		"number": "123",
	}

	tt.NoError(m.Validate("email", newTestValidator("required", "email")))
	tt.NoError(m.Validate("age", newTestValidator("required", "int", "min18")))
	tt.NoError(m.Validate("name", newTestValidator("required", "min2")))
	tt.NoError(m.Validate("phone", newTestValidator("required", "mobile")))
	tt.NoError(m.Validate("number", newTestValidator("int")))

	tt.EqualTrue(m.Validate("nonexistent", newTestValidator("required")) != nil)
	tt.EqualTrue(m.Validate("empty", newTestValidator("required")) != nil)
	age16 := Map{"age": "16"}
	tt.EqualTrue(age16.Validate("age", newTestValidator("required", "int", "min18")) != nil)
	var nilMap Map
	tt.EqualTrue(nilMap.Validate("test", newTestValidator("required")) != nil)
}

func TestMapValidateAll(t *testing.T) {
	tt := zlsgo.NewTest(t)

	m := Map{
		"email": "test@example.com",
		"age":   "25",
		"name":  "张三",
	}
	rules := map[string]Validator{
		"email": newTestValidator("required", "email"),
		"age":   newTestValidator("required", "int", "min18"),
		"name":  newTestValidator("required", "min2"),
	}
	tt.NoError(m.ValidateAll(rules))

	failMap := Map{
		"email": "test@example.com",
		"age":   "16",
		"name":  "张三",
	}
	tt.EqualTrue(failMap.ValidateAll(rules) != nil)

	missingKeyRules := map[string]Validator{
		"missing": newTestValidator("required"),
	}
	tt.EqualTrue(m.ValidateAll(missingKeyRules) != nil)
	var nilMap Map
	tt.EqualTrue(nilMap.ValidateAll(rules) != nil)
}

func TestMapValidateEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	m := Map{
		"empty": "",
		"space": "   ",
		"zero":  "0",
	}

	tt.EqualTrue(m.Validate("empty", newTestValidator("required")) != nil)
	tt.NoError(m.Validate("space", newTestValidator("required")))
	tt.NoError(m.Validate("zero", newTestValidator("required")))

	mixedMap := Map{
		"stringNum": "123",
		"intValue":  123,
		"floatVal":  12.34,
		"boolVal":   true,
	}

	tt.NoError(mixedMap.Validate("stringNum", newTestValidator("int")))
	tt.NoError(mixedMap.Validate("intValue", newTestValidator("int")))
	tt.NoError(mixedMap.Validate("floatVal", newTestValidator("required")))
	tt.NoError(mixedMap.Validate("boolVal", newTestValidator("required")))
}

func TestMapValidateErrorHandling(t *testing.T) {
	tt := zlsgo.NewTest(t)

	m := Map{"test": "value"}

	tt.EqualTrue(m.Validate("test", nil) != nil)

	nullVal := &nullValidator{}
	tt.EqualTrue(m.Validate("test", nullVal) != nil)
	tt.NoError(m.ValidateAll(map[string]Validator{}))
	tt.NoError(m.ValidateAll(nil))
}

func TestMapValidateWithOptionsSafety(t *testing.T) {
	tt := zlsgo.NewTest(t)

	m := Map{"test": "value"}

	rules := map[string]Validator{
		"test": newTestValidator("required"),
	}
	tt.NoError(m.ValidateWithOptions(rules))
	tt.NoError(m.ValidateWithOptions(rules, ValidateOptions{FastPath: true}))
	tt.NoError(m.ValidateWithOptions(rules, ValidateOptions{UnsafeMode: true}))

	errorRules := map[string]Validator{
		"nonexistent": newTestValidator("required"),
	}
	tt.EqualTrue(m.ValidateWithOptions(errorRules) != nil)

	nullVal := &nullValidator{}
	tt.EqualTrue(m.ValidateWithOptions(map[string]Validator{
		"test": nullVal,
	}) != nil)
}

func TestMapConcurrentValidationSafety(t *testing.T) {
	m := Map{
		"name":  "John",
		"email": "john@example.com",
		"age":   "25",
		"city":  "New York",
	}

	rules := map[string]Validator{
		"name":  newTestValidator("required"),
		"email": newTestValidator("required", "email"),
		"age":   newTestValidator("required", "int", "min18"),
		"city":  newTestValidator("required"),
	}

	const numGoroutines = 50
	const numIterations = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				err := m.ValidateAll(rules)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent validation error: %v", err)
		errorCount++
	}

	if errorCount != 0 {
		t.Errorf("Should not have any concurrent validation errors, but found %d errors", errorCount)
	}
}

func TestValidateWithOptionsConcurrentSafety(t *testing.T) {
	m := Map{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	const numGoroutines = 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*5)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				key := "field" + strconv.Itoa((j%3)+1)
				validator := newTestValidator("required")
				rules := map[string]Validator{key: validator}

				err := m.ValidateWithOptions(rules)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Errorf("ValidateWithOptions concurrent safety test failed with error: %v", err)
		errorCount++
	}

	if errorCount != 0 {
		t.Errorf("ValidateWithOptions concurrent safety test failed with %d errors", errorCount)
	}
}

func TestKeyNotFoundErrorPoolConcurrency(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 1000

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				key := "test_key_" + strconv.Itoa(id*1000+j)

				err1 := newKeyNotFoundError(key)
				if err1 == nil {
					errors <- fmt.Errorf("newKeyNotFoundError returned nil")
					continue
				}

				err2 := newKeyNotFoundError(key)
				if err2 == nil {
					errors <- fmt.Errorf("newKeyNotFoundError returned nil")
					continue
				}

				expectedMsg := "key '" + key + "' not found"
				if err1.Error() != expectedMsg {
					errors <- fmt.Errorf("error message does not match: %s", err1.Error())
				}
				if err2.Error() != expectedMsg {
					errors <- fmt.Errorf("pool error message does not match: %s", err2.Error())
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Errorf("concurrent error pool test failed with error: %v", err)
		errorCount++
	}

	if errorCount != 0 {
		t.Errorf("should not have any concurrent error pool errors, but found %d errors", errorCount)
	}
}

func TestMapConcurrentValidationStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip stress test")
	}

	m := Map{}
	for i := 0; i < 1000; i++ {
		m["field_"+strconv.Itoa(i)] = strconv.Itoa(i)
	}

	rules := make(map[string]Validator)
	for i := 0; i < 1000; i++ {
		key := "field_" + strconv.Itoa(i)
		rules[key] = newTestValidator("required", "int")
	}

	const numGoroutines = 50
	const numIterations = 10

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				err := m.ValidateAll(rules)
				if err != nil {
					t.Errorf("stress test validation failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("stress test completed: %d goroutines, %d iterations, duration: %v",
		numGoroutines, numIterations, duration)
}
