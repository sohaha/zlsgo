# ZLSGo 文档中心

欢迎使用 ZLSGo 文档中心。这是一个功能丰富的 Go 语言工具库集合，提供了日常开发中常用的各种功能模块。

## 模块概览

### 核心模块
- [zarray - 数组操作库](zarray.md) - 提供丰富的数组操作方法、泛型支持、高性能的哈希映射、排序映射等功能
- [zcache - 缓存库](zcache.md) - 提供简单缓存、表缓存、快速缓存、文件缓存等功能
- [zcli - 命令行界面库](zcli.md) - 提供命令行解析、参数绑定、服务管理、信号处理等功能
- [zdi - 依赖注入库](zdi.md) - 提供依赖注入容器、类型映射、依赖解析、函数调用注入等功能
- [zerror - 错误处理库](zerror.md) - 提供错误包装、错误码、堆栈跟踪、panic 恢复、错误标签等功能

### 文件和数据模块
- [zfile - 文件操作库](zfile.md) - 提供文件操作、压缩解压、内存文件、文件锁、文件句柄等功能
- [zhttp - HTTP 客户端库](zhttp.md) - 提供HTTP引擎配置、请求方法、响应处理、SSE、HTML解析等功能
- [zjson - JSON 处理库](zjson.md) - 提供JSON解析、查询、设置、格式化、验证、修复、转换等功能
- [ztype - 类型处理库](ztype.md) - 提供灵活的类型转换工具和动态类型系统

### 网络和Web模块
- [znet - Web 框架](znet.md) - 提供HTTP服务器、路由、中间件、模板引擎、SSE、RPC等功能
- [zpool - 资源池管理库](zpool.md) - 提供工作池、负载均衡器、资源注入等功能
- [zpprof - 性能分析库](zpprof.md) - 提供pprof处理器注册、系统信息收集等功能

### 工具和辅助模块
- [zlog - 日志库](zlog.md) - 提供丰富的日志功能、颜色支持、文件输出、调试工具等
- [zreflect - 反射工具库](zreflect.md) - 提供反射操作、方法调用、字段访问、类型检查等功能的便捷封装
- [zshell - Shell 命令执行库](zshell.md) - 提供跨平台的命令执行、管道操作、后台运行、回调处理等功能
- [zstring - 字符串处理库](zstring.md) - 提供字符串操作、正则表达式、加密解密、编码解码、随机生成、模板处理等功能
- [zsync - 同步原语库](zsync.md) - 扩展了标准库的sync包，提供额外的功能和优化
- [ztime - 时间处理库](ztime.md) - 提供时间获取、格式化、计算、转换、时区管理、定时任务等功能
- [zutil - 通用工具库](zutil.md) - 提供反射工具、原子操作、重试机制、通道管理、缓冲区池、环境变量、参数解析、工具函数、Once 模式、选项模式等功能
- [zvalid - 数据验证库](zvalid.md) - 提供灵活的验证规则链、多种验证方法、自定义验证函数等功能
- [zlocale - 国际化库](zlocale.md) - 提供多语言支持、参数化翻译、缓存机制和回退策略等功能

## 快速开始

### 安装
```bash
go get github.com/sohaha/zlsgo
```

### 基本使用
```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zstring"
    "github.com/sohaha/zlsgo/zarray"
    "github.com/sohaha/zlsgo/zlocale"
)

func main() {
    // 字符串处理
    result := zstring.Ucfirst("hello world")
    fmt.Println(result) // "Hello world"

    // 数组操作
    arr := zarray.NewArray(5)
    arr.Push("张三", "李四", "王五")
    fmt.Printf("数组长度: %d\n", arr.Length())

    // 国际化翻译
    zlocale.LoadLanguage("zh-CN", "简体中文", map[string]string{
        "welcome": "欢迎",
        "user.name": "用户: {0}",
    })
    zlocale.SetLanguage("zh-CN")
    fmt.Println(zlocale.T("welcome")) // "欢迎"
    fmt.Println(zlocale.T("user.name", "张三")) // "用户: 张三"
}
```

## 特性

- **高性能**: 所有模块都经过性能优化，适合生产环境使用
- **类型安全**: 充分利用 Go 语言的类型系统，提供类型安全的API
- **易于使用**: 简洁的API设计，降低学习成本
- **功能丰富**: 覆盖日常开发中的大部分需求
- **文档完善**: 每个模块都有详细的使用文档和示例代码
- **持续维护**: 活跃的开发和维护，定期更新和修复

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个项目。

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](../LICENSE) 文件。

## 相关链接

- [GitHub 仓库](https://github.com/sohaha/zlsgo)
- [Go 模块](https://pkg.go.dev/github.com/sohaha/zlsgo)
- [在线文档](https://docs.73zls.com/zlsgo/)