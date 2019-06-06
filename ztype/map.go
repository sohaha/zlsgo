package ztype

// MapKeyExists Whether the dictionary key exists
func MapKeyExists(key interface{}, m map[interface{}]interface{}) bool {
	_, ok := m[key]
	return ok
}
