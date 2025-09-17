# zstring 模块

`zstring` 提供了字符串操作、正则表达式、加密解密、编码解码、随机生成、模板处理等功能，用于各种字符串和文本处理任务。

## 功能概览

- **字符串操作**: 字符串的基本操作和转换
- **正则表达式**: 正则表达式匹配和处理
- **加密解密**: AES、RSA 等加密算法
- **编码解码**: Base64、序列化等编码方式
- **随机生成**: 随机字符串、UUID、雪花ID等
- **模板处理**: 字符串模板解析和替换
- **字符匹配**: 模式匹配和通配符支持

## 核心功能

### 字符串操作

```go
// 使用指定内容填充字符添加到字符串中以达到指定的长度
func Pad(raw string, length int, padStr string, padType PadType) string
// 以 Unicode 字符数计算字符串长度
func Len(str string) int
// 截取 UTF-8 字符串，支持负起始位置与可选长度
func Substr(str string, start int, length ...int) string
// 零拷贝将 []byte 转为 string（只读）
func Bytes2String(b []byte) string
// 零拷贝将 string 转为 []byte（只读）
func String2Bytes(s string) []byte
// 将首字母转为大写
func Ucfirst(str string) string
// 将首字母转为小写
func Lcfirst(str string) string
// 判断首字母是否为大写
func IsUcfirst(str string) bool
// 判断首字母是否为小写
func IsLcfirst(str string) bool
// 去除字节流开头的 UTF-8 BOM
func TrimBOM(fileBytes []byte) []byte
// 下划线命名转驼峰，ucfirst 控制是否首字母大写
func SnakeCaseToCamelCase(str string, ucfirst bool, delimiter ...string) string
// 驼峰命名转下划线，支持自定义分隔符
func CamelCaseToSnakeCase(str string, delimiter ...string) string
// 清除可能的 XSS 内容（移除标签与脚本）
func XSSClean(str string) string
// 压缩多余空白并去掉行首尾空白
func TrimLine(s string) string
// 去除字符串首尾空白（扩展支持更多空白符）
func TrimSpace(s string) string
// 判断 rune 是否为空白字符
func IsSpace(r rune) bool
// Go 1.10+ 返回 *strings.Builder；更早版本返回 *bytes.Buffer
// 创建可选初始容量的字符串缓冲区（签名随 Go 版本不同而不同）
// Go>=1.10: func Buffer(size ...int) *strings.Builder
// Go<1.10:  func Buffer(size ...int) *bytes.Buffer
```

### 正则表达式

```go
// 判断字符串是否匹配正则表达式
func RegexMatch(pattern string, str string) bool
// 提取第一个匹配结果及分组
func RegexExtract(pattern string, str string) ([]string, error)
// 提取所有匹配结果及分组，可选数量限制
func RegexExtractAll(pattern string, str string, count ...int) ([][]string, error)
// 返回所有匹配的位置区间
func RegexFind(pattern string, str string, count int) [][]int
// 使用替换字符串替换所有匹配内容
func RegexReplace(pattern string, str string, repl string) (string, error)
// 使用回调函数替换每一个匹配内容
func RegexReplaceFunc(pattern string, str string, repl func(string) string) (string, error)
// 依据正则表达式分割字符串
func RegexSplit(pattern string, str string) ([]string, error)
```

### 字符匹配

```go
// 支持 * 与 ? 的通配模式匹配，可选忽略大小写
func Match(str string, pattern string, equalFold ...bool) bool
// 判断字符串是否包含通配符，是否为模式串
func IsPattern(str string) bool
```

### MD5 哈希

```go
// 获取当前可执行文件的 MD5
func ProjectMd5() string
// 计算字符串的 MD5（hex 编码）
func Md5(s string) string
// 计算字节切片的 MD5（hex 编码）
func Md5Byte(s []byte) string
// 计算文件内容的 MD5（hex 编码）
func Md5File(path string) (string, error)
```

### Base64 编码

```go
// 标准 Base64 编码（[]byte -> []byte）
func Base64Encode(value []byte) []byte
// 标准 Base64 编码（string -> string）
func Base64EncodeString(value string) string
// 标准 Base64 解码（[]byte -> []byte）
func Base64Decode(data []byte) ([]byte, error)
// 标准 Base64 解码（string -> string）
func Base64DecodeString(data string) (string, error)
// 使用 gob 将任意值序列化为字节
func Serialize(value interface{}) ([]byte, error)
// 使用 gob 从字节反序列化为原始值
func UnSerialize(valueBytes []byte, registers ...interface{}) (interface{}, error)
// 读取图片文件并返回 data URL（base64）
func Img2Base64(path string) (string, error)
```

### URL 编码

```go
// 对字符串进行 URL 编码（application/x-www-form-urlencoded）
func UrlEncode(str string) string
// 对字符串进行 URL 解码
func UrlDecode(str string) (string, error)
// 原始 URL 编码（不转义空格为 +）
func UrlRawEncode(str string) string
// 原始 URL 解码
func UrlRawDecode(str string) (string, error)
```

### AES 加密

```go
// AES-CBC 加密（可选 IV，默认使用处理后的 key 作为 IV）
func AesEncrypt(plainText []byte, key string, iv ...string) ([]byte, error)
// AES-CBC 解密（可选 IV）
func AesDecrypt(cipherText []byte, key string, iv ...string) ([]byte, error)
// AES-CBC 加密字符串，返回 Base64
func AesEncryptString(plainText string, key string, iv ...string) (string, error)
// AES-CBC 解密 Base64 字符串
func AesDecryptString(cipherText string, key string, iv ...string) (string, error)
// AES-GCM 加密，返回带随机 nonce 的密文
func AesGCMEncrypt(plaintext []byte, key string) ([]byte, error)
// AES-GCM 解密
func AesGCMDecrypt(ciphertext []byte, key string) ([]byte, error)
// AES-GCM 加密字符串，返回 Base64
func AesGCMEncryptString(plainText string, key string) (string, error)
// AES-GCM 解密 Base64 字符串
func AesGCMDecryptString(cipherText string, key string) (string, error)
// PKCS#7 填充
func PKCS7Padding(ciphertext []byte, blockSize int) []byte
// 移除 PKCS#7 填充
func PKCS7UnPadding(origData []byte) ([]byte, error)
```

### RSA 加密

```go
// 生成 RSA 私钥与公钥（bits 指定位数，默认 1024）
func GenRSAKey(bits ...int) (prvkey []byte, pubkey []byte, err error)
// 使用公钥加密，支持大数据分块
func RSAEncrypt(plainText []byte, publicKey string, bits ...int) ([]byte, error)
// 使用公钥对象加密
func RSAKeyEncrypt(plainText []byte, publicKey *rsa.PublicKey, bits ...int) ([]byte, error)
// 使用公钥加密字符串，返回 Base64
func RSAEncryptString(plainText string, publicKey string) (string, error)
// 使用私钥签名（按加密形式返回 Base64）
func RSAPriKeyEncrypt(plainText []byte, privateKey string) ([]byte, error)
// 使用私钥签名字符串（返回 Base64）
func RSAPriKeyEncryptString(plainText string, privateKey string) (string, error)
// 使用私钥解密（支持大数据分块）
func RSADecrypt(cipherText []byte, privateKey string, bits ...int) ([]byte, error)
// 使用私钥对象解密
func RSAKeyDecrypt(cipherText []byte, privateKey *rsa.PrivateKey, bits ...int) ([]byte, error)
// 使用私钥解密 Base64 字符串
func RSADecryptString(cipherText string, privateKey string) (string, error)
// 使用公钥验证/解密由私钥加密的数据
func RSAPubKeyDecrypt(cipherText []byte, publicKey string) ([]byte, error)
// 使用公钥验证/解密 Base64 字符串
func RSAPubKeyDecryptString(cipherText string, publicKey string) (string, error)
```

### 随机生成

```go
// 返回 [0..^uint32(0)] 范围的伪随机数
func RandUint32() uint32
// 返回 [0..max) 范围的伪随机数
func RandUint32Max(max uint32) uint32
// 返回 [min..max] 范围的随机整数
func RandInt(min int, max int) int
// 生成指定长度的随机字符串，可自定义字符模板
func Rand(n int, tpl ...string) string
// 生成至少指定长度的唯一 ID（优先使用加密随机）
func UniqueID(n int) string
// 进行一次带权随机选择
func WeightedRand(choices map[interface{}]uint32) (interface{}, error)
// 构造带权随机选择器
func NewWeightedRand(choices map[interface{}]uint32) (*Weighteder, error)
// 生成指定长度的 NanoID，可自定义字母表
func NewNanoID(size int, alphabet ...string) (string, error)
// 生成 UUID v4
func UUID() string
// 从加权选择器中选择一个元素
func (w *Weighteder) Pick() interface{}
```

### 雪花ID

```go
// 创建 Snowflake ID 生成器
func NewIDWorker(workerid int64) (*IDWorker, error)
// 生成一个全局唯一的 ID
func (w *IDWorker) ID() (int64, error)
// 解析 ID，返回时间、原始时间戳、工作节点与序列
func ParseID(id int64) (time.Time, int64, int64, int64)
```

### 字符串模板

```go
// 使用自定义开始/结束标签创建模板
func NewTemplate(template string, startTag string, endTag string) (*Template, error)
// 将模板渲染到 io.Writer，回调返回写入字节数与错误
func (t *Template) Process(w io.Writer, fn func(io.Writer, string) (int, error)) (int64, error)
// 重设模板字符串并重新解析
func (t *Template) ResetTemplate(template string) error
```

### 字符串展开

```go
// 以 ${var} 或 $var 形式展开变量，通过回调提供变量值
func Expand(s string, process func(string) string) string
```

### 字符串过滤

```go
// 创建敏感词过滤器，mask 为替换字符
func NewFilter(words []string, mask ...rune) *filterNode
```

### 字符串替换

```go
// 创建基于字典树的多项替换器
func NewReplacer(mapping map[string]string) *replacer
```

## 使用示例

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/sohaha/zlsgo/zstring"
    "io"
)

func main() {
    // 字符串操作示例
    str := "hello world"
    
    // 首字母大写
    ucfirst := zstring.Ucfirst(str)
    fmt.Printf("首字母大写: %s\n", ucfirst)
    
    // 驼峰转下划线
    camelCase := "HelloWorld"
    snakeCase := zstring.CamelCaseToSnakeCase(camelCase)
    fmt.Printf("驼峰转下划线: %s\n", snakeCase)
    
    // 下划线转驼峰
    snakeToCamel := zstring.SnakeCaseToCamelCase(snakeCase, true)
    fmt.Printf("下划线转驼峰: %s\n", snakeToCamel)
    
    // 字符串长度（支持中文）
    chineseStr := "你好世界"
    length := zstring.Len(chineseStr)
    fmt.Printf("中文字符串长度: %d\n", length)
    
    // 字符串截取
    substr := zstring.Substr(chineseStr, 0, 2)
    fmt.Printf("字符串截取: %s\n", substr)
    
    // 正则表达式示例
    text := "手机号码: 13812345678, 邮箱: test@example.com"
    
    // 提取手机号
    phonePattern := `1[3-9]\d{9}`
    if zstring.RegexMatch(phonePattern, text) {
        phone, _ := zstring.RegexExtract(phonePattern, text)
        fmt.Printf("提取手机号: %v\n", phone)
    }
    
    // 提取邮箱
    emailPattern := `\w+@\w+\.\w+`
    emails, _ := zstring.RegexExtractAll(emailPattern, text, -1)
    fmt.Printf("提取邮箱: %v\n", emails)
    
    // 替换敏感词
    sensitiveText := "这是一个测试内容，包含敏感词"
    replaced, _ := zstring.RegexReplace("敏感词", sensitiveText, "***")
    fmt.Printf("替换后: %s\n", replaced)
    
    // MD5 哈希示例
    originalText := "要加密的文本"
    md5Hash := zstring.Md5(originalText)
    fmt.Printf("MD5 哈希: %s\n", md5Hash)
    
    // Base64 编码示例
    base64Encoded := zstring.Base64EncodeString(originalText)
    fmt.Printf("Base64 编码: %s\n", base64Encoded)
    
    // Base64 解码
    base64Decoded, _ := zstring.Base64DecodeString(base64Encoded)
    fmt.Printf("Base64 解码: %s\n", base64Decoded)
    
    // AES 加密示例
    key := "1234567890123456" // 16字节密钥
    plaintext := "要加密的敏感数据"
    
    // AES 加密
    encrypted, err := zstring.AesEncryptString(plaintext, key)
    if err != nil {
        fmt.Printf("AES 加密失败: %v\n", err)
    } else {
        fmt.Printf("AES 加密结果: %s\n", encrypted)
        
        // AES 解密
        decrypted, err := zstring.AesDecryptString(encrypted, key)
        if err != nil {
            fmt.Printf("AES 解密失败: %v\n", err)
        } else {
            fmt.Printf("AES 解密结果: %s\n", decrypted)
        }
    }
    
    // RSA 加密示例
    // 生成RSA密钥对
    privateKey, publicKey, err := zstring.GenRSAKey(2048)
    if err != nil {
        fmt.Printf("生成RSA密钥失败: %v\n", err)
    } else {
        // RSA 加密
        rsaEncrypted, err := zstring.RSAEncryptString(plaintext, string(publicKey))
        if err != nil {
            fmt.Printf("RSA 加密失败: %v\n", err)
        } else {
            fmt.Printf("RSA 加密成功\n")
            
            // RSA 解密
            rsaDecrypted, err := zstring.RSADecryptString(rsaEncrypted, string(privateKey))
            if err != nil {
                fmt.Printf("RSA 解密失败: %v\n", err)
            } else {
                fmt.Printf("RSA 解密结果: %s\n", rsaDecrypted)
            }
        }
    }
    
    // 随机生成示例
    // 随机字符串
    randomStr := zstring.Rand(10)
    fmt.Printf("随机字符串: %s\n", randomStr)
    
    // 随机数字字符串
    randomNum := zstring.Rand(6, "0123456789")
    fmt.Printf("随机数字: %s\n", randomNum)
    
    // 唯一ID
    uniqueID := zstring.UniqueID(16)
    fmt.Printf("唯一ID: %s\n", uniqueID)
    
    // UUID
    uuid := zstring.UUID()
    fmt.Printf("UUID: %s\n", uuid)
    
    // NanoID
    nanoID, _ := zstring.NewNanoID(21)
    fmt.Printf("NanoID: %s\n", nanoID)
    
    // 雪花ID
    worker, err := zstring.NewIDWorker(1)
    if err != nil {
        fmt.Printf("创建雪花ID生成器失败: %v\n", err)
    } else {
        snowflakeID, err := worker.ID()
        if err != nil {
            fmt.Printf("生成雪花ID失败: %v\n", err)
            return
        }
        fmt.Printf("雪花ID: %d\n", snowflakeID)
        
        // 解析雪花ID
        t, ts, workerId, seq := zstring.ParseID(snowflakeID)
        fmt.Printf("雪花ID解析: 时间=%v, 时间戳=%d, 工作ID=%d, 序列号=%d\n", t, ts, workerId, seq)
    }
    
    // 权重随机
    choices := map[interface{}]uint32{
        "苹果": 10,
        "香蕉": 20,
        "橘子": 30,
        "葡萄": 40,
    }
    
    result, err := zstring.WeightedRand(choices)
    if err != nil {
        fmt.Printf("权重随机失败: %v\n", err)
    } else {
        fmt.Printf("权重随机结果: %v\n", result)
    }
    
    // 字符串模板示例
    template, err := zstring.NewTemplate("Hello {{name}}, you are {{age}} years old!", "{{", "}}")
    if err != nil {
        fmt.Printf("创建模板失败: %v\n", err)
    } else {
        // 使用模板处理
        var buf bytes.Buffer
        _, err := template.Process(&buf, func(w io.Writer, tag string) (int, error) {
            switch tag {
            case "name":
                return w.Write([]byte("张三"))
            case "age":
                return w.Write([]byte("25"))
            default:
                return 0, fmt.Errorf("未知标签: %s", tag)
            }
        })
        
        if err != nil {
            fmt.Printf("执行模板失败: %v\n", err)
        } else {
            result := buf.String()
            fmt.Printf("模板结果: %s\n", result)
        }
    }
    
    // 字符串展开示例
    expanded := zstring.Expand("Hello $USER, today is $DATE", func(key string) string {
        switch key {
        case "USER":
            return "张三"
        case "DATE":
            return "2023-12-25"
        default:
            return ""
        }
    })
    fmt.Printf("字符串展开: %s\n", expanded)
    
    // 模式匹配示例
    pattern := "test*.txt"
    filename := "test123.txt"
    
    if zstring.Match(filename, pattern) {
        fmt.Printf("文件名 %s 匹配模式 %s\n", filename, pattern)
    }
    
    // 敏感词过滤示例
    sensitiveWords := []string{"敏感词1", "敏感词2", "违禁内容"}
    filter := zstring.NewFilter(sensitiveWords, '*')
    
    testText := "这里包含敏感词1和违禁内容"
    filteredText := filter.Filter(testText)
    fmt.Printf("过滤后文本: %s\n", filteredText)
    
    // 检查是否包含敏感词
    isValid := filter.Validate(testText)
    fmt.Printf("文本是否合规: %t\n", isValid)
    
    // 字符串替换器
    replacements := map[string]string{
        "apple":  "苹果",
        "banana": "香蕉",
        "orange": "橘子",
    }
    
    replacer := zstring.NewReplacer(replacements)
    englishText := "I like apple and banana"
    chineseText := replacer.Replace(englishText)
    fmt.Printf("替换结果: %s\n", chineseText)
    
    // URL 编码示例
    url := "https://example.com/search?q=Go语言"
    encoded := zstring.UrlEncode(url)
    fmt.Printf("URL 编码: %s\n", encoded)
    
    decoded, _ := zstring.UrlDecode(encoded)
    fmt.Printf("URL 解码: %s\n", decoded)
    
    // 实际应用示例
    // 用户信息验证
    userInput := "用户输入的内容包含<script>alert('xss')</script>"
    cleanInput := zstring.XSSClean(userInput)
    fmt.Printf("XSS 清理后: %s\n", cleanInput)
    
    // 配置文件处理
    configTemplate := "server_name={{host}}\nport={{port}}\ndebug={{debug}}"
    config := map[string]interface{}{
        "host":  "localhost",
        "port":  8080,
        "debug": true,
    }
    
    tmpl, _ := zstring.NewTemplate(configTemplate, "{{", "}}")
    var configBuf bytes.Buffer
    _, err := tmpl.Process(&configBuf, func(w io.Writer, tag string) (int, error) {
        if value, exists := config[tag]; exists {
            return w.Write([]byte(fmt.Sprintf("%v", value)))
        }
        return 0, fmt.Errorf("未知配置项: %s", tag)
    })
    
    if err != nil {
        fmt.Printf("处理配置模板失败: %v\n", err)
    } else {
        configResult := configBuf.String()
        fmt.Printf("配置文件:\n%s\n", configResult)
    }
    
    fmt.Println("zstring 模块示例完成")
}
