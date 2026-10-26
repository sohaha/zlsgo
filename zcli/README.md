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
- **进度条**: 支持终端进度条和 spinner
- **执行引擎**: 本地 / embed / 远程下载二进制的统一执行入口

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

### 进度条

```go
// 创建进度条
func NewProgressBar(total int64, opts ...func(o *ProgressOptions)) *ProgressBar

// 更新进度
func (p *ProgressBar) Add(delta int64)
func (p *ProgressBar) Increment()
func (p *ProgressBar) SetTotal(total int64)
func (p *ProgressBar) Set(value int64)
func (p *ProgressBar) Finish()
func (p *ProgressBar) Close() error

// 读取状态
func (p *ProgressBar) Current() int64
func (p *ProgressBar) Total() int64
func (p *ProgressBar) String() string
```

常用选项：

```go
type ProgressOptions struct {
    Writer        io.Writer
    Width         int
    Prefix        string
    Suffix        string
    Fill          byte
    Empty         byte
    Spinner       []rune
    FlushInterval time.Duration
}
```

可以按需传多个 `func(o *zcli.ProgressOptions)`，未设置字段会继续使用默认值。

#### 已知总量的进度条

```go
pb := zcli.NewProgressBar(100, func(o *zcli.ProgressOptions) {
    o.Prefix = "上传"
    o.Width = 30
})
defer pb.Close()

for i := 0; i < 100; i++ {
    pb.Add(1)
}
```

#### 未知总量的 spinner

```go
pb := zcli.NewProgressBar(0, func(o *zcli.ProgressOptions) {
    o.Prefix = "处理中"
    o.Spinner = []rune{'⠁', '⠂', '⠄', '⠂'}
})
defer pb.Close()

for i := 0; i < 10; i++ {
    pb.Add(1)
}
```

说明：

- `total <= 0` 时自动进入 spinner 模式。
- 默认会按刷新间隔输出，避免高频写终端。
- ETA 使用平滑速率估算，长任务中的跳动会比简单累计平均更小。
- 真实终端下会尝试读取终端宽度并自动收缩进度条，避免窄终端换行。
- 非终端 writer 会按行输出，不再写入 `\r` 覆盖控制符，便于日志和文件消费。
- `ProgressOptions.FlushInterval = 0` 会在每次状态变化时立即输出，包含同百分比下的计数变化，适合日志采集或测试。
- 对于已知总量的进度条，内部会将进度限制在 `0 ~ total`，不会输出越界进度。
- 当终端非常窄时，会优先保留核心进度信息，其次保留前缀，后缀会最后参与省略。
- 当已知总量模式下的 meta 过长时，会优先缩写 `Elapsed/ETA`，必要时继续只保留百分比与计数等核心字段。
- 宽度估算会优先兼容常见中文、emoji、keycap 和 ZWJ emoji 组合，但仍不是完整的 grapheme 布局引擎。

### 执行引擎（NewExec）

`NewExec` 提供统一的二进制执行入口，支持三种来源（按优先级）：

1. `BinFile` 指向的本地文件已存在：直接使用
2. 本地文件不存在但传入了 `embed.FS`：落盘到临时目录后使用
3. embed 内容为空且配置了 `DownloadUrl`：下载（支持 `.zip` / `.tar.gz` / `.tgz` / 单文件 `.gz`）后使用

```go
// 创建执行引擎
func NewExec(fs embed.FS, fn ...func(options *ExecOptions)) (*Exec, error)

// 同步执行，返回退出码、标准输出、标准错误与错误
func (e *Exec) Run(ctx context.Context, args []string, opt ...func(o *ExecOptions)) (code int, outStr string, errStr string, err error)
// 流式执行，逐行回调输出，返回退出码通道与 stdin 写入函数
func (e *Exec) KeepRun(ctx context.Context, args []string, callback func(str string, isStdout bool), opt ...func(o *ExecOptions)) (<-chan int, func(string), error)
// 释放引擎上下文，并清理引擎自己创建的临时目录（外部 BinFile 不会被删除）
func (e *Exec) Release() error
// 引擎当前存放二进制的目录
func (e *Exec) Dir() string
// 解析后的二进制路径
func (e *Exec) BinFile() string
```

`ExecOptions` 可配置项：

```go
type ExecOptions struct {
    // BinFile 既可以是已存在的二进制路径，也可以是解压/下载后的目标文件名
    BinFile string
    // 环境变量与工作目录覆盖（nil Env 表示继承父进程）
    Env []string
    Dir string
    // KeepRun 时立即关闭子进程 stdin，适合读到 EOF 才退出的命令
    CloseStdin bool
    // 取消或进程退出后管道排空的上限，防止孤儿孙进程卡住 Wait。默认 5s，负值禁用
    WaitDelay time.Duration
    // 下载来源与进度回调
    DownloadUrl string
    DownloadProgress func(percent float64, size int64)
    // 自定义下载产物处理（WithDir 版本优先）
    DownloadHandler func(file string) error
    DownloadHandlerWithDir func(file string, dir string) error
}
```

#### 执行本地二进制

```go
engine, err := zcli.NewExec(embed.FS{}, func(o *zcli.ExecOptions) {
    o.BinFile = "/path/to/tool"
    o.Dir = "/path/to/workdir"
    o.Env = []string{"APP_MODE=prod"}
})
if err != nil {
    panic(err)
}
defer engine.Release()

code, out, errOut, err := engine.Run(context.Background(), []string{"--version"})
fmt.Println(code, out, errOut, err)
```

#### 使用 embed 二进制

```go
//go:embed bin/*
var binFS embed.FS

engine, err := zcli.NewExec(binFS, func(o *zcli.ExecOptions) {
    o.BinFile = "tool"
})
if err != nil {
    panic(err)
}
defer engine.Release()
```

#### 下载远程二进制

```go
engine, err := zcli.NewExec(embed.FS{}, func(o *zcli.ExecOptions) {
    o.BinFile = "tool"
    o.DownloadUrl = "https://example.com/tool.zip"
    o.DownloadProgress = func(percent float64, size int64) {
        fmt.Printf("progress: %.2f%% size=%d\n", percent, size)
    }
})
if err != nil {
    panic(err)
}
defer engine.Release()
```

#### 流式执行与 stdin 交互

```go
codeCh, input, err := engine.KeepRun(ctx, []string{"run"}, func(line string, isStdout bool) {
    fmt.Println(line)
})
if err != nil {
    panic(err)
}
input("hello\n")
code := <-codeCh
```

说明：

- 参数按 argv 原样传递，不经 shell 二次解析，含空格或特殊字符也安全。
- 下载会先缓存到本地临时目录再解压，压缩包会校验魔数与解压路径，避免目录穿越与错误页面被缓存复用。
- 解压保留 tar 中的硬链接与符号链接，并拒绝逃逸出解压目录的链接目标。
- `Run` / `KeepRun` 可通过额外的 `func(o *ExecOptions)` 按次覆盖 `Env`、`Dir`、`CloseStdin`、`WaitDelay`。

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
7. 大批量任务中优先使用默认刷新间隔，避免频繁输出影响性能
