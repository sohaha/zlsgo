[English](./README.EN.md) | 简体中文

![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/sohaha/zlsgo?tab=subdirectories)
[![UnitTest](https://github.com/sohaha/zlsgo/actions/workflows/go.yml/badge.svg)](https://github.com/sohaha/zlsgo/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

![luckything](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fa4bcc6b2-32ef-4a7d-ba1c-65a0330f632d%2Flogo.png?table=block&id=37f366ec-0593-4a21-94c0-c24023a85354&width=590&cache=v2)

## 📚 文档

[在线文档](https://docs.73zls.com/zls-go/#) | [本地文档](./docs/README.md)

建议搭配 [zzz](https://github.com/sohaha/zzz) 的 `zzz watch` 指令使用

## ✨ 特性

- **轻量高效**：避免过多的外部依赖，最低兼容 Windows 7 等老系统
- **模块化设计**：按需引入，减少不必要的代码体积
- **类型安全**：充分利用 Go 类型系统，提供类型安全的 API
- **高性能**：优化底层实现，追求极致的性能表现
- **简单易用**：提供简洁直观的 API 设计

## 🚀 快速开始

### 安装

```bash
go get github.com/sohaha/zlsgo
```

### HTTP 服务

```go
// main.go
package main

import (
    "github.com/sohaha/zlsgo/znet"
)

func main(){
    // 获取一个实例
    r := znet.New()

    // 注册路由
    r.GET("/hi", func(c *znet.Context) {
        c.String(200, "Hello world")
     })
    // 隐性路由（结构体绑定）请参考文档
    // 启动
    znet.Run()
}
```

![znet](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2F1d7f2372-5d58-4848-85ca-1bedf8ad14ae%2FUntitled.png?table=block&id=18fdfaa9-5dab-4cb8-abb3-f19ff37ed3f0&width=2210&userId=&cache=v2)

### 日志工具

```go
package main

import (
    "github.com/sohaha/zlsgo/zlog"
)

func main(){
    logs := []string{"这是一个测试","这是一个错误"}
    zlog.Debug(logs[0])
    zlog.Error(logs[1])
    zlog.Dump(logs)
    // zlog...
}
```

![zlog](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fd8cc2527-8d9d-466c-b5c8-96e706ee0691%2FUntitled.png?table=block&id=474726aa-05fd-47ba-b270-59017c59817b&width=2560&cache=v2)

### HTTP 客户端

```go
// main.go
package main

import (
    "github.com/sohaha/zlsgo/zhttp"
    "github.com/sohaha/zlsgo/zlog"
)

func main(){
    data, err := zhttp.Get("https://github.com")
    if err != nil {
      zlog.Error(err)
      return
    }
    res := data.String()
    zlog.Debug(res)

}
```

### 更多功能

请阅读文档 [https://docs.73zls.com/zls-go/#](https://docs.73zls.com/zls-go/#)

## Todo

- [x] HTTP 服务端
- [x] Http 客户端
- [x] JSON RPC
- [x] 日志功能
- [x] Json 处理
- [x] 字符串处理
- [x] 验证器
- [x] 热重启
- [x] 守护进程
- [x] 异常上报
- [x] 终端应用
- [x] 协程池
- [x] HTML 解析
- [x] 依赖注入
- [x] Server Sent 推送
- [x] 高性能 HashMap
- [x] 国际化
- [ ] [数据库操作](https://github.com/sohaha/zdb)
- [ ] ...(更多请阅读[文档](https://docs.73zls.com/zls-go/#))

## LICENSE

[MIT](LICENSE)
