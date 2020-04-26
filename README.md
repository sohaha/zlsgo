<!--
 * @Author: seekwe
 * @Date: 2020-01-03 12:52:27
 * @Last Modified by:: seekwe
 * @Last Modified time: 2020-04-26 17:55:08
 -->

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

### HTTP 服务

![znet](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/1d7f2372-5d58-4848-85ca-1bedf8ad14ae/Untitled.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAT73L2G45O3KS52Y5%2F20200426%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20200426T094654Z&X-Amz-Expires=86400&X-Amz-Signature=92f6cebbf76b4ae5a1190e107ead1b0ca07c760f2b230a0865dd8320168e7fd1&X-Amz-SignedHeaders=host&response-content-disposition=filename%20%3D%22Untitled.png%22)

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

![zlog](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/76a0d6e2-8fda-43a1-b900-91160ce9cbd6/Untitled.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAT73L2G45O3KS52Y5%2F20200426%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20200426T095429Z&X-Amz-Expires=86400&X-Amz-Signature=73b2e4ed47431ae72a16e3f22577a8537ba6c6fc4621ec5cfa08cd73bed749fe&X-Amz-SignedHeaders=host&response-content-disposition=filename%20%3D%22Untitled.png%22)

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

### 更多功能

请阅读文档 [https://docs.73zls.com/zls-go/#](https://docs.73zls.com/zls-go/#)

## Todo

- [x] HttpServer
- [x] HttpClient
- [x] Logger
- [x] Json processing
- [x] String processing
- [x] Validator
- [ ] ...

## LICENSE

[MIT](LICENSE)
