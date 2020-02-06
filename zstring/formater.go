package zstring

import "strings"

// SnakeCaseToCamelCase snakeCase To CamelCase: hello_world => helloWorld
func SnakeCaseToCamelCase(str string, ucfirst bool, delimiter ...string) string {
	if str == "" {
		return ""
	}
	sep := "_"
	if len(delimiter) > 0 {
		sep = delimiter[0]
	}
	slice := strings.Split(str, sep)
	for i := range slice {
		if ucfirst || i > 0 {
			slice[i] = strings.Title(slice[i])
		}
	}
	return strings.Join(slice, "")
}

// CamelCaseToSnakeCase camelCase To SnakeCase helloWorld/HelloWorld => hello_world
func CamelCaseToSnakeCase(str string, delimiter ...string) string {
	if str == "" {
		return ""
	}
	sep := []byte("_")
	if len(delimiter) > 0 {
		sep = []byte(delimiter[0])
	}
	strLen := len(str)
	result := make([]byte, 0, strLen*2)
	j := false
	for i := 0; i < strLen; i++ {
		char := str[i]
		if i > 0 && char >= 'A' && char <= 'Z' && j {
			result = append(result, sep...)
		}
		if char != '_' {
			j = true
		}
		result = append(result, char)
	}
	return strings.ToLower(string(result))
}
