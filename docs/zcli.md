# zcli 模块

`zcli` 提供了命令行解析、参数绑定、服务管理、信号处理、工具函数等功能，用于构建功能丰富的命令行应用程序。

## 功能概览

- **命令行解析**: 命令行参数和标志解析
- **参数绑定**: 参数绑定到结构体和变量
- **服务管理**: 服务生命周期管理
- **信号处理**: 系统信号处理
- **工具函数**: 终端交互和工具函数
- **子命令支持**: 支持多级子命令
- **帮助文档**: 自动生成帮助信息
- **交互式输入**: 支持用户交互输入

## 核心功能

### 命令行解析

```go
// 添加命令到CLI系统
func Add(name, description string, command Cmd) *cmdCont
// 获取指定名称的命令
func GetCommand(name string) (cmd *cmdCont, ok bool)
// 获取所有命令
func GetAllCommand() map[string]*cmdCont
// 设置未知命令处理函数
func SetUnknownCommand(fn func(string))
// 解析命令行参数
func Parse(arg ...[]string) bool
// 获取命令行参数
func Args() []string
// 获取标志集合
func Flag() *flag.FlagSet
// 获取标志集合（别名）
func GetFlag() *flag.FlagSet
// 运行命令
func Run(runFunc ...runFunc) bool
// 启动CLI系统
func Start(runFunc ...runFunc)
// 停止CLI系统
func Stop()
// 等待CLI完成
func Wait()
// 致命错误处理
func Fatal(format string, v ...interface{})
```

### 国际化支持

```go
// 设置语言文本
func SetLangText(lang, key, value string)
// 获取语言文本
func GetLangText(key string, def ...string) string
```

### 参数绑定

```go
// 设置变量
func SetVar(name, usage string) *v
// 设置必需参数
func (v *v) Required() *v
// 字符串类型
func (v *v) String(def ...string) *string
// 整数类型
func (v *v) Int(def ...int) *int
// 64位整数类型
func (v *v) Int64(def ...int64) *int64
// 无符号整数类型
func (v *v) Uint(def ...uint) *uint
// 64位无符号整数类型
func (v *v) Uint64(def ...uint64) *uint64
// 64位浮点数类型
func (v *v) Float64(def ...float64) *float64
// 布尔类型
func (v *v) Bool(def ...bool) *bool
// 时间间隔类型
func (v *v) Duration(def ...time.Duration) *time.Duration
// 自定义类型
func (v *v) Var(value flag.Value, name string) *flag.Flag
// 自定义函数
func (v *v) Func(fn func(s string) error) *flag.Flag
```

### 服务管理

```go
// 启动服务运行
func LaunchServiceRun(name string, description string, fn func(), config ...*daemon.Config) error
// 启动服务
func LaunchService(name string, description string, fn func(), config ...*daemon.Config) (daemon.ServiceIface, error)
// 获取服务
func GetService() (daemon.ServiceIface, error)
```

### 信号处理

```go
// 单次终止信号
func SingleKillSignal() <-chan bool
```

### 工具函数

```go
// 获取用户输入
func Input(problem string, required bool) string
// 获取用户输入（换行）
func Inputln(problem string, required bool) string
// 获取当前命令
func Current() (interface{}, bool)
// 检查是否为sudo
func IsSudo() bool
// 检查是否为双击启动
func IsDoubleClickStartUp() bool
// 锁定实例
func LockInstance() (clean func(), ok bool)
// 检查错误
func CheckErr(err error, exit ...bool)
// 显示帮助
func Help()
// 显示错误
func Error(format string, v ...interface{})
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zcli"
)

// 定义命令结构体
type testCmd struct{}

func (cmd *testCmd) Flags(sub *zcli.Subcommand) {
    // 设置命令标志
    sub.SetVar("name", "测试名称").String("default")
    sub.SetVar("count", "测试次数").Int(1)
}

func (cmd *testCmd) Run(args []string) {
    fmt.Println("执行测试命令")
    fmt.Printf("参数: %v\n", args)
}

func main() {
    // 添加命令
    zcli.Add("test", "测试命令", &testCmd{})
    
    // 设置未知命令处理
    zcli.SetUnknownCommand(func(cmd string) {
        fmt.Printf("未知命令: %s\n", cmd)
        zcli.Help()
    })
    
    // 启动 CLI
    zcli.Start()
}
```

## 命令结构

### 命令接口
```go
type Cmd interface {
    Flags(sub *Subcommand)
    Run(args []string)
}
```

### 子命令结构
```go
type Subcommand struct {
    // 子命令相关字段
}
```

## 高级功能

### 多级子命令
```go
// 支持多级子命令结构
zcli.Add("db", "数据库操作", &dbCmd{})
zcli.Add("db:migrate", "数据库迁移", &migrateCmd{})
zcli.Add("db:seed", "数据库填充", &seedCmd{})
```

### 参数验证
```go
// 参数验证和转换
name := sub.SetVar("name", "用户名").Required().String()
count := sub.SetVar("count", "数量").Int(1)
if *count < 0 {
    zcli.Fatal("数量不能为负数")
}
```

### 服务管理
```go
// 启动后台服务
ctx, cancel := context.WithCancel(context.Background())
err := zcli.LaunchServiceRun("myapp", "我的应用", func() {
    // 服务逻辑 ...
    
    // 等待结束
    <-zcli.SingleKillSignal()
    
    // 释放
    cancel()
}, &daemon.Config{
		Context: ctx,
})
```

## 最佳实践

1. 使用结构化的命令定义
2. 实现适当的参数验证
3. 提供详细的帮助文档
4. 处理异常情况
5. 使用有意义的命令名称
6. 提供清晰的参数说明
