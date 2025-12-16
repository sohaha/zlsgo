English | [简体中文](./README.md)

![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/sohaha/zlsgo?tab=subdirectories)
[![UnitTest](https://github.com/sohaha/zlsgo/actions/workflows/go.yml/badge.svg)](https://github.com/sohaha/zlsgo/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sohaha/zlsgo)](https://goreportcard.com/report/github.com/sohaha/zlsgo)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

![luckything](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fa4bcc6b2-32ef-4a7d-ba1c-65a0330f632d%2Flogo.png?table=block&id=37f366ec-0593-4a21-94c0-c24023a85354&width=590&cache=v2)

## 📚 Documentation

[Online Documentation](https://docs.73zls.com/zls-go/#)

For detailed documentation of each module, please refer to the README.md file in the corresponding module directory, for example:
- [znet - Web Framework](./znet/)
- [zlog - Logger](./zlog/)
- [zhttp - HTTP Client](./zhttp/)
- [zjson - JSON Processing](./zjson/)
- [More modules...](#-module-list)

Recommended to use with the `zzz watch` command of [zzz](https://github.com/sohaha/zzz)

## ✨ Features

- **Lightweight & Efficient**: Avoid excessive external dependencies, minimum compatible with old systems such as Windows 7
- **Modular Design**: Import on demand, reduce unnecessary code volume
- **Type Safe**: Fully utilize Go's type system, provide type-safe APIs
- **High Performance**: Optimize underlying implementation, pursue ultimate performance
- **Simple & Easy to Use**: Provide concise and intuitive API design

## 🚀 QuickStart

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
    // Get an instance
    r := znet.New()

    // Register route
    r.GET("/hi", func(c *znet.Context) {
        c.String(200, "Hello world")
     })
    // Implicit routing (struct binding) please refer to the document
    // Start
    znet.Run()
}
```

![znet](https://www.notion.so/image/https%3A%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2F1d7f2372-5d58-4848-85ca-1bedf8ad14ae%2FUntitled.png?table=block&id=18fdfaa9-05fd-47ba-b270-59017c59817b&width=2210&userId=&cache=v2)

### Logger

```go
package main

import (
    "github.com/sohaha/zlsgo/zlog"
)

func main(){
    logs := []string{"This is a test", "This is an error"}
    zlog.Debug(logs[0])
    zlog.Error(logs[1])
    zlog.Dump(logs)
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

## 📦 Module List

### Core Modules
- [zarray](./zarray/) - Array operations
- [zcache](./zcache/) - Cache library
- [zcli](./zcli/) - Command-line interface
- [zdi](./zdi/) - Dependency injection
- [zerror](./zerror/) - Error handling

### File and Data Modules
- [zfile](./zfile/) - File operations
- [zhttp](./zhttp/) - HTTP client
- [zjson](./zjson/) - JSON processing
- [ztype](./ztype/) - Type handling

### Network and Web Modules
- [znet](./znet/) - Web framework
- [zpool](./zpool/) - Resource pool management
- [zpprof](./zpprof/) - Performance profiling

### Utility Modules
- [zlog](./zlog/) - Logger
- [zreflect](./zreflect/) - Reflection utilities
- [zshell](./zshell/) - Shell command execution
- [zstring](./zstring/) - String processing
- [zsync](./zsync/) - Synchronization primitives
- [ztime](./ztime/) - Time handling
- [zutil](./zutil/) - Common utilities
- [zvalid](./zvalid/) - Data validation
- [zlocale](./zlocale/) - Internationalization

## Todo

- [x] HTTP Server
- [x] HTTP Client
- [x] JSON RPC
- [x] Logger
- [x] JSON processing
- [x] String processing
- [x] Validator
- [x] Hot Restart
- [x] Daemon
- [x] Exception reporting
- [x] Terminal application
- [x] Goroutine pool
- [x] HTML Parse
- [x] Dependency injection
- [x] Server Sent Event
- [x] High-performance HashMap
- [x] Internationalization
- [ ] [Database operations](https://github.com/sohaha/zdb)
- [ ] ...(Read more [documentation](https://docs.73zls.com/zls-go/#))

## LICENSE

[MIT](LICENSE)
