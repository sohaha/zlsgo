# zutil 模块

`zutil` 提供了反射工具、原子操作、重试机制、通道管理、缓冲区池、环境变量、参数解析、工具函数、Once 模式、选项模式等功能，用于通用工具和辅助函数。

## 功能概览

- **反射工具**: 反射相关的工具函数
- **原子操作**: 原子值操作和类型
- **重试机制**: 函数重试和延迟
- **通道管理**: 通道创建和操作
- **缓冲区池**: 缓冲区复用管理
- **环境变量**: 环境变量和系统信息
- **参数解析**: 命令行参数解析
- **工具函数**: 通用工具函数
- **Once 模式**: 单次执行保证
- **选项模式**: 灵活的函数配置

## 核心功能

### 反射工具

```go
// 设置反射值，支持类型转换
func SetValue(vTypeOf reflect.Kind, vValueOf reflect.Value, value interface{}) error
// 遍历结构体字段
func ReflectStructField(v reflect.Type, fn func(fieldName, fieldTag string, field reflect.StructField) error) error
// 遍历结构体字段和值
func ReflectForNumField(v reflect.Value, fn func(fieldName, fieldTag string, field reflect.StructField, fieldValue reflect.Value) error) error
// 获取所有方法
func GetAllMethod(s interface{}, fn func(numMethod int, m reflect.Method) error) error
// 运行所有方法
func RunAllMethod(st interface{}, args ...interface{}) error
// 运行指定方法
func RunAssignMethod(st interface{}, filter func(methodName string) bool, args ...interface{}) error
```

### 原子操作

```go
func NewBool(b bool) *Bool
func (b *Bool) Store(val bool)
func (b *Bool) Load() bool
func (b *Bool) Toggle() bool
func (b *Bool) CAS(old, new bool) bool

func NewInt32(i int32) *Int32
func (i32 *Int32) Add(i int32) int32
func (i32 *Int32) Sub(i int32) int32
func (i32 *Int32) Swap(i int32) int32
func (i32 *Int32) Load() int32
func (i32 *Int32) Store(i int32)
func (i32 *Int32) CAS(old, new int32) bool
func (i32 *Int32) String() string

func NewUint32(i uint32) *Uint32
func (u32 *Uint32) Add(i uint32) uint32
func (u32 *Uint32) Sub(i uint32) uint32
func (u32 *Uint32) Swap(i uint32) uint32
func (u32 *Uint32) Load() uint32
func (u32 *Uint32) Store(i uint32)
func (u32 *Uint32) CAS(old, new uint32) bool
func (u32 *Uint32) String() string

func NewUint64(i uint64) *Uint64
func (u64 *Uint64) Add(i uint64) uint64
func (u64 *Uint64) Sub(i uint64) uint64
func (u64 *Uint64) Swap(i uint64) uint64
func (u64 *Uint64) Load() uint64
func (u64 *Uint64) Store(i uint64)
func (u64 *Uint64) CAS(old, new uint64) bool
func (u64 *Uint64) String() string

func NewInt64(i int64) *Int64
func (i64 *Int64) Add(i int64) int64
func (i64 *Int64) Sub(i int64) int64
func (i64 *Int64) Swap(i int64) int64
func (i64 *Int64) Load() int64
func (i64 *Int64) Store(i int64)
func (i64 *Int64) CAS(old, new int64) bool
func (i64 *Int64) String() string

func NewUintptr(i uintptr) *Uintptr
func (uptr *Uintptr) Add(i uintptr) uintptr
func (uptr *Uintptr) Sub(i uintptr) uintptr
func (uptr *Uintptr) Swap(i uintptr) uintptr
func (uptr *Uintptr) Load() uintptr
func (uptr *Uintptr) Store(i uintptr)
func (uptr *Uintptr) CAS(old, new uintptr) bool
func (uptr *Uintptr) String() string

func NewPointer(p unsafe.Pointer) *Pointer
func (ptr *Pointer) Load() unsafe.Pointer
func (ptr *Pointer) Store(p unsafe.Pointer)
func (ptr *Pointer) Swap(new unsafe.Pointer) unsafe.Pointer
func (ptr *Pointer) CAS(old, new unsafe.Pointer) bool
```

### 重试机制

```go
func DoRetry(sum int, fn func() error, opt ...func(*RetryConf)) error
func BackOffDelay(attempt int, retryInterval, maxRetryInterval time.Duration) time.Duration
```

### 缓冲区池

```go
func NewBufferPool(left, right uint) *BufferPool
func GetBuff(size ...uint) *bytes.Buffer
func PutBuff(buffer *bytes.Buffer, noreset ...bool)
```

### 通道管理

```go
// 创建通道，支持无缓冲、缓冲和无界通道
func NewChan[T any](cap ...int) *Chan[T]
// 获取发送通道
func (ch *Chan[T]) In() chan<- T
// 获取接收通道
func (ch *Chan[T]) Out() <-chan T
// 关闭通道
func (ch *Chan[T]) Close()
// 获取通道当前元素数量
func (ch *Chan[T]) Len() int
```

**NewChan 参数说明：**
- `cap == 0`: 创建无缓冲通道（发送会阻塞直到被接收）
- `cap > 0`: 创建指定容量的缓冲通道
- `cap < 0` 或未提供: 创建无界通道（发送永不阻塞）

### 环境变量

```go
func Getenv(name string, def ...string) string
func GOROOT() string
func Loadenv(filenames ...string) error
func GetOs() string
func IsWin() bool
func IsMac() bool
func IsLinux() bool
func Is32BitArch() bool
```

### 参数解析

```go
func NewArgs(opt ...ArgsOpt) *Args
func (args *Args) Var(arg interface{}) string
func (args *Args) CompileString(format string, initialValue ...interface{}) string
func (args *Args) Compile(format string, initialValue ...interface{}) (query string, values []interface{})
```

### Once 模式

```go
func Once[T any](fn func() T) func() T
func OnceWithError[T any](fn func() (T, error)) func() (T, error)
func Guard[T any](fn func() T) func() (T, error)
// Go 1.17 及以下版本
func Once(fn func() interface{}) func() interface{}
```

### 选项模式

```go
func Optional[T interface{}](o T, fn ...func(*T)) T
```

### 条件值

```go
func IfVal[T interface{}](condition bool, trueVal, falseVal T) T
```

### 系统与运行时工具

```go
func Named(name string, arg interface{}) interface{}
func WithRunContext(handler func()) (time.Duration, uint64)
func TryCatch(fn func() error) error
func Try(fn func(), catch func(e interface{}), finally ...func())
func CheckErr(err error, exit ...bool)
func Callers(skip ...int) Stack
func (s Stack) Format(f func(fn *runtime.Func, file string, line int) bool)
func IsDoubleClickStartUp() bool
func GetParentProcessName() (string, error)
func MaxRlimit() (int, error)
func GetGid() uint64
func UnescapeHTML(s string) string
```

## 使用示例

```go
package main

import (
    "fmt"
    "reflect"
    "time"
    "github.com/sohaha/zlsgo/zutil"
)

func main() {
    // 反射工具示例
    type Person struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
    
    person := Person{Name: "张三", Age: 25}
    
    // 遍历结构体字段
    err := zutil.ReflectForNumField(reflect.ValueOf(person), func(fieldName, fieldTag string, field reflect.StructField, fieldValue reflect.Value) error {
        fmt.Printf("字段: %s, 标签: %s, 值: %v\n", fieldName, fieldTag, fieldValue.Interface())
        return nil
    })
    
    if err != nil {
        fmt.Printf("反射遍历失败: %v\n", err)
    }
    
    // 获取所有方法
    err = zutil.GetAllMethod(&person, func(numMethod int, m reflect.Method) error {
        fmt.Printf("方法 %d: %s\n", numMethod, m.Name)
        return nil
    })
    
    if err != nil {
        fmt.Printf("获取方法失败: %v\n", err)
    }
    
    // 原子操作示例
    // 布尔值原子操作
    atomicBool := zutil.NewBool(true)
    
    // 存储值
    atomicBool.Store(false)
    fmt.Printf("存储布尔值: false\n")
    
    // 加载值
    currentVal := atomicBool.Load()
    fmt.Printf("当前布尔值: %t\n", currentVal)
    
    // 切换值
    oldVal = atomicBool.Toggle()
    fmt.Printf("切换布尔值，旧值: %t\n", oldVal)
    
    // 比较并交换
    swapped := atomicBool.CAS(false, true)
    fmt.Printf("CAS 操作: %t\n", swapped)
    
    // 整数原子操作
    atomicInt := zutil.NewInt32(100)
    
    // 加法操作
    newVal := atomicInt.Add(50)
    fmt.Printf("整数加法: %d\n", newVal)
    
    // 减法操作
    newVal = atomicInt.Sub(25)
    fmt.Printf("整数减法: %d\n", newVal)
    
    // 交换值
    oldVal32 := atomicInt.Swap(200)
    fmt.Printf("整数交换，旧值: %d\n", oldVal32)
    
    // 加载值
    currentInt := atomicInt.Load()
    fmt.Printf("当前整数值: %d\n", currentInt)
    
    // 比较并交换
    swapped = atomicInt.CAS(200, 300)
    fmt.Printf("整数 CAS 操作: %t\n", swapped)
    
    // 重试机制示例
    var attemptCount int
    
    err = zutil.DoRetry(3, func() error {
        attemptCount++
        fmt.Printf("尝试第 %d 次\n", attemptCount)
        
        // 模拟失败
        if attemptCount < 3 {
            return fmt.Errorf("模拟失败")
        }
        
        return nil
    })
    
    if err != nil {
        fmt.Printf("重试失败: %v\n", err)
    } else {
        fmt.Println("重试成功")
    }
    
    // 计算退避延迟
    delay := zutil.BackOffDelay(2, time.Second, time.Minute)
    fmt.Printf("第2次重试延迟: %v\n", delay)
    
    // 参数解析示例
    args := zutil.NewArgs()
    
    // 添加参数
    args.Var("张三")
    args.Var(25)
    args.Var(true)
    
    // 编译字符串
    result := args.CompileString("姓名: {}, 年龄: {}, 激活: {}")
    fmt.Printf("编译结果: %s\n", result)
    
    // 编译查询
    query, values := args.Compile("SELECT * FROM users WHERE name = ? AND age = ?")
    fmt.Printf("查询: %s, 参数: %v\n", query, values)
    
    // 工具函数示例
    // 命名参数
    namedArg := zutil.Named("age", 25)
    fmt.Printf("命名参数: %v\n", namedArg)
    
    // 性能测量
    duration, memory := zutil.WithRunContext(func() {
        // 模拟耗时操作
        time.Sleep(100 * time.Millisecond)
    })
    
    fmt.Printf("执行耗时: %v, 内存分配: %d\n", duration, memory)
    
    // 异常捕获
    err = zutil.TryCatch(func() error {
        // 模拟可能出错的函数
        if time.Now().Second()%2 == 0 {
            return fmt.Errorf("随机错误")
        }
        return nil
    })
    
    if err != nil {
        fmt.Printf("捕获到错误: %v\n", err)
    }
    
    // Try-Catch-Finally
    zutil.Try(
        func() {
            fmt.Println("执行主要逻辑")
            panic("模拟 panic")
        },
        func(e interface{}) {
            fmt.Printf("捕获异常: %v\n", e)
        },
        func() {
            fmt.Println("执行清理逻辑")
        },
    )
    
    // 错误检查
    testErr := fmt.Errorf("测试错误")
    zutil.CheckErr(testErr, false) // 不退出程序
    
    // 获取调用栈
    stack := zutil.Callers(1)
    stack.Format(func(fn *runtime.Func, file string, line int) bool {
        fmt.Printf("调用: %s, 文件: %s, 行: %d\n", fn.Name(), file, line)
        return true // 继续遍历
    })
    
    // Once 模式示例
    // 泛型版本
    onceFunc := zutil.Once[string](func() string {
        fmt.Println("执行一次初始化")
        return "初始化完成"
    })
    
    // 多次调用，只执行一次
    result1 := onceFunc()
    result2 := onceFunc()
    result3 := onceFunc()
    
    fmt.Printf("结果1: %s\n", result1)
    fmt.Printf("结果2: %s\n", result2)
    fmt.Printf("结果3: %s\n", result3)
    
    // 带错误的 Once
    onceWithError := zutil.OnceWithError[int](func() (int, error) {
        fmt.Println("执行一次计算")
        return 42, nil
    })
    
    val1, err1 := onceWithError()
    val2, err2 := onceWithError()
    
    fmt.Printf("值1: %d, 错误1: %v\n", val1, err1)
    fmt.Printf("值2: %d, 错误2: %v\n", val2, err2)
    
    // Guard 模式
    guardFunc := zutil.Guard[string](func() string {
        return "受保护的操作"
    })
    
    val, err := guardFunc()
    if err != nil {
        fmt.Printf("Guard 错误: %v\n", err)
    } else {
        fmt.Printf("Guard 结果: %s\n", val)
    }
    
    // 条件值示例
    condition := true
    trueValue := "真值"
    falseValue := "假值"
    
    result = zutil.IfVal(condition, trueValue, falseValue)
    fmt.Printf("条件值: %s\n", result)
    
    // 系统工具示例
    // 检查是否双击启动
    isDoubleClick := zutil.IsDoubleClickStartUp()
    fmt.Printf("是否双击启动: %t\n", isDoubleClick)
    
    // 获取父进程名称
    parentProcess, err := zutil.GetParentProcessName()
    if err == nil {
        fmt.Printf("父进程名称: %s\n", parentProcess)
    }
    
    // 获取最大资源限制
    maxRlimit, err := zutil.MaxRlimit()
    if err == nil {
        fmt.Printf("最大资源限制: %d\n", maxRlimit)
    }
    
    // 环境变量示例
    os := zutil.GetOs()
    fmt.Printf("当前操作系统: %s\n", os)
    
    if zutil.IsWin() {
        fmt.Println("运行在 Windows 系统上")
    } else if zutil.IsMac() {
        fmt.Println("运行在 macOS 系统上")
    } else if zutil.IsLinux() {
        fmt.Println("运行在 Linux 系统上")
    }
    
    // 通道管理示例
    ch := zutil.NewChan[int](10)  // 创建容量为10的缓冲通道
    
    // 发送数据
    go func() {
        ch.In() <- 42  // 通过 In() 获取发送通道
        ch.Close()
    }()
    
    // 接收数据
    select {
    case val := <-ch.Out():  // 通过 Out() 获取接收通道
        fmt.Printf("接收到值: %d\n", val)
    default:
        fmt.Println("无数据可接收")
    }
    
    // 不同类型通道示例
    // 1. 无缓冲通道
    unbuffered := zutil.NewChan[string](0)
    go func() {
        unbuffered.In() <- "同步消息"
        unbuffered.Close()
    }()
    
    // 2. 无界通道 (永不阻塞)
    unbounded := zutil.NewChan[int]() // 默认为无界通道
    for i := 0; i < 1000; i++ {
        unbounded.In() <- i  // 永远不会阻塞
    }
    fmt.Printf("无界通道长度: %d\n", unbounded.Len())
    unbounded.Close()
    
    // 缓冲区池示例
    pool := zutil.NewBufferPool(1024, 4096)
    buffer := pool.Get()
    buffer.WriteString("Hello, World!")
    fmt.Printf("缓冲区内容: %s\n", buffer.String())
    pool.Put(buffer)
    
    // 实际应用示例
    // 配置管理
    type Config struct {
        MaxRetries int
        Timeout    time.Duration
        Debug      bool
    }
    
    var config Config
    
    // 使用 Once 确保配置只加载一次
    loadConfig := zutil.Once[Config](func() Config {
        fmt.Println("加载配置...")
        return Config{
            MaxRetries: 3,
            Timeout:    time.Second * 30,
            Debug:      true,
        }
    })
    
    // 多次获取配置，只加载一次
    for i := 0; i < 3; i++ {
        cfg := loadConfig()
        fmt.Printf("配置 %d: %+v\n", i+1, cfg)
    }
    
    // 原子计数器
    counter := zutil.NewInt64(0)
    
    // 并发增加计数
    for i := 0; i < 10; i++ {
        go func() {
            counter.Add(1)
        }()
    }
    
    // 等待所有 goroutine 完成
    time.Sleep(100 * time.Millisecond)
    
    finalCount := counter.Load()
    fmt.Printf("最终计数: %d\n", finalCount)
    
    // 重试机制应用
    var successCount int
    
    err = zutil.DoRetry(5, func() error {
        successCount++
        fmt.Printf("尝试操作 %d\n", successCount)
        
        // 模拟成功率 30%
        if successCount < 4 {
            return fmt.Errorf("操作失败，重试中...")
        }
        
        return nil
    })
    
    if err != nil {
        fmt.Printf("所有重试都失败了: %v\n", err)
    } else {
        fmt.Printf("操作成功，尝试了 %d 次\n", successCount)
    }
    
    fmt.Println("工具库示例完成")
}