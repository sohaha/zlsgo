# ztime 模块

`ztime` 提供了时间获取、格式化、计算、转换、时区管理、定时任务等功能，用于高效的时间操作和日期处理。

## 功能概览

- **时间获取**: 当前时间和时间对象获取
- **时间格式化**: 时间格式化和解析
- **时间计算**: 时间差计算和范围查找
- **时间转换**: 时间转换和变换
- **时区管理**: 时区设置和转换
- **定时任务**: 定时任务管理

## 核心功能

### 时间获取

```go
// 获取当前时间，支持自定义格式
func Now(format ...string) string
// 获取当前时间对象
func Time() time.Time
// 获取当前时间戳（秒）
func Clock() int64
// 获取当前时间戳（微秒）
func ClockMicro() int64
```

### 时间格式化

```go
// 格式化时间，支持自定义格式
func FormatTime(t time.Time, format ...string) string
// 格式化时间戳，支持自定义格式
func FormatTimestamp(timestamp int64, format ...string) string
```

### 时间解析

```go
// 解析时间字符串，支持自定义格式
func Parse(str string, format ...string) (time.Time, error)
// 解析时间戳
func Unix(tt int64) time.Time
// 解析微秒时间戳
func UnixMicro(tt int64) time.Time
```

### 时间计算

```go
// 计算两个时间的差值
func Diff[T time.Time | string](t1, t2 T, format ...string) (time.Duration, error)
// 查找时间范围
func FindRange[T time.Time | string](times []T, format ...string) (time.Time, time.Time, error)
// 生成时间序列
func Sequence[T time.Time | string](start, end T, stepFn func(time.Time) time.Time, format ...string) ([]string, error)
```

### 时间转换

```go
// 带上下文的睡眠
func Sleep(ctx context.Context, duration time.Duration) error
// 获取星期几
func Week(t time.Time) int
// 获取月份范围
func MonthRange(year, month int) (beginTime, endTime int64, err error)
```

### 时区管理

```go
func Zone(zone ...int) *time.Location
func SetTimeZone(zone int) *TimeEngine
func GetTimeZone() *time.Location
func In(tt time.Time) time.Time
```

### 时间引擎

```go
func New(zone ...int) *TimeEngine
func (e *TimeEngine) SetTimeZone(zone int) *TimeEngine
func (e *TimeEngine) GetTimeZone() *time.Location
func (e *TimeEngine) In(t time.Time) time.Time
func (e *TimeEngine) Now(format ...string) string
func (e *TimeEngine) Time(realTime ...bool) time.Time
func (e *TimeEngine) Clock() int64
func (e *TimeEngine) FormatTime(t time.Time, format ...string) string
func (e *TimeEngine) FormatTimestamp(timestamp int64, format ...string) string
func (e *TimeEngine) Unix(tt int64) time.Time
func (e *TimeEngine) UnixMicro(tt int64) time.Time
func (e *TimeEngine) Parse(str string, format ...string) (time.Time, error)
func (e *TimeEngine) Week(t time.Time) int
func (e *TimeEngine) MonthRange(year, month int) (beginTime, endTime int64, err error)
```

### 本地时间

```go
type LocalTime time.Time
func (t LocalTime) MarshalJSON() ([]byte, error)
func (t *LocalTime) UnmarshalJSON(data []byte) (err error)
func (t LocalTime) Value() (driver.Value, error)
func (t *LocalTime) Scan(v interface{}) error
func (t LocalTime) String() string
func (t LocalTime) Format(layout string) string
```

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/ztime"
)

func main() {
    // 时间获取示例
    currentTime := ztime.Now()
    fmt.Printf("当前时间: %s\n", currentTime)
    
    currentTime = ztime.Now("2006-01-02 15:04:05")
    fmt.Printf("当前时间: %s\n", currentTime)
    
    timeObj := ztime.Time()
    fmt.Printf("时间对象: %v\n", timeObj)
    
    clock := ztime.Clock()
    fmt.Printf("当前时间戳: %d\n", clock)
    
    // 时间格式化示例
    now := time.Now()
    formatted := ztime.FormatTime(now, "2006-01-02 15:04:05")
    fmt.Printf("格式化时间: %s\n", formatted)
    
    timestamp := now.Unix()
    formattedTimestamp := ztime.FormatTimestamp(timestamp, "2006-01-02 15:04:05")
    fmt.Printf("格式化时间戳: %s\n", formattedTimestamp)
    
    // 时间解析示例
    timeStr := "2023-12-25 10:30:00"
    parsed, err := ztime.Parse(timeStr, "2006-01-02 15:04:05")
    if err == nil {
        fmt.Printf("解析时间: %v\n", parsed)
    }
    
    // 时间计算示例
    start := time.Now()
    time.Sleep(100 * time.Millisecond)
    end := time.Now()
    
    duration, err := ztime.Diff(start, end)
    if err == nil {
        fmt.Printf("时间差: %v\n", duration)
    }
    
    // 时间序列示例
    stepFn := func(t time.Time) time.Time {
        return t.Add(24 * time.Hour)
    }
    
    sequence, err := ztime.Sequence(start, start.Add(5*24*time.Hour), stepFn)
    if err == nil {
        fmt.Printf("时间序列: %v\n", sequence)
    }
    
    // 时区管理示例
    // 设置时区
    engine := ztime.New(8) // UTC+8
    
    // 获取时区
    zone := engine.GetTimeZone()
    fmt.Printf("当前时区: %v\n", zone)
    
    // 时区转换
    utcTime := time.Now().UTC()
    localTime := engine.In(utcTime)
    fmt.Printf("UTC 时间: %v\n", utcTime)
    fmt.Printf("本地时间: %v\n", localTime)
    
    // 时间引擎示例
    timeEngine := ztime.New(8)
    
    // 设置时区
    timeEngine.SetTimeZone(9) // UTC+9
    
    // 格式化时间
    formatted = timeEngine.FormatTime(now, "2006-01-02 15:04:05")
    fmt.Printf("引擎格式化时间: %s\n", formatted)
    
    // 解析时间
    parsed, err = timeEngine.Parse("2023-12-25 10:30:00", "2006-01-02 15:04:05")
    if err == nil {
        fmt.Printf("引擎解析时间: %v\n", parsed)
    }
    
    // 获取星期
    week := timeEngine.Week(now)
    fmt.Printf("星期: %d\n", week)
    
    // 获取月份范围
    beginTime, endTime, err := timeEngine.MonthRange(2023, 12)
    if err == nil {
        fmt.Printf("12月时间范围: %d 到 %d\n", beginTime, endTime)
    }
    
    // 本地时间示例
    localTimeType := ztime.LocalTime(now)
    
    // JSON 序列化
    jsonData, err := localTimeType.MarshalJSON()
    if err == nil {
        fmt.Printf("JSON 数据: %s\n", string(jsonData))
    }
    
    // 字符串表示
    str := localTimeType.String()
    fmt.Printf("字符串: %s\n", str)
    
    // 格式化
    formatted = localTimeType.Format("2006-01-02 15:04:05")
    fmt.Printf("格式化: %s\n", formatted)
    
    // 时间工具示例
    // 睡眠
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    err = ztime.Sleep(ctx, 200*time.Millisecond)
    if err != nil {
        fmt.Printf("睡眠被中断: %v\n", err)
    }
    
    // 获取星期
    week = ztime.Week(now)
    fmt.Printf("当前星期: %d\n", week)
    
    // 获取月份范围
    beginTime, endTime, err = ztime.MonthRange(2023, 12)
    if err == nil {
        fmt.Printf("2023年12月时间范围: %d 到 %d\n", beginTime, endTime)
    }
    
    // Unix 时间转换
    unixTime := ztime.Unix(1703481600) // 2023-12-25 10:00:00 UTC
    fmt.Printf("Unix 时间: %v\n", unixTime)
    
    unixMicroTime := ztime.UnixMicro(1703481600000000)
    fmt.Printf("Unix 微秒时间: %v\n", unixMicroTime)
    
    // 实际应用示例
    // 日志时间戳
    logTime := ztime.Now("2006-01-02 15:04:05.000")
    fmt.Printf("[%s] 日志消息\n", logTime)
    
    // 缓存过期时间
    cacheExpire := ztime.Now("2006-01-02 15:04:05")
    fmt.Printf("缓存过期时间: %s\n", cacheExpire)
    
    // 定时任务时间
    cronTime := ztime.Now("15:04")
    fmt.Printf("定时任务时间: %s\n", cronTime)
    
    // 数据库时间字段
    dbTime := ztime.LocalTime(now)
    fmt.Printf("数据库时间: %v\n", dbTime)
    
    // 时区转换服务
    timeService := ztime.New(0) // UTC
    
    // 转换不同时区的时间
    times := []int{-8, -5, 0, 1, 8, 9}
    for _, tz := range times {
        timeService.SetTimeZone(tz)
        localTime := timeService.In(now)
        formatted := timeService.FormatTime(localTime, "15:04")
        fmt.Printf("UTC%+d: %s\n", tz, formatted)
    }
    
    fmt.Println("时间处理示例完成")
}
```

## 时间格式

### 标准格式
- **2006-01-02**: 日期格式
- **15:04:05**: 时间格式
- **2006-01-02 15:04:05**: 日期时间格式
- **2006-01-02T15:04:05Z07:00**: RFC3339 格式

### 自定义格式
```go
// 自定义时间格式
customFormat := "2006年01月02日 15时04分05秒"
formatted := ztime.FormatTime(now, customFormat)
```

## 时区支持

### 常用时区
- **UTC**: 协调世界时
- **UTC+8**: 北京时间
- **UTC+9**: 东京时间
- **UTC-5**: 纽约时间
- **UTC-8**: 洛杉矶时间

### 时区操作
```go
// 创建时区引擎
engine := ztime.New(8)

// 设置时区
engine.SetTimeZone(9)

// 时区转换
localTime := engine.In(utcTime)
```

## 最佳实践

1. 使用标准时间格式
2. 实现适当的时区管理
3. 缓存时间引擎实例
4. 注意时间格式的一致性