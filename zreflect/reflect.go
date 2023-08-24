package zreflect

import (
	"unsafe"
)

type (
	Type  = *rtype
	rtype struct{}
	flag  uintptr
	Value struct {
		typ Type
		ptr unsafe.Pointer
		flag
	}
	// Method struct {
	// 	Name    string
	// 	PkgPath string
	// 	Type    Type
	// 	Func    Value
	// 	Index   int
	// }
)

// func toMethod(v reflect.Method) Method {
// 	return Method{
// 		Name:    v.Name,
// 		PkgPath: v.PkgPath,
// 		Type:    rtypeToType(v.Type),
// 		Func:    NewValue(v.Func),
// 		Index:   v.Index,
// 	}
// }
