English | [简体中文](./README.md)

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/sohaha/zlsgo?tab=subdirectories)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

![luckything](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fa4bcc6b2-32ef-4a7d-ba1c-65a0330f632d%2Flogo.png?table=block&id=37f366ec-0593-4a21-94c0-c24023a85354&width=590&cache=v2)

Golang daily development common function library

## Documentation

[Check out the documentation](https://docs.73zls.com/zls-go/#)

Recommended to use with the `zzz watch` command of [zzz](https://github.com/sohaha/zzz)

## Why Zara

Simple and easy to use, lightweight enough to avoid excessive external dependencies, 
minimum compatible with old systems such as Window 7.

## QuickStart

### Install

```bash
$ go get github.com/sohaha/zlsgo
```

### HTTP Service

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
    // Implicit routing (struct binding) please refer to the document
    znet.Run()
}
```

![znet](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2F1d7f2372-5d58-4848-85ca-1bedf8ad14ae%2FUntitled.png?table=block&id=18fdfaa9-5dab-4cb8-abb3-f19ff37ed3f0&width=2210&userId=&cache=v2)

### Logger

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

![zlog](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fd8cc2527-8d9d-466c-b5c8-96e706ee0691%2FUntitled.png?table=block&id=474726aa-05fd-47ba-b270-59017c59817b&width=2560&cache=v2)

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
- [x] Abnormal report
- [x] Terminal application
- [ ] [Database](https://github.com/sohaha/zdb)
- [ ] ...

## LICENSE

[MIT](LICENSE)
