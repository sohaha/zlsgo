# zvalid 模块

`zvalid` 提供了灵活的验证规则链、多种验证方法、自定义验证函数等功能，用于数据验证和格式检查。

## 功能概览

- **验证引擎**: 链式验证规则构建
- **基础验证**: 必需字段、非空等基本验证
- **格式验证**: 邮箱、手机号、URL、IP等格式验证
- **类型验证**: 数字、字母、中文等类型验证
- **内容验证**: 包含字母、数字、符号等内容验证
- **长度验证**: 字符串长度和UTF8长度验证
- **数值验证**: 整数和浮点数范围验证
- **枚举验证**: 字符串、整数、浮点数枚举验证
- **格式处理**: 字符串格式化和清理
- **自定义验证**: 支持自定义验证函数
- **JSON验证**: JSON数据的批量验证

## 核心功能

### 验证引擎创建

```go
// 创建新的验证引擎
func New() Engine
// 创建整数验证引擎
func Int(value int, name ...string) Engine
// 创建文本验证引擎
func Text(value string, name ...string) Engine
```

### 基础验证规则

```go
// 设置字段为必填
func (v Engine) Required(customError ...string) Engine
// 验证值
func (v Engine) Verifi(value string, name ...string) Engine
// 验证任意类型值
func (v Engine) VerifiAny(value interface{}, name ...string) Engine
// 自定义验证函数
func (v Engine) Customize(fn func(rawValue string, err error) (newValue string, newErr error)) Engine
// 正则表达式验证
func (v Engine) Regex(pattern string, customError ...string) Engine
```

### 类型验证规则

```go
func (v Engine) IsBool(customError ...string) Engine
func (v Engine) IsLower(customError ...string) Engine
func (v Engine) IsUpper(customError ...string) Engine
func (v Engine) IsLetter(customError ...string) Engine
func (v Engine) IsNumber(customError ...string) Engine
func (v Engine) IsInteger(customError ...string) Engine
func (v Engine) IsLowerOrDigit(customError ...string) Engine
func (v Engine) IsUpperOrDigit(customError ...string) Engine
func (v Engine) IsLetterOrDigit(customError ...string) Engine
func (v Engine) IsChinese(customError ...string) Engine
```

### 格式验证规则

```go
func (v Engine) IsMobile(customError ...string) Engine
func (v Engine) IsMail(customError ...string) Engine
func (v Engine) IsURL(customError ...string) Engine
func (v Engine) IsIP(customError ...string) Engine
func (v Engine) IsJSON(customError ...string) Engine
func (v Engine) IsChineseIDNumber(customError ...string) Engine
```

### 内容验证规则

```go
func (v Engine) HasLetter(customError ...string) Engine
func (v Engine) HasLower(customError ...string) Engine
func (v Engine) HasUpper(customError ...string) Engine
func (v Engine) HasNumber(customError ...string) Engine
func (v Engine) HasSymbol(customError ...string) Engine
func (v Engine) HasString(sub string, customError ...string) Engine
func (v Engine) HasPrefix(sub string, customError ...string) Engine
func (v Engine) HasSuffix(sub string, customError ...string) Engine
func (v Engine) Password(customError ...string) Engine
func (v Engine) StrongPassword(customError ...string) Engine
```

### 长度验证规则

```go
func (v Engine) MinLength(min int, customError ...string) Engine
func (v Engine) MaxLength(max int, customError ...string) Engine
func (v Engine) MinUTF8Length(min int, customError ...string) Engine
func (v Engine) MaxUTF8Length(max int, customError ...string) Engine
```

### 数值验证规则

```go
func (v Engine) MinInt(min int, customError ...string) Engine
func (v Engine) MaxInt(max int, customError ...string) Engine
func (v Engine) MinFloat(min float64, customError ...string) Engine
func (v Engine) MaxFloat(max float64, customError ...string) Engine
```

### 枚举验证规则

```go
func (v Engine) EnumString(slice []string, customError ...string) Engine
func (v Engine) EnumInt(i []int, customError ...string) Engine
func (v Engine) EnumFloat64(f []float64, customError ...string) Engine
```

### 格式处理规则

```go
func (v Engine) Trim() Engine
func (v Engine) RemoveSpace() Engine
func (v Engine) Replace(old, new string, n int) Engine
func (v Engine) ReplaceAll(old, new string) Engine
func (v Engine) XSSClean() Engine
func (v Engine) SnakeCaseToCamelCase(ucfirst bool, delimiter ...string) Engine
func (v Engine) CamelCaseToSnakeCase(delimiter ...string) Engine
func (v Engine) EncryptPassword(cost ...int) Engine
```

### 绑定和控制

```go
func (v Engine) Silent() Engine
func (v Engine) Default(value interface{}) Engine
func (v Engine) Separator(sep string) Engine
func (v Engine) SetAlias(name string) Engine
```

### JSON验证

```go
func JSON(json *zjson.Res, rules map[string]Engine) error
```

### 值获取

```go
func (v Engine) Ok() bool
func (v Engine) Error() error
func (v Engine) Value() string
func (v Engine) String() (string, error)
func (v Engine) Bool() (bool, error)
func (v Engine) Int() (int, error)
func (v Engine) Int64() (int64, error)
func (v Engine) Float32() (float32, error)
func (v Engine) Float64() (float64, error)
func (v Engine) Time(layout ...string) (time.Time, error)
func (v Engine) Split(sep string) ([]string, error)
```

### 批量验证

```go
func BatchError(rules ...Engine) error
func Batch(elements ...*ValidEle) error
func BatchVar(target interface{}, source Engine) *ValidEle
func Var(target interface{}, source Engine, name ...string) error
```

### 密码验证

```go
func (v Engine) CheckPassword(password string, customError ...string) Engine
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zvalid"
)

func main() {
    // 基础验证示例
    // 验证必需字段
    result := zvalid.Text("", "用户名").Required("用户名不能为空")
    if !result.Ok() {
        fmt.Printf("验证失败: %v\n", result.Error())
    }
    
    // 验证邮箱格式
    result = zvalid.Text("test@example.com", "邮箱").IsMail("邮箱格式不正确")
    if result.Ok() {
        email, _ := result.String()
        fmt.Printf("邮箱验证通过: %s\n", email)
    }
    
    // 验证手机号
    result = zvalid.Text("13812345678", "手机号").IsMobile("手机号格式不正确")
    if result.Ok() {
        phone, _ := result.String()
        fmt.Printf("手机号验证通过: %s\n", phone)
    }
    
    // 验证数字范围
    result = zvalid.Int(25, "年龄").MinInt(18, "年龄不能小于18岁").MaxInt(65, "年龄不能大于65岁")
    if result.Ok() {
        age, _ := result.Int()
        fmt.Printf("年龄验证通过: %d\n", age)
    }
    
    // 验证字符串长度
    result = zvalid.Text("张三", "姓名").MinLength(2, "姓名长度不能少于2个字符").MaxLength(10, "姓名长度不能超过10个字符")
    if result.Ok() {
        name, _ := result.String()
        fmt.Printf("姓名验证通过: %s\n", name)
    }
    
    // 链式验证
    result = zvalid.Text("test123", "密码").
        Required("密码不能为空").
        MinLength(6, "密码长度不能少于6位").
        MaxLength(20, "密码长度不能超过20位").
        HasLetter("密码必须包含字母").
        HasNumber("密码必须包含数字")
    
    if result.Ok() {
        password, _ := result.String()
        fmt.Printf("密码验证通过: %s\n", password)
    } else {
        fmt.Printf("密码验证失败: %v\n", result.Error())
    }
    
    // 强密码验证
    result = zvalid.Text("MyPass123!", "密码").
        Required("密码不能为空").
        StrongPassword("密码强度不够")
    
    if result.Ok() {
        fmt.Println("强密码验证通过")
    } else {
        fmt.Printf("强密码验证失败: %v\n", result.Error())
    }
    
    // 自定义验证
    engine := zvalid.Text("admin", "用户名").
        Required("用户名不能为空").
        MinLength(3, "用户名至少3个字符").
        MaxLength(20, "用户名最多20个字符").
        IsLetterOrDigit("用户名只能包含字母和数字")
    
    // 自定义验证
    engine.Customize(func(rawValue string, err error) (string, error) {
        if rawValue == "admin" {
            return "", fmt.Errorf("不能使用 admin 作为用户名")
        }
        return rawValue, nil
    })
    
    // 执行验证
    result, err := engine.String()
    if err != nil {
        fmt.Printf("验证失败: %v\n", err)
    } else {
        fmt.Printf("验证成功: %s\n", result)
    }
    
    // 枚举验证
    validRoles := []string{"admin", "user", "guest"}
    result = zvalid.Text("admin", "角色").EnumString(validRoles, "角色必须是有效的角色类型")
    if result.Ok() {
        role, _ := result.String()
        fmt.Printf("角色验证通过: %s\n", role)
    }
    
    // 格式处理
    result = zvalid.Text("  Hello World  ", "文本").
        Trim().
        RemoveSpace().
        ReplaceAll("World", "Go")
    
    if result.Ok() {
        processedText := result.Value()
        fmt.Printf("格式处理结果: %s\n", processedText)
    }
    
    // 实际应用示例
    // 用户注册验证
    type User struct {
        Username string `valid:"username"`
        Password string `valid:"password"`
        Email    string `valid:"email"`
        Phone    string `valid:"phone"`
        Age      int    `valid:"age"`
    }
    
    user := &User{}
    
    // 创建验证规则
    usernameRule := zvalid.Text("zhangsan123", "用户名").
        Required("用户名不能为空").
        MinLength(3, "用户名至少3个字符").
        MaxLength(20, "用户名最多20个字符").
        IsLetterOrDigit("用户名只能包含字母和数字")
    
    passwordRule := zvalid.Text("MyPass123!", "密码").
        Required("密码不能为空").
        MinLength(8, "密码至少8个字符").
        HasLetter("密码必须包含字母").
        HasNumber("密码必须包含数字").
        HasSymbol("密码必须包含特殊字符")
    
    emailRule := zvalid.Text("zhangsan@example.com", "邮箱").
        Required("邮箱不能为空").
        IsMail("邮箱格式不正确")
    
    phoneRule := zvalid.Text("13800138000", "手机号").
        Required("手机号不能为空").
        IsMobile("手机号格式不正确")
    
    ageRule := zvalid.Int(25, "年龄").
        Required("年龄不能为空").
        MinInt(18, "年龄不能小于18岁").
        MaxInt(65, "年龄不能大于65岁")
    
    // 执行批量验证
    err := zvalid.Batch(
        zvalid.BatchVar(&user.Username, usernameRule),
        zvalid.BatchVar(&user.Password, passwordRule),
        zvalid.BatchVar(&user.Email, emailRule),
        zvalid.BatchVar(&user.Phone, phoneRule),
        zvalid.BatchVar(&user.Age, ageRule),
    )
    
    if err != nil {
        fmt.Printf("表单验证失败: %v\n", err)
    } else {
        fmt.Printf("用户注册表单验证通过: %+v\n", user)
    }
    
    // 配置文件验证
    configData := map[string]interface{}{
        "server_port": 8080,
        "db_host":     "127.0.0.1",
        "debug":       true,
    }
    
    // 验证服务器端口
    portResult := zvalid.Int(configData["server_port"].(int), "服务器端口").
        MinInt(1024, "端口必须大于1024").
        MaxInt(65535, "端口必须小于65535")
    
    if portResult.Ok() {
        fmt.Println("服务器端口验证通过")
    }
    
    // 验证数据库主机
    hostResult := zvalid.Text(configData["db_host"].(string), "数据库主机").
        Required("数据库主机不能为空").
        IsIP("数据库主机必须是有效的IP地址")
    
    if hostResult.Ok() {
        fmt.Println("数据库主机验证通过")
    }
    
    fmt.Println("数据验证示例完成")
}
```

## 验证规则说明

### 验证链
```go
// 验证规则可以链式调用
engine := zvalid.Text("value", "字段名").
    Required("不能为空").
    MinLength(3, "长度不能少于3位").
    IsMail("必须是邮箱格式")
```

### 错误处理
```go
// 获取验证错误
if !engine.Ok() {
    err := engine.Error()
    // 处理错误
}

// 获取验证后的值
value, err := engine.String()
```

### 自定义验证
```go
// 使用自定义验证函数
engine.Customize(func(rawValue string, err error) (string, error) {
    // 自定义验证逻辑
    if rawValue == "invalid" {
        return rawValue, fmt.Errorf("无效值")
    }
    return rawValue, err
})
```

### 批量验证
```go
// 批量验证多个字段
err := zvalid.Batch(
    zvalid.BatchVar(&target1, rule1),
    zvalid.BatchVar(&target2, rule2),
)
```

## 最佳实践

1. 使用链式验证提高可读性
2. 实现适当的错误处理
3. 缓存验证引擎实例
4. 合理使用自定义验证函数