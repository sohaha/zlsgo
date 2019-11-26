[English](./README.EN.md) | 简体中文

[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![GoDoc](https://godoc.org/github.com/sohaha/zlsgo?status.svg)](https://godoc.org/github.com/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

## 文档

[查看文档](https://docs.73zls.com/zls-go/#)

建议搭配 [zzz](https://github.com/sohaha/zzz) 的 `zzz watch` 指令使用

## 特性

- 简单易用、足够轻量

## 快速上手

### 安装

```bash
$ go get github.com/sohaha/zlsgo
```

### HTTP服务

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

### 日志工具
```go
package main 

import (
    "github.com/sohaha/zlsgo/zlog"
)

func main(){
    zlog.Debug("这是一个测试")
    zlog.Error("这是一个错误")
    // zlog...
}
```

### 更多功能

请阅读文档 [https://docs.73zls.com/zls-go/#](https://docs.73zls.com/zls-go/#)

## LICENSE

[MIT](LICENSE)

