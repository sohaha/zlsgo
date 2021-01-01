[English](./README.EN.md) | 简体中文

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/sohaha/zlsgo?tab=subdirectories)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![Build Status](https://www.travis-ci.com/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.com/sohaha/zlsgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

![luckything](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fa4bcc6b2-32ef-4a7d-ba1c-65a0330f632d%2Flogo.png?table=block&id=37f366ec-0593-4a21-94c0-c24023a85354&width=590&cache=v2)

## 文档

[查看文档](https://docs.73zls.com/zls-go/#)

建议搭配 [zzz](https://github.com/sohaha/zzz) 的 `zzz watch` 指令使用

## 特性

简单易用、足够轻量，避免过多的外部依赖，最低兼容 Window 7 等老系统

## 快速上手

### 安装

```bash
$ go get github.com/sohaha/zlsgo
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
- [ ] [数据库操作](https://github.com/sohaha/zdb)
- [ ] ...

## LICENSE

[MIT](LICENSE)
