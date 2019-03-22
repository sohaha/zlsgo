English | [简体中文](./README.ZH.md)

[![Build Status](https://www.travis-ci.org/sohaha/zlsgo.svg?branch=master)](https://www.travis-ci.org/sohaha/zlsgo)
[![GoDoc](https://godoc.org/github.com/sohaha/zlsgo?status.svg)](https://godoc.org/github.com/sohaha/zlsgo)
![flat](https://img.shields.io/github/languages/top/sohaha/zlsgo.svg?style=flat)
[![codecov](https://codecov.io/gh/sohaha/zlsgo/branch/master/graph/badge.svg)](https://codecov.io/gh/sohaha/zlsgo)

Golang daily development common function library

## Why Zara

- Convenient for daily development

## QuickStart

### Install

```bash
$ go get github.com/sohaha/zlsgo
```

### Usage

```go
package main

import(
  "github.com/sohaha/zlsgo/gvar"
  "fmt
  )

func main()  {
  name := "hi"
  fmt.Println("This is a string",gvar.IsString(name))
}
```

## LICENSE

[MIT](LICENSE)
