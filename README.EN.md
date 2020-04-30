English | [简体中文](./README.md)

[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![GoDoc](https://godoc.org/github.com/sohaha/zlsgo?status.svg)](https://godoc.org/github.com/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

Golang daily development common function library

## Documentation

[Check out the documentation](https://docs.73zls.com/zls-go/#)

Recommended to use with the `zzz watch` command of [zzz](https://github.com/sohaha/zzz)

## Why Zara

Simple to use and light enough to avoid excessive external dependence

## QuickStart

### Install

```bash
$ go get github.com/sohaha/zlsgo
```

### HTTP Service

![znet](https://www.notion.so/signed/https:%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2F1d7f2372-5d58-4848-85ca-1bedf8ad14ae%2FUntitled.png)

```go
// main.go
package main

import (
    "github.com/sohaha/zlsgo/znet"
)

func main(){
    r := znet.New()

    r.GET("/hi", func(c *znet.Context) {
        c.String(200, "Hello world")
     })

    znet.Run()
}
```

### Logger

![zlog](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fd8cc2527-8d9d-466c-b5c8-96e706ee0691%2FUntitled.png?table=block&id=474726aa-05fd-47ba-b270-59017c59817b&width=2560&cache=v2)

```go
package main

import (
    "github.com/sohaha/zlsgo/zlog"
)

func main(){
    zlog.Debug("This is a debug")
    zlog.Error("This is a error")
    // zlog...
}
```

### HTTP Client

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

### More features

Please read the documentation [https://docs.73zls.com/zls-go/#](https://docs.73zls.com/zls-go/#)

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
