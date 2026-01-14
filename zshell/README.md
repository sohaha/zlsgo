# zshell 模块

`zshell` 提供了跨平台的命令执行、管道操作、后台运行、回调处理等功能，用于系统命令的自动化执行。

## 功能概览

- **基本命令执行**: 同步和异步命令执行
- **管道操作**: 多命令管道连接
- **后台运行**: 非阻塞命令执行
- **回调处理**: 命令执行结果回调
- **跨平台支持**: Windows、Linux、macOS 兼容
- **输出控制**: 灵活的输出处理

## 核心功能

### 命令执行

```go
// 同步执行命令
func Run(command string, opt ...func(o *Options)) (code int, outStr, errStr string, err error)
// 带上下文的同步执行命令
func RunContext(ctx context.Context, command string, opt ...func(o *Options)) (code int, outStr, errStr string, err error)
// 执行命令并返回结果
func ExecCommand(ctx context.Context, command []string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, opt ...func(o *Options)) (code int, outStr, errStr string, err error)
// 执行命令并处理回调
func ExecCommandHandle(ctx context.Context, command []string, bef func(cmd *exec.Cmd) error, aft func(cmd *exec.Cmd, err error)) (code int, err error)
```

### 管道操作

```go
func PipeExecCommand(ctx context.Context, commands [][]string, opt ...func(o *Options)) (code int, outStr, errStr string, err error)
```

### 后台运行

```go
func BgRun(command string, opt ...func(o *Options)) error
func BgRunContext(ctx context.Context, command string, opt ...func(o *Options)) error
```

### 回调处理

```go
func CallbackRun(command string, callback func(out string, isBasic bool), opt ...func(o *Options)) (<-chan int, func(string), error)
func CallbackRunContext(ctx context.Context, command string, callback func(str string, isStdout bool), opt ...func(o *Options)) (<-chan int, func(string), error)
```

### 跨平台支持

```go
func RunNewProcess(file string, args []string) (pid int, err error)
func RunBash(ctx context.Context, command string) (code int, outStr, errStr string, err error)
```

### 输出控制

```go
func OutRun(command string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, opt ...func(o Options) Options) (code int, outStr, errStr string, err error)
```

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "io"
    "os"
    "time"
    "github.com/sohaha/zlsgo/zshell"
)

func main() {
    // 基本命令执行
    code, output, errStr, err := zshell.Run("ls -la")
    if err == nil {
        fmt.Printf("命令输出: %s\n", output)
        fmt.Printf("退出码: %d\n", code)
    } else {
        fmt.Printf("命令执行失败: %v, 错误输出: %s\n", err, errStr)
    }
    
    // 带上下文的命令执行
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    code, output, errStr, err = zshell.RunContext(ctx, "sleep 10")
    if err != nil {
        fmt.Printf("超时命令: %v\n", err)
    }
    
    // 执行带参数的命令
    command := []string{"echo", "Hello", "World"}
    code, output, errStr, err = zshell.ExecCommand(ctx, command, nil, nil, nil)
    if err == nil {
        fmt.Printf("参数命令输出: %s\n", output)
    }
    
    // 管道操作
    commands := [][]string{
        {"ls", "-la"},
        {"grep", ".go"},
        {"wc", "-l"},
    }
    code, output, errStr, err = zshell.PipeExecCommand(ctx, commands)
    if err == nil {
        fmt.Printf("管道命令输出: %s\n", output)
    }
    
    // 后台运行
    err = zshell.BgRun("sleep 30")
    if err == nil {
        fmt.Println("后台任务已启动")
    }
    
    // 带上下文的后台运行
    err = zshell.BgRunContext(ctx, "sleep 60")
    if err == nil {
        fmt.Println("带上下文的后台任务已启动")
    }
    
    // 回调处理
    done, _, err := zshell.CallbackRun("ping -c 3 localhost", func(output string, isBasic bool) {
        if isBasic {
            fmt.Printf("标准输出: %s", output)
        } else {
            fmt.Printf("标准错误: %s", output)
        }
    })
    if err == nil {
        // 等待命令完成
        <-done
        fmt.Println("回调命令执行完成")
    }
    
    // 带上下文的回调处理
    done, _, err = zshell.CallbackRunContext(ctx, "ping -c 3 localhost", func(output string, isStdout bool) {
        if isStdout {
            fmt.Printf("标准输出: %s", output)
        } else {
            fmt.Printf("标准错误: %s", output)
        }
    })
    if err == nil {
        // 等待命令完成
        <-done
        fmt.Println("带上下文的回调命令执行完成")
    }
    
    // 跨平台支持
    // 运行新进程
    pid, err := zshell.RunNewProcess("echo", []string{"Hello", "World"})
    if err == nil {
        fmt.Printf("新进程 PID: %d\n", pid)
    }
    
    // 运行 Bash 命令
    code, output, errStr, err = zshell.RunBash(ctx, "echo 'Hello from Bash'")
    if err == nil {
        fmt.Printf("Bash 命令输出: %s\n", output)
    }
    
    // 输出控制
    code, output, errStr, err = zshell.OutRun("ls -la", nil, os.Stdout, os.Stderr)
    if err == nil {
        fmt.Printf("输出控制命令完成，退出码: %d\n", code)
    }
    
    // 实际应用示例
    // 系统监控脚本
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            // 检查系统负载
            _, output, _, err := zshell.Run("uptime")
            if err == nil {
                fmt.Printf("系统负载: %s\n", output)
            }
            
            // 检查磁盘使用
            _, output, _, err = zshell.Run("df -h")
            if err == nil {
                fmt.Printf("磁盘使用: %s\n", output)
            }
        }
    }()
    
    // 文件处理管道
    go func() {
        commands := [][]string{
            {"find", ".", "-name", "*.go"},
            {"xargs", "wc", "-l"},
        }
        
        code, output, errStr, err := zshell.PipeExecCommand(context.Background(), commands)
        if err == nil {
            fmt.Printf("Go 文件行数统计: %s\n", output)
        }
    }()
    
    // 保持程序运行
    time.Sleep(2 * time.Minute)
}
```

## 命令执行模式

### 同步执行
```go
// 同步执行命令
code, output, errStr, err := zshell.Run("ls -la")
```

### 异步执行
```go
// 后台执行命令
err := zshell.BgRun("long-running-command")
```

### 管道执行
```go
// 执行管道命令
commands := [][]string{
    {"ls", "-la"},
    {"grep", ".go"},
    {"wc", "-l"},
}
code, output, errStr, err := zshell.PipeExecCommand(ctx, commands)
```

## Options 配置选项

```go
type Options struct {
    Dir        string   // 工作目录
    Env        []string // 环境变量
    CloseStdin bool     // 启动后立即关闭 stdin
}
```

### 使用示例

```go
// 设置工作目录
zshell.Run("ls", func(o *zshell.Options) {
    o.Dir = "/tmp"
})

// 关闭 stdin（用于等待 stdin EOF 的命令）
zshell.CallbackRunContext(ctx, "wc -l", callback, func(o *zshell.Options) {
    o.CloseStdin = true
})

// 组合选项
zshell.CallbackRunContext(ctx, cmd, callback, func(o *zshell.Options) {
    o.Dir = "/tmp"
    o.CloseStdin = true
})
```

## 最佳实践

1. 使用上下文控制超时
2. 实现适当的错误处理
3. 合理使用管道操作
4. 合理使用后台执行