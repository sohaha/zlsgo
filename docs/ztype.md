# ztype 模块

`ztype` 提供了灵活的类型转换工具和动态类型系统，允许安全地访问值并进行自动类型转换。该模块经过全面的性能优化，提供卓越的执行效率和内存管理。

## 功能概览

- **类型包装**: 安全的类型包装和访问
- **类型转换**: 自动类型转换和验证
- **路径访问**: 嵌套值的路径表达式访问
- **切片处理**: 切片类型的安全操作
- **映射处理**: 映射类型的安全操作
- **结构体构建**: 动态结构体构建
- **性能优化**: 高性能的内存管理和并发支持

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

#### 基础类型转换
```go
func ToString(i interface{}) string         // 转换为字符串，优化类型顺序
func ToBytes(i interface{}) []byte          // 转换为字节数组
func ToBool(i interface{}) bool             // 转换为布尔值
func ToInt(i interface{}) int               // 转换为整数
func ToInt8(i interface{}) int8             // 转换为8位整数
func ToInt16(i interface{}) int16           // 转换为16位整数
func ToInt32(i interface{}) int32           // 转换为32位整数
func ToInt64(i interface{}) int64           // 转换为64位整数，支持分隔符数字
func ToUint(i interface{}) uint             // 转换为无符号整数
func ToUint8(i interface{}) uint8           // 转换为8位无符号整数
func ToUint16(i interface{}) uint16         // 转换为16位无符号整数
func ToUint32(i interface{}) uint32         // 转换为32位无符号整数
func ToUint64(i interface{}) uint64         // 转换为64位无符号整数
func ToFloat32(i interface{}) float32       // 转换为32位浮点数
func ToFloat64(i interface{}) float64       // 转换为64位浮点数，支持百分比
func ToTime(i interface{}, format ...string) (time.Time, error)  // 转换为时间
```

#### 复合类型转换
```go
func ToSlice(i interface{}, noConv ...bool) SliceType  // 转换为切片类型
func ToMap(i interface{}) Map                          // 转换为映射类型
func ToMaps(i interface{}) Maps                        // 转换为映射切片
func ToStruct(v interface{}, outVal interface{}) error // 转换为结构体
```

#### 泛型工具函数
```go
func ToPointer[T any](value T) *T           // 返回值的指针（Go 1.18+）
```

#### 类型检查函数
```go
func IsEmpty(value interface{}) bool        // 检查值是否为空
func IsString(v interface{}) bool           // 检查是否为字符串
func IsBool(v interface{}) bool             // 检查是否为布尔值
func IsInt(v interface{}) bool              // 检查是否为整数
func IsFloat64(v interface{}) bool          // 检查是否为64位浮点数
func IsStruct(v interface{}) bool           // 检查是否为结构体
func GetType(s interface{}) string          // 获取变量类型字符串
```

### 结构体构建

#### 构建器创建
```go
func NewStruct() *StruBuilder                                   // 创建普通结构体构建器
func NewStructFromValue(v interface{}) (*StruBuilder, error)    // 从现有结构体创建构建器
func NewMapStruct(key interface{}) *StruBuilder                 // 创建map[T]struct构建器
func NewSliceStruct() *StruBuilder                              // 创建[]struct构建器
```

#### 构建器方法
```go
func (b *StruBuilder) AddField(name string, fieldType interface{}, tag ...string) *StruBuilder  // 添加字段
func (b *StruBuilder) RemoveField(name string) *StruBuilder                                      // 移除字段
func (b *StruBuilder) HasField(name string) bool                                                 // 检查字段是否存在
func (b *StruBuilder) GetField(name string) *StruField                                           // 获取字段信息
func (b *StruBuilder) FieldNames() []string                                                      // 获取所有字段名
func (b *StruBuilder) Copy(v *StruBuilder) *StruBuilder                                          // 复制构建器配置
func (b *StruBuilder) Merge(values ...interface{}) error                                         // 合并结构体字段
func (b *StruBuilder) Type() reflect.Type                                                        // 获取构建的类型
func (b *StruBuilder) Value() reflect.Value                                                      // 获取构建的值
func (b *StruBuilder) Interface() interface{}                                                    // 获取构建的接口
```

#### 字段方法
```go
func (f *StruField) SetType(typ interface{}) *StruField        // 设置字段类型
func (f *StruField) SetTag(tag string) *StruField              // 设置字段标签
```

### 高级转换配置

#### 转换器配置
```go
// To 和 ValueConv 函数支持自定义转换配置
func To(input, out interface{}, opt ...func(*Conver)) error
func ValueConv(input interface{}, out reflect.Value, opt ...func(*Conver)) error

// Conver 配置选项
type Conver struct {
    MatchName     func(mapKey, fieldName string) bool    // 字段名匹配函数
    ConvHook      func(name string, i reflect.Value, o reflect.Type) (reflect.Value, bool) // 转换钩子
    TagName       string                                 // 结构体标签名
    IgnoreTagName bool                                   // 是否忽略标签
    ZeroFields    bool                                   // 是否写入零值
    Squash        bool                                   // 是否压平嵌套结构体
    Deep          bool                                   // 是否深度复制
    Merge         bool                                   // 是否合并而非替换
}
```

## 使用示例

### 基础示例

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

    // 切片操作（高性能优化）
    scores := t.Get("scores").SliceInt()
    fmt.Printf("所有分数: %v\n", scores)

    // 类型转换
    ageStr := t.Get("age").String()
    fmt.Printf("年龄字符串: %s\n", ageStr)

    // 映射操作
    addressMap := t.Get("address").Map()
    street := addressMap.Get("street").String()
    fmt.Printf("街道: %s\n", street)
    
    // 数字解析（支持分隔符）
    num1 := ztype.ToInt64("1,234,567")      // 支持逗号分隔
    num2 := ztype.ToInt64("1_234_567")      // 支持下划线分隔
    fmt.Printf("数字1: %d, 数字2: %d\n", num1, num2)

    // 百分比解析
    percent := ztype.ToFloat64("85.5%")     // 自动转换为 0.855
    fmt.Printf("百分比: %f\n", percent)

    // 泛型指针工具（Go 1.18+）
    value := 42
    ptr := ztype.ToPointer(value)
    fmt.Printf("指针值: %d\n", *ptr)

    // 类型检查
    if ztype.IsEmpty("") {
        fmt.Println("字符串为空")
    }

    if ztype.IsInt(123) {
        fmt.Println("是整数类型")
    }

    // 结构体转换
    type User struct {
        Name  string `json:"name"`
        Age   int    `json:"age"`
        Email string `json:"email"`
    }

    userData := map[string]interface{}{
        "name": "李四",
        "age":  30,
        "email": "lisi@example.com",
    }

    var user User
    err := ztype.ToStruct(userData, &user)
    if err == nil {
        fmt.Printf("用户: %+v\n", user)
    }
}
```

### 动态结构体构建示例

```go
package main

import (
    "fmt"
    "reflect"
    "github.com/sohaha/zlsgo/ztype"
)

func main() {
    // 创建动态结构体
    builder := ztype.NewStruct()
    builder.AddField("Name", reflect.TypeOf(""), `json:"name"`)
    builder.AddField("Age", reflect.TypeOf(0), `json:"age"`)
    builder.AddField("Active", reflect.TypeOf(true), `json:"active"`)

    // 获取构建的类型
    structType := builder.Type()
    fmt.Printf("动态结构体类型: %v\n", structType)

    // 创建实例
    instance := reflect.New(structType).Elem()

    // 设置字段值
    instance.FieldByName("Name").SetString("动态用户")
    instance.FieldByName("Age").SetInt(25)
    instance.FieldByName("Active").SetBool(true)

    fmt.Printf("动态实例: %v\n", instance.Interface())

    // 从现有结构体创建构建器
    type ExistingStruct struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

    existing := ExistingStruct{ID: 1, Name: "存在的结构体"}
    existingBuilder, _ := ztype.NewStructFromValue(existing)

    // 添加新字段
    existingBuilder.AddField("Email", reflect.TypeOf(""), `json:"email"`)

    newType := existingBuilder.Type()
    fmt.Printf("扩展后的类型: %v\n", newType)
}
```

### 高级转换配置示例

```go
package main

import (
    "fmt"
    "reflect"
    "strings"
    "github.com/sohaha/zlsgo/ztype"
)

func main() {
    // 自定义转换配置
    input := map[string]interface{}{
        "user_name": "张三",
        "user_age":  25,
        "is_active": true,
    }

    type User struct {
        Name   string `custom:"user_name"`
        Age    int    `custom:"user_age"`
        Active bool   `custom:"is_active"`
    }

    var user User
    err := ztype.To(input, &user, func(c *ztype.Conver) {
        // 自定义标签名
        c.TagName = "custom"

        // 自定义字段名匹配（下划线转驼峰）
        c.MatchName = func(mapKey, fieldName string) bool {
            return strings.EqualFold(
                strings.ReplaceAll(mapKey, "_", ""),
                fieldName,
            )
        }

        // 启用深度复制
        c.Deep = true

        // 自定义转换钩子
        c.ConvHook = func(name string, inputVal reflect.Value, outputType reflect.Type) (reflect.Value, bool) {
            if outputType.Kind() == reflect.String {
                // 所有字符串都转为大写
                str := ztype.ToString(inputVal.Interface())
                return reflect.ValueOf(strings.ToUpper(str)), false
            }
            return inputVal, true
        }
    })

    if err == nil {
        fmt.Printf("自定义转换结果: %+v\n", user)
    }
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

### 性能优化建议 🚀

1. **利用对象池优化** - 模块内置对象池自动优化 slice 和 map 的内存分配
2. **高频路径缓存** - 路径表达式会被自动缓存，重复访问相同路径时性能更佳
3. **批量类型转换** - 使用 `ToSlice()` 和 `ToMaps()` 进行批量转换比单个转换更高效
4. **结构体缓存利用** - 结构体字段信息会被缓存，相同类型的转换会更快

### 功能使用建议

1. **使用路径表达式简化访问**
   ```go
   // 推荐：使用路径表达式
   value := t.Get("user.profile.email").String()

   // 不推荐：多层嵌套访问
   user := t.Get("user").Map()
   profile := user.Get("profile").Map()
   email := profile.Get("email").String()
   ```

2. **提供合适的默认值**
   ```go
   // 推荐：提供有意义的默认值
   name := t.Get("name").String("未知用户")
   age := t.Get("age").Int(0)

   // 避免：不提供默认值导致空值
   name := t.Get("name").String() // 可能返回空字符串
   ```

3. **检查值的存在性**
   ```go
   // 推荐：先检查存在性
   if t.Get("optional_field").Exists() {
       value := t.Get("optional_field").String()
       // 处理存在的值
   }

   // 或使用 IsEmpty 进行全面检查
   if !ztype.IsEmpty(t.Get("field").Value()) {
       // 处理非空值
   }
   ```

4. **正确处理空值情况**
   ```go
   // 推荐：使用 IsEmpty 检查
   if ztype.IsEmpty(data) {
       // 处理空值情况
       return
   }

   // 推荐：使用默认值机制
   result := t.Get("field").String("默认值")
   ```

5. **利用转换功能**
   ```go
   // 推荐：使用支持分隔符的数字解析
   amount := ztype.ToInt64("1,234,567")

   // 推荐：使用百分比解析
   rate := ztype.ToFloat64("85.5%")

   // 推荐：使用泛型指针工具（Go 1.18+）
   ptr := ztype.ToPointer(value)
   ```

6. **缓存频繁访问的结果**
   ```go
   // 推荐：对于复杂转换，缓存结果
   type UserCache struct {
       userData ztype.Type
       userMap  ztype.Map
   }

   func (c *UserCache) GetUserMap() ztype.Map {
       if c.userMap == nil {
           c.userMap = c.userData.Map()
       }
       return c.userMap
   }
   ```

7. **合理使用结构体转换配置**
   ```go
   // 推荐：使用配置选项优化转换
   err := ztype.To(input, &output, func(c *ztype.Conver) {
       c.TagName = "json"           // 指定标签名
       c.ZeroFields = false         // 跳过零值字段
       c.Deep = true               // 启用深度复制
   })
   ```

### 内存管理建议

1. **避免在循环中创建大量临时对象**
2. **对于大型数据结构，考虑使用引用而非复制**
3. **利用模块内置的对象池，无需手动管理**
4. **对于频繁访问的路径，让缓存机制自动优化**

### 并发安全建议

1. **Type 对象是并发安全的**，可以在多个 goroutine 中安全使用
2. **结构体缓存是并发安全的**，支持高并发访问
3. **路径缓存是并发安全的**，多线程解析相同路径不会有竞争
