English | [简体中文](./README.md)

[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![GoDoc](https://godoc.org/github.com/sohaha/zlsgo?status.svg)](https://godoc.org/github.com/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

Golang daily development common function library

## Documentation

[Check out the documentation](https://docs.73zls.com/zls-go/#)

## Why Zara

- Easy to use, lightweight enough

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

    znet.Run()
}
```

### Logging Tool
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
## LICENSE

[MIT](LICENSE)
