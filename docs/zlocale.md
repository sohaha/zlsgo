# zlocale 模块

`zlocale` 是一个 Go 语言国际化（i18n）库，提供了多语言支持、参数化翻译、缓存机制和回退策略等功能。

## 功能概览

- **多语言支持**: 支持加载和管理多种语言的翻译数据
- **参数化翻译**: 支持 printf 风格和占位符风格的参数替换
- **高性能缓存**: 提供两种缓存系统（FastCache 和 Legacy）以优化模板处理性能
- **线程安全**: 使用读偏向的互斥锁保证并发安全
- **回退机制**: 当翻译不存在时自动回退到默认语言
- **内存管理**: 提供缓存统计和内存使用监控功能
- **加载器模式**: 支持从映射表和自定义源加载翻译数据

## 核心功能

### I18n 实例管理

```go
// 创建新的 I18n 实例，指定默认语言
func New(defaultLang string) *I18n

// 创建默认的 I18n 实例（默认语言为 "en"）
func NewDefault() *I18n

// 加载语言翻译数据
func (i *I18n) LoadLanguage(langCode, langName string, data map[string]string) error

// 使用配置加载语言翻译数据
func (i *I18n) LoadLanguageWithConfig(langCode, langName string, data map[string]string, cache TemplateCache) error

// 设置当前活跃语言
func (i *I18n) SetLanguage(langCode string) error

// 获取当前活跃语言代码
func (i *I18n) GetLanguage() string

// 获取已加载的语言列表
func (i *I18n) GetLoadedLanguages() map[string]string
```

### 翻译功能

```go
// 使用当前语言翻译
func (i *I18n) T(key string, args ...interface{}) string

// 使用指定语言翻译
func (i *I18n) TWithLang(langCode, key string, args ...interface{}) string

// 检查语言是否已加载
func (i *I18n) HasLanguage(langCode string) bool

// 检查翻译键是否存在
func (i *I18n) HasKey(langCode, key string) bool

// 移除语言
func (i *I18n) RemoveLanguage(langCode string) error
```

### 缓存和内存管理

```go
// 获取内存使用统计
func (i *I18n) GetMemoryUsage() map[string]interface{}

// 获取详细缓存统计
func (i *I18n) GetCacheStats() map[string]CacheStats

// 清理模板缓存
func (i *I18n) ClearTemplateCache()

// 关闭所有缓存资源
func (i *I18n) Close()
```

### 全局便捷函数

```go
// 加载语言到全局实例
func LoadLanguage(langCode, langName string, data map[string]string) error

// 设置全局实例的活跃语言
func SetLanguage(langCode string) error

// 获取全局实例的当前活跃语言
func GetLanguage() string

// 使用全局实例翻译
func T(key string, args ...interface{}) string

// 使用指定语言在全局实例中翻译
func TWithLang(langCode, key string, args ...interface{}) string

// 检查全局实例中是否已加载语言
func HasLanguage(langCode string) bool

// 检查全局实例中翻译键是否存在
func HasKey(langCode, key string) bool

// 获取全局实例中已加载的语言
func GetLoadedLanguages() map[string]string
```

### 加载器功能

```go
// 创建新的加载器
func NewLoader(i18n *I18n) *Loader

// 从映射表加载翻译
func (l *Loader) LoadTranslationsFromMap(langCode, langName string, translations map[string]string) error

// 添加自定义翻译
func (l *Loader) AddCustomTranslation(langCode, key, value string) error

// 获取可用语言列表
func (l *Loader) GetAvailableLanguages() []string

// 全局便捷函数
func LoadTranslationsFromMap(langCode, langName string, translations map[string]string) error
func AddCustomTranslation(langCode, key, value string) error
```

### 缓存接口

```go
// 模板缓存接口
type TemplateCache interface {
    Get(key string) (*TemplateCacheEntry, bool)
    Set(key string, template *TemplateCacheEntry)
    Delete(key string)
    Clear()
    Count() int
    Stats() CacheStats
    Close()
}

// 模板缓存条目
type TemplateCacheEntry struct {
    Template interface{}
    Created  time.Time
    Accessed time.Time
    Hits     int64
}

// 缓存统计信息
type CacheStats struct {
    TotalItems       int
    HitCount         int64
    MissCount        int64
    HitRate          float64
    MemoryUsage      int64
    EvictionCount    int64
    CleanerLevel     int32
    IdleDuration     time.Duration
    IsCleanerRunning bool
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zlocale"
)

func main() {
    // 创建 I18n 实例
    i18n := zlocale.New("en")

    // 加载英文翻译
    enData := map[string]string{
        "welcome":      "Welcome",
        "goodbye":      "Goodbye",
        "user.profile": "User Profile",
        "hello.name":   "Hello, {0}!",
        "items.count":  "You have %d items",
    }
    err := i18n.LoadLanguage("en", "English", enData)
    if err != nil {
        panic(err)
    }

    // 加载中文翻译
    zhData := map[string]string{
        "welcome":      "欢迎",
        "goodbye":      "再见",
        "user.profile": "用户资料",
        "hello.name":   "你好，{0}！",
        "items.count":  "你有 %d 个项目",
    }
    err = i18n.LoadLanguage("zh-CN", "简体中文", zhData)
    if err != nil {
        panic(err)
    }

    // 使用英文翻译
    i18n.SetLanguage("en")
    fmt.Println(i18n.T("welcome"))           // "Welcome"
    fmt.Println(i18n.T("hello.name", "张三")) // "Hello, 张三!"
    fmt.Println(i18n.T("items.count", 5))    // "You have 5 items"

    // 切换到中文
    i18n.SetLanguage("zh-CN")
    fmt.Println(i18n.T("welcome"))           // "欢迎"
    fmt.Println(i18n.T("hello.name", "张三")) // "你好，张三！"
    fmt.Println(i18n.T("items.count", 5))    // "你有 5 个项目"

    // 使用指定语言翻译
    fmt.Println(i18n.TWithLang("en", "welcome"))  // "Welcome"
    fmt.Println(i18n.TWithLang("zh-CN", "welcome")) // "欢迎"
}
```

### 使用全局便捷函数

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zlocale"
)

func main() {
    // 加载翻译到全局实例
    enData := map[string]string{
        "app.title": "My Application",
        "app.version": "Version {0}",
    }
    zlocale.LoadLanguage("en", "English", enData)

    zhData := map[string]string{
        "app.title": "我的应用程序",
        "app.version": "版本 {0}",
    }
    zlocale.LoadLanguage("zh-CN", "简体中文", zhData)

    // 设置当前语言
    zlocale.SetLanguage("zh-CN")

    // 使用全局函数翻译
    fmt.Println(zlocale.T("app.title"))           // "我的应用程序"
    fmt.Println(zlocale.T("app.version", "1.0.0")) // "版本 1.0.0"

    // 检查语言和键
    if zlocale.HasLanguage("en") {
        fmt.Println("English is loaded")
    }

    if zlocale.HasKey("zh-CN", "app.title") {
        fmt.Println("Chinese app.title exists")
    }

    // 获取已加载的语言
    languages := zlocale.GetLoadedLanguages()
    for code, name := range languages {
        fmt.Printf("Language: %s (%s)\n", code, name)
    }
}
```

### 使用加载器

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zlocale"
)

func main() {
    i18n := zlocale.New("en")
    loader := zlocale.NewLoader(i18n)

    // 使用加载器批量加载翻译
    translations := map[string]string{
        "button.save":   "Save",
        "button.cancel": "Cancel",
        "error.required": "This field is required",
    }

    err := loader.LoadTranslationsFromMap("en", "English", translations)
    if err != nil {
        panic(err)
    }

    // 添加自定义翻译
    err = loader.AddCustomTranslation("en", "custom.welcome", "Welcome to our app!")
    if err != nil {
        panic(err)
    }

    // 获取可用语言
    languages := loader.GetAvailableLanguages()
    fmt.Printf("Available languages: %v\n", languages)

    i18n.SetLanguage("en")
    fmt.Println(i18n.T("custom.welcome")) // "Welcome to our app!"
}
```

### 高级缓存配置

```go
package main

import (
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zlocale"
)

func main() {
    i18n := zlocale.New("en")

    // 使用传统缓存系统（向后兼容）
    translations := map[string]string{
        "template.test": "Hello {0}, today is {1}",
    }

    err := i18n.LoadLanguageWithConfig("en", "English", translations, i18n.NewLegacyCacheAdapter(500)) // 最大500个模板
    if err != nil {
        panic(err)
    }

    i18n.SetLanguage("en")
    fmt.Println(i18n.T("template.test", "Alice", "Monday"))

    // 获取内存使用统计
    stats := i18n.GetMemoryUsage()
    fmt.Printf("Memory usage: %+v\n", stats)

    // 获取缓存统计
    cacheStats := i18n.GetCacheStats()
    for lang, stats := range cacheStats {
        fmt.Printf("Cache stats for %s: %+v\n", lang, stats)
    }

    // 清理缓存
    i18n.ClearTemplateCache()
    fmt.Println("Template cache cleared")

    // 关闭实例
    defer i18n.Close()
}
```

### 错误处理和回退机制

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zlocale"
)

func main() {
    i18n := zlocale.New("en") // 设置默认语言为英语

    // 只加载部分中文翻译
    zhData := map[string]string{
        "welcome": "欢迎",
        // 故意不加载 "goodbye" 键
    }
    i18n.LoadLanguage("zh-CN", "简体中文", zhData)

    // 英文翻译（完整）
    enData := map[string]string{
        "welcome": "Welcome",
        "goodbye": "Goodbye",
        "hello":   "Hello, {0}!",
    }
    i18n.LoadLanguage("en", "English", enData)

    // 切换到中文
    i18n.SetLanguage("zh-CN")

    // 存在的翻译
    fmt.Println(i18n.T("welcome")) // "欢迎"

    // 不存在的翻译，回退到默认语言（英语）
    fmt.Println(i18n.T("goodbye")) // "Goodbye" (从英语回退)
    fmt.Println(i18n.T("hello", "World")) // "Hello, World!" (从英语回退)

    // 完全不存在的键，返回键本身
    fmt.Println(i18n.T("nonexistent.key")) // "nonexistent.key"

    // 检查键是否存在
    fmt.Printf("Chinese 'goodbye' exists: %v\n", i18n.HasKey("zh-CN", "goodbye")) // false
    fmt.Printf("English 'goodbye' exists: %v\n", i18n.HasKey("en", "goodbye"))    // true
}
```

## 最佳实践

### 1. 语言文件组织

建议按模块组织翻译数据：

```go
var userTranslations = map[string]string{
    "profile.title":    "User Profile",
    "profile.edit":     "Edit Profile",
    "profile.save":     "Save Changes",
    "profile.email":    "Email Address",
    "profile.password": "Password",
}

var errorTranslations = map[string]string{
    "required":      "This field is required",
    "invalid.email": "Please enter a valid email address",
    "min.length":    "Must be at least {0} characters",
}
```

### 2. 键命名规范

使用点号分隔的层次结构：

- `module.section.action` - 模块.部分.操作
- `error.type.code` - 错误.类型.代码
- `button.action` - 按钮.动作

### 3. 参数化翻译

优先使用占位符风格 `{0}`, `{1}`，因为它们更直观：

```go
// 推荐
"user.welcome": "Welcome, {0}!"
"items.count":  "You have {0} items in {1}"

// 也可以使用 printf 风格
"user.welcome": "Welcome, %s!"
"items.count":  "You have %d items in %s"
```

### 4. 缓存策略

- 对于小型应用，使用默认缓存配置即可
- 对于大型应用，考虑使用 `LoadLanguageWithConfig` + `CacheAdapter` 并调整最大模板数量
- 定期监控内存使用情况，必要时清理缓存

### 5. 错误处理

始终检查加载错误：

```go
if err := i18n.LoadLanguage("en", "English", data); err != nil {
    log.Printf("Failed to load English translations: %v", err)
    // 实施回退策略或使用默认值
}
```

### 6. 并发安全

该模块是线程安全的，可以在多个 goroutine 中安全使用：

```go
go func() {
    fmt.Println(i18n.T("welcome"))
}()

go func() {
    i18n.SetLanguage("zh-CN")
}()
```

### 7. 内存管理

对于长期运行的应用，定期清理缓存和监控内存使用：

```go
// 定期清理（例如每小时）
go func() {
    ticker := time.NewTicker(time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        i18n.ClearTemplateCache()

        stats := i18n.GetMemoryUsage()
        if memoryUsage, ok := stats["total"].(map[string]interface{}); ok {
            log.Printf("Memory usage: %+v", memoryUsage)
        }
    }
}()
```

## 性能特性

- **读偏向锁**: 使用读偏向的互斥锁优化读多写少的场景
- **模板缓存**: 自动缓存已处理的模板以提高性能
- **缓冲池**: 使用 zutil 的缓冲池减少内存分配
- **并发安全**: 支持高并发访问而不影响性能

## 错误类型

```go
var (
    ErrLanguageNotFound = fmt.Errorf("language not found")
    ErrKeyNotFound      = fmt.Errorf("translation key not found")
)
```

## 注意事项

1. 翻译数据应该在应用启动时加载，避免运行时频繁加载
2. 键名应该保持一致性和可读性
3. 合理设置缓存大小以平衡性能和内存使用
4. 在应用关闭时调用 `Close()` 方法清理资源
5. 对于高并发场景，优先使用全局便捷函数以提高性能