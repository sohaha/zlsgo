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
    Invoker
    TypeMapper
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

// Invoker 提供调用与解析能力
type Invoker interface {
    Apply(Pointer) error
    Resolve(...Pointer) error
    Invoke(interface{}) ([]reflect.Value, error)
    InvokeWithErrorOnly(interface{}) error
}

// TypeMapper 提供注册能力
type TypeMapper interface {
    Map(interface{}, ...Option) reflect.Type
    Maps(...interface{}) []reflect.Type
    Provide(interface{}, ...Option) []reflect.Type
}

// 创建新的注入器
func New(parent ...Injector) Injector
```

### 类型映射与选项

```go
type Pointer interface{}
type Option func(*mapOption)

// 使用接口指针映射
func WithInterface(ifacePtr Pointer) Option
// 使用名称映射
func Name(name string) Option
```

### 命名注入

```go
func GetNamed(inj Injector, name string, t reflect.Type) (reflect.Value, bool)
func GetNamedErr(inj Injector, name string, t reflect.Type) (reflect.Value, bool, error)
```

### 泛型解析

```go
func Resolve[T any](inj Injector) (T, error)
func ResolveNamed[T any](inj Injector, name string) (T, error)
```

### 预调用优化

```go
type PreInvoker interface {
    Invoke([]interface{}) ([]reflect.Value, error)
}

func IsPreInvoker(handler interface{}) bool
```

### 构造函数注入 (Provide)

`Provide` 注册构造函数，构造出来的值按返回类型注册为可注入依赖：

```go
func (inj *injector) Provide(provider interface{}, opt ...Option) []reflect.Type
```

```go
di := zdi.New()

// 返回 (T, error)：error 是构造控制流，不会被注册成依赖
di.Provide(func() (*DB, error) {
    return openDB()
})

// 返回 (T1, T2)：同时提供两个依赖
di.Provide(func() (*DB, *Cache) {
    return newDB(), newCache()
})

// 命名注册：不参与按类型解析，只能用 Name 取
di.Provide(func() (*DB, error) { return openPrimary() }, zdi.Name("primary"))

var db *DB
if err := di.Resolve(&db); err != nil {
    // 构造函数返回的错误会原样出现在这里
}
```

## error 语义

`zdi` 把 `error` 当构造控制流，而不是可注入的值，规则如下：

| 规则 | 行为 |
|---|---|
| `error` 永不成为依赖 key | `Provide` 时跳过最后一个 `error` 返回值，不会出现「谁的 error 覆盖谁」 |
| 构造失败对调用方可见 | `Resolve` / `Apply` / `Invoke` 返回该错误；`Get` 返回 `ok=false` |
| 失败不缓存、可重试 | 失败时不写入任何值、不删除 provider，下次调用重新执行构造函数 |
| 构造恐慌不逃逸 | provider 内部 panic 被转成 error 返回 |
| 多返回值全成功才入库 | 先完整检查 error，再一次性写入，不会留下半成品单例 |
| 循环依赖返回错误 | 解析链上重复出现同一类型时返回 circular dependency 错误，而不是死锁或栈溢出 |
| 重入自身不会死锁 | provider 内部再次请求自己（按类型或按名称）会立即报错，不会永久阻塞 |

并发请求同一依赖时只会执行一次构造函数，其余调用者等待并共享结果（失败也共享同一个 error）。

因此 `Provide` 需要满足这些约束，否则直接 panic（属于编程错误，不是运行时错误）：

- 必须是函数，且至少有一个返回值
- 若有 `error` 返回值，必须是最后一个，且只能有一个
- `func() error` 不是合法 provider，请用 `InvokeWithErrorOnly`

需要拿到构造错误又只用底层接口时，用 `GetErr` / `GetNamedErr`：

```go
func GetErr(inj Injector, t reflect.Type) (reflect.Value, bool, error)

v, ok, err := zdi.GetErr(di, reflect.TypeOf((*DB)(nil)))
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
        Config map[string]interface{}
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

1. 避免循环依赖，循环引用会在解析时返回错误
2. 合理使用层次化注入器
3. **始终检查 `Resolve` / `Apply` / `Invoke` 返回的 error**，构造函数的失败就是通过它暴露的
4. 构造函数保持无副作用可重试，因为失败后会在下次解析时重新执行
5. 需要按名称区分同一类型的多个实例时用 `zdi.Name`
