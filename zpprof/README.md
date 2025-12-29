# zpprof 模块

`zpprof` 提供了 pprof 处理器注册、系统信息收集等功能，用于应用程序的性能监控和调试。

## 使用 Go pprof

### 使用方式 1

```go
import (
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zpprof"
)

func main(){
	r := znet.New("Go")

	// 注册pprof路由，如果 token 设置为空表示不需要验证 token
	zpprof.Register(r, "mytoken")

	znet.Run()
}

// 启动服务后直接访问 http://127.0.0.1:3788/debug?token=mytoken
```

### 使用方式 2

```go
// 使用另外端口(原始版本)

go zpprof.ListenAndServe("0.0.0.0:8082")

// 启动服务后直接访问 http://127.0.0.1:8082/debug/pprof/
```

## 使用 zpprof 升级版

运行时性能监控与自动分析工具，支持 CPU/内存/Goroutine/线程/GCHeap 自动触发 pprof dump，并可将性能数据上报至 Pyroscope。

具体配置请参考 [zpprof](https://github.com/zlsgo/zpprof)

```go
package main

import (
    "time"
    "github.com/zlsgo/zpprof"
)

func main() {
    engine, _ := zpprof.New()

    engine.Start()

    // 运行业务逻辑
    select {}
}
```

### 通过 Web 界面分析

查看当前总览：访问  `http://127.0.0.1:3788/debug/pprof/`   (如设置了token自行填上)

```bash
/debug/pprof/

profiles:
0    block
5    goroutine
3    heap
0    mutex
9    threadcreate

full goroutine stack dump
```

这个页面中有许多子页面。

- cpu（CPU Profiling）: /debug/pprof/profile，默认进行 30s 的 CPU Profiling，得到一个分析用的 profile 文件

- block（Block Profiling）：/debug/pprof/block，查看导致阻塞同步的堆栈跟踪

- goroutine：/debug/pprof/goroutine，查看当前所有运行的 goroutines 堆栈跟踪

- heap（Memory Profiling）: /debug/pprof/heap，查看活动对象的内存分配情况

- mutex（Mutex Profiling）：/debug/pprof/mutex，查看导致互斥锁的竞争持有者的堆栈跟踪

- threadcreate：/debug/pprof/threadcreate，查看创建新OS线程的堆栈跟踪

### 通过交互式终端分析

终端执行  `go tool pprof http://127.0.0.1:3788/debug/pprof/profile?seconds=60`  

执行该命令后，需等待 60 秒（可调整 seconds 的值），pprof 会进行 CPU Profiling。

结束后将默认进入 pprof 的交互式命令模式，可以对分析的结果进行查看或导出。

具体可执行 pprof help 查看命令说明

### 可视化界面

终端执行  `go tool pprof -http=:8080 http://127.0.0.1:3788/debug/pprof/profile?seconds=60`  

## 实时监控

statsviz 是一款可视化实时运行时统计，我们可以很方便的集成进来。

```go
package main

import (
	"github.com/arl/statsviz"
	"github.com/sohaha/zlsgo/znet"
)

func main() {
	r := znet.New()

	srv, _ := statsviz.NewServer()

	ws := srv.Ws()
	index := srv.Index()

	r.GET(`/debug/statsviz{*:[\S]*}`, func(c *znet.Context) {
		q := c.GetParam("*")
		if q == "" {
			c.Redirect("/debug/statsviz/")
			return
		}

		if q == "/ws" {
			ws(c.Writer, c.Request)
			return
		}

		index(c.Writer, c.Request)
	})

	// 启动服务后直接访问 http://127.0.0.1:3788/debug/statsviz/
	znet.Run()
}
```