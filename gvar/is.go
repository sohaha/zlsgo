package gvar

// IsInterface 是否interface{}
func IsInterface(v interface{}) bool {
	return GetType(v) == "interface{}"
}

// IsByte 是否[]byte
func IsByte(v interface{}) bool {
	return GetType(v) == "[]byte"
}

// IsString 是否字符串
func IsString(v interface{}) bool {
	return GetType(v) == "string"
}

// IsBool 是否布尔值
func IsBool(v interface{}) bool {
	return GetType(v) == "bool"
}

// IsFloat64 是否float64
func IsFloat64(v interface{}) bool {
	return GetType(v) == "float64"
}

// IsFloat32 是否float32
func IsFloat32(v interface{}) bool {
	return GetType(v) == "float32"
}

// IsUint64 是否uint64
func IsUint64(v interface{}) bool {
	return GetType(v) == "uint64"
}

// IsUint32 是否uint32
func IsUint32(v interface{}) bool {
	return GetType(v) == "uint32"
}

// IsUint16 是否uint16
func IsUint16(v interface{}) bool {
	return GetType(v) == "uint16"
}

// IsUint8 是否uint8
func IsUint8(v interface{}) bool {
	return GetType(v) == "uint8"
}

// IsUint 是否uint
func IsUint(v interface{}) bool {
	return GetType(v) == "uint"
}

// IsInt64 是否int64
func IsInt64(v interface{}) bool {
	return GetType(v) == "int64"
}

// IsInt32 是否int32
func IsInt32(v interface{}) bool {
	return GetType(v) == "int32"
}

// IsInt16 是否int16
func IsInt16(v interface{}) bool {
	return GetType(v) == "int16"
}

// IsInt8 是否int8
func IsInt8(v interface{}) bool {
	return GetType(v) == "int8"
}

// IsInt 是否int
func IsInt(v interface{}) bool {
	return GetType(v) == "int"
}

// IsStruct 是否结构体
func IsStruct(v interface{}) bool {
	return GetType(v) == "struct"
}

// IsDir 是否是一个存在目录
func IsDir(path string) bool {
	state, _ := PathExists(path)
	return state == 1
}

// IsFile 是否是一个存在文件
func IsFile(path string) bool {
	state, _ := PathExists(path)
	return state == 2
}
