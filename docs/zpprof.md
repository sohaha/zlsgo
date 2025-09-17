# zpprof 模块

`zpprof` 提供了 pprof 处理器注册、系统信息收集等功能，用于应用程序的性能监控和调试。

## 功能概览

- **pprof 注册**: 自动注册 pprof 处理器到 znet 引擎
- **系统信息**: 运行时系统信息收集
- **性能分析**: 支持多种分析类型
- **认证支持**: 可选的访问控制
- **独立服务器**: 支持独立的 pprof 服务器
- **系统监控**: 实时系统性能监控

## 核心功能

### pprof 集成

```go
// 注册 pprof 处理器到 znet 引擎
func Register(r *znet.Engine, token string) *znet.Engine
// 启动独立的 pprof 服务器
func ListenAndServe(addr ...string) error
```



### 系统信息

```go
// 创建系统信息收集器
func NewSystemInfo(startTime time.Time) *SystemInfo
```

### 系统信息字段

```go
type SystemInfo struct {
    ServerName   string // 服务器名称
    Runtime      string // 运行时间
    GoroutineNum string // Goroutine 数量
    CPUNum       string // CPU 核心数
    UsedMem      string // 当前内存使用
    TotalMem     string // 总分配内存
    SysMem       string // 系统内存使用
    Lookups      string // 指针查找次数
    Mallocs      string // 内存分配次数
    Frees        string // 内存释放次数
    LastGCTime   string // 上次 GC 时间
    NextGC       string // 下次 GC 阈值
    PauseTotalNs string // GC 暂停总时间
    PauseNs      string // 上次 GC 暂停时间
    HeapInuse    string // 堆内存使用
}
```

## 使用示例

```go
package main

import (
    "fmt"
    "net/http"
    "time"
    "github.com/sohaha/zlsgo/zpprof"
    "github.com/sohaha/zlsgo/znet"
)

func main() {
    // 创建 znet 引擎
    app := znet.New()
    
    // 注册 pprof 处理器（带认证令牌）
    zpprof.Register(app, "your-secret-token")
    
    // 启动主应用服务器
    go func() {
        znet.Run()
    }()
    
    // 启动独立的 pprof 服务器
    go func() {
        err := zpprof.ListenAndServe(":6060")
        if err != nil {
            fmt.Printf("启动 pprof 服务器失败: %v\n", err)
        }
    }()
    
    // 系统信息收集
    startTime := time.Now()
    sysInfo := zpprof.NewSystemInfo(startTime)
    
    // 访问系统信息
    fmt.Printf("服务器名称: %s\n", sysInfo.ServerName)
    fmt.Printf("运行时间: %s\n", sysInfo.Runtime)
    fmt.Printf("Goroutine 数量: %s\n", sysInfo.GoroutineNum)
    fmt.Printf("CPU 核心数: %s\n", sysInfo.CPUNum)
    fmt.Printf("堆内存使用: %s\n", sysInfo.HeapInuse)
    fmt.Printf("已用内存: %s\n", sysInfo.UsedMem)
    fmt.Printf("总分配内存: %s\n", sysInfo.TotalMem)
    fmt.Printf("系统内存: %s\n", sysInfo.SysMem)
    fmt.Printf("指针查找次数: %s\n", sysInfo.Lookups)
    fmt.Printf("内存分配次数: %s\n", sysInfo.Mallocs)
    fmt.Printf("内存释放次数: %s\n", sysInfo.Frees)
    fmt.Printf("上次 GC 时间: %s\n", sysInfo.LastGCTime)
    fmt.Printf("下次 GC 阈值: %s\n", sysInfo.NextGC)
    fmt.Printf("GC 暂停总时间: %s\n", sysInfo.PauseTotalNs)
    fmt.Printf("上次 GC 暂停时间: %s\n", sysInfo.PauseNs)
    
    // 实际应用示例
    // 监控服务性能
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            currentInfo := zpprof.NewSystemInfo(startTime)
            fmt.Printf("=== 性能监控 ===\n")
            fmt.Printf("Goroutine: %s, 内存: %s\n", 
                currentInfo.GoroutineNum, currentInfo.UsedMem)
        }
    }()
    
    // 保持程序运行
    select {}
}
```

## pprof 处理器说明

### 自动注册的处理器
- **信息页面**: `/debug/` - 系统信息概览
- **CPU 分析**: `/debug/pprof/profile` - CPU 性能分析
- **内存分析**: `/debug/pprof/heap` - 堆内存分析
- **阻塞分析**: `/debug/pprof/block` - 阻塞分析
- **互斥锁分析**: `/debug/pprof/mutex` - 互斥锁分析
- **Goroutine 分析**: `/debug/pprof/goroutine` - Goroutine 分析
- **线程创建分析**: `/debug/pprof/threadcreate` - 线程创建分析
- **命令行**: `/debug/pprof/cmdline` - 命令行参数
- **符号**: `/debug/pprof/symbol` - 符号表
- **跟踪**: `/debug/pprof/trace` - 执行跟踪

### 使用方法
1. 注册到 znet 引擎
2. 使用浏览器或工具访问相应端点
3. 下载分析文件进行离线分析

## 集成示例

### 与现有 HTTP 服务器集成

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/sohaha/zlsgo/zpprof"
    "github.com/sohaha/zlsgo/znet"
)

func main() {
    // 创建 znet 引擎
    app := znet.New()
    
    // 添加业务路由
    app.GET("/", func(c *znet.Context) {
        c.String(200, "Hello World")
    })
    
    // 注册 pprof 处理器
    zpprof.Register(app, "debug-token")
    
    // 启动服务器
    znet.Run()
}
```

### 与 Gin 框架集成

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/sohaha/zlsgo/zpprof"
    "github.com/sohaha/zlsgo/znet"
)

func main() {
    // 创建 Gin 路由
    ginRouter := gin.Default()
    
    // 创建 znet 引擎用于 pprof
    pprofEngine := znet.New()
    zpprof.Register(pprofEngine, "gin-debug")
    
    // 将 pprof 路由添加到 Gin
    for _, route := range pprofEngine.Routes() {
        ginRouter.Handle(route.Method, route.Path, func(c *gin.Context) {
            // 处理 pprof 请求
        })
    }
    
    // 启动 Gin 服务器
    ginRouter.Run(":8080")
}
```

### 与标准库 http.Server 集成

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/sohaha/zlsgo/zpprof"
    "github.com/sohaha/zlsgo/znet"
)

func main() {
    // 创建 znet 引擎用于 pprof
    pprofEngine := znet.New()
    zpprof.Register(pprofEngine, "debug-token")
    
    // 创建标准库服务器
    mux := http.NewServeMux()
    
    // 添加业务路由
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World"))
    })
    
    // 将 pprof 路由添加到标准库服务器
    for _, route := range pprofEngine.Routes() {
        mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
            // 处理 pprof 请求
        })
    }
    
    // 启动服务器
    fmt.Println("服务器启动在 :8080")
    http.ListenAndServe(":8080", mux)
}
```

## 性能监控示例

### 实时性能监控

```go
package main

import (
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zpprof"
)

func main() {
    startTime := time.Now()
    
    // 启动性能监控
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            info := zpprof.NewSystemInfo(startTime)
            fmt.Printf("=== 性能报告 ===\n")
            fmt.Printf("运行时间: %s\n", info.Runtime)
            fmt.Printf("Goroutine: %s\n", info.GoroutineNum)
            fmt.Printf("内存使用: %s\n", info.UsedMem)
            fmt.Printf("堆内存: %s\n", info.HeapInuse)
            fmt.Printf("GC 状态: %s\n", info.LastGCTime)
            fmt.Println("==================")
        }
    }()
    
    // 模拟工作负载
    for i := 0; i < 100; i++ {
        go func(id int) {
            for {
                time.Sleep(time.Second)
                // 模拟工作
            }
        }(i)
    }
    
    // 保持程序运行
    select {}
}
```

### 内存泄漏检测

```go
package main

import (
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zpprof"
)

func main() {
    startTime := time.Now()
    
    // 启动内存监控
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        var lastMem uint64
        for range ticker.C {
            info := zpprof.NewSystemInfo(startTime)
            
            // 检测内存增长
            if lastMem > 0 {
                growth := info.UsedMem - lastMem
                if growth > 10*1024*1024 { // 10MB
                    fmt.Printf("警告: 内存增长过快: %s\n", growth)
                }
            }
            lastMem = info.UsedMem
            
            fmt.Printf("内存使用: %s\n", info.UsedMem)
        }
    }()
    
    // 模拟内存分配
    for i := 0; i < 1000; i++ {
        go func() {
            data := make([]byte, 1024*1024) // 1MB
            time.Sleep(time.Second)
            _ = data
        }()
    }
    
    select {}
}
```

## 安全配置

### 生产环境配置

```go
package main

import (
    "github.com/sohaha/zlsgo/zpprof"
    "github.com/sohaha/zlsgo/znet"
)

func main() {
    app := znet.New()
    
    // 生产环境使用强密码
    zpprof.Register(app, "your-very-strong-password-here")
    
    // 限制访问IP
    app.Use(func(c *znet.Context) {
        clientIP := c.GetClientIP()
        if !isAllowedIP(clientIP) {
            c.AbortWithStatus(403)
            return
        }
        c.Next()
    })
    
    znet.Run()
}

func isAllowedIP(ip string) bool {
    allowedIPs := []string{"127.0.0.1", "192.168.1.0/24"}
    // 实现IP白名单检查
    return true
}
```

## 最佳实践

1. 仅在开发/测试环境启用
2. 使用强密码作为认证令牌
3. 定期检查系统性能指标
4. 结合日志系统进行监控
5. 实现IP白名单访问控制
6. 监控 pprof 端点的访问频率
7. 定期清理性能分析文件
8. 设置合理的监控间隔