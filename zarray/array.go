// Package zarray 数组操作
package zarray

import (
	"errors"
	"fmt"
)

// Array 数组的插入、删除、按照下标随机访问操作，数据是interface类型的
type Array struct {
	data []interface{}
	size int
}

// New 数组初始化内存
func New(capacity ...int) (array *Array) {
	if len(capacity) >= 1 && capacity[0] != 0 {
		array = &Array{
			data: make([]interface{}, capacity[0]),
			size: 0,
		}
	} else {
		array = &Array{
			data: make([]interface{}, 5),
			size: 0,
		}
	}

	return
}

// Copy 复制一个数组
func Copy(arr interface{}) (array *Array, err error) {
	array = New()
	data, ok := arr.([]interface{})
	if ok {
		l := len(data)
		for i := 0; i < l; i++ {
			_ = array.Push(data[i])
		}
	} else {
		err = errors.New("type of error")
	}

	return
}

// 判断索引是否越界
func (array *Array) checkIndex(index int) (bool, int) {
	size := array.size
	if index < 0 || index >= size {
		return true, size
	}

	return false, size
}

// 数组扩容
func (array *Array) resize(capacity int) {
	newArray := make([]interface{}, capacity)
	for i := 0; i < array.size; i++ {
		newArray[i] = array.data[i]
	}
	array.data = newArray
	newArray = nil
}

// CapLength 获取数组容量
func (array *Array) CapLength() int {
	return cap(array.data)
}

// Length 获取数组长度
func (array *Array) Length() int {
	return array.size
}

// IsEmpty 判断数组是否为空
func (array *Array) IsEmpty() bool {
	return array.size == 0
}

// Unshift 向数组头插入元素
func (array *Array) Unshift(value interface{}) error {
	return array.Add(0, value)
}

// Push 向数组尾插入元素
func (array *Array) Push(value interface{}) error {
	return array.Add(array.size, value)
}

// Add 在 index 位置，插入元素
func (array *Array) Add(index int, value interface{}) (err error) {
	if index < 0 || index > array.size {
		err = errors.New("sdd failed. Require index >= 0 and index <= size")
		return
	}

	// 如果当前元素个数等于数组容量，则将数组扩容为原来的2倍
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

// Get 获取对应 index 位置的元素
func (array *Array) Get(index int, defaultValue ...interface{}) (value interface{}, err error) {
	if r, _ := array.checkIndex(index); r {
		err = errors.New("get failed. Illegal index")
		if dValue, dErr := GetInterface(defaultValue, 0, nil); dErr == nil {
			value = dValue
		}
		return
	}

	value = array.data[index]
	return
}

// Set 修改 index 位置的元素
func (array *Array) Set(index int, value interface{}) (err error) {
	if r, _ := array.checkIndex(index); r {
		err = errors.New("set failed. Illegal index")
		return
	}

	array.data[index] = value
	return
}

// Includes 查找数组中是否有元素
func (array *Array) Includes(value interface{}) bool {
	for i := 0; i < array.size; i++ {
		if array.data[i] == value {
			return true
		}
	}

	return false
}

// IndexOf 通过索引查找数组，索引范围[0,n-1]（未找到，返回 -1）
func (array *Array) IndexOf(value interface{}) int {
	for i := 0; i < array.size; i++ {
		if array.data[i] == value {
			return i
		}
	}

	return -1
}

// Remove 删除 index 位置的元素，并返回
func (array *Array) Remove(index int, l ...int) (value []interface{}, err error) {
	r, size := array.checkIndex(index)
	if r {
		err = errors.New("remove failed. Illegal index")
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

// Shift 删除数组首个元素
func (array *Array) Shift() (interface{}, error) {
	return array.Remove(0)
}

// Pop 删除末尾元素
func (array *Array) Pop() (interface{}, error) {
	return array.Remove(int(array.size - 1))
}

// RemoveValue 从数组中删除指定元素
func (array *Array) RemoveValue(value interface{}) (e interface{}, err error) {
	index := array.IndexOf(value)
	if index != -1 {
		e, err = array.Remove(index)
	}
	return
}

// Clear 清空数组
func (array *Array) Clear() {
	array.data = make([]interface{}, array.size)
	array.size = 0
}

// Raw 原始数组
func (array *Array) Raw() []interface{} {
	return array.data
}

// Format 输出数列
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

// GetInterface  获取 []interface{} 对应 index 位置的元素
func GetInterface(arr []interface{}, index int, defaultValue ...interface{}) (value interface{}, err error) {
	arrLen := len(arr)
	if arrLen > 0 && index < arrLen {
		value = arr[index]
	} else {
		err = errors.New("getInterface failed. Illegal index")
		var dValue interface{}
		if len(defaultValue) > 0 {
			dValue = defaultValue[0]
		}
		value = dValue
	}
	return
}
