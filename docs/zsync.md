# zsync 模块

`zsync` 提供了扩展了标准库的 `sync` 包，提供了额外的功能和优化，用于常见的并发模式。

## 功能概览

- **WaitGroup**: 扩展的等待组，支持错误处理和并发限制
- **RBMutex**: 读偏向的读写互斥锁，针对读密集型工作负载优化
- **Promise**: Go 语言实现的 Promise 模式，用于异步操作
- **Context 工具**: 用于处理和合并多个上下文的工具
- **原子操作**: 类型安全的原子值操作
- **对象池**: 高效的对象复用管理

## 核心功能

### WaitGroup

```go
func NewWaitGroup(max ...uint) *WaitGroup
func (wg *WaitGroup) Add(delta int)
func (wg *WaitGroup) Done()
func (wg *WaitGroup) Go(f func())
func (wg *WaitGroup) Wait() error
```

### RBMutex

```go
func NewRBMutex() *RBMutex
func (m *RBMutex) Lock()
func (m *RBMutex) Unlock()
func (m *RBMutex) RLock() RBToken
func (m *RBMutex) RUnlock(token RBToken)
```

说明：RBMutex 为读偏向锁。在读多写少场景下，读锁开销更低；在写入后短时间内会根据“写入动量”抑制读偏向以避免写方饥饿。非 64 位架构上自动回退为标准 RWMutex 实现，API 保持一致。

示例：

```go
mu := zsync.NewRBMutex()

// 写
mu.Lock()
shared = 42
mu.Unlock()

// 读
tok := mu.RLock()
_ = shared
mu.RUnlock(tok)
```

### SeqLock（泛型）

提供基于序列号的无锁读方案，读路径在无写竞争时无需加锁，适合“多读少写”的只读快照场景。

```go
type SeqLockT[T any] struct{}

func NewSeqLock[T any]() *SeqLockT[T]
func (s *SeqLockT[T]) Write(v T)
func (s *SeqLockT[T]) Read() (T, bool)
```

示例：

```go
s := zsync.NewSeqLock[*Data]()
s.Write(&Data{A: 1})
if v, ok := s.Read(); ok {
    _ = v.A
}
```

### Promise

```go
func NewPromise[T any](fn func() (T, error)) *Promise[T]
func (p *Promise[T]) Then(fn func(T) (T, error)) *Promise[T]
func (p *Promise[T]) Catch(fn func(error) (T, error)) *Promise[T]
func (p *Promise[T]) Done() (T, error)
func (p *Promise[T]) State() PromiseState
```

### Promise 聚合操作

```go
// 等待所有 Promise 完成
func PromiseAll[T any](promises ...*Promise[T]) *Promise[[]T]
// 竞争 Promise，返回最快完成的结果
func PromiseRace[T any](promises ...*Promise[T]) *Promise[T]
// 任意 Promise 完成即可
func PromiseAny[T any](promises ...*Promise[T]) *Promise[T]
// 带上下文的 Promise 聚合
func PromiseAllContext[T any](ctx context.Context, promises ...*Promise[T]) *Promise[[]T]
// 带上下文的 Promise 竞争
func PromiseRaceContext[T any](ctx context.Context, promises ...*Promise[T]) *Promise[T]
// 带上下文的 Promise 任意
func PromiseAnyContext[T any](ctx context.Context, promises ...*Promise[T]) *Promise[T]
```

### 原子操作

```go
func NewValue[T any](initial T) *AtomicValue[T]
func (v *AtomicValue[T]) Load() T
func (v *AtomicValue[T]) Store(value T)
func (v *AtomicValue[T]) Swap(new T) T
func (v *AtomicValue[T]) CAS(old, new T) bool
```

### 对象池

```go
func NewPool[T any](n func() T) *Pool[T]
func (p *Pool[T]) Get() T
func (p *Pool[T]) Put(x T)
```

### 上下文合并

```go
func MergeContext(ctxs ...context.Context) Context
```

### Promise 状态

```go
type PromiseState uint8

const (
    PromiseStatePending PromiseState = iota
    PromiseStateFulfilled
    PromiseStateRejected
)
```

### 错误处理

```go
var (
    ErrWaitGroupClosed = errors.New("wait group is closed")
    ErrMaxConcurrency  = errors.New("max concurrency exceeded")
)
```

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zsync"
)

func main() {
    // 原子操作示例
    atomicValue := zsync.NewValue[int](100)
    
    // 加载值
    value := atomicValue.Load()
    fmt.Printf("当前值: %d\n", value)
    
    // 存储值
    atomicValue.Store(200)
    fmt.Printf("存储后值: %d\n", atomicValue.Load())
    
    // 交换值
    oldValue := atomicValue.Swap(300)
    fmt.Printf("旧值: %d, 新值: %d\n", oldValue, atomicValue.Load())
    
    // 比较并交换
    swapped := atomicValue.CAS(300, 400)
    if swapped {
        fmt.Printf("比较并交换成功: %d\n", atomicValue.Load())
    }
    
    // 互斥锁示例
    rwMutex := zsync.NewRBMutex()
    
    // 写锁
    rwMutex.Lock()
    fmt.Println("获取写锁")
    time.Sleep(100 * time.Millisecond)
    rwMutex.Unlock()
    fmt.Println("释放写锁")
    
    // 读锁
    token := rwMutex.RLock()
    fmt.Println("获取读锁")
    time.Sleep(50 * time.Millisecond)
    rwMutex.RUnlock(token)
    fmt.Println("释放读锁")
    
    // 等待组示例
    wg := zsync.NewWaitGroup(10) // 最大并发数 10
    
    // 添加任务
    for i := 0; i < 5; i++ {
        wg.Add(1)
        wg.Go(func() {
            defer wg.Done()
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("任务 %d 完成\n", i)
        })
    }
    
    // 等待所有任务完成
    err := wg.Wait()
    if err == nil {
        fmt.Println("所有任务完成")
    }
    
    // Promise 示例
    promise := zsync.NewPromise[string](func() (string, error) {
        time.Sleep(100 * time.Millisecond)
        return "异步操作完成", nil
    })
    
    // 链式调用
    promise2 := promise.Then(func(result string) (string, error) {
        return result + " 并已处理", nil
    })
    
    // 等待结果
    result, err := promise2.Done()
    if err != nil {
        fmt.Printf("Promise 执行失败: %v\n", err)
    } else {
        fmt.Printf("Promise 结果: %s\n", result)
    }
    
    // Promise 聚合示例
    promises := make([]*zsync.Promise[int], 3)
    
    for i := 0; i < 3; i++ {
        idx := i
        promises[i] = zsync.NewPromise[int](func() (int, error) {
            time.Sleep(time.Duration(idx+1) * 100 * time.Millisecond)
            return idx * 10, nil
        })
    }
    
    // 等待所有 Promise 完成
    allPromise := zsync.PromiseAll(promises...)
    allResults, err := allPromise.Done()
    if err != nil {
        fmt.Printf("聚合失败: %v\n", err)
    } else {
        fmt.Printf("所有结果: %v\n", allResults)
    }
    
    // 竞争 Promise
    racePromise := zsync.PromiseRace(promises...)
    raceResult, err := racePromise.Done()
    if err != nil {
        fmt.Printf("竞争失败: %v\n", err)
    } else {
        fmt.Printf("最快结果: %d\n", raceResult)
    }
    
    // 任意 Promise
    anyPromise := zsync.PromiseAny(promises...)
    anyResult, err := anyPromise.Done()
    if err != nil {
        fmt.Printf("任意 Promise 失败: %v\n", err)
    } else {
        fmt.Printf("任意结果: %d\n", anyResult)
    }
    
    // 带上下文的 Promise
    ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
    defer cancel()
    
    ctxPromise := zsync.PromiseAllContext(ctx, promises...)
    ctxResults, err := ctxPromise.Done()
    if err != nil {
        fmt.Printf("带上下文的 Promise 失败: %v\n", err)
    } else {
        fmt.Printf("上下文结果: %v\n", ctxResults)
    }
    
    // 对象池示例
    pool := zsync.NewPool[string](func() string {
        return "新对象"
    })
    
    // 获取对象
    obj1 := pool.Get()
    fmt.Printf("获取对象: %s\n", obj1)
    
    // 放回对象
    pool.Put(obj1)
    
    // 再次获取
    obj2 := pool.Get()
    fmt.Printf("再次获取对象: %s\n", obj2)
    
    // 上下文合并示例
    ctx1, cancel1 := context.WithCancel(context.Background())
    defer cancel1()
    
    ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel2()
    
    mergedCtx := zsync.MergeContext(ctx1, ctx2)
    
    select {
    case <-mergedCtx.Done():
        fmt.Println("合并上下文已取消")
    case <-time.After(500 * time.Millisecond):
        fmt.Println("500ms 后超时")
    }
    
    // 实际应用示例
    // 并发下载器
    urls := []string{
        "https://example.com/1",
        "https://example.com/2",
        "https://example.com/3",
    }
    
    downloadPromises := make([]*zsync.Promise[string], len(urls))
    
    for i, url := range urls {
        idx := i
        downloadPromises[i] = zsync.NewPromise[string](func() (string, error) {
            // 模拟下载
            time.Sleep(time.Duration(idx+1) * 100 * time.Millisecond)
            return fmt.Sprintf("下载完成: %s", url), nil
        })
    }
    
    // 等待所有下载完成
    downloadResults, err := zsync.PromiseAll(downloadPromises...).Done()
    if err != nil {
        fmt.Printf("下载失败: %v\n", err)
    } else {
        fmt.Printf("所有下载完成: %v\n", downloadResults)
    }
    
    // 并发处理器
    processor := zsync.NewWaitGroup(5) // 最多5个并发处理器
    
    for i := 0; i < 20; i++ {
        processor.Add(1)
        processor.Go(func() {
            defer processor.Done()
            // 模拟处理
            time.Sleep(50 * time.Millisecond)
            fmt.Printf("处理任务 %d\n", i)
        })
    }
    
    err = processor.Wait()
    if err != nil {
        fmt.Printf("处理失败: %v\n", err)
    } else {
        fmt.Println("所有任务处理完成")
    }
    
    fmt.Println("zsync 模块示例完成")
    
    // 高级用法示例
    fmt.Println("\n=== 高级用法示例 ===")
    
    // 带错误处理的 WaitGroup
    errorWg := zsync.NewWaitGroup(3)
    
    for i := 0; i < 5; i++ {
        idx := i
        errorWg.Add(1)
        errorWg.Go(func() {
            defer errorWg.Done()
            
            // 模拟可能出错的任务
            if idx == 2 {
                panic("任务2出错")
            }
            
            time.Sleep(50 * time.Millisecond)
            fmt.Printf("任务 %d 完成\n", idx)
        })
    }
    
    // 等待并处理错误
    err = errorWg.Wait()
    if err != nil {
        fmt.Printf("等待组执行出错: %v\n", err)
    }
    
    // 带超时的 Promise
    timeoutPromise := zsync.NewPromise[string](func() (string, error) {
        time.Sleep(2 * time.Second)
        return "长时间操作完成", nil
    })
    
    // 设置超时
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    select {
    case result := <-func() chan string {
        ch := make(chan string, 1)
        go func() {
            if result, err := timeoutPromise.Done(); err == nil {
                ch <- result
            }
            close(ch)
        }()
        return ch
    }():
        fmt.Printf("Promise 结果: %s\n", result)
    case <-ctx.Done():
        fmt.Println("Promise 超时")
    }
    
    // 读写锁性能测试
    fmt.Println("\n=== 读写锁性能测试 ===")
    
    testMutex := zsync.NewRBMutex()
    var counter int
    
    // 启动多个读协程
    for i := 0; i < 10; i++ {
        go func() {
            for j := 0; j < 100; j++ {
                token := testMutex.RLock()
                _ = counter // 读取计数器
                testMutex.RUnlock(token)
                time.Sleep(time.Microsecond)
            }
        }()
    }
    
    // 启动写协程
    go func() {
        for i := 0; i < 10; i++ {
            testMutex.Lock()
            counter++
            testMutex.Unlock()
            time.Sleep(time.Millisecond)
        }
    }()
    
    time.Sleep(2 * time.Second)
    fmt.Printf("最终计数器值: %d\n", counter)
    
    // 对象池性能优化
    fmt.Println("\n=== 对象池性能优化 ===")
    
    // 创建字符串缓冲区池
    bufferPool := zsync.NewPool[[]byte](func() []byte {
        return make([]byte, 0, 1024)
    })
    
    // 并发使用对象池
    poolWg := zsync.NewWaitGroup(20)
    
    for i := 0; i < 100; i++ {
        poolWg.Add(1)
        go func() {
            defer poolWg.Done()
            
            // 获取缓冲区
            buffer := bufferPool.Get()
            
            // 使用缓冲区
            buffer = append(buffer, "数据"...)
            
            // 放回缓冲区
            poolWg.Put(buffer)
        }()
    }
    
    poolWg.Wait()
    fmt.Println("对象池测试完成")
    
    // 上下文合并的高级用法
    fmt.Println("\n=== 上下文合并高级用法 ===")
    
    // 创建多个上下文
    ctx1, cancel1 := context.WithCancel(context.Background())
    ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
    ctx3, cancel3 := context.WithDeadline(context.Background(), time.Now().Add(300*time.Millisecond))
    
    defer func() {
        cancel1()
        cancel2()
        cancel3()
    }()
    
    // 合并上下文
    mergedCtx := zsync.MergeContext(ctx1, ctx2, ctx3)
    
    // 监听合并后的上下文
    go func() {
        select {
        case <-mergedCtx.Done():
            fmt.Printf("合并上下文已取消: %v\n", mergedCtx.Err())
        }
    }()
    
    // 模拟不同的取消情况
    go func() {
        time.Sleep(200 * time.Millisecond)
        cancel1() // 第一个上下文取消
    }()
    
    go func() {
        time.Sleep(400 * time.Millisecond)
        cancel2() // 第二个上下文超时
    }()
    
    time.Sleep(600 * time.Millisecond)
    fmt.Println("上下文合并测试完成")
}
```

## 同步模式说明

### 原子操作模式
```go
// 原子值操作
value := zsync.NewValue[int](100)
current := value.Load()
value.Store(200)
old := value.Swap(300)
swapped := value.CAS(300, 400)
```

### 读写锁模式
```go
// 读写锁使用
mutex := zsync.NewRBMutex()
mutex.Lock()           // 写锁
mutex.Unlock()         // 释放写锁
token := mutex.RLock() // 读锁
mutex.RUnlock(token)   // 释放读锁
```

### 序列锁模式
```go
// 无锁读快照（多读少写）
type Data struct{ A int }
s := zsync.NewSeqLock[Data]()
s.Write(Data{A: 1})
v, ok := s.Read()
if ok { _ = v.A }
```

### Promise 模式
```go
// Promise 链式调用
promise := zsync.NewPromise[string](func() (string, error) {
    return "结果", nil
})

promise.Then(func(result string) (string, error) {
    return result + " - 处理", nil
}).Catch(func(err error) (string, error) {
    return "", err
}).Finally(func() {
    // 清理工作
})
```

## 最佳实践

1. 优先使用原子操作
2. 合理选择锁类型
3. 合理设置等待组大小
4. 避免 Promise 链过长