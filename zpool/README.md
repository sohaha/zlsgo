# zpool 模块

`zpool` 提供了工作池、负载均衡器、资源注入等功能，用于高效管理并发资源和任务。

## 功能概览

- **工作池**: 并发任务执行和资源管理
- **负载均衡器**: 多种负载均衡策略
- **资源注入**: 依赖注入和资源管理

## 核心功能

### 工作池

```go
func New(size int, max ...int) *WorkPool
func (wp *WorkPool) Do(fn Task) error
func (wp *WorkPool) DoWithTimeout(fn Task, t time.Duration) error
func (wp *WorkPool) PanicFunc(handler PanicFunc)
func (wp *WorkPool) IsClosed() bool
func (wp *WorkPool) Close()
func (wp *WorkPool) Wait()
func (wp *WorkPool) Pause()
func (wp *WorkPool) Continue(workerNum ...int)
func (wp *WorkPool) Cap() uint
func (wp *WorkPool) AdjustSize(workSize int)
func (wp *WorkPool) PreInit() error
```

### 负载均衡器

```go
func NewBalancer[T any]() *Balancer[T]
func (b *Balancer[T]) Get(key string) (node T, available bool, exists bool)
func (b *Balancer[T]) Add(key string, node T, opt ...func(opts *BalancerNodeOptions)) error
func (b *Balancer[T]) Remove(key string)
func (b *Balancer[T]) Mark(key string, available bool)
func (b *Balancer[T]) Run(fn func(node T) (normal bool, err error), strategy ...BalancerStrategy) error
func (b *Balancer[T]) RunByKeys(keys []string, fn func(node T) (normal bool, err error), strategy ...BalancerStrategy) error
func (b *Balancer[T]) WalkNodes(fn func(node T, available bool) (normal bool))
func (b *Balancer[T]) Keys() []string
func (b *Balancer[T]) Len() int

// 权重管理
func (b *Balancer[T]) GetWeight(key string) (uint64, bool)
func (b *Balancer[T]) SetWeight(key string, weight uint64) error
func (b *Balancer[T]) GetNodeInfo(key string) (BalancerNodeInfo[T], bool)
```

### 资源注入

```go
func (wp *WorkPool) Injector() zdi.TypeMapper
```

## 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zpool"
)

func main() {
    // 工作池示例
    pool := zpool.New(10)
    
    // 设置 panic 处理函数
    pool.PanicFunc(func(err error) {
        fmt.Printf("任务执行出错: %v\n", err)
    })
    
    // 提交任务
    err := pool.Do(func() {
        fmt.Println("执行任务")
    })
    if err != nil {
        fmt.Printf("提交任务失败: %v\n", err)
    }
    
    // 提交带超时的任务
    err = pool.DoWithTimeout(func() {
        time.Sleep(100 * time.Millisecond)
        fmt.Println("超时任务完成")
    }, 50*time.Millisecond)
    if err != nil {
        fmt.Printf("超时任务失败: %v\n", err)
    }
    
    // 获取工作池信息
    fmt.Printf("工作池容量: %d\n", pool.Cap())
    fmt.Printf("工作池是否关闭: %t\n", pool.IsClosed())
    
    // 调整工作池大小
    pool.AdjustSize(15)
    fmt.Printf("调整后容量: %d\n", pool.Cap())
    
    // 暂停和继续
    pool.Pause()
    fmt.Println("工作池已暂停")
    
    pool.Continue(20)
    fmt.Println("工作池已继续，工作线程数: 20")
    
    // 等待所有任务完成
    pool.Wait()
    
    // 关闭工作池
    pool.Close()
    
    // 负载均衡器示例
    balancer := zpool.NewBalancer[string]()
    
    // 添加节点
    err = balancer.Add("node1", "服务器1", func(opts *zpool.BalancerNodeOptions) {
        opts.Weight = 1
        opts.MaxConns = 10
    })
    if err != nil {
        fmt.Printf("添加节点失败: %v\n", err)
    }
    
    err = balancer.Add("node2", "服务器2", func(opts *zpool.BalancerNodeOptions) {
        opts.Weight = 2
        opts.MaxConns = 20
    })
    if err != nil {
        fmt.Printf("添加节点失败: %v\n", err)
    }
    
    // 获取节点
    if node, available, exists := balancer.Get("node1"); exists {
        fmt.Printf("节点1: %s, 可用: %t\n", node, available)
    }
    
    // 标记节点状态
    balancer.Mark("node1", false)
    fmt.Println("节点1已标记为不可用")
    
    // 运行任务
    err = balancer.Run(func(node string) (normal bool, err error) {
        fmt.Printf("在节点 %s 上运行任务\n", node)
        return true, nil
    })
    if err != nil {
        fmt.Printf("运行任务失败: %v\n", err)
    }
    
    // 按指定键运行任务
    keys := []string{"node1", "node2"}
    err = balancer.RunByKeys(keys, func(node string) (normal bool, err error) {
        fmt.Printf("在指定节点 %s 上运行任务\n", node)
        return true, nil
    })
    if err != nil {
        fmt.Printf("按键运行任务失败: %v\n", err)
    }
    
    // 遍历所有节点
    balancer.WalkNodes(func(node string, available bool) bool {
        fmt.Printf("节点: %s, 可用: %t\n", node, available)
        return true // 继续遍历
    })
    
    // 获取节点信息
    allKeys := balancer.Keys()
    fmt.Printf("所有节点: %v\n", allKeys)
    
    nodeCount := balancer.Len()
    fmt.Printf("节点数量: %d\n", nodeCount)
    
    // 移除节点
    balancer.Remove("node1")
    fmt.Println("节点1已移除")

    // 权重管理示例
    // 获取节点权重
    if weight, exists := balancer.GetWeight("node2"); exists {
        fmt.Printf("node2 当前权重: %d\n", weight)
    }

    // 动态调整权重
    err = balancer.SetWeight("node2", 8)
    if err != nil {
        fmt.Printf("调整权重失败: %v\n", err)
    } else {
        fmt.Println("node2 权重已调整为 8")
    }

    // 错误处理示例
    err = balancer.SetWeight("node2", 0) // 无效权重
    if err != nil {
        fmt.Printf("无效权重错误: %v\n", err)
    }

    // 获取完整节点信息
    if info, exists := balancer.GetNodeInfo("node2"); exists {
        fmt.Printf("节点信息: 节点=%s, 权重=%d, 最大连接=%d, 可用=%t, 活跃连接=%d\n",
            info.Node, info.Weight, info.MaxConns, info.Available, info.Active)
    }
    
    // 实际应用示例
    // 数据库连接池
    dbPool := zpool.New(5, 20)
    
    // 设置 panic 处理
    dbPool.PanicFunc(func(err error) {
        fmt.Printf("数据库操作出错: %v\n", err)
    })
    
    // 模拟数据库操作
    for i := 0; i < 10; i++ {
        dbPool.Do(func() {
            fmt.Printf("执行数据库查询 %d\n", i)
            time.Sleep(10 * time.Millisecond)
        })
    }
    
    // 等待所有操作完成
    dbPool.Wait()
    
    // 关闭连接池
    dbPool.Close()
    
    // 服务负载均衡
    serviceBalancer := zpool.NewBalancer[string]()
    
    // 添加服务实例
    services := []string{"service1", "service2", "service3"}
    for i, service := range services {
        weight := i + 1
        serviceBalancer.Add(service, service, func(opts *zpool.BalancerNodeOptions) {
            opts.Weight = uint64(weight)
            opts.MaxConns = 50
        })
    }
    
    // 模拟服务调用
    for i := 0; i < 5; i++ {
        serviceBalancer.Run(func(service string) (normal bool, err error) {
            fmt.Printf("调用服务: %s\n", service)
            return true, nil
        })
    }
    
    // 获取服务统计
    fmt.Printf("服务数量: %d\n", serviceBalancer.Len())
    fmt.Printf("服务列表: %v\n", serviceBalancer.Keys())
    
    // 高级负载均衡示例
    // 使用不同策略的负载均衡器
    advancedBalancer := zpool.NewBalancer[string]()
    
    // 添加不同权重的节点
    advancedBalancer.Add("primary", "主服务器", func(opts *zpool.BalancerNodeOptions) {
        opts.Weight = 3
        opts.MaxConns = 100
        opts.Cooldown = 5000 // 5秒冷却期
    })
    
    advancedBalancer.Add("secondary", "备用服务器", func(opts *zpool.BalancerNodeOptions) {
        opts.Weight = 2
        opts.MaxConns = 50
        opts.Cooldown = 3000 // 3秒冷却期
    })
    
    advancedBalancer.Add("fallback", "故障转移服务器", func(opts *zpool.BalancerNodeOptions) {
        opts.Weight = 1
        opts.MaxConns = 25
        opts.Cooldown = 1000 // 1秒冷却期
    })
    
    // 测试不同策略
    strategies := []zpool.BalancerStrategy{
        zpool.StrategyRandom,
        zpool.StrategyLeastConn,
        zpool.StrategyRoundRobin,
        zpool.StrategyWeighted,
    }
    
    for _, strategy := range strategies {
        fmt.Printf("\n=== 使用策略: %v ===\n", strategy)
        
        // 运行任务测试策略
        err = advancedBalancer.Run(func(node string) (normal bool, err error) {
            fmt.Printf("策略 %v 选择节点: %s\n", strategy, node)
            return true, nil
        }, strategy)
        
        if err != nil {
            fmt.Printf("策略 %v 执行失败: %v\n", strategy, err)
        }
    }
    
    // 节点健康检查示例
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            advancedBalancer.WalkNodes(func(node string, available bool) bool {
                if !available {
                    fmt.Printf("节点 %s 不可用，尝试恢复...\n", node)
                    // 模拟健康检查
                    if time.Now().Unix()%60 == 0 { // 每分钟恢复一个节点
                        advancedBalancer.Mark(node, true)
                        fmt.Printf("节点 %s 已恢复\n", node)
                    }
                }
                return true
            })
        }
    }()
    
    // 依赖注入示例
    // 工作池支持依赖注入
    injectorPool := zpool.New(5)
    
    // 注册依赖
    injectorPool.Injector().Map("数据库连接")
    injectorPool.Injector().Map(time.Now())
    
    // 使用依赖注入的任务
    for i := 0; i < 3; i++ {
        injectorPool.Do(func(dbConn string, startTime time.Time) {
            fmt.Printf("任务 %d 使用依赖: %s, 开始时间: %v\n", i, dbConn, startTime)
        })
    }
    
    // 等待任务完成
    injectorPool.Wait()
    
    // 关闭注入器工作池
    injectorPool.Close()
}
```

## 负载均衡策略说明

### 支持的策略
- **StrategyRandom (随机策略)**: 随机选择可用节点
- **StrategyLeastConn (最少连接策略)**: 选择当前连接数最少的节点
- **StrategyRoundRobin (轮询策略)**: 按顺序循环选择节点
- **StrategyWeighted (权重策略)**: 根据节点权重进行选择

### 策略选择建议
- **简单场景**: 使用轮询策略 (StrategyRoundRobin)
- **性能敏感**: 使用权重策略 (StrategyWeighted)
- **高可用**: 使用最少连接策略 (StrategyLeastConn)
- **负载分散**: 使用随机策略 (StrategyRandom)

### 节点配置选项
```go
type BalancerNodeOptions struct {
    // 每个节点的最大并发连接数，不保证公平性
    MaxConns int64
    // 节点权重，默认为 1
    Weight uint64
    // 节点失败后的冷却期，默认为 1000ms
    Cooldown int64
}
```

### 错误处理
```go
var (
    ErrKeyRequired      = errors.New("key is required")
    ErrNodeExists       = errors.New("node already exists")
    ErrNodeNotFound     = errors.New("node not found")
    ErrNoAvailableNodes = errors.New("no available nodes")
    ErrEmptyCallback    = errors.New("callback function cannot be empty")
    ErrNoNodesAdded     = errors.New("please add nodes first")
)
```

### 权重管理最佳实践
- 权重范围：1-1000000，推荐使用 1-100 范围
- 权重为 0 或超过范围会返回错误信息（不需要外部处理，主要是开发配置错误）
- 节点不存在时会返回 ErrNodeNotFound 错误
- 建议根据服务器性能动态调整权重
- 使用原子操作确保并发安全

### 节点信息结构体
```go
type BalancerNodeInfo[T any] struct {
    Node      T     // 节点数据
    Weight    uint64 // 当前权重
    MaxConns  int64  // 最大连接数
    Cooldown  int64  // 冷却时间（毫秒）
    Available bool   // 是否可用
    Active    int64  // 当前活跃连接数
}
```

## 最佳实践

1. 根据任务类型调整工作池大小
2. 实现适当的错误处理
3. 定期检查和调整负载均衡策略
4. 注意资源释放时机
5. 权重管理建议：
   - 在服务启动时设置合理的初始权重
   - 根据服务器性能动态调整权重（高性能服务器设置更高权重）
   - 监控节点健康状态，及时调整不可用节点的权重
   - 使用 GetNodeInfo 定期检查节点状态和活跃连接数
   - 权重值支持 1-1000000 范围，推荐使用 1-100 以获得最佳性能
   - 权重设置是原子操作，支持高并发环境下的安全调整