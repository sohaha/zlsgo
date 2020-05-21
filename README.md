[English](./README.EN.md) | 简体中文

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/sohaha/zlsgo?tab=subdirectories)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

## 文档

[查看文档](https://docs.73zls.com/zls-go/#)

建议搭配 [zzz](https://github.com/sohaha/zzz) 的 `zzz watch` 指令使用

## 特性

简单易用、足够轻量，避免过多的外部依赖

## 快速上手

### 安装

```bash
$ go get github.com/sohaha/zlsgo
```

### 示例

<details>
<summary>HTTP 服务</summary>
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

    // 启动
    znet.Run()
}
```
</details>
![znet](https://www.notion.so/signed/https:%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2F1d7f2372-5d58-4848-85ca-1bedf8ad14ae%2FUntitled.png)

### 日志工具

![zlog](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fd8cc2527-8d9d-466c-b5c8-96e706ee0691%2FUntitled.png?table=block&id=474726aa-05fd-47ba-b270-59017c59817b&width=2560&cache=v2)

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

- [x] HttpServer
- [x] HttpClient
- [x] Logger
- [x] Json processing
- [x] String processing
- [x] Validator
- [x] Hot Restart
- [x] Daemon
- [ ] ...

## LICENSE

[MIT](LICENSE)
