# zarray 模块

`zarray` 是一个 Go 语言数组和切片操作库，提供了丰富的数组操作方法、泛型支持、高性能的哈希映射、排序映射等功能，用于数组数据处理和集合操作。

## 功能概览

- **数组操作**: 动态数组的增删改查操作
- **切片工具**: 切片的高效操作和转换
- **哈希映射**: 高性能的键值对存储
- **排序映射**: 有序的键值对存储
- **泛型支持**: 类型安全的集合操作
- **并行处理**: 支持并行的数组操作

## 核心功能

### 数组操作

```go
// 按容量创建一个新的动态数组
func NewArray(capacity ...int) *Array
// 创建新数组（NewArray 的别名）
func New(capacity ...int) *Array
// 从 interface{} 数组拷贝生成 Array
func CopyArray(arr interface{}) (*Array, error)
// 从 interface{} 数组拷贝生成 Array（CopyArray 的别名）
func Copy(arr interface{}) (*Array, error)
// 从 []interface{} 中安全获取 index 位置元素，支持默认值
func GetInf(arr []interface{}, index int, def ...interface{}) (interface{}, error)
```

### 数组方法

```go
// 在指定位置插入元素
func (a *Array) Add(index int, value interface{}) error
// 在头部插入一个元素
func (a *Array) Unshift(value interface{}) error
// 在尾部追加一个或多个元素
func (a *Array) Push(values ...interface{})
// 获取指定位置元素，支持默认值
func (a *Array) Get(index int, def ...interface{}) (interface{}, error)
// 设置指定位置元素
func (a *Array) Set(index int, value interface{}) error
// 从指定位置删除若干个元素，返回被删除的元素切片
func (a *Array) Remove(index int, l ...int) ([]interface{}, error)
// 删除第一个等于 value 的元素，返回被删除的元素
func (a *Array) RemoveValue(value interface{}) (interface{}, error)
// 删除并返回头部元素
func (a *Array) Shift() (interface{}, error)
// 删除并返回尾部元素
func (a *Array) Pop() (interface{}, error)
// 判断是否包含某个值
func (a *Array) Contains(value interface{}) bool
// 查找某个值的索引，不存在返回 -1
func (a *Array) Index(value interface{}) int
// 返回当前元素个数
func (a *Array) Length() int
// 返回当前容量
func (a *Array) CapLength() int
// 是否为空
func (a *Array) IsEmpty() bool
// 生成一个新的数组，对每个元素应用映射函数
func (a *Array) Map(fn func(int, interface{}) interface{}) *Array
// 打乱元素顺序，返回新数组
func (a *Array) Shuffle() *Array
// 返回底层原始切片的副本
func (a *Array) Raw() []interface{}
// 清空数组
func (a *Array) Clear()
// 返回包含大小、容量与内容的字符串
func (a *Array) Format() string
```

### 切片工具

```go
// 拷贝一个切片副本
func CopySlice[T any](l []T) []T
// 从切片中随机取一个值
func Rand[T any](collection []T) T
// 从切片中随机取 n 个值
func RandPickN[T any](collection []T, n int) []T
// 对切片做映射转换，可选并行度
func Map[T any, R any](collection []T, iteratee func(int, T) R, parallel ...uint) []R
// 并行映射（已废弃，建议用 Map 的并行参数）
func ParallelMap[T any, R any](collection []T, iteratee func(int, T) R, workers uint) []R
// 打乱切片并返回新切片
func Shuffle[T any](collection []T) []T
// 反转切片并返回新切片
func Reverse[T any](collection []T) []T
// 过滤切片，保留满足条件的元素
func Filter[T any](slice []T, predicate func(index int, item T) bool) []T
// 判断切片是否包含指定值
func Contains[T comparable](collection []T, v T) bool
// 查找并返回第一个满足条件的元素
func Find[T any](collection []T, predicate func(index int, item T) bool) (res T, ok bool)
// 去重，保留首次出现的元素
func Unique[T comparable](collection []T) []T
// 求两个切片的差集
func Diff[T comparable](list1 []T, list2 []T) ([]T, []T)
// 从切片尾部弹出一个元素（修改原切片）
func Pop[T comparable](list *[]T) (v T)
// 从切片头部弹出一个元素（修改原切片）
func Shift[T comparable](list *[]T) (v T)
// 将切片按固定大小分块
func Chunk[T any](slice []T, size int) [][]T
// 构造随机移位函数，每次调用返回并移除一个随机元素
func RandShift[T comparable](list []T) func() (T, error)
// 按优先列表排序，first 先排前，last 排后
func SortWithPriority[T comparable](slice []T, first, last []T) []T
// 求两个切片的交集
func Intersection[T comparable](list1 []T, list2 []T) []T
```

### 排序映射

```go
// 创建维护插入顺序的映射
func NewSortMap[K hashable, V any](size ...uintptr) *SortMaper[K, V]
// 设置键值，首次出现的键会追加到有序键列表
func (s *SortMaper[K, V]) Set(key K, value V)
// 按键获取值
func (s *SortMaper[K, V]) Get(key K) (value V, ok bool)
// 判断是否存在指定键
func (s *SortMaper[K, V]) Has(key K) (ok bool)
// 删除一个或多个键
func (s *SortMaper[K, V]) Delete(key ...K)
// 返回元素数量
func (s *SortMaper[K, V]) Len() int
// 返回按插入顺序排列的键列表
func (s *SortMaper[K, V]) Keys() []K
// 以插入顺序遍历键值对，返回 true 继续，false 终止
func (s *SortMaper[K, V]) ForEach(lambda func(K, V) bool)
```

### 哈希映射

```go
// 创建并发安全的泛型哈希映射
func NewHashMap[K hashable, V any](size ...uintptr) *Maper[K, V]
// 设置键值
func (m *Maper[K, V]) Set(key K, value V)
// 获取键值
func (m *Maper[K, V]) Get(key K) (value V, ok bool)
// 获取或设置（若不存在则设置并返回）
func (m *Maper[K, V]) GetOrSet(key K, value V) (actual V, loaded bool)
// 获取并删除
func (m *Maper[K, V]) GetAndDelete(key K) (value V, ok bool)
// 提供函数计算值（可缓存），返回值、是否已存在、是否计算得到
func (m *Maper[K, V]) ProvideGet(key K, provide func() (V, bool)) (actual V, loaded, computed bool)
// 删除一个或多个键
func (m *Maper[K, V]) Delete(keys ...K)
// 判断键是否存在
func (m *Maper[K, V]) Has(key K) (ok bool)
// 返回元素数量
func (m *Maper[K, V]) Len() int
// 遍历所有键值对，返回 true 继续，false 终止
func (m *Maper[K, V]) Range(f func(key K, value V) bool)
// 清空映射
func (m *Maper[K, V]) Clear()
// 交换键对应的值，返回旧值与是否成功
func (m *Maper[K, V]) Swap(key K, newValue V) (oldValue V, swapped bool)
// 比较并交换
func (m *Maper[K, V]) CAS(key K, oldValue, newValue V) bool
```

### 工具函数

```go
// 获取 map 的所有键
func Keys[K comparable, V any](in map[K]V) []K
// 获取 map 的所有值
func Values[K comparable, V any](in map[K]V) []V
// 根据转换函数将切片转为以键为索引的 map
func IndexMap[K comparable, V any](arr []V, toKey func(V) (K, V)) (map[K]V, error)
// 将 map 的值按函数映射为切片
func FlatMap[K comparable, V any](m map[K]V, fn func(key K, value V) V) []V
```

### 字符串处理

```go
// 将字符串按分隔符分割为切片
func Slice[T comparable](s, sep string, n ...int) []T
// 将切片按分隔符连接为字符串
func Join[T comparable](s []T, sep string) string
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zarray"
)

func main() {
    // 数组操作示例
    arr := zarray.NewArray(5)
    // 头部插入
    _ = arr.Unshift("赵六")
    // 尾部追加
    arr.Push("张三", "李四", "王五")
    // 指定位置插入
    _ = arr.Add(2, "钱七")
    _ = arr.Add(3, "孙八")
    fmt.Printf("数组长度: %d\n", arr.Length())
    fmt.Printf("数组容量: %d\n", arr.CapLength())
    // 获取元素（带默认值）
    if v, err := arr.Get(1); err == nil {
        fmt.Printf("索引1的值: %v\n", v)
    }
    // 映射生成新数组
    mapped := arr.Map(func(i int, v interface{}) interface{} { 
        return fmt.Sprintf("[%d]%v", i, v) 
    })
    fmt.Printf("映射后: %v\n", mapped.Raw())
    
    // 切片工具示例
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    // 随机选择
    randomNum := zarray.Rand(numbers)
    fmt.Printf("随机数: %d\n", randomNum)
    
    // 随机选择多个
    randomNums := zarray.RandPickN(numbers, 3)
    fmt.Printf("随机选择3个: %v\n", randomNums)
    
    // 映射操作
    doubled := zarray.Map(numbers, func(index int, item int) int {
        return item * 2
    })
    fmt.Printf("翻倍后: %v\n", doubled)
    
    // 并行映射
    squared := zarray.Map(numbers, func(index int, item int) int {
        return item * item
    }, 4) // 使用4个工作协程
    fmt.Printf("平方后: %v\n", squared)
    
    // 过滤操作
    evens := zarray.Filter(numbers, func(index int, item int) bool {
        return item%2 == 0
    })
    fmt.Printf("偶数: %v\n", evens)
    
    // 查找操作
    if found, ok := zarray.Find(numbers, func(index int, item int) bool {
        return item > 5
    }); ok {
        fmt.Printf("第一个大于5的数: %d\n", found)
    }
    
    // 去重操作
    duplicates := []int{1, 2, 2, 3, 3, 3, 4, 4, 4, 4}
    unique := zarray.Unique(duplicates)
    fmt.Printf("去重后: %v\n", unique)
    
    // 差集操作
    list1 := []int{1, 2, 3, 4, 5}
    list2 := []int{4, 5, 6, 7, 8}
    onlyIn1, onlyIn2 := zarray.Diff(list1, list2)
    fmt.Printf("只在list1中: %v\n", onlyIn1)
    fmt.Printf("只在list2中: %v\n", onlyIn2)
    
    // 分块操作
    chunks := zarray.Chunk(numbers, 3)
    fmt.Printf("分块结果: %v\n", chunks)
    
    // 随机移位
    randShift := zarray.RandShift(numbers)
    for i := 0; i < 3; i++ {
        if val, err := randShift(); err == nil {
            fmt.Printf("随机移位 %d: %d\n", i+1, val)
        }
    }
    
    // 优先级排序
    priorityFirst := []int{1, 3, 5}
    priorityLast := []int{2, 4, 6}
    sorted := zarray.SortWithPriority(numbers, priorityFirst, priorityLast)
    fmt.Printf("优先级排序: %v\n", sorted)
    
    // 交集操作
    intersection := zarray.Intersection(list1, list2)
    fmt.Printf("交集: %v\n", intersection)
    
    // 排序映射示例
    sortMap := zarray.NewSortMap[string, int]()
    
    // 添加键值对
    sortMap.Set("apple", 5)
    sortMap.Set("banana", 3)
    sortMap.Set("cherry", 7)
    sortMap.Set("date", 1)
    
    fmt.Printf("排序映射长度: %d\n", sortMap.Len())
    
    // 遍历排序映射
    sortMap.ForEach(func(key string, value int) bool {
        fmt.Printf("键: %s, 值: %d\n", key, value)
        return true
    })
    
    // 获取所有键
    keys := sortMap.Keys()
    fmt.Printf("所有键: %v\n", keys)
    
    // 哈希映射示例
    hashMap := zarray.NewHashMap[string, int]()
    
    // 设置值
    hashMap.Set("Alice", 25)
    hashMap.Set("Bob", 30)
    hashMap.Set("Charlie", 35)
    
    // 获取值
    if age, ok := hashMap.Get("Alice"); ok {
        fmt.Printf("Alice的年龄: %d\n", age)
    }
    
    // 获取或设置值
    actualAge, loaded, computed := hashMap.ProvideGet("Eve", func() (int, bool) {
        return 28, true
    })
    fmt.Printf("Eve的年龄: %d, 是否已存在: %t, 是否计算得出: %t\n", actualAge, loaded, computed)
    
    // 比较并交换
    swapped := hashMap.CAS("Bob", 30, 31)
    fmt.Printf("CAS操作: %t\n", swapped)
    
    // 工具函数示例
    userMap := map[string]int{
        "张三": 25,
        "李四": 30,
        "王五": 35,
    }
    
    // 获取所有键
    userKeys := zarray.Keys(userMap)
    fmt.Printf("用户键: %v\n", userKeys)
    
    // 获取所有值
    userValues := zarray.Values(userMap)
    fmt.Printf("用户值: %v\n", userValues)
    
    // 索引映射
    users := []string{"张三", "李四", "王五"}
    indexedUsers, err := zarray.IndexMap(users, func(user string) (string, string) {
        return user, user
    })
    if err == nil {
        fmt.Printf("索引映射: %v\n", indexedUsers)
    }
    
    // 扁平化映射
    flatValues := zarray.FlatMap(userMap, func(key string, value int) int {
        return value
    })
    fmt.Printf("扁平化值: %v\n", flatValues)
    
    // 字符串转换
    str := "Hello, World!"
    bytes := zarray.Str2bytes(str)
    fmt.Printf("字符串转字节: %v\n", bytes)
    
    strBack := zarray.Bytes2str(bytes)
    fmt.Printf("字节转字符串: %s\n", strBack)
    
    // 字符串切片转换
    strSlice := zarray.ToStringSlice([]interface{}{"a", "b", "c"})
    fmt.Printf("字符串切片: %v\n", strSlice)
    
    // 实际应用示例
    // 用户管理系统
    type User struct {
        Name string
        Age  int
    }
    
    usersData := []User{
        {Name: "张三", Age: 25},
        {Name: "李四", Age: 30},
        {Name: "王五", Age: 35},
        {Name: "赵六", Age: 28},
        {Name: "钱七", Age: 32},
    }
    
    // 过滤年轻用户
    youngUsers := zarray.Filter(usersData, func(index int, user User) bool {
        return user.Age < 30
    })
    fmt.Printf("年轻用户: %v\n", youngUsers)
    
    // 提取用户名
    names := zarray.Map(usersData, func(index int, user User) string {
        return user.Name
    })
    fmt.Printf("用户名列表: %v\n", names)
    
    // 按年龄索引
    userIndex, err := zarray.IndexMap(usersData, func(user User) (int, User) {
        return user.Age, user
    })
    if err == nil {
        fmt.Printf("按年龄索引: %v\n", userIndex)
    }
    
    // 缓存系统
    cache := zarray.NewHashMap[string, interface{}]()
    
    // 提供默认值的获取
    user, loaded, computed := cache.ProvideGet("user:3", func() (interface{}, bool) {
        return User{Name: "默认用户", Age: 0}, true
    })
    fmt.Printf("用户: %v, 是否已存在: %t, 是否计算得出: %t\n", user, loaded, computed)
    
    // 年龄统计
    ages := zarray.Map(usersData, func(index int, user User) int {
        return user.Age
    })
    fmt.Printf("年龄列表: %v\n", ages)
    
    fmt.Println("zarray 模块示例完成")
}
```

## 类型定义

### 可哈希类型
```go
type hashable interface {
    comparable
}
```

### 排序映射器
```go
type SortMaper[K hashable, V any] struct {
    // 排序映射结构
}
```

### 哈希映射器
```go
type Maper[K hashable, V any] struct {
    // 哈希映射结构
}
```

## 最佳实践

1. 使用类型安全的操作
2. 合理选择数据结构
3. 实现适当的错误处理
4. 监控操作性能
