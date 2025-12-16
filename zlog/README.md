# zlog 模块

`zlog` 提供了丰富的日志功能、颜色支持、文件输出、调试工具等，用于应用程序的日志记录和调试。

## 功能概览

- **日志记录**: 支持多种日志级别的记录
- **颜色支持**: 终端颜色输出和格式化
- **文件输出**: 日志文件写入和归档
- **调试工具**: 变量转储和调用栈跟踪
- **模块化**: 支持多个模块的独立日志配置
- **性能优化**: 高效的日志处理和输出

## 核心功能

### 日志器创建

```go
func New(moduleName ...string) *Logger
func NewZLog(out io.Writer, prefix string, flag int, level int, color bool, calldDepth int) *Logger
func CleanLog(log *Logger)
```

### 全局配置

```go
func SetDefault(l *Logger)
func SetLogLevel(level int)
func GetLogLevel() int
func SetPrefix(prefix string)
func SetFlags(flag int)
func GetFlags() int
func AddFlag(flag int)
func ResetFlags(flag int)
func SetFile(filepath string, archive ...bool)
func SetSaveFile(filepath string, archive ...bool)
```

### 颜色控制

```go
func DisableConsoleColor()
func ForceConsoleColor()
func IsSupportColor() bool
func ColorTextWrap(color Color, text string) string
func ColorBackgroundWrap(color Color, backgroundColor Color, text string) string
func OpTextWrap(op Op, text string) string
func OutAllColor()
func GetAllColorText() map[string]Color
func TrimAnsi(str string) string
```

### 日志级别函数

```go
// 调试级别日志，支持格式化
func Debugf(format string, v ...interface{})
// 成功级别日志，支持格式化
func Successf(format string, v ...interface{})
// 信息级别日志，支持格式化
func Infof(format string, v ...interface{})
// 提示级别日志，支持格式化
func Tipsf(format string, v ...interface{})
// 警告级别日志，支持格式化
func Warnf(format string, v ...interface{})
// 错误级别日志，支持格式化
func Errorf(format string, v ...interface{})
// 致命级别日志，支持格式化
func Fatalf(format string, v ...interface{})
// 恐慌级别日志，支持格式化
func Panicf(format string, v ...interface{})
// 打印级别日志，支持格式化
func Printf(format string, v ...interface{})
// 打印级别日志，自动换行
func Println(v ...interface{})
```

### 调试工具

```go
func Dump(v ...interface{})
func Track(v string, i ...int)
func Stack(v interface{})
func Discard()
```

### 日志器方法

```go
func (log *Logger) SetLogLevel(level int)
func (log *Logger) GetLogLevel() int
func (log *Logger) SetPrefix(prefix string)
func (log *Logger) GetPrefix() string
func (log *Logger) ResetFlags(flag int)
func (log *Logger) SetFlags(flag int)
func (log *Logger) GetFlags() int
func (log *Logger) AddFlag(flag int)
func (log *Logger) SetFile(filepath string, archive ...bool)
func (log *Logger) SetSaveFile(filepath string, archive ...bool)
func (log *Logger) SetFormatter(formatter Formatter)
func (log *Logger) SetIgnoreLog(logs ...string)
func (log *Logger) Write(b []byte) (n int, err error)
func (log *Logger) Writer() io.Writer
func (log *Logger) WriteBefore(fn ...func(level int, log string) bool)
func (log *Logger) DisableConsoleColor()
func (log *Logger) ForceConsoleColor()
func (log *Logger) ColorTextWrap(color Color, text string) string
func (log *Logger) ColorBackgroundWrap(color Color, backgroundColor Color, text string) string
func (log *Logger) OpTextWrap(op Op, text string) string
```

### 日志器日志方法

```go
func (log *Logger) Debugf(format string, v ...interface{})
func (log *Logger) Debug(v ...interface{})
func (log *Logger) Successf(format string, v ...interface{})
func (log *Logger) Success(v ...interface{})
func (log *Logger) Infof(format string, v ...interface{})
func (log *Logger) Info(v ...interface{})
func (log *Logger) Tipsf(format string, v ...interface{})
func (log *Logger) Tips(v ...interface{})
func (log *Logger) Warnf(format string, v ...interface{})
func (log *Logger) Warn(v ...interface{})
func (log *Logger) Errorf(format string, v ...interface{})
func (log *Logger) Error(v ...interface{})
func (log *Logger) Fatalf(format string, v ...interface{})
func (log *Logger) Fatal(v ...interface{})
func (log *Logger) Panicf(format string, v ...interface{})
func (log *Logger) Panic(v ...interface{})
func (log *Logger) Printf(format string, v ...interface{})
func (log *Logger) Println(v ...interface{})
func (log *Logger) Dump(v ...interface{})
func (log *Logger) Stack(v interface{})
func (log *Logger) Track(v string, i ...int)
```

## 使用示例

```go
package main

import (
    "fmt"
    "os"
    "time"
    "github.com/sohaha/zlsgo/zlog"
)

func main() {
    // 创建默认日志器
    logger := zlog.New("main")
    
    // 设置日志级别
    logger.SetLogLevel(zlog.LevelDebug)
    
    // 设置前缀
    logger.SetPrefix("[APP]")
    
    // 设置文件输出
    err := logger.SetFile("logs/app.log", true)
    if err != nil {
        fmt.Printf("设置日志文件失败: %v\n", err)
    }
    
    // 基本日志记录
    logger.Info("应用程序启动")
    logger.Debug("调试信息")
    logger.Success("操作成功")
    logger.Warn("警告信息")
    logger.Error("错误信息")
    
    // 格式化日志
    logger.Infof("当前时间: %s", time.Now().Format("2006-01-02 15:04:05"))
    logger.Errorf("发生错误: %v", fmt.Errorf("模拟错误"))
    
    // 变量转储
    user := map[string]interface{}{
        "name": "张三",
        "age":  25,
        "city": "北京",
    }
    logger.Dump(user)
    
    // 调用栈跟踪
    logger.Track("函数调用跟踪", 1)
    
    // 堆栈信息
    logger.Stack("堆栈信息")
    
    // 颜色支持
    if zlog.IsSupportColor() {
        logger.Info("终端支持颜色")
        
        // 使用颜色包装文本
        coloredText := logger.ColorTextWrap(zlog.ColorRed, "红色文本")
        logger.Info(coloredText)
        
        // 背景色
        bgText := logger.ColorBackgroundWrap(zlog.ColorBlue, zlog.ColorWhite, "蓝底白字")
        logger.Info(bgText)
    } else {
        logger.Info("终端不支持颜色")
    }
    
    // 全局配置
    zlog.SetDefault(logger)
    zlog.SetLogLevel(zlog.LevelInfo)
    zlog.SetPrefix("[GLOBAL]")
    
    // 使用全局日志器
    zlog.Info("全局日志信息")
    zlog.Success("全局成功信息")
    
    // 禁用控制台颜色
    zlog.DisableConsoleColor()
    zlog.Info("禁用颜色后的日志")
    
    // 强制控制台颜色
    zlog.ForceConsoleColor()
    zlog.Info("强制颜色后的日志")
    
    // 自定义格式化器
    logger.SetFormatter(func(level int, msg string) string {
        return fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
    })
    
    logger.Info("使用自定义格式化器")
    
    // 写入前处理
    logger.WriteBefore(func(level int, log string) bool {
        fmt.Printf("写入前处理: 级别=%d, 内容=%s\n", level, log)
        return true // 继续写入
    })
    
    logger.Info("测试写入前处理")
    
    // 忽略特定日志
    logger.SetIgnoreLog("debug", "trace")
    logger.Debug("这条日志将被忽略")
    logger.Info("这条日志正常显示")
    
    // 文件输出测试
    fileLogger := zlog.New("file")
    fileLogger.SetFile("logs/test.log", false)
    
    for i := 0; i < 10; i++ {
        fileLogger.Infof("测试日志 %d", i)
        time.Sleep(100 * time.Millisecond)
    }
    
    // 错误处理
    defer func() {
        if r := recover(); r != nil {
            logger.Errorf("程序恢复: %v", r)
        }
    }()
    
    // 模拟 panic
    logger.Panic("模拟程序崩溃")
    
    // 实际应用示例
    // Web 服务器日志
    webLogger := zlog.New("web")
    webLogger.SetFile("logs/web.log", true)
    
    // 模拟请求日志
    webLogger.Infof("收到请求: GET /api/users")
    webLogger.Infof("请求来源: 192.168.1.100")
    webLogger.Infof("用户代理: Mozilla/5.0...")
    
    // 模拟响应日志
    webLogger.Success("请求处理成功")
    webLogger.Infof("响应时间: 45ms")
    webLogger.Infof("响应状态: 200")
    
    // 数据库操作日志
    dbLogger := zlog.New("database")
    dbLogger.SetFile("logs/db.log", true)
    
    dbLogger.Info("连接数据库")
    dbLogger.Debug("执行查询: SELECT * FROM users")
    dbLogger.Success("查询完成，返回 100 条记录")
    
    // 性能监控日志
    perfLogger := zlog.New("performance")
    perfLogger.SetFile("logs/performance.log", true)
    
    start := time.Now()
    time.Sleep(100 * time.Millisecond) // 模拟耗时操作
    duration := time.Since(start)
    
    perfLogger.Infof("操作耗时: %v", duration)
    
    if duration > 50*time.Millisecond {
        perfLogger.Warn("操作耗时较长")
    }
    
    // 错误日志示例
    errorLogger := zlog.New("error")
    errorLogger.SetFile("logs/error.log", true)
    
    // 模拟各种错误
    errorLogger.Error("数据库连接失败")
    errorLogger.Errorf("文件读取错误: %v", os.ErrNotExist)
    errorLogger.Errorf("网络请求超时: %v", "timeout after 30s")
    
    // 调试日志
    debugLogger := zlog.New("debug")
    debugLogger.SetLogLevel(zlog.LevelDebug)
    
    // 变量调试
    config := map[string]interface{}{
        "port":     8080,
        "host":     "localhost",
        "debug":    true,
        "timeout":  30,
        "database": "mysql",
    }
    
    debugLogger.Dump(config)
    
    // 函数调用跟踪
    debugLogger.Track("main函数", 1)
    debugLogger.Track("配置加载", 2)
    debugLogger.Track("服务启动", 3)
    
    // 清理日志
    zlog.CleanLog(logger)
    
    fmt.Println("zlog 模块示例完成")
}
```