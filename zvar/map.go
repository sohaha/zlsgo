package zvar

// MapKeyExists 字典下标是否存在
func MapKeyExists(key interface{}, m map[interface{}]interface{}) bool {
	_, ok := m[key]
	return ok
}
