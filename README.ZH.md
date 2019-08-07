[English](./README.md) | 简体中文

[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![GoDoc](https://godoc.org/github.com/sohaha/zlsgo?status.svg)](https://godoc.org/github.com/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

[详细文档](https://docs.73zls.com/zls-go/#)

## 特性

- 简单易用、足够轻量

## 快速上手

### 安装

```bash
$ go get github.com/sohaha/zlsgo
```

### 用法

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

## LICENSE

[MIT](LICENSE)
