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

	tt.Log(ToMap(nil))
	tt.Log(ToMap(1))
	tt.Log(ToMap("1"))
}

func TestMapPick(t *testing.T) {
	tt := zlsgo.NewTest(t)
	origin := Map{"a": 1, "b": 2, "c": 3}
	selected := origin.Pick("a", "c", "none")

	tt.Equal(2, len(selected))
	tt.Equal(1, selected.Get("a").Int())
	tt.Equal(3, selected.Get("c").Int())
	tt.EqualTrue(!selected.Valid("none"))
	tt.Equal(3, len(origin))

	emptyPick := origin.Pick()
	tt.Equal(0, len(emptyPick))
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
	tt.Log(m, m2, m3)
	tt.EqualTrue(m.Get("z.a").String() == m3.Get("z.a").String())
}

func TestMapNil(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var m Map
	tt.Equal(true, m.IsEmpty())
	err := m.Delete("no")
	tt.Log(err)

	err = m.Set("val", "99")
	tt.Log(err)
	tt.EqualTrue(err != nil)

	m2 := &Map{}
	tt.Equal(true, m2.IsEmpty())
	err = m.Delete("no")
	tt.Log(err)

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

func TestToMapsNilInput(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var nilInterface interface{}
	tt.EqualTrue(ToMaps(nilInterface) == nil)

	var nilSlice []map[string]interface{}
	result := ToMaps(nilSlice)
	tt.Equal(0, len(result))

	single := ToMaps(Map{"key": "value"})
	tt.Equal(1, len(single))
	tt.Equal("value", single.Index(0).Get("key").String())
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

func TestMapHasAndForEach(t *testing.T) {
	m := Map{"a": 1, "b": 2}
	if !m.Has("a") || m.Has("c") {
		t.Fatalf("Has failed: %v", m)
	}

	sum := 0
	visited := 0
	m.ForEach(func(k string, v Type) bool {
		visited++
		sum += v.Int()
		return visited < 1
	})
	if visited != 1 || sum == 0 {
		t.Fatalf("ForEach short-circuit failed, visited=%d sum=%d", visited, sum)
	}
}

func TestMapsForEach(t *testing.T) {
	ms := Maps{{"x": 1}, {"y": 2}, {"z": 3}}
	count := 0
	acc := 0
	ms.ForEach(func(i int, value Map) bool {
		count++
		acc += len(value)
		return i < 1
	})
	if count != 2 || acc == 0 {
		t.Fatalf("Maps.ForEach behavior unexpected: count=%d acc=%d", count, acc)
	}
}

func TestHasOptionAndValidateMapKeyType(t *testing.T) {
	f := fieldInfo{Options: []string{"omitempty", "squash"}}
	if !f.hasOption("squash") || f.hasOption("remain") {
		t.Fatalf("hasOption failed: %#v", f.Options)
	}

	if err := validateMapKeyType("", reflect.TypeOf(map[string]int{})); err != nil {
		t.Fatalf("unexpected error for string-key map: %v", err)
	}
	if err := validateMapKeyType("", reflect.TypeOf(map[int]int{})); err == nil {
		t.Fatal("expected error for non-string-key map, got nil")
	}
}

func TestProcessFieldTagAndRemainField(t *testing.T) {
	type Emb struct{ X int }
	type T struct {
		Rem map[interface{}]interface{} `z:"Rem,remain"`
		Emb `z:",squash"`
		N   int `z:"n"`
	}
	c := &Conver{TagName: tagName, Squash: true}

	val := reflect.New(reflect.TypeOf(T{})).Elem()
	sfEmb := val.Type().Field(1)
	fvEmb := val.Field(1)
	squash, remain := c.processFieldTag(sfEmb, fvEmb)
	if !squash || remain {
		t.Fatalf("processFieldTag squash failed: squash=%v remain=%v", squash, remain)
	}

	sfRem := val.Type().Field(0)
	fvRem := val.Field(0)
	squash, remain = c.processFieldTag(sfRem, fvRem)
	if squash || !remain {
		t.Fatalf("processFieldTag remain failed: squash=%v remain=%v", squash, remain)
	}

	data := map[string]interface{}{`n`: 1, `a`: 2, `b`: 3}
	unused := map[interface{}]struct{}{"a": {}, "b": {}}
	rf := &structFieldInfo{val: fvRem}
	if err := c.processRemainField(rf, reflect.ValueOf(data), unused, ""); err != nil {
		t.Fatalf("processRemainField error: %v", err)
	}
	rem := val.Field(0).Interface().(map[interface{}]interface{})
	if len(rem) != 2 || rem["a"].(int) != 2 || rem["b"].(int) != 3 {
		t.Fatalf("unexpected remain: %#v", rem)
	}
}

func TestExecuteArrayAccessVariants(t *testing.T) {
	if v, ok := executeArrayAccess(1, []string{"x", "y"}); !ok || v.(string) != "y" {
		t.Fatalf("array access on []string failed: %v %v", v, ok)
	}
	if v, ok := executeArrayAccess(0, []interface{}{"a"}); !ok || v.(string) != "a" {
		t.Fatalf("array access on []interface{} failed: %v %v", v, ok)
	}

	if v, ok := executeArrayAccess(1, [2]int{5, 6}); !ok || v.(int) != 6 {
		t.Fatalf("array access on array via ToSlice failed: %v %v", v, ok)
	}

	tok := pathToken{kind: 1, index: 0}
	if v, ok := executePathToken(tok, []string{"p"}); !ok || v.(string) != "p" {
		t.Fatalf("executePathToken(kind=1) failed: %v %v", v, ok)
	}
}

func TestMapDeleteAndIndexBounds(t *testing.T) {
	m := Map{"a": 1, "b": 2}
	if err := m.Delete("a"); err != nil {
		t.Fatalf("delete existing failed: %v", err)
	}
	if _, ok := m["a"]; ok {
		t.Fatal("key a should be deleted")
	}
	ms := Maps{{"x": 1}}
	if v := ms.Index(-1); len(v) != 0 {
		t.Fatal("negative index should return empty Map")
	}
	if v := ms.Index(99); len(v) != 0 {
		t.Fatal("oob index should return empty Map")
	}
	var empty Maps
	if v := empty.Last(); len(v) != 0 {
		t.Fatal("Last on empty should be empty Map")
	}
}

func TestExecuteFieldAccessNumericKey(t *testing.T) {
	v, ok := executeFieldAccess("1", []string{"a", "b"})
	if !ok || v.(string) != "b" {
		t.Fatalf("numeric-key field access failed: %v %v", v, ok)
	}
	if v, ok := executeFieldAccess("k", map[string]int{"k": 3}); !ok || v.(int) != 3 {
		t.Fatalf("map[string]int field access failed: %v %v", v, ok)
	}
}

func TestToMapStringReflectSliceStruct(t *testing.T) {
	type Inner struct {
		Z int `z:"z"`
	}
	type Outer struct {
		L []Inner `z:"l"`
	}
	o := Outer{L: []Inner{{Z: 1}, {Z: 2}}}
	mp := ToMap(o)
	l, ok := mp["l"].([]map[string]interface{})
	if !ok || len(l) != 2 || l[0]["z"].(int) != 1 || l[1]["z"].(int) != 2 {
		t.Fatalf("unexpected mapped slice-of-struct: %#v", mp["l"])
	}
}

func TestMapDeepCopyMore(t *testing.T) {
	var nilMap Map = nil
	src := Map{
		"n": nilMap,
		"m": map[string]interface{}{"k": 1},
		"r": map[int]int{1: 2},
	}
	cp := src.DeepCopy()
	if v, ok := cp["n"].(Map); !ok || v != nil {
		t.Fatal("typed nil Map should be preserved in copy")
	}
	src["m"].(map[string]interface{})["k"] = 9
	getK := func(x interface{}) int {
		switch v := x.(type) {
		case Map:
			return v["k"].(int)
		case map[string]interface{}:
			return v["k"].(int)
		default:
			t.Fatalf("unexpected type for m: %T", x)
			return 0
		}
	}
	if getK(cp["m"]) == 9 {
		t.Fatal("deep copy should be independent")
	}
}

func TestMapGetDisabled(t *testing.T) {
	m := Map{"a.b": 3, "a": Map{"b": 4}}
	if m.Get("a.b").Int() != 4 {
		t.Fatal("path parsing should access nested value 4")
	}
	if m.Get("a.b", true).Int() != 3 {
		t.Fatal("disabled path parsing should access raw key 3")
	}
}

func TestExecuteArrayAccessOnMapSlice(t *testing.T) {
	m := []Map{{"a": 1}, {"b": 2}}
	v, ok := executeArrayAccess(1, m)
	if !ok {
		t.Fatal("executeArrayAccess on []Map failed")
	}
	mv, ok := v.(Map)
	if !ok || mv["b"].(int) != 2 {
		t.Fatalf("unexpected value: %#v", v)
	}
}

func TestToMapEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var nilMap Map
	result := ToMap(nilMap)
	tt.EqualTrue(result == nil)

	m := Map{"key": "value"}
	result = ToMap(m)
	tt.Equal("value", result.Get("key").String())

	stdMap := map[string]interface{}{"test": 123}
	result = ToMap(stdMap)
	tt.Equal(123, result.Get("test").Int())

	type nested struct {
		Inner string `z:"inner"`
	}
	type outer struct {
		Nested nested `z:"nested"`
		Value  int    `z:"value"`
	}
	o := outer{
		Nested: nested{Inner: "test"},
		Value:  100,
	}
	result = ToMap(o)
	tt.Equal("test", result.Get("nested.inner").String())
	tt.Equal(100, result.Get("value").Int())
}

func TestToMapFromMapEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	source := map[string]interface{}{
		"nested": map[string]interface{}{
			"deep": "value",
		},
		"array": []int{1, 2, 3},
	}

	var target map[string]interface{}
	err := To(source, &target)
	tt.NoError(err)
	tt.NotNil(target)

	sourceIntKey := map[int]string{
		1: "one",
		2: "two",
	}

	var targetIntKey map[int]string
	err = To(sourceIntKey, &targetIntKey)
	tt.NoError(err)
	tt.Equal("one", targetIntKey[1])

	emptySource := map[string]int{}
	var emptyTarget map[string]int
	err = To(emptySource, &emptyTarget)
	tt.NoError(err)
	tt.NotNil(emptyTarget)
}

func TestToMapStringEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type Inner struct {
		Value string `z:"value"`
	}
	type Outer struct {
		Inner *Inner `z:"inner"`
		Name  string `z:"name"`
	}

	obj := Outer{
		Inner: &Inner{Value: "test"},
		Name:  "outer",
	}

	result := ToMap(&obj)
	tt.Equal("outer", result.Get("name").String())
	tt.Equal("test", result.Get("inner.value").String())

	type Item struct {
		Name string `z:"name"`
		ID   int    `z:"id"`
	}

	items := []Item{
		{ID: 1, Name: "first"},
		{ID: 2, Name: "second"},
	}

	result2 := ToMap(items)
	tt.NotNil(result2)

	intKeyMap := map[int]string{
		1: "one",
		2: "two",
	}
	result3 := ToMap(intKeyMap)
	tt.NotNil(result3)
}

func TestValidEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	m := Map{"key": "value"}
	exists := m.Valid("nonexistent")
	tt.EqualTrue(!exists)

	exists = m.Valid("key")
	tt.EqualTrue(exists)

	m2 := Map{"a": 1, "b": 2, "c": 3}
	exists = m2.Valid("a", "b", "c")
	tt.EqualTrue(exists)

	exists = m2.Valid("a", "b", "d")
	tt.EqualTrue(!exists)
}

func TestMapSetAndDeleteEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var m Map
	err := m.Set("key", "value")
	tt.EqualTrue(err != nil)

	err = m.Delete("key")
	tt.EqualTrue(err != nil)

	m = Map{"key": "value", "other": "value"}
	err = m.Delete("key")
	tt.NoError(err)
	tt.EqualTrue(!m.Valid("key"))

	m = Map{"key": "old"}
	err = m.Set("key", "new")
	tt.NoError(err)
	tt.Equal("new", m.Get("key").String())
}

func TestToMapFromMapWithDifferentKeyTypes(t *testing.T) {
	tt := zlsgo.NewTest(t)

	source := map[interface{}]interface{}{
		"key1": "value1",
		"key2": 123,
		1:      "number key",
	}

	var target map[string]interface{}
	err := To(source, &target)
	tt.NoError(err)
	tt.NotNil(target)

	sourceBool := map[bool]string{
		true:  "yes",
		false: "no",
	}

	var targetBool map[bool]string
	err = To(sourceBool, &targetBool)
	tt.NoError(err)
	tt.Equal("yes", targetBool[true])
}

func TestToMapStringWithComplexTypes(t *testing.T) {
	tt := zlsgo.NewTest(t)

	intKeyMap := map[int]interface{}{
		1: "one",
		2: "two",
		3: map[string]string{"nested": "value"},
	}
	result := ToMap(intKeyMap)
	tt.NotNil(result)

	sliceMap := map[string][]int{
		"numbers": {1, 2, 3},
		"more":    {4, 5},
	}
	result2 := ToMap(sliceMap)
	tt.NotNil(result2)

	ptrMap := &map[string]string{"key": "value"}
	result3 := ToMap(ptrMap)
	tt.Equal("value", result3.Get("key").String())
}

func TestMapValidWithNilMap(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var m Map
	exists := m.Valid("key")
	tt.EqualTrue(!exists)

	exists = m.Valid()
	tt.EqualTrue(!exists)
}
