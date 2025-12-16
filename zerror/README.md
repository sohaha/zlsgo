# zerror 模块

`zerror` 提供了错误包装、错误码、堆栈跟踪、panic 恢复、错误标签、自定义格式化、错误链等功能，用于高效的错误处理和调试。

## 功能概览

- **错误包装**: 错误信息包装和增强
- **错误码**: 标准化的错误码系统
- **堆栈跟踪**: 详细的错误堆栈信息
- **Panic 恢复**: panic 异常捕获和恢复
- **错误标签**: 错误分类和标签管理
- **自定义格式化**: 灵活的错误格式化
- **错误链**: 错误链式传递和处理

## 核心功能

### 错误创建与包装

```go
// 创建新的错误
func New(code ErrCode, message string, w ...External) error
// 包装现有错误
func Wrap(err error, code ErrCode, text string, w ...External) error
// 包装错误并添加上下文
func With(err error, text string, w ...External) error
// 重用错误对象
func Reuse(err error) error
```

### 错误码操作

```go
func Unwrap(err error, code ErrCode) (error, bool)
func Is(err error, code ...ErrCode) bool
func UnwrapCode(err error) (ErrCode, bool)
func UnwrapCodes(err error) (codes []ErrCode)
func UnwrapErrors(err error) (errs []string)
func UnwrapFirst(err error) (ferr error)
func UnwrapFirstCode(err error) (code ErrCode)
```

### 错误标签

```go
func WrapTag(tag TagKind) External
func GetTag(err error) TagKind
func (t TagKind) Wrap(err error, text string) error
func (t TagKind) Text(text string) error
```

### Panic 恢复

```go
func TryCatch(fn func() error) (err error)
func Panic(err error)
```

### 错误格式化

```go
func (e *Error) Error() string
func (e *Error) Unwrap() error
func (e *Error) Format(s fmt.State, verb rune)
func (e *Error) Stack() string
```

## 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zerror"
)

func main() {
    // 基本错误创建示例
    // 创建新错误
    err := zerror.New(500, "服务器内部错误")
    fmt.Printf("新错误: %v\n", err)
    
    // 重用错误
    reusedErr := zerror.Reuse(err)
    fmt.Printf("重用错误: %v\n", reusedErr)
    
    // 错误包装示例
    originalErr := fmt.Errorf("数据库连接失败")
    
    // 包装错误
    wrappedErr := zerror.Wrap(originalErr, 500, "服务调用失败")
    fmt.Printf("包装后的错误: %v\n", wrappedErr)
    
    // 添加额外文本
    errWithText := zerror.SupText(wrappedErr, "用户操作失败")
    fmt.Printf("带额外文本的错误: %v\n", errWithText)
    
    // 使用 With 包装
    errWith := zerror.With(wrappedErr, "请求处理失败")
    fmt.Printf("With 包装的错误: %v\n", errWith)
    
    // 错误码操作示例
    // 检查错误码
    if zerror.Is(err, 500) {
        fmt.Println("这是一个 500 错误")
    }
    
    // 解包特定错误码
    if unwrappedErr, ok := zerror.Unwrap(err, 500); ok {
        fmt.Printf("解包的 500 错误: %v\n", unwrappedErr)
    }
    
    // 获取错误码
    if code, ok := zerror.UnwrapCode(err); ok {
        fmt.Printf("错误码: %d\n", code)
    }
    
    // 获取所有错误码
    codes := zerror.UnwrapCodes(err)
    fmt.Printf("所有错误码: %v\n", codes)
    
    // 获取所有错误消息
    errors := zerror.UnwrapErrors(err)
    fmt.Printf("所有错误消息: %v\n", errors)
    
    // 获取第一个错误
    firstErr := zerror.UnwrapFirst(err)
    fmt.Printf("第一个错误: %v\n", firstErr)
    
    // 获取第一个错误码
    firstCode := zerror.UnwrapFirstCode(err)
    fmt.Printf("第一个错误码: %d\n", firstCode)
    
    // 错误标签示例
    // 使用预定义标签
    // 使用标签包装错误
    dbErr := zerror.Wrap(fmt.Errorf("连接超时"), 500, "数据库操作失败", zerror.WrapTag(zerror.Internal))
    fmt.Printf("内部错误标签: %v\n", dbErr)
    
    // 获取错误标签
    tag := zerror.GetTag(dbErr)
    fmt.Printf("错误标签: %s\n", tag)
    
    // 使用标签创建错误
    networkErr := zerror.InvalidInput.Wrap(fmt.Errorf("网络不可达"), "网络连接失败")
    fmt.Printf("无效输入标签错误: %v\n", networkErr)
    
    // 使用标签创建纯文本错误
    authErr := zerror.Unauthorized.Text("认证失败")
    fmt.Printf("未授权标签错误: %v\n", authErr)
    
    // Panic 恢复示例
    // TryCatch 示例
    err = zerror.TryCatch(func() error {
        // 模拟可能出错的函数
        if time.Now().Second()%2 == 0 {
            return fmt.Errorf("随机错误")
        }
        return nil
    })
    
    if err != nil {
        fmt.Printf("TryCatch 捕获到错误: %v\n", err)
    }
    
    // Panic 示例
    // zerror.Panic(fmt.Errorf("这是一个 panic 错误"))
    
    // 错误链示例
    // 创建错误链
    chainErr := zerror.New(400, "请求参数错误")
    chainErr = zerror.Wrap(chainErr, 500, "服务处理失败")
    chainErr = zerror.Wrap(chainErr, 502, "网关错误")
    
    fmt.Printf("错误链: %v\n", chainErr)
    
    // 获取错误链中的所有错误码
    allCodes := zerror.UnwrapCodes(chainErr)
    fmt.Printf("错误链中的所有错误码: %v\n", allCodes)
    
    // 获取错误链中的所有错误消息
    allErrors := zerror.UnwrapErrors(chainErr)
    fmt.Printf("错误链中的所有错误消息: %v\n", allErrors)
    
    // 检查特定错误码
    if zerror.Is(chainErr, 400, 500, 502) {
        fmt.Println("错误链包含指定的错误码")
    }
    
    // 错误格式化示例
    // 创建带堆栈的错误
    stackErr := zerror.New(500, "运行时错误")
    
    // 获取错误堆栈
    if errorWithStack, ok := stackErr.(*zerror.Error); ok {
        stack := errorWithStack.Stack()
        fmt.Printf("错误堆栈:\n%s\n", stack)
    }
    
    // 实际应用示例
    // API 错误处理
    type APIError struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
        Details string `json:"details,omitempty"`
    }
    
    // 处理数据库错误
    func handleDatabaseError(err error) *APIError {
        if zerror.Is(err, 500) {
            return &APIError{
                Code:    500,
                Message: "数据库操作失败",
                Details: err.Error(),
            }
        }
        
        return &APIError{
            Code:    500,
            Message: "未知数据库错误",
            Details: err.Error(),
        }
    }
    
    // 处理网络错误
    func handleNetworkError(err error) *APIError {
        if zerror.Is(err, 502) {
            return &APIError{
                Code:    502,
                Message: "网络连接失败",
                Details: err.Error(),
            }
        }
        
        return &APIError{
            Code:    502,
            Message: "网络服务不可用",
            Details: err.Error(),
        }
    }
    
    // 模拟错误处理
    dbError := zerror.Wrap(fmt.Errorf("connection refused"), 500, "数据库连接失败", zerror.WrapTag(zerror.Internal))
    apiErr := handleDatabaseError(dbError)
    fmt.Printf("API 错误: %+v\n", apiErr)
    
    networkError := zerror.Wrap(fmt.Errorf("timeout"), 502, "网络超时", zerror.WrapTag(zerror.InvalidInput))
    apiErr = handleNetworkError(networkError)
    fmt.Printf("API 错误: %+v\n", apiErr)
    
    // 错误分类和统计
    var databaseErrors, networkErrors, authErrors int
    
    errors := []error{dbError, networkError, authErr}
    
    for _, err := range errors {
        tag := zerror.GetTag(err)
        switch tag {
        case zerror.Internal:
            databaseErrors++
        case zerror.InvalidInput:
            networkErrors++
        case zerror.Unauthorized:
            authErrors++
        }
    }
    
    fmt.Printf("错误统计 - 数据库: %d, 网络: %d, 认证: %d\n", 
        databaseErrors, networkErrors, authErrors)
    
    // 错误恢复策略
    func retryOperation(operation func() error, maxRetries int) error {
        var lastErr error
        
        for i := 0; i < maxRetries; i++ {
            err := operation()
            if err == nil {
                return nil
            }
            
            lastErr = err
            
            // 检查是否是可重试的错误
            if zerror.Is(err, 500) {
                fmt.Printf("尝试 %d/%d 失败，准备重试...\n", i+1, maxRetries)
                time.Sleep(time.Duration(i+1) * time.Second)
                continue
            }
            
            // 不可重试的错误，直接返回
            break
        }
        
        return zerror.Wrap(lastErr, 500, "操作重试失败")
    }
    
    // 模拟重试操作
    operation := func() error {
        if time.Now().Second()%3 == 0 {
            return zerror.New(500, "临时错误")
        }
        return nil
    }
    
    err = retryOperation(operation, 3)
    if err != nil {
        fmt.Printf("重试操作失败: %v\n", err)
    } else {
        fmt.Println("重试操作成功")
    }
    
    fmt.Println("错误处理示例完成")
}
```

## 错误码定义

### 标准错误码
```go
type ErrCode int

const (
    // 客户端错误 (4xx)
    ErrBadRequest          ErrCode = 400
    ErrUnauthorized        ErrCode = 401
    ErrForbidden           ErrCode = 403
    ErrNotFound            ErrCode = 404
    ErrMethodNotAllowed    ErrCode = 405
    ErrRequestTimeout      ErrCode = 408
    ErrConflict            ErrCode = 409
    ErrTooManyRequests     ErrCode = 429
    
    // 服务器错误 (5xx)
    ErrInternalServer      ErrCode = 500
    ErrNotImplemented      ErrCode = 501
    ErrBadGateway          ErrCode = 502
    ErrServiceUnavailable  ErrCode = 503
    ErrGatewayTimeout      ErrCode = 504
)
```

### 自定义错误码
```go
const (
    // 业务错误码 (1000+)
    ErrUserNotFound        ErrCode = 1001
    ErrInvalidPassword     ErrCode = 1002
    ErrUserExists          ErrCode = 1003
    ErrDatabaseError       ErrCode = 2001
    ErrNetworkError        ErrCode = 3001
)
```

## 错误标签系统

### 标签类型
```go
type TagKind string

const (
    None             TagKind = ""
    Internal         TagKind = "INTERNAL"
    Cancelled        TagKind = "CANCELLED"
    InvalidInput     TagKind = "INVALID_INPUT"
    NotFound         TagKind = "NOT_FOUND"
    PermissionDenied TagKind = "PERMISSION_DENIED"
    Unauthorized     TagKind = "UNAUTHORIZED"
)
```

### 标签操作
- **WrapTag**: 创建标签包装器
- **GetTag**: 获取错误标签
- **TagKind.Wrap**: 使用标签包装错误
- **TagKind.Text**: 使用标签创建文本错误

## 错误格式化

### 格式化选项
```go
// 支持标准 fmt 格式化
%v  // 默认格式
%s  // 字符串格式
%+v // 详细格式（包含堆栈）
```

### 堆栈跟踪
- **Stack()**: 获取错误堆栈信息
- **Format()**: 自定义格式化输出
- **Unwrap()**: 获取原始错误

## 最佳实践

1. 合理使用错误码
2. 正确处理错误链
3. 避免过度包装
4. 监控错误性能
