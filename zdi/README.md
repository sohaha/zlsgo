# zdi 模块

`zdi` 提供了依赖注入容器、类型映射、依赖解析、函数调用注入、结构体字段注入等功能，用于实现松耦合的依赖管理。

## 功能概览

- **依赖注入容器**: 核心依赖注入容器
- **类型映射**: 接口到实现的映射
- **依赖解析**: 依赖解析和实例化
- **函数调用**: 函数参数的自动注入
- **结构体注入**: 结构体字段的自动注入
- **层次化注入**: 支持父子注入器的层次结构
- **命名注入**: 支持按名称进行注入和检索
- **延迟注入**: 支持服务的延迟加载（懒加载）

## 核心功能

### 依赖注入容器 (Injector)

`Injector` 是依赖注入容器的核心接口。

```go
// Injector 定义了依赖注入容器的接口
type Injector interface {
    // 映射值到容器
    Map(val interface{}, opt ...Option) reflect.Type
    // 映射多个值到容器
    Maps(values ...interface{}) []reflect.Type
    // 设置类型和值的映射
    Set(typ reflect.Type, val reflect.Value)
    // 获取类型对应的值
    Get(t reflect.Type) (reflect.Value, bool)
    // 调用函数并注入依赖
    Invoke(f interface{}) ([]reflect.Value, error)
    // 只检查错误的调用
    InvokeWithErrorOnly(f interface{}) error
    // 应用依赖到指针
    Apply(p Pointer) error
    // 解析依赖到指针
    Resolve(v ...Pointer) error
    // 设置父注入器
    SetParent(parent Injector)
}

// 创建新的注入器
func New(parent ...Injector) Injector
```

### 类型映射与选项

```go
type Pointer interface{}
type Option func(o *option)

// 使用接口指针映射
func WithInterface(ifacePtr Pointer) Option
// 使用名称映射
func WithName(name string) Option
```

### 命名注入

```go
func GetNamed(inj Injector, name string, t reflect.Type) (reflect.Value, bool)
```

### 延迟注入 (Lazy Injection)

```go
func Lazy(getter interface{}) interface{}
func IsLazy(v interface{}) bool
```

### 预调用优化

```go
type PreInvoker interface {
    Invoke([]interface{}) ([]reflect.Value, error)
}

func IsPreInvoker(handler interface{}) bool
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zdi"
)

// 定义接口
type Database interface {
    Connect() string
}

type Cache interface {
    Get(key string) string
}

type Logger interface {
    Log(message string)
}

// 实现接口
type MySQLDatabase struct{}

func (db *MySQLDatabase) Connect() string {
    return "MySQL 数据库已连接"
}

type RedisCache struct{}

func (db *RedisCache) Get(key string) string {
    return fmt.Sprintf("从 Redis 获取: %s", key)
}

type ConsoleLogger struct{}

func (l *ConsoleLogger) Log(message string) {
    fmt.Printf("日志: %s\n", message)
}

// 服务结构体
type UserService struct {
    DB    Database `di:""`
    Cache Cache   `di:""`
    Log   Logger  `di:""`
}

func (s *UserService) GetUser(id string) string {
    s.Log.Log(fmt.Sprintf("获取用户: %s", id))
    result := s.DB.Connect()
    cached := s.Cache.Get("user:" + id)
    return fmt.Sprintf("用户服务: %s, 缓存: %s", result, cached)
}

// 业务逻辑函数
func ProcessUser(db Database, cache Cache, logger Logger, userID string) string {
    logger.Log("开始处理用户")
    result := fmt.Sprintf("处理用户 %s: %s, %s", userID, db.Connect(), cache.Get("user:"+userID))
    logger.Log("用户处理完成")
    return result
}

// 预调用处理器
type OptimizedHandler struct{}

func (h *OptimizedHandler) Invoke(args []interface{}) ([]reflect.Value, error) {
    db := args[0].(Database)
    cache := args[1].(Cache)
    logger := args[2].(Logger)
    userID := args[3].(string)
    
    result := fmt.Sprintf("优化处理用户 %s: %s, %s", userID, db.Connect(), cache.Get("user:"+userID))
    return []reflect.Value{reflect.ValueOf(result)}, nil
}

func main() {
    // 创建注入器
    injector := zdi.New()
    
    // 映射接口到实现
    injector.Map(&MySQLDatabase{})
    injector.Map(&RedisCache{})
    injector.Map(&ConsoleLogger{})
    
    // 使用 WithInterface 选项映射到接口
    injector.Map(&MySQLDatabase{}, zdi.WithInterface((*Database)(nil)))
    injector.Map(&RedisCache{}, zdi.WithInterface((*Cache)(nil)))
    injector.Map(&ConsoleLogger{}, zdi.WithInterface((*Logger)(nil)))
    
    // 映射多个值
    injector.Maps(
        &MySQLDatabase{},
        &RedisCache{},
        &ConsoleLogger{},
    )
    
    // 解析依赖到指针
    var db Database
    var cache Cache
    var logger Logger
    
    err := injector.Resolve(&db, &cache, &logger)
    if err != nil {
        fmt.Printf("依赖解析失败: %v\n", err)
        return
    }
    
    fmt.Println(db.Connect())
    fmt.Println(cache.Get("test"))
    logger.Log("依赖解析成功")
    
    // 调用函数
    results, err := injector.Invoke(ProcessUser, "user123")
    if err != nil {
        fmt.Printf("函数调用失败: %v\n", err)
        return
    }
    
    if len(results) > 0 {
        fmt.Printf("函数调用结果: %v\n", results[0].Interface())
    }
    
    // 只检查错误的调用
    err = injector.InvokeWithErrorOnly(ProcessUser, "user456")
    if err != nil {
        fmt.Printf("调用错误: %v\n", err)
    }
    
    // 结构体注入
    userService := &UserService{}
    err = injector.Apply(userService)
    if err != nil {
        fmt.Printf("结构体注入失败: %v\n", err)
        return
    }
    
    result := userService.GetUser("user789")
    fmt.Println(result)
    
    // 创建层次化注入器
    parentInjector := zdi.New()
    childInjector := zdi.New(parentInjector)
    
    // 在父注入器中映射共享依赖
    parentInjector.Map(&ConsoleLogger{})
    
    // 在子注入器中映射特定依赖
    childInjector.Map(&MySQLDatabase{})
    
    // 子注入器可以访问父注入器的依赖
    var sharedLogger Logger
    err = childInjector.Resolve(&sharedLogger)
    if err == nil {
        sharedLogger.Log("从父注入器获取的日志器")
    }
    
    // 预调用优化
    if zdi.IsPreInvoker(&OptimizedHandler{}) {
        fmt.Println("支持预调用优化")
        
        // 调用预调用处理器
        results, err := injector.Invoke(&OptimizedHandler{}, "user999")
        if err == nil && len(results) > 0 {
            fmt.Printf("预调用结果: %v\n", results[0].Interface())
        }
    }
    
    // 实际应用示例
    // 配置服务
    type ConfigService struct {
        Config map[string]interface{}:
    }
    
    type AppService struct {
        Config *ConfigService `di:""`
        DB     Database      `di:""`
        Cache  Cache         `di:""`
    }
    
    // 设置配置
    configService := &ConfigService{
        Config: map[string]interface{}{
            "database": "mysql",
            "cache":    "redis",
            "port":     3306,
        },
    }
    
    injector.Map(configService)
    
    appService := &AppService{}
    injector.Apply(appService)
    
    fmt.Printf("应用服务配置: %+v\n", appService.Config.Config)
    fmt.Printf("应用服务数据库: %s\n", appService.DB.Connect())
    fmt.Printf("应用服务缓存: %s\n", appService.Cache.Get("app:config"))
}
```

## 依赖注入模式

### 构造函数注入
```go
type Service struct {
    db Database
}

func NewService(db Database) *Service {
    return &Service{db: db}
}
```

### 属性注入
```go
type Service struct {
    DB Database `di:""`
}
```

### 方法注入
```go
func (s *Service) Process(db Database) {
    // 使用方法注入的依赖
}
```

### 接口注入
```go
type ServiceInterface interface {
    SetDatabase(db Database)
}
```

## 最佳实践

1. 避免循环依赖
2. 合理使用层次化注入器
3. 正确处理错误
4. 注意内存泄漏