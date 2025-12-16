# zcache 模块

`zcache` 是一个 Go 语言缓存库，提供了简单缓存、表缓存、快速缓存、文件缓存等功能，用于高效的数据缓存和内存管理。

## 功能概览

- **简单缓存**: 基本的键值对缓存
- **表缓存**: 表结构数据缓存
- **快速缓存**: 高性能缓存实现
- **文件缓存**: 持久化文件缓存

## 核心功能

### 简单缓存

```go
// 设置缓存值
func Set(key string, val interface{}, expiration ...time.Duration)
// 删除缓存
func Delete(key string)
// 获取缓存值
func Get(key string) (interface{}, bool)
// 获取类型化缓存值
func GetAny(key string) (ztype.Type, bool)
// 提供者模式获取缓存
func ProvideGet(key string, provide func() (interface{}, bool), expiration ...time.Duration) (interface{}, bool)
```

### 快速缓存

```go
// 创建新的快速缓存实例，支持配置选项
func NewFast(opt ...func(o *Options)) *FastCache
// 设置缓存项，支持过期时间
func (l *FastCache) Set(key string, val interface{}, expiration ...time.Duration)
// 设置字节数组缓存项
func (l *FastCache) SetBytes(key string, b []byte)
// 获取缓存项的值
func (l *FastCache) Get(key string) (interface{}, bool)
// 获取缓存项的类型化值
func (l *FastCache) GetAny(key string) (ztype.Type, bool)
// 获取字节数组缓存项
func (l *FastCache) GetBytes(key string) ([]byte, bool)
// 提供者模式获取缓存，如果不存在则执行回调函数
func (l *FastCache) ProvideGet(key string, provide func() (interface{}, bool), expiration ...time.Duration) (interface{}, bool)
// 删除指定键的缓存项
func (l *FastCache) Delete(key string)
// 遍历所有缓存项
func (l *FastCache) ForEach(walker func(key string, iface interface{}) bool)
// 关闭缓存并清理资源
func (l *FastCache) Close()
```

## 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zcache"
)

func main() {
    // 简单缓存示例
    // 设置缓存
    zcache.Set("name", "张三")
    zcache.Set("age", 25)
    zcache.Set("city", "北京", time.Hour) // 设置过期时间
    
    // 获取缓存
    if name, exists := zcache.Get("name"); exists {
        fmt.Printf("姓名: %v\n", name)
    }
    
    if age, exists := zcache.Get("age"); exists {
        fmt.Printf("年龄: %v\n", age)
    }
    
    if city, exists := zcache.Get("city"); exists {
        fmt.Printf("城市: %v\n", city)
    }
    
    // 获取类型化值
    if ageType, exists := zcache.GetAny("age"); exists {
        ageInt := ageType.Int(0)
        fmt.Printf("年龄整数: %d\n", ageInt)
    }
    
    // 提供者模式
    userData, exists := zcache.ProvideGet("user:123", func() (interface{}, bool) {
        // 模拟从数据库获取用户数据
        user := map[string]interface{}{
            "id":   123,
            "name": "张三",
            "email": "zhangsan@example.com",
        }
        return user, true
    }, time.Minute*30)
    
    if exists {
        fmt.Printf("用户数据: %v\n", userData)
    }
    
    // 删除缓存
    zcache.Delete("age")
    
    // 检查是否还存在
    if _, exists := zcache.Get("age"); !exists {
        fmt.Println("年龄缓存已删除")
    }
    
    // 快速缓存示例
    fastCache := zcache.NewFast()
    
    // 设置缓存
    fastCache.Set("key1", "value1", time.Minute)
    fastCache.Set("key2", 42, time.Hour)
    
    // 设置字节数据
    fastCache.SetBytes("bytes", []byte("hello world"))
    
    // 获取缓存
    if val, exists := fastCache.Get("key1"); exists {
        fmt.Printf("快速缓存值1: %v\n", val)
    }
    
    if val, exists := fastCache.Get("key2"); exists {
        fmt.Printf("快速缓存值2: %v\n", val)
    }
    
    // 获取字节数据
    if bytes, exists := fastCache.GetBytes("bytes"); exists {
        fmt.Printf("字节数据: %s\n", string(bytes))
    }
    
    // 获取类型化值
    if valType, exists := fastCache.GetAny("key2"); exists {
        intVal := valType.Int(0)
        fmt.Printf("类型化整数值: %d\n", intVal)
    }
    
    // 提供者模式
    fastData, exists := fastCache.ProvideGet("dynamic", func() (interface{}, bool) {
        return fmt.Sprintf("动态数据: %d", time.Now().Unix()), true
    }, time.Second*30)
    
    if exists {
        fmt.Printf("快速缓存动态数据: %v\n", fastData)
    }
    
    // 遍历快速缓存
    fastCache.ForEach(func(key string, iface interface{}) bool {
        fmt.Printf("快速缓存 - 键: %s, 值: %v\n", key, iface)
        return true
    })
    
    // 删除缓存项
    fastCache.Delete("key1")
    
    // 关闭快速缓存
    fastCache.Close()
    
    fmt.Println("缓存示例完成")
}
```

## 最佳实践

1. 使用合适的缓存类型
2. 实现缓存预热策略
3. 监控缓存命中率
4. 定期清理过期缓存
