// Package zarray provides comprehensive array operations and utilities for working with
// dynamic arrays in Go. It implements various array manipulation functions including
// insertion, deletion, searching, and transformation operations.
package zarray

import (
	"errors"
	"fmt"
	"math/rand"
)

// Array represents a dynamic array that supports insertion, deletion, and random access
// operations. All elements are stored as interface{} type for maximum flexibility.
type Array struct {
	data []interface{}
	size int
}

// ErrIllegalIndex is returned when an operation is attempted with an invalid array index
var ErrIllegalIndex = errors.New("illegal index")

// NewArray initializes a new Array with the specified capacity.
// If no capacity is provided, a default capacity of 5 is used.
func NewArray(capacity ...int) (array *Array) {
	c := 5
	if len(capacity) >= 1 && capacity[0] != 0 {
		c = capacity[0]
	}

	return &Array{
		data: make([]interface{}, c),
		size: 0,
	}
}

// Deprecated: New is deprecated, please use NewArray instead
func New(capacity ...int) (array *Array) {
	return NewArray(capacity...)
}

// CopyArray creates a new Array by copying all elements from the provided array.
// Returns the new Array and any error that occurred during copying.
func CopyArray(arr interface{}) (array *Array, err error) {
	data, ok := arr.([]interface{})
	if ok {
		l := len(data)
		array = NewArray(l)
		for i := 0; i < l; i++ {
			array.Push(data[i])
		}
	} else {
		err = errors.New("type of error")
	}
	return
}

// Deprecated: Copy is deprecated, please use CopyArray instead
func Copy(arr interface{}) (array *Array, err error) {
	return CopyArray(arr)
}

// checkIndex determines whether the provided index is out of bounds.
// Returns true if the index is invalid, along with the current size of the array.
func (arr *Array) checkIndex(index int) (bool, int) {
	size := arr.size
	if index < 0 || index >= size {
		return true, size
	}

	return false, size
}

// resize expands the array's capacity to the specified size by creating
// a new underlying array and copying all existing elements.
func (arr *Array) resize(capacity int) {
	newArray := make([]interface{}, capacity)
	for i := 0; i < arr.size; i++ {
		newArray[i] = arr.data[i]
	}
	arr.data = newArray
}

// CapLength returns the current capacity of the array
func (arr *Array) CapLength() int {
	return cap(arr.data)
}

// Length returns the current number of elements in the array
func (arr *Array) Length() int {
	return arr.size
}

// IsEmpty returns true if the array contains no elements, false otherwise
func (arr *Array) IsEmpty() bool {
	return arr.size == 0
}

// Unshift inserts an element at the beginning of the array.
// Returns an error if the operation fails.
func (arr *Array) Unshift(value interface{}) error {
	return arr.Add(0, value)
}

// Push appends one or more elements to the end of the array
func (arr *Array) Push(values ...interface{}) {
	for i := 0; i < len(values); i++ {
		_ = arr.Add(arr.size, values[i])
	}
}

// Add inserts an element at the specified index position.
// Returns an error if the index is out of bounds or if the operation fails.
func (arr *Array) Add(index int, value interface{}) (err error) {
	if index < 0 || index > arr.size {
		err = errors.New("sdd failed. Require index >= 0 and index <= size")
		return
	}

	// If the current number of elements is equal to the arr capacity,
	// the arr will be expanded to twice the original size
	capLen := arr.CapLength()
	if arr.size == capLen {
		arr.resize(capLen * 2)
	}

	for i := arr.size - 1; i >= index; i-- {
		arr.data[i+1] = arr.data[i]
	}

	arr.data[index] = value
	arr.size++
	return
}

// Map creates a new array by applying the provided function to each element.
// The function receives the index and value of each element and returns the transformed value.
func (arr *Array) Map(fn func(int, interface{}) interface{}) *Array {
	values, _ := Copy(arr.data)
	for i := 0; i < values.Length(); i++ {
		value, _ := values.Get(i)
		_ = values.Set(i, fn(i, value))
	}
	return values
}

// Get retrieves the element at the specified index position.
// If the index is invalid and a default value is provided, returns the default value.
// Otherwise returns the element and any error that occurred.
func (arr *Array) Get(index int, def ...interface{}) (value interface{}, err error) {
	if r, _ := arr.checkIndex(index); r {
		err = ErrIllegalIndex
		if dValue, dErr := GetInf(def, 0, nil); dErr == nil {
			value = dValue
		}
		return
	}

	value = arr.data[index]
	return
}

// Set modifies the element at the specified index position.
// Returns an error if the index is out of bounds.
func (arr *Array) Set(index int, value interface{}) (err error) {
	if r, _ := arr.checkIndex(index); r {
		return ErrIllegalIndex
	}

	arr.data[index] = value
	return
}

// Contains checks if the specified value exists in the array.
// Returns true if found, false otherwise.
func (arr *Array) Contains(value interface{}) bool {
	for i := 0; i < arr.size; i++ {
		if arr.data[i] == value {
			return true
		}
	}

	return false
}

// Index finds the position of the specified value in the array.
// Returns the index (in range [0, n-1]) if found, or -1 if not found.
func (arr *Array) Index(value interface{}) int {
	for i := 0; i < arr.size; i++ {
		if arr.data[i] == value {
			return i
		}
	}

	return -1
}

// Remove deletes one or more elements starting at the specified index position.
// Returns the removed elements and any error that occurred during the operation.
func (arr *Array) Remove(index int, l ...int) (value []interface{}, err error) {
	r, size := arr.checkIndex(index)

	if r {
		err = ErrIllegalIndex
		return
	}
	removeL := 1
	if len(l) > 0 && l[0] > 1 {
		removeL = l[0]
	}

	value = make([]interface{}, removeL)
	copy(value, arr.data[index:index+removeL])
	for i := index + removeL; i < arr.size; i++ {
		arr.data[i-removeL] = arr.data[i]
		arr.data[i] = nil
	}

	arr.size = size - removeL
	capLen := arr.CapLength()
	if arr.size == capLen/4 && capLen/2 != 0 {
		arr.resize(capLen / 2)
	}
	return
}

// Shift removes and returns the first element of the array.
// Returns the removed element and any error that occurred during the operation.
func (arr *Array) Shift() (interface{}, error) {
	return arr.Remove(0)
}

// Pop removes and returns the last element of the array.
// Returns the removed element and any error that occurred during the operation.
func (arr *Array) Pop() (interface{}, error) {
	return arr.Remove(arr.size - 1)
}

// RemoveValue removes the first occurrence of the specified element from the array.
// Returns the removed element and any error that occurred during the operation.
func (arr *Array) RemoveValue(value interface{}) (e interface{}, err error) {
	index := arr.Index(value)
	if index != -1 {
		e, err = arr.Remove(index)
	}
	return
}

// Clear removes all elements from the array, resetting it to an empty state
func (arr *Array) Clear() {
	arr.data = make([]interface{}, arr.size)
	arr.size = 0
}

// Raw returns a copy of the underlying array data as a slice of interface{} values
func (arr *Array) Raw() []interface{} {
	v := make([]interface{}, arr.size)
	copy(v, arr.data)
	return v
}

// Format returns a string representation of the array including its size, capacity, and elements
func (arr *Array) Format() (format string) {
	format = fmt.Sprintf("Array: size = %d , capacity = %d\n", arr.size, cap(arr.data))
	format += "["
	for i := 0; i < arr.Length(); i++ {
		format += fmt.Sprintf("%+v", arr.data[i])
		if i != arr.size-1 {
			format += ", "
		}
	}
	format += "]"
	return
}

// Shuffle creates a new array with the same elements in random order
func (arr *Array) Shuffle() (array *Array) {
	data := arr.Raw()
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})
	array, _ = Copy(data)
	return
}

// GetInf retrieves the element at the specified index from a slice of interface{} values.
// If the index is invalid and a default value is provided, returns the default value.
// Otherwise returns the element and any error that occurred.
func GetInf(arr []interface{}, index int, def ...interface{}) (value interface{}, err error) {
	arrLen := len(arr)
	if arrLen > 0 && index < arrLen {
		value = arr[index]
	} else {
		err = ErrIllegalIndex
		var dValue interface{}
		if len(def) > 0 {
			dValue = def[0]
		}
		value = dValue
	}
	return
}
