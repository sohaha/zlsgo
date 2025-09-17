# ztype 模块

`ztype` 提供了灵活的类型转换工具和动态类型系统，允许安全地访问值并进行自动类型转换。

## 功能概览

- **类型包装**: 安全的类型包装和访问
- **类型转换**: 自动类型转换和验证
- **路径访问**: 嵌套值的路径表达式访问
- **切片处理**: 切片类型的安全操作
- **映射处理**: 映射类型的安全操作
- **结构体构建**: 动态结构体构建

## 核心功能

### 类型包装

```go
// 创建新的类型包装器
func New(v interface{}) Type
// 获取包装的值
func (t Type) Value() interface{}
// 检查值是否存在
func (t Type) Exists() bool
```

### 路径访问

```go
// 使用路径表达式获取嵌套值
func (t Type) Get(path string) Type
```

### 类型转换

```go
// 转换为字符串，支持默认值
func (t Type) String(def ...string) string
// 转换为字节数组，支持默认值
func (t Type) Bytes(def ...[]byte) []byte
// 转换为布尔值，支持默认值
func (t Type) Bool(def ...bool) bool
// 转换为整数，支持默认值
func (t Type) Int(def ...int) int
// 转换为8位整数，支持默认值
func (t Type) Int8(def ...int8) int8
// 转换为16位整数，支持默认值
func (t Type) Int16(def ...int16) int16
// 转换为32位整数，支持默认值
func (t Type) Int32(def ...int32) int32
// 转换为64位整数，支持默认值
func (t Type) Int64(def ...int64) int64
// 转换为无符号整数，支持默认值
func (t Type) Uint(def ...uint) uint
// 转换为8位无符号整数，支持默认值
func (t Type) Uint8(def ...uint8) uint8
// 转换为16位无符号整数，支持默认值
func (t Type) Uint16(def ...uint16) uint16
// 转换为32位无符号整数，支持默认值
func (t Type) Uint32(def ...uint32) uint32
// 转换为64位无符号整数，支持默认值
func (t Type) Uint64(def ...uint64) uint64
// 转换为32位浮点数，支持默认值
func (t Type) Float32(def ...float32) float32
// 转换为64位浮点数，支持默认值
func (t Type) Float64(def ...float64) float64
// 转换为时间，支持默认值
func (t Type) Time(format ...string) (time.Time, error)
```

### 切片操作

```go
func (t Type) Slice(noConv ...bool) SliceType
func (t Type) SliceValue(noConv ...bool) []interface{}
func (t Type) SliceString(noConv ...bool) []string
func (t Type) SliceInt(noConv ...bool) []int
```

### 映射操作

```go
func (t Type) Map() Map
func (t Type) Maps() Maps
```

### 工具函数

```go
func ToString(i interface{}) string
func ToBytes(i interface{}) []byte
func ToBool(i interface{}) bool
func ToInt(i interface{}) int
func ToInt8(i interface{}) int8
func ToInt16(i interface{}) int16
func ToInt32(i interface{}) int32
func ToInt64(i interface{}) int64
func ToUint(i interface{}) uint
func ToUint8(i interface{}) uint8
func ToUint16(i interface{}) uint16
func ToUint32(i interface{}) uint32
func ToUint64(i interface{}) uint64
func ToFloat32(i interface{}) float32
func ToFloat64(i interface{}) float64
func ToTime(i interface{}, format ...string) (time.Time, error)
func ToSlice(i interface{}, noConv ...bool) SliceType
func ToMap(i interface{}) Map
```

### 结构体构建

```go
func NewStruct() *StruBuilder
func NewStructFromValue(v interface{}) (*StruBuilder, error)
func NewMapStruct(key interface{}) *StruBuilder
func NewSliceStruct() *StruBuilder
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/ztype"
)

func main() {
    // 创建类型包装器
    data := map[string]interface{}{
        "name": "张三",
        "age":  25,
        "scores": []int{85, 90, 78},
        "address": map[string]interface{}{
            "city": "北京",
            "street": "中关村大街",
        },
    }
    
    // 包装数据
    t := ztype.New(data)
    
    // 检查值是否存在
    if t.Exists() {
        fmt.Println("数据存在")
    }
    
    // 获取简单值
    name := t.Get("name").String()
    fmt.Printf("姓名: %s\n", name)
    
    age := t.Get("age").Int()
    fmt.Printf("年龄: %d\n", age)
    
    // 使用默认值
    email := t.Get("email").String("无邮箱")
    fmt.Printf("邮箱: %s\n", email)
    
    // 路径访问
    city := t.Get("address.city").String()
    fmt.Printf("城市: %s\n", city)
    
    // 数组访问
    firstScore := t.Get("scores.0").Int()
    fmt.Printf("第一个分数: %d\n", firstScore)
    
    // 切片操作
    scores := t.Get("scores").SliceInt()
    fmt.Printf("所有分数: %v\n", scores)
    
    // 类型转换
    ageStr := t.Get("age").String()
    fmt.Printf("年龄字符串: %s\n", ageStr)
    
    // 映射操作
    addressMap := t.Get("address").Map()
    street := addressMap.Get("street").String()
    fmt.Printf("街道: %s\n", street)
    
    // 工具函数
    str := ztype.ToString(123)
    fmt.Printf("数字转字符串: %s\n", str)
    
    num := ztype.ToInt("456")
    fmt.Printf("字符串转数字: %d\n", num)
    
    // 结构体构建
    builder := ztype.NewStruct()
    builder.AddField("Name", "string")
    builder.AddField("Age", "int")
    
    // 获取结构体类型和值
    structType := builder.Type()
    structValue := builder.Value()
    fmt.Printf("构建的结构体类型: %v\n", structType)
    fmt.Printf("构建的结构体值: %v\n", structValue)
    
    fmt.Println("ztype 模块示例完成")
}
```

## 类型定义

### Type
```go
type Type struct {
    v interface{}
}
```

### SliceType
```go
type SliceType struct {
    // 切片类型结构
}
```

### Map
```go
type Map struct {
    // 映射类型结构
}
```

### StruBuilder
```go
type StruBuilder struct {
    // 结构体构建器
}
```

## 路径表达式语法

### 基本语法
- **字段访问**: `name`, `address.city`
- **数组索引**: `scores.0`, `users.1.name`
- **嵌套访问**: `user.addresses.0.street`

### 示例
```go
// 访问嵌套值
city := t.Get("user.address.city").String()

// 访问数组元素
firstUser := t.Get("users.0.name").String()

// 访问嵌套数组
street := t.Get("users.0.addresses.1.street").String()
```

## 最佳实践

1. 使用路径表达式简化访问
2. 提供合适的默认值
3. 检查值的存在性
4. 正确处理空值情况
5. 缓存频繁访问的结果