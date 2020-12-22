// Package zarray provides array operations
package zarray

import (
	"errors"
	"fmt"
)

// Array Array insert, delete, random access according to the subscript operation, the data is interface type
type Array struct {
	data []interface{}
	size int
}

// ERR_ILLEGAL_INDEX illegal index
var ERR_ILLEGAL_INDEX = errors.New("illegal index")

// New array initialization memory
func New(capacity ...int) (array *Array) {
	if len(capacity) >= 1 && capacity[0] != 0 {
		array = &Array{
			data: make([]interface{}, capacity[0]),
			size: 0,
		}
	} else {
		array = &Array{
			data: make([]interface{}, 10),
			size: 0,
		}
	}

	return
}

// Copy copy an array
func Copy(arr interface{}) (array *Array, err error) {
	data, ok := arr.([]interface{})
	if ok {
		l := len(data)
		array = New(l)
		for i := 0; i < l; i++ {
			array.Push(data[i])
		}
	} else {
		err = errors.New("type of error")
	}

	return
}

// determine whether the index is out of bounds
func (array *Array) checkIndex(index int) (bool, int) {
	size := array.size
	if index < 0 || index >= size {
		return true, size
	}

	return false, size
}

// array expansion
func (array *Array) resize(capacity int) {
	newArray := make([]interface{}, capacity)
	for i := 0; i < array.size; i++ {
		newArray[i] = array.data[i]
	}
	array.data = newArray
}

// CapLength get array capacity
func (array *Array) CapLength() int {
	return cap(array.data)
}

// Length get array length
func (array *Array) Length() int {
	return array.size
}

// IsEmpty determine whether the array is empty
func (array *Array) IsEmpty() bool {
	return array.size == 0
}

// Unshift insert element into array header
func (array *Array) Unshift(value interface{}) error {
	return array.Add(0, value)
}

// Push insert element to end of array
func (array *Array) Push(values ...interface{}) {
	for i := 0; i < len(values); i++ {
		_ = array.Add(array.size, values[i])
	}
}

// Set in the index position insert the element
func (array *Array) Add(index int, value interface{}) (err error) {
	if index < 0 || index > array.size {
		err = errors.New("sdd failed. Require index >= 0 and index <= size")
		return
	}

	// If the current number of elements is equal to the array capacity,
	// the array will be expanded to twice the original size
	capLen := array.CapLength()
	if array.size == capLen {
		array.resize(capLen * 2)
	}

	for i := array.size - 1; i >= index; i-- {
		array.data[i+1] = array.data[i]
	}

	array.data[index] = value
	array.size++
	return
}

// ForEcho traversing generates a new array
func (array *Array) Map(fn func(interface{}) interface{}) *Array {
	values, _ := Copy(array.data)
	for i := 0; i < values.Length(); i++ {
		value, _ := values.Get(i)
		_ = values.Set(i, fn(value))
	}
	return values
}

// Get Gets the element corresponding to the index position
func (array *Array) Get(index int, def ...interface{}) (value interface{}, err error) {
	if r, _ := array.checkIndex(index); r {
		err = ERR_ILLEGAL_INDEX
		if dValue, dErr := GetInterface(def, 0, nil); dErr == nil {
			value = dValue
		}
		return
	}

	value = array.data[index]
	return
}

// Set modify the element at the index position
func (array *Array) Set(index int, value interface{}) (err error) {
	if r, _ := array.checkIndex(index); r {
		return ERR_ILLEGAL_INDEX
	}

	array.data[index] = value
	return
}

// Contains find if there are elements in the array
func (array *Array) Contains(value interface{}) bool {
	for i := 0; i < array.size; i++ {
		if array.data[i] == value {
			return true
		}
	}

	return false
}

// Index Find array by index, index range [0, n-1] (not found, return - 1)
func (array *Array) Index(value interface{}) int {
	for i := 0; i < array.size; i++ {
		if array.data[i] == value {
			return i
		}
	}

	return -1
}

// Remove delete the element at index position and return
func (array *Array) Remove(index int, l ...int) (value []interface{}, err error) {
	r, size := array.checkIndex(index)

	if r {
		err = ERR_ILLEGAL_INDEX
		return
	}
	removeL := 1
	if len(l) > 0 && l[0] > 1 {
		removeL = l[0]
	}

	value = make([]interface{}, removeL)
	copy(value, array.data[index:index+removeL])
	for i := index + removeL; i < array.size; i++ {
		array.data[i-removeL] = array.data[i]
		array.data[i] = nil
	}

	array.size = size - removeL
	capLen := array.CapLength()
	if array.size == capLen/4 && capLen/2 != 0 {
		array.resize(capLen / 2)
	}
	return
}

// Shift delete the first element of the array
func (array *Array) Shift() (interface{}, error) {
	return array.Remove(0)
}

// Pop delete end element
func (array *Array) Pop() (interface{}, error) {
	return array.Remove(array.size - 1)
}

// RemoveValue removes the specified element from the array
func (array *Array) RemoveValue(value interface{}) (e interface{}, err error) {
	index := array.Index(value)
	if index != -1 {
		e, err = array.Remove(index)
	}
	return
}

// Clear clear array
func (array *Array) Clear() {
	array.data = make([]interface{}, array.size)
	array.size = 0
}

// Raw original array
func (array *Array) Raw() []interface{} {
	return array.data[:array.size]
}

// Format output sequence
func (array *Array) Format() (format string) {
	format = fmt.Sprintf("Array: size = %d , capacity = %d\n", array.size, cap(array.data))
	format += "["
	for i := 0; i < array.Length(); i++ {
		format += fmt.Sprintf("%+v", array.data[i])
		if i != array.size-1 {
			format += ", "
		}
	}
	format += "]"
	return
}

// GetInterface  Get the element corresponding to the index position of [] interface {}
func GetInterface(arr []interface{}, index int, def ...interface{}) (value interface{}, err error) {
	arrLen := len(arr)
	if arrLen > 0 && index < arrLen {
		value = arr[index]
	} else {
		err = ERR_ILLEGAL_INDEX
		var dValue interface{}
		if len(def) > 0 {
			dValue = def[0]
		}
		value = dValue
	}
	return
}
