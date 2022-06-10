// Package zarray provides array operations
package zarray

import (
	"errors"
	"fmt"
	"math/rand"
)

// Array insert, delete, random access according to the subscript operation, the data is interface type
type Array struct {
	data []interface{}
	size int
}

// ErrIllegalIndex illegal index
var ErrIllegalIndex = errors.New("illegal index")

// New array initialization memory
func New(capacity ...int) (array *Array) {
	c := 5
	if len(capacity) >= 1 && capacity[0] != 0 {
		c = capacity[0]
	}

	return &Array{
		data: make([]interface{}, c),
		size: 0,
	}
}

// Copy an array
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
func (arr *Array) checkIndex(index int) (bool, int) {
	size := arr.size
	if index < 0 || index >= size {
		return true, size
	}

	return false, size
}

// array expansion
func (arr *Array) resize(capacity int) {
	newArray := make([]interface{}, capacity)
	for i := 0; i < arr.size; i++ {
		newArray[i] = arr.data[i]
	}
	arr.data = newArray
}

// CapLength get array capacity
func (arr *Array) CapLength() int {
	return cap(arr.data)
}

// Length get array length
func (arr *Array) Length() int {
	return arr.size
}

// IsEmpty determine whether the array is empty
func (arr *Array) IsEmpty() bool {
	return arr.size == 0
}

// Unshift insert element into array header
func (arr *Array) Unshift(value interface{}) error {
	return arr.Add(0, value)
}

// Push insert element to end of array
func (arr *Array) Push(values ...interface{}) {
	for i := 0; i < len(values); i++ {
		_ = arr.Add(arr.size, values[i])
	}
}

// Add in the index position insert the element
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

// Map ForEcho traversing generates a new array
func (arr *Array) Map(fn func(int, interface{}) interface{}) *Array {
	values, _ := Copy(arr.data)
	for i := 0; i < values.Length(); i++ {
		value, _ := values.Get(i)
		_ = values.Set(i, fn(i, value))
	}
	return values
}

// Get the element corresponding to the index position
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

// Set modify the element at the index position
func (arr *Array) Set(index int, value interface{}) (err error) {
	if r, _ := arr.checkIndex(index); r {
		return ErrIllegalIndex
	}

	arr.data[index] = value
	return
}

// Contains find if there are elements in the array
func (arr *Array) Contains(value interface{}) bool {
	for i := 0; i < arr.size; i++ {
		if arr.data[i] == value {
			return true
		}
	}

	return false
}

// Index find array by index, index range [0, n-1] (not found, return - 1)
func (arr *Array) Index(value interface{}) int {
	for i := 0; i < arr.size; i++ {
		if arr.data[i] == value {
			return i
		}
	}

	return -1
}

// Remove delete the element at index position and return
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

// Shift delete the first element of the array
func (arr *Array) Shift() (interface{}, error) {
	return arr.Remove(0)
}

// Pop delete end element
func (arr *Array) Pop() (interface{}, error) {
	return arr.Remove(arr.size - 1)
}

// RemoveValue removes the specified element from the array
func (arr *Array) RemoveValue(value interface{}) (e interface{}, err error) {
	index := arr.Index(value)
	if index != -1 {
		e, err = arr.Remove(index)
	}
	return
}

// Clear array
func (arr *Array) Clear() {
	arr.data = make([]interface{}, arr.size)
	arr.size = 0
}

// Raw original array
func (arr *Array) Raw() []interface{} {
	v := make([]interface{}, arr.size)
	copy(v, arr.data)
	return v
}

// Format output sequence
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

// Shuffle creates an slice of shuffled values
func (arr *Array) Shuffle() (array *Array) {
	data := arr.Raw()
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})
	array, _ = Copy(data)
	return
}

// GetInf Get the element corresponding to the index position of [] interface {}
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
