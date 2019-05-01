[English](./README.md) | 简体中文

[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![GoDoc](https://godoc.org/github.com/sohaha/zlsgo?status.svg)](https://godoc.org/github.com/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

golang 日常开发常用函数库

## 特性

- 方便日常开发

## 快速上手

### 安装

```bash
$ go get github.com/sohaha/zlsgo
```

### 用法

```go
package main

import(
  "github.com/sohaha/zlsgo/zvar"
  "fmt
  )

func main()  {
  name := "hi"
  fmt.Println("This is a string",zvar.IsString(name))
}
```

## LICENSE

[MIT](LICENSE)
