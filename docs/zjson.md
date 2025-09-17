# zjson 模块

`zjson` 提供了 JSON 解析、查询、设置、格式化、验证、修复、转换等功能，用于高效的 JSON 数据操作。

## 功能概览

- **JSON 解析**: JSON 字符串和字节数组解析
- **JSON 查询**: 使用路径语法查询 JSON 数据
- **JSON 设置**: 修改和设置 JSON 数据
- **JSON 格式化**: JSON 数据格式化和美化
- **JSON 验证**: JSON 数据有效性验证
- **JSON 修复**: 修复损坏的 JSON 数据
- **JSON 转换**: JSON 数据类型转换
- **高级功能**: 修饰符、遍历等高级特性

## 核心功能

### JSON 解析与查询

```go
// 解析JSON字符串
func Parse(data string) *Res
// 解析JSON字节数组
func ParseBytes(data []byte) *Res
// 反序列化JSON到结构体
func Unmarshal(data []byte, v interface{}) error
// 获取字节数组值
func (j *Res) GetBytes(key string) []byte
// 获取多个值
func (j *Res) GetMultiple(keys ...string) []interface{}
// 获取多个字节数组值
func (j *Res) GetMultipleBytes(keys ...string) [][]byte
// 检查键是否存在
func (j *Res) Exists(key string) bool
// 获取值（支持路径）
func (j *Res) Get(path string) *Res
// 获取值（支持路径，字节数组版本）
func (j *Res) GetBytes(path string) []byte
// 获取多个值（支持路径）
func GetMultiple(json, path string, keys ...string) []interface{}
// 获取多个值（字节数组版本）
func GetMultipleBytes(json []byte, path string, keys ...string) [][]byte
// 检查路径是否存在
func Exists(json, path string) bool
// 检查路径是否存在（字节数组版本）
func ExistsBytes(json []byte, path string) bool
```

### JSON 设置

```go
// 设置JSON值
func Set(json, path string, value interface{}) (string, error)
// 设置JSON值（字节数组版本）
func SetBytes(json []byte, path string, value interface{}) ([]byte, error)
// 使用选项设置JSON值
func SetOptions(json, path string, value interface{}, opts *Options) (string, error)
// 设置原始JSON值
func SetRaw(json, path, value string) (string, error)
// 设置原始JSON值（字节数组版本）
func SetRawBytes(json []byte, path string, value []byte) ([]byte, error)
// 使用选项设置原始JSON值
func SetRawOptions(json, path, value string, opts *Options) (string, error)
// 使用选项设置JSON值（字节数组版本）
func SetBytesOptions(json []byte, path string, value interface{}, opts *Options) ([]byte, error)
// 设置值（Res方法）
func (r *Res) Set(path string, value interface{}) error
// 删除值（Res方法）
func (r *Res) Delete(path string) error
// 设置值（Res方法，字节数组版本）
func (r *Res) SetBytes(path string, value []byte) error
// 设置原始值（Res方法）
func (r *Res) SetRaw(path, value string) error
// 设置原始值（Res方法，字节数组版本）
func (r *Res) SetRawBytes(path string, value []byte) error
```

### JSON 格式化

```go
// 格式化JSON（美化）
func Format(json []byte) []byte
// 使用选项格式化JSON
func FormatOptions(json []byte, opts *StFormatOptions) []byte
// 压缩JSON
func Ugly(json []byte) []byte
// 丢弃特定内容
func Discard(json string) (string, error)
// 格式化JSON（字符串版本）
func FormatString(json string) string
// 使用选项格式化JSON（字符串版本）
func FormatStringOptions(json string, opts *StFormatOptions) string
// 压缩JSON（字符串版本）
func UglyString(json string) string
```

### JSON 验证

```go
// 验证JSON是否有效
func Valid(data []byte) bool
// 验证JSON是否有效（字符串版本）
func ValidString(data string) bool
// 验证JSON路径是否存在
func ValidPath(json, path string) bool
// 验证JSON路径是否存在（字节数组版本）
func ValidPathBytes(json []byte, path string) bool
```

### JSON 修复

```go
// 修复损坏的JSON
func Repair(src string, opt ...func(*RepairOptions)) (dst string, err error)
// 修复损坏的JSON（字节数组版本）
func RepairBytes(src []byte, opt ...func(*RepairOptions)) (dst []byte, err error)
// 修复JSON并返回结果
func RepairToRes(src string, opt ...func(*RepairOptions)) (*Res, error)
// 修复JSON并返回结果（字节数组版本）
func RepairBytesToRes(src []byte, opt ...func(*RepairOptions)) (*Res, error)
```

### JSON 转换

```go
// 将值转换为JSON字符串
func Stringify(value interface{}) string
// 序列化JSON
func Marshal(json interface{}) ([]byte, error)
// 将值转换为JSON字符串（带选项）
func StringifyOptions(value interface{}, opts *Options) string
// 序列化JSON（带选项）
func MarshalOptions(json interface{}, opts *Options) ([]byte, error)

// 类型转换方法
func (r *Res) String(def ...string) string
func (r *Res) Bool(def ...bool) bool
func (r *Res) Int(def ...int) int
func (r *Res) Int8(def ...int8) int8
func (r *Res) Int16(def ...int16) int16
func (r *Res) Int32(def ...int32) int32
func (r *Res) Int64(def ...int64) int64
func (r *Res) Uint(def ...uint) uint
func (r *Res) Uint8(def ...uint8) uint8
func (r *Res) Uint16(def ...uint16) uint16
func (r *Res) Uint32(def ...uint32) uint32
func (r *Res) Uint64(def ...uint64) uint64
func (r *Res) Float64(def ...float64) float64
func (r *Res) Float(def ...float64) float64
func (r *Res) Float32(def ...float32) float32
func (r *Res) Time(format ...string) time.Time
func (r *Res) Array() []*Res
func (r *Res) Slice() ztype.SliceType
func (r *Res) SliceValue(noConv ...bool) []interface{}
func (r *Res) SliceString() []string
func (r *Res) SliceInt() []int
func (r *Res) Maps() ztype.Maps
func (r *Res) MapRes() map[string]*Res
func (r *Res) Map() ztype.Map
func (r *Res) MapKeys(exclude ...string) []string
func (r *Res) IsObject() bool
func (r *Res) IsArray() bool
func (r *Res) Raw() string
func (r *Res) Bytes() []byte
func (r *Res) Unmarshal(v interface{}) error
// 获取值
func (r *Res) Value() interface{}
// 获取值（带默认值）
func (r *Res) ValueOr(def interface{}) interface{}
// 检查是否存在
func (r *Res) Exists() bool
// 检查是否为空
func (r *Res) IsEmpty() bool
// 检查是否为null
func (r *Res) IsNull() bool
// 检查是否为数字
func (r *Res) IsNumber() bool
// 检查是否为字符串
func (r *Res) IsString() bool
// 检查是否为布尔值
func (r *Res) IsBool() bool
```

### 高级功能

```go
// 遍历JSON对象
func (r *Res) ForEach(fn func(key, value *Res) bool)
// 匹配键
func (r *Res) MatchKeys(keys []string) *Res
// 过滤JSON对象
func (r *Res) Filter(fn func(key, value *Res) bool) *Res
// 逐行处理JSON
func ForEachLine(json string, fn func(line *Res) bool)
// 遍历JSON对象（字节数组版本）
func ForEachLineBytes(json []byte, fn func(line *Res) bool)
// 合并JSON对象
func Merge(json1, json2 string) (string, error)
// 合并JSON对象（字节数组版本）
func MergeBytes(json1, json2 []byte) ([]byte, error)
// 比较JSON对象
func Equal(json1, json2 string) bool
// 比较JSON对象（字节数组版本）
func EqualBytes(json1, json2 []byte) bool
```

### 删除操作

```go
// 删除JSON字段
func Delete(json, path string) (string, error)
// 删除JSON字段（字节数组版本）
func DeleteBytes(json []byte, path string) ([]byte, error)
// 使用选项删除JSON字段
func DeleteOptions(json, path string, opts *Options) (string, error)
// 使用选项删除JSON字段（字节数组版本）
func DeleteBytesOptions(json []byte, path string, opts *Options) ([]byte, error)
// 删除多个字段
func DeleteMultiple(json string, paths ...string) (string, error)
// 删除多个字段（字节数组版本）
func DeleteMultipleBytes(json []byte, paths ...string) ([]byte, error)
```

### 内存池管理

```go
// 设置内存池
func SetPool(p Pool)
// 获取内存池
func GetPool() Pool
// 创建新的内存池
func NewPool() Pool
// 重置内存池
func ResetPool()
// 获取内存池统计信息
func GetPoolStats() PoolStats
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zjson"
)

func main() {
    // JSON 解析示例
    jsonStr := `{
        "name": "张三",
        "age": 25,
        "city": "北京",
        "hobbies": ["读书", "游泳", "编程"],
        "address": {
            "street": "中关村大街",
            "district": "海淀区",
            "postcode": "100080"
        },
        "active": true,
        "score": 95.5
    }`
    
    // 解析 JSON
    result := zjson.Parse(jsonStr)
    if result.Exists() {
        fmt.Println("JSON 解析成功")
    }
    
    // JSON 查询示例
    // 获取简单值
    name := result.Get("name").String()
    fmt.Printf("姓名: %s\n", name)
    
    age := result.Get("age").Int()
    fmt.Printf("年龄: %d\n", age)
    
    city := result.Get("city").String()
    fmt.Printf("城市: %s\n", city)
    
    active := result.Get("active").Bool()
    fmt.Printf("是否活跃: %t\n", active)
    
    score := result.Get("score").Float64()
    fmt.Printf("分数: %.1f\n", score)
    
    // 获取数组
    hobbies := result.Get("hobbies").Array()
    fmt.Printf("爱好数量: %d\n", len(hobbies))
    for i, hobby := range hobbies {
        fmt.Printf("爱好 %d: %s\n", i+1, hobby.String())
    }
    
    // 获取嵌套对象
    street := result.Get("address.street").String()
    fmt.Printf("街道: %s\n", street)
    
    district := result.Get("address.district").String()
    fmt.Printf("区域: %s\n", district)
    
    postcode := result.Get("address.postcode").String()
    fmt.Printf("邮编: %s\n", postcode)
    
    // 检查字段是否存在
    if result.Get("email").Exists() {
        email := result.Get("email").String()
        fmt.Printf("邮箱: %s\n", email)
    } else {
        fmt.Println("邮箱字段不存在")
    }
    
    // 使用默认值
    email := result.Get("email").String("无邮箱")
    fmt.Printf("邮箱: %s\n", email)
    
    // 类型转换示例
    // 转换为整数
    ageInt := result.Get("age").Int(0)
    fmt.Printf("年龄(整数): %d\n", ageInt)
    
    // 转换为浮点数
    scoreFloat := result.Get("score").Float64(0.0)
    fmt.Printf("分数(浮点): %.2f\n", scoreFloat)
    
    // 转换为布尔值
    activeBool := result.Get("active").Bool(false)
    fmt.Printf("活跃状态: %t\n", activeBool)
    
    // 转换为时间
    if result.Get("created_at").Exists() {
        createdAt := result.Get("created_at").Time("2006-01-02 15:04:05")
        fmt.Printf("创建时间: %v\n", createdAt)
    }
    
    // 数组操作示例
    hobbiesArray := result.Get("hobbies").Array()
    fmt.Printf("爱好数组长度: %d\n", len(hobbiesArray))
    
    // 获取第一个爱好
    if len(hobbiesArray) > 0 {
        firstHobby := hobbiesArray[0].String()
        fmt.Printf("第一个爱好: %s\n", firstHobby)
    }
    
    // 获取特定索引的爱好
    if len(hobbiesArray) > 1 {
        secondHobby := hobbiesArray[1].String()
        fmt.Printf("第二个爱好: %s\n", secondHobby)
    }
    
    // 映射操作示例
    addressMap := result.Get("address").MapRes()
    fmt.Printf("地址映射: %+v\n", addressMap)
    
    // 获取所有键
    addressKeys := result.Get("address").MapKeys()
    fmt.Printf("地址字段: %v\n", addressKeys)
    
    // 排除某些字段
    excludeKeys := result.Get("address").MapKeys("postcode")
    fmt.Printf("排除邮编后的字段: %v\n", excludeKeys)
    
    // 遍历操作示例
    fmt.Println("=== 遍历所有字段 ===")
    result.ForEach(func(key, value *zjson.Res) bool {
        fmt.Printf("键: %s, 值: %s\n", key.String(), value.String())
        return true // 继续遍历
    })
    
    // 匹配键示例
    fmt.Println("=== 匹配特定键 ===")
    matched := result.MatchKeys([]string{"name", "age", "city"})
    matched.ForEach(func(key, value *zjson.Res) bool {
        fmt.Printf("匹配键: %s, 值: %s\n", key.String(), value.String())
        return true
    })
    
    // 过滤示例
    fmt.Println("=== 过滤字段 ===")
    filtered := result.Filter(func(key, value *zjson.Res) bool {
        // 只保留字符串类型的字段
        return value.IsObject() == false && value.IsArray() == false
    })
    filtered.ForEach(func(key, value *zjson.Res) bool {
        fmt.Printf("过滤后: %s = %s\n", key.String(), value.String())
        return true
    })
    
    // JSON 设置示例
    // 设置新值
    newJson, err := zjson.Set(jsonStr, "email", "zhangsan@example.com")
    if err == nil {
        fmt.Printf("设置邮箱后的 JSON: %s\n", newJson)
    }
    
    // 设置嵌套值
    newJson, err = zjson.Set(jsonStr, "address.country", "中国")
    if err == nil {
        fmt.Printf("设置国家后的 JSON: %s\n", newJson)
    }
    
    // 设置数组值
    newJson, err = zjson.Set(jsonStr, "hobbies.3", "旅行")
    if err == nil {
        fmt.Printf("添加爱好后的 JSON: %s\n", newJson)
    }
    
    // 使用选项设置
    opts := &zjson.Options{
        // 设置选项
    }
    newJson, err = zjson.SetOptions(jsonStr, "status", "在线", opts)
    if err == nil {
        fmt.Printf("使用选项设置后的 JSON: %s\n", newJson)
    }
    
    // 设置原始值
    newJson, err = zjson.SetRaw(jsonStr, "raw_data", `{"key": "value"}`)
    if err == nil {
        fmt.Printf("设置原始值后的 JSON: %s\n", newJson)
    }
    
    // 字节数组操作
    jsonBytes := []byte(jsonStr)
    newBytes, err := zjson.SetBytes(jsonBytes, "version", "1.0")
    if err == nil {
        fmt.Printf("字节数组设置后的长度: %d\n", len(newBytes))
    }
    
    // JSON 删除示例
    // 删除字段
    deletedJson, err := zjson.Delete(jsonStr, "score")
    if err == nil {
        fmt.Printf("删除分数后的 JSON: %s\n", deletedJson)
    }
    
    // 删除嵌套字段
    deletedJson, err = zjson.Delete(jsonStr, "address.postcode")
    if err == nil {
        fmt.Printf("删除邮编后的 JSON: %s\n", deletedJson)
    }
    
    // 删除数组元素
    deletedJson, err = zjson.Delete(jsonStr, "hobbies.1")
    if err == nil {
        fmt.Printf("删除第二个爱好后的 JSON: %s\n", deletedJson)
    }
    
    // 字节数组删除
    deletedBytes, err := zjson.DeleteBytes(jsonBytes, "age")
    if err == nil {
        fmt.Printf("删除年龄后的字节长度: %d\n", len(deletedBytes))
    }
    
    // JSON 格式化示例
    // 美化 JSON
    formatted := zjson.Format(jsonBytes)
    fmt.Printf("美化后的 JSON:\n%s\n", string(formatted))
    
    // 压缩 JSON
    ugly := zjson.Ugly(jsonBytes)
    fmt.Printf("压缩后的 JSON: %s\n", string(ugly))
    
    // 使用选项格式化
    formatOpts := &zjson.StFormatOptions{
        // 格式化选项
    }
    formattedWithOpts := zjson.FormatOptions(jsonBytes, formatOpts)
    fmt.Printf("使用选项格式化后的 JSON:\n%s\n", string(formattedWithOpts))
    
    // 丢弃特定内容
    discarded, err := zjson.Discard(jsonStr)
    if err == nil {
        fmt.Printf("丢弃后的 JSON: %s\n", discarded)
    }
    
    // JSON 验证示例
    // 验证 JSON 有效性
    isValid := zjson.Valid(jsonBytes)
    fmt.Printf("JSON 是否有效: %t\n", isValid)
    
    // 验证无效 JSON
    invalidJSON := `{"name": "张三", "age": 25,}`
    isValid = zjson.Valid([]byte(invalidJSON))
    fmt.Printf("无效 JSON 是否有效: %t\n", isValid)
    
    // JSON 修复示例
    // 修复损坏的 JSON
    brokenJSON := `{"name": "张三", "age": 25, "city": "北京",}`
    fixedJSON, err := zjson.Repair(brokenJSON)
    if err == nil {
        fmt.Printf("修复后的 JSON: %s\n", fixedJSON)
    }
    
    // 使用修复选项
    repairOpts := func(opts *zjson.RepairOptions) {
        // 设置修复选项
    }
    fixedJSON, err = zjson.Repair(brokenJSON, repairOpts)
    if err == nil {
        fmt.Printf("使用选项修复后的 JSON: %s\n", fixedJSON)
    }
    
    // JSON 转换示例
    // 转换为字符串
    jsonString := zjson.Stringify(result.Value())
    fmt.Printf("转换后的字符串: %s\n", jsonString)
    
    // 序列化
    marshaled, err := zjson.Marshal(result.Value())
    if err == nil {
        fmt.Printf("序列化后的字节长度: %d\n", len(marshaled))
    }
    
    // 反序列化
    var userData map[string]interface{}
    err = zjson.Unmarshal(jsonBytes, &userData)
    if err == nil {
        fmt.Printf("反序列化后的数据: %+v\n", userData)
    }
    
    // 多路径查询示例
    paths := []string{"name", "age", "city", "address.street"}
    multipleResults := zjson.GetMultiple(jsonStr, paths...)
    fmt.Printf("多路径查询结果数量: %d\n", len(multipleResults))
    
    for i, path := range paths {
        if i < len(multipleResults) {
            fmt.Printf("路径 %s: %s\n", path, multipleResults[i].String())
        }
    }
    
    // 字节数组多路径查询
    multipleBytes := zjson.GetMultipleBytes(jsonBytes, paths...)
    fmt.Printf("字节数组多路径查询结果数量: %d\n", len(multipleBytes))
    
    // 逐行处理示例
    multiLineJSON := `{"name": "张三"}
{"name": "李四"}
{"name": "王五"}`
    
    zjson.ForEachLine(multiLineJSON, func(line *zjson.Res) bool {
        name := line.Get("name").String()
        fmt.Printf("处理行: %s\n", name)
        return true // 继续处理
    })
    
    // 实际应用示例
    // 用户配置文件处理
    userConfig := `{ 
        "user": { 
            "id": 1001, 
            "name": "张三", 
            "settings": { 
                "theme": "dark", 
                "language": "zh-CN", 
                "notifications": true 
            } 
        }, 
        "system": { 
            "version": "1.0.0", 
            "debug": false 
        } 
    }`
    
    config := zjson.Parse(userConfig)
    
    // 获取用户设置
    theme := config.Get("user.settings.theme").String("light")
    language := config.Get("user.settings.language").String("en")
    notifications := config.Get("user.settings.notifications").Bool(true)
    
    fmt.Printf("用户设置 - 主题: %s, 语言: %s, 通知: %t\n", 
        theme, language, notifications)
    
    // 更新用户设置
    updatedConfig, err := zjson.Set(userConfig, "user.settings.theme", "light")
    if err == nil {
        fmt.Printf("更新主题后的配置: %s\n", updatedConfig)
    }
    
    // 添加新设置
    updatedConfig, err = zjson.Set(userConfig, "user.settings.timezone", "Asia/Shanghai")
    if err == nil {
        fmt.Printf("添加时区后的配置: %s\n", updatedConfig)
    }
    
    // 删除设置
    updatedConfig, err = zjson.Delete(userConfig, "user.settings.notifications")
    if err == nil {
        fmt.Printf("删除通知设置后的配置: %s\n", updatedConfig)
    }
    
    // API 响应处理
    apiResponse := `{ 
        "code": 200, 
        "message": "success", 
        "data": { 
            "users": [ 
                {"id": 1, "name": "用户1", "email": "user1@example.com"},
                {"id": 2, "name": "用户2", "email": "user2@example.com"},
                {"id": 3, "name": "用户3", "email": "user3@example.com"}
            ], 
            "total": 3, 
            "page": 1, 
            "size": 10 
        } 
    }`
    
    response := zjson.Parse(apiResponse)
    
    // 检查响应状态
    code := response.Get("code").Int()
    message := response.Get("message").String()
    
    if code == 200 {
        fmt.Printf("API 调用成功: %s\n", message)
        
        // 处理用户数据
        users := response.Get("data.users").Array()
        total := response.Get("data.total").Int()
        page := response.Get("data.page").Int()
        size := response.Get("data.size").Int()
        
        fmt.Printf("用户列表 (第 %d 页，每页 %d 条，共 %d 条):\n", page, size, total)
        
        for i, user := range users {
            id := user.Get("id").Int()
            name := user.Get("name").String()
            email := user.Get("email").String()
            fmt.Printf("  %d. ID: %d, 姓名: %s, 邮箱: %s\n", i+1, id, name, email)
        }
        
        // 过滤特定用户
        filteredUsers := response.Get("data.users").Filter(func(key, value *zjson.Res) bool {
            // 只保留 ID 大于 1 的用户
            return value.Get("id").Int() > 1
        })
        
        fmt.Printf("过滤后的用户数量: %d\n", len(filteredUsers.Array()))
        
    } else {
        fmt.Printf("API 调用失败: %s (代码: %d)\n", message, code)
    }
    
    // 日志数据处理
    logData := `{ 
        "timestamp": "2024-01-15T10:30:00Z", 
        "level": "INFO", 
        "message": "用户登录成功", 
        "user_id": 1001, 
        "ip": "192.168.1.100", 
        "metadata": { 
            "browser": "Chrome", 
            "os": "Windows", 
            "version": "120.0.0.0" 
        } 
    }`
    
    log := zjson.Parse(logData)
    
    // 提取日志信息
    timestamp := log.Get("timestamp").Time("2006-01-02T15:04:05Z")
    level := log.Get("level").String()
    message := log.Get("message").String()
    userID := log.Get("user_id").Int()
    ip := log.Get("ip").String()
    
    fmt.Printf("日志记录:\n")
    fmt.Printf("  时间: %v\n", timestamp)
    fmt.Printf("  级别: %s\n", level)
    fmt.Printf("  消息: %s\n", message)
    fmt.Printf("  用户ID: %d\n", userID)
    fmt.Printf("  IP地址: %s\n", ip)
    
    // 处理元数据
    metadata := log.Get("metadata")
    if metadata.Exists() {
        browser := metadata.Get("browser").String()
        os := metadata.Get("os").String()
        version := metadata.Get("version").String()
        
        fmt.Printf("  浏览器: %s %s\n", browser, version)
        fmt.Printf("  操作系统: %s\n", os)
    }
    
    fmt.Println("JSON 处理示例完成")
}
```

## 类型定义

### JSON 结果
```go
type Res struct {
    // JSON 解析结果
}
```

### 格式化选项
```go
type StFormatOptions struct {
    // 格式化选项
}
```

### 修复选项
```go
type RepairOptions struct {
    // 修复选项
}
```

### 设置选项
```go
type Options struct {
    // 设置选项
}
```

### 内存池接口
```go
type Pool interface {
    // 内存池接口
}
```

### 内存池统计信息
```go
type PoolStats struct {
    // 内存池统计信息
}
```

## JSON 路径语法

### 基本语法
- **对象字段**: `name`, `address.street`
- **数组索引**: `hobbies.0`, `users.1.name`
- **通配符**: `users.*.name`
- **范围**: `users.0:2.name`
- **条件**: `users[?(@.age > 18)]`

### 高级语法
- **管道操作**: `users.*.name | length`
- **函数调用**: `users.*.age | avg`
- **正则匹配**: `users[?(@.name =~ /张.*/)]`

## 最佳实践

1. 使用适当的错误处理
2. 缓存频繁查询的结果
3. 优化 JSON 路径表达式
4. 监控内存使用情况