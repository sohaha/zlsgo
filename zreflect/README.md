# zreflect 模块

`zreflect` 提供了反射操作、方法调用、字段访问、类型检查等功能的便捷封装。

## 功能概览

- **反射工具**: 反射值和类型的创建与操作
- **方法操作**: 方法获取、调用和遍历
- **字段操作**: 字段访问、设置和遍历
- **类型工具**: 类型检查、标签获取等工具函数
- **值操作**: 反射值的操作和转换
- **工具函数**: 反射相关的辅助函数

## 核心功能

### 反射工具

```go
// 获取反射值
func ValueOf(v interface{}) reflect.Value
// 获取反射类型
func TypeOf(v interface{}) reflect.Type
// 创建新的反射值包装器
func NewValue(v interface{}) Value
// 创建新的反射类型包装器
func NewType(v interface{}) Type
```

### Type 类型操作

```go
func (t Type) Native() reflect.Type
func (t Type) NumMethod() int
func (t Type) CanExpand() bool
func (t Type) CanInline() bool
func (t Type) IsLabel() bool
```

### Value 类型操作

```go
func (v Value) Native() reflect.Value
func (v Value) Type() Type
```

### 方法操作

```go
func GetAllMethod(s interface{}, fn func(numMethod int, m reflect.Method) error) error
func RunAssignMethod(st interface{}, filter func(methodName string) bool, args ...interface{}) error
func ForEachMethod(valof reflect.Value, fn func(index int, method reflect.Method, value reflect.Value) error) error
```

### 字段操作

```go
func GetUnexportedField(v reflect.Value, field string) (interface{}, error)
func SetUnexportedField(v reflect.Value, field string, value interface{}) error
func ForEach(typ reflect.Type, fn func(parent []string, index int, tag string, field reflect.StructField) error) error
func ForEachValue(val reflect.Value, fn func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error) error
```

### 工具函数

```go
func Nonzero(v reflect.Value) bool
func GetAbbrKind(val reflect.Value) reflect.Kind
func GetStructTag(field reflect.StructField, tags ...string) (tagValue, tagOpts string)
func ReflectStructField(v reflect.Type, fn func(fieldName, fieldTag string, field reflect.StructField) error) error
func ReflectForNumField(v reflect.Value, fn func(fieldName, fieldTag string, field reflect.StructField, fieldValue reflect.Value) error) error
func SetValue(vTypeOf reflect.Kind, vValueOf reflect.Value, value interface{}) error
```

### 特殊常量

```go
// SkipChild 是一个特殊的错误值，用于在 ForEach 和 ForEachValue 回调中
// 指示跳过当前结构体字段的子字段
var SkipChild = errors.New("skip struct")
```

## 使用示例

```go
package main

import (
    "fmt"
    "reflect"
    "time"
    "github.com/sohaha/zlsgo/zreflect"
)

// 示例结构体
type Person struct {
    Name    string    `json:"name" z:"name"`
    Age     int       `json:"age" z:"age"`
    Address Address   `json:"address" z:"address"`
    private string    `z:"private"` // 私有字段
}

type Address struct {
    City  string `json:"city" z:"city"`
    State string `json:"state" z:"state"`
}

func main() {
    // 创建示例数据
    person := Person{
        Name: "张三",
        Age:  30,
        Address: Address{
            City:  "北京",
            State: "北京",
        },
    }
    
    // 反射工具示例
    fmt.Println("=== 反射工具示例 ===")
    
    // 获取反射类型
    typ := zreflect.TypeOf(person)
    fmt.Printf("类型: %v\n", typ.Native())
    fmt.Printf("方法数量: %d\n", typ.NumMethod())
    
    // 获取反射值
    val := zreflect.ValueOf(person)
    fmt.Printf("值类型: %v\n", val.Type().Native())
    
    // 方法操作示例
    fmt.Println("\n=== 方法操作示例 ===")
    
    // 获取所有方法
    err := zreflect.GetAllMethod(person, func(numMethod int, m reflect.Method) error {
        fmt.Printf("方法 %d: %s\n", numMethod, m.Name)
        return nil
    })
    if err != nil {
        fmt.Printf("获取方法失败: %v\n", err)
    }
    
    // 字段操作示例
    fmt.Println("\n=== 字段操作示例 ===")
    
    // 遍历结构体字段
    err = zreflect.ForEach(reflect.TypeOf(person), func(parent []string, index int, tag string, field reflect.StructField) error {
        fmt.Printf("字段: %s, 标签: %s, 类型: %v\n", field.Name, tag, field.Type)
        return nil
    })
    if err != nil {
        fmt.Printf("遍历字段失败: %v\n", err)
    }
    
    // 遍历结构体值
    err = zreflect.ForEachValue(reflect.ValueOf(person), func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error {
        fmt.Printf("字段: %s, 标签: %s, 值: %v\n", field.Name, tag, val.Interface())
        return nil
    })
    if err != nil {
        fmt.Printf("遍历值失败: %v\n", err)
    }
    
    // 访问私有字段（谨慎使用）
    fmt.Println("\n=== 私有字段访问示例 ===")
    
    // 获取私有字段
    privateValue, err := zreflect.GetUnexportedField(reflect.ValueOf(&person), "private")
    if err != nil {
        fmt.Printf("获取私有字段失败: %v\n", err)
    } else {
        fmt.Printf("私有字段值: %v\n", privateValue)
    }
    
    // 设置私有字段
    err = zreflect.SetUnexportedField(reflect.ValueOf(&person), "private", "新的私有值")
    if err != nil {
        fmt.Printf("设置私有字段失败: %v\n", err)
    } else {
        fmt.Println("私有字段设置成功")
    }
    
    // 工具函数示例
    fmt.Println("\n=== 工具函数示例 ===")
    
    // 检查值是否为零值
    zeroValue := reflect.ValueOf("")
    if zreflect.Nonzero(zeroValue) {
        fmt.Println("零值检查: 非零值")
    } else {
        fmt.Println("零值检查: 零值")
    }
    
    // 获取结构体标签
    personType := reflect.TypeOf(person)
    nameField, _ := personType.FieldByName("Name")
    tagValue, tagOpts := zreflect.GetStructTag(nameField, "z", "json")
    fmt.Printf("Name字段标签: %s, 选项: %s\n", tagValue, tagOpts)
    
    // 获取缩写类型
    ageField, _ := personType.FieldByName("Age")
    ageKind := zreflect.GetAbbrKind(reflect.ValueOf(person.Age))
    fmt.Printf("Age字段类型: %v\n", ageKind)
    
    // 反射值包装器示例
    fmt.Println("\n=== 反射值包装器示例 ===")
    
    // 创建新的反射值
    newVal := zreflect.NewValue(person)
    fmt.Printf("新反射值类型: %v\n", newVal.Type().Native())
    
    // 创建新的反射类型
    newType := zreflect.NewType(person)
    fmt.Printf("新反射类型: %v\n", newType.Native())
    
    // 实际应用示例
    fmt.Println("\n=== 实际应用示例 ===")
    
    // 动态设置结构体字段
    config := struct {
        Host     string `z:"host"`
        Port     int    `z:"port"`
        Timeout  time.Duration `z:"timeout"`
        Enabled  bool   `z:"enabled"`
    }{}
    
    // 从配置映射设置字段
    configMap := map[string]interface{}{
        "host":     "localhost",
        "port":     8080,
        "timeout":  "30s",
        "enabled":  "true",
    }
    
    configVal := reflect.ValueOf(&config).Elem()
    configType := configVal.Type()
    
    for i := 0; i < configVal.NumField(); i++ {
        field := configVal.Field(i)
        fieldType := configType.Field(i)
        
        // 获取字段标签
        tagValue, _ := zreflect.GetStructTag(fieldType, "z")
        if tagValue == "" {
            continue
        }
        
        // 从配置映射获取值
        if configValue, exists := configMap[tagValue]; exists {
            // 设置字段值
            err := zreflect.SetValue(field.Kind(), field, configValue)
            if err != nil {
                fmt.Printf("设置字段 %s 失败: %v\n", fieldType.Name, err)
            } else {
                fmt.Printf("字段 %s 设置成功: %v\n", fieldType.Name, field.Interface())
            }
        }
    }
    
    fmt.Printf("最终配置: %+v\n", config)
    
    // 结构体验证示例
    fmt.Println("\n=== 结构体验证示例 ===")
    
    // 验证必填字段
    requiredFields := []string{"Name", "Age"}
    personVal := reflect.ValueOf(person)
    
    for _, fieldName := range requiredFields {
        field := personVal.FieldByName(fieldName)
        if !field.IsValid() {
            fmt.Printf("字段 %s 不存在\n", fieldName)
            continue
        }
        
        if zreflect.Nonzero(field) {
            fmt.Printf("字段 %s 验证通过\n", fieldName)
        } else {
            fmt.Printf("字段 %s 验证失败: 不能为空\n", fieldName)
        }
    }
    
    fmt.Println("反射工具示例完成")
}
```

## 高级用法

### 递归字段遍历
```go
// 使用 SkipChild 跳过某些字段的子字段
err := zreflect.ForEachValue(reflect.ValueOf(data), func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error {
    // 跳过某些字段的子字段
    if field.Name == "SkipField" {
        return zreflect.SkipChild
    }
    
    // 处理字段
    fmt.Printf("字段: %s, 路径: %v\n", field.Name, parent)
    return nil
})
```

### 动态方法调用
```go
// 根据条件调用特定方法
err := zreflect.RunAssignMethod(obj, func(methodName string) bool {
    // 只调用以 "Get" 开头的方法
    return strings.HasPrefix(methodName, "Get")
}, arg1, arg2)
```

### 类型安全的值设置
```go
// 安全地设置不同类型的值
err := zreflect.SetValue(reflect.String, stringField, "新值")
err = zreflect.SetValue(reflect.Int, intField, "42")
err = zreflect.SetValue(reflect.Bool, boolField, "true")
```

## 最佳实践

1. 缓存反射结果
2. 使用类型断言优化性能
3. 实现适当的错误处理
4. 谨慎使用私有字段访问