# zhttp 模块

`zhttp` 提供了 HTTP 引擎配置、请求方法、响应处理、SSE、HTML 解析、查询选择器等功能，用于 HTTP 通信和 Web 数据获取。

## 功能概览

- **HTTP 引擎**: 可配置的 HTTP 客户端引擎
- **请求方法**: 支持所有 HTTP 方法
- **响应处理**: 响应数据解析和转换
- **JSON-RPC**: 支持 JSON-RPC 2.0 客户端
- **SSE**: Server-Sent Events 客户端
- **HTML 解析**: HTML 文档解析和查询
- **查询选择器**: CSS 选择器查询
- **调试**: 内置的请求和响应调试工具

## 核心功能

### HTTP 引擎

```go
// 创建新的 HTTP 引擎
func New() *Engine
// 禁用分块传输
func (e *Engine) DisableChunke(enable ...bool)
// 启用不安全的 TLS
func (e *Engine) EnableInsecureTLS(enable bool)
// 设置 TLS 证书
func (e *Engine) TlsCertificate(certs ...Certificate) error
// 启用 Cookie 支持
func (e *Engine) EnableCookie(enable bool)
// 设置超时时间
func (e *Engine) SetTimeout(d time.Duration)
// 设置传输配置
func (e *Engine) SetTransport(transport func(*http.Transport)) error
// 设置代理 URL
func (e *Engine) SetProxyUrl(proxyUrl ...string) error
// 设置代理函数
func (e *Engine) SetProxy(proxy func(*http.Request) (*url.URL, error)) error
// 移除代理
func (e *Engine) RemoveProxy() error
// 设置 JSON 转义 HTML
func (e *Engine) SetJSONEscapeHTML(escape bool)
// 设置 JSON 缩进
func (e *Engine) SetJSONIndent(prefix, indent string)
// 设置 XML 缩进
func (e *Engine) SetXMLIndent(prefix, indent string)
// 设置用户代理
func (e *Engine) SetUserAgent(fn func() string)
// 设置标志
func (e *Engine) SetFlags(flags int)
// 获取标志
func (e *Engine) Flags() int
// 获取 HTTP 客户端
func (e *Engine) Client() *http.Client
// 设置 HTTP 客户端
func (e *Engine) SetClient(client *http.Client)
```

### 请求方法

```go
func Get(url string, v ...interface{}) (*Res, error)
func Post(url string, v ...interface{}) (*Res, error)
func Put(url string, v ...interface{}) (*Res, error)
func Head(url string, v ...interface{}) (*Res, error)
func Options(url string, v ...interface{}) (*Res, error)
func Delete(url string, v ...interface{}) (*Res, error)
func Patch(url string, v ...interface{}) (*Res, error)
func Connect(url string, v ...interface{}) (*Res, error)
func Trace(url string, v ...interface{}) (*Res, error)
func Do(method, rawurl string, v ...interface{}) (resp *Res, err error)
func DoRetry(attempt int, sleep time.Duration, fn func() (*Res, error)) (*Res, error)
```

### 响应处理

```go
// 获取原始请求对象
func (r *Response) Request() *Request
// 获取响应对象
func (r *Response) Response() *http.Response
// 获取状态码
func (r *Response) StatusCode() int
// 获取所有Cookie
func (r *Response) GetCookie() map[string]*http.Cookie
// 获取响应字节
func (r *Response) Bytes() []byte
// 获取响应流
func (r *Response) Stream() io.ReadCloser
// 转换为字节数组
func (r *Response) ToBytes() ([]byte, error)
// 获取响应体
func (r *Response) Body() io.ReadCloser
// 获取HTML内容
func (r *Response) HTML() (string, error)
// 获取字符串内容
func (r *Response) String() (string, error)
// 获取JSON数组
func (r *Response) JSONs() ([]interface{}, error)
// 获取JSON对象
func (r *Response) JSON() (interface{}, error)
// 转换为字符串
func (r *Response) ToString() (string, error)
// 转换为XML
func (r *Response) ToXML() (string, error)
```

### JSON-RPC 客户端

```go
func NewJsonRPC(url string, opts ...func(rpc *JsonRPC)) *JsonRPC
func (rpc *JsonRPC) Call(method string, params interface{}, result interface{}) error
func (rpc *JsonRPC) SetClient(client *http.Client)
func (rpc *JsonRPC) SetHeader(key, value string)
```

### SSE

```go
func (e *Engine) SSE(url string, opt func(*SSEOption), v ...interface{}) (*SSEEngine, error)
func (sse *SSEEngine) Event() <-chan *SSEEvent
func (sse *SSEEngine) Close()
func (sse *SSEEngine) Done() <-chan struct{}
func (sse *SSEEngine) Error() <-chan error
func (sse *SSEEngine) VerifyHeader(fn func(http.Header) bool)
func (sse *SSEEngine) OnMessage(fn func(*SSEEvent)) (<-chan struct{}, error)
```

### HTML 解析

```go
func HTMLParse(HTML []byte) (doc QueryHTML, err error)
func (r *QueryHTML) SelectChild(el string, args ...map[string]string) QueryHTML
func (r *QueryHTML) SelectAllChild(el string, args ...map[string]string) (arrEls)
func (r *QueryHTML) Child() (childs []QueryHTML)
func (r *QueryHTML) ForEachChild(f func(index int, child QueryHTML) bool)
func (r *QueryHTML) NthChild(index int) QueryHTML
func (r *QueryHTML) Select(el string, args ...map[string]string) QueryHTML
func (r *QueryHTML) SelectAll(el string, args ...map[string]string) (arrEls)
func (r *QueryHTML) SelectBrother(el string, args ...map[string]string) QueryHTML
func (r *QueryHTML) SelectParent(el string, args ...map[string]string) QueryHTML
func (r *QueryHTML) Find(text string) QueryHTML
func (r *QueryHTML) Filter(el ...QueryHTML) QueryHTML
```

### 查询选择器

```go
func (eEls) ForEach(f func(index int, el QueryHTML) bool)
func (r QueryHTML) String() string
func (r QueryHTML) Exist() bool
func (r QueryHTML) Attr(key string) string
func (r QueryHTML) Attrs() map[string]string
func (r QueryHTML) Name() string
func (r QueryHTML) Text(trimSpace ...bool) string
func (r QueryHTML) FullText(trimSpace ...bool) string
func (r QueryHTML) HTML(trimSpace ...bool) string
```

### Requester

```go
func NewRequester(e *Engine) *Requester
func (r *Requester) Send(method, rawurl string, v ...interface{}) (*Res, error)
```

### 调试

```go
func OpenDebug(key ...string)
```

### 工具函数

```go
// 禁用分块传输
func DisableChunke(enable ...bool)
// 转换 Cookie 字符串为映射
func ConvertCookie(cookiesRaw string) map[string]*http.Cookie
// 生成随机用户代理
func RandomUserAgent() Header
// 创建 JSON 请求体
func BodyJSON(v interface{}) *bodyJson
// 创建 XML 请求体
func BodyXML(v interface{}) *bodyXml
// 创建文件请求体
func File(path string, field ...string) interface{}
// 创建新的 JSON-RPC 客户端
func NewJSONRPC(address string, path string, opts ...func(o *JSONRPCOptions)) (*JSONRPC, error)
// 创建 SSE 连接
func SSE(url string, v ...interface{}) (*SSEEngine, error)
```

## 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/sohaha/zlsgo/zhttp"
)

func main() {
    // 创建 HTTP 引擎
    engine := zhttp.New()
    
    // 配置引擎
    engine.EnableCookie(true)
    engine.SetTimeout(30 * time.Second)
    engine.EnableInsecureTLS(true)
    
    // 设置代理
    err := engine.SetProxyUrl("http://proxy.example.com:8080")
    if err != nil {
        fmt.Printf("设置代理失败: %v\n", err)
    }
    
    // GET 请求
    resp, err := zhttp.Get("https://httpbin.org/get")
    if err == nil {
        fmt.Printf("状态码: %d\n", resp.StatusCode())
        fmt.Printf("响应体: %s\n", resp.String())
    }
    
    // POST 请求
    resp, err = zhttp.Post("https://httpbin.org/post", map[string]interface{}{
        "name": "张三",
        "age":  25,
    })
    if err == nil {
        fmt.Printf("POST 响应: %s\n", resp.String())
    }
    
    // PUT 请求
    resp, err = zhttp.Put("https://httpbin.org/put", "更新数据")
    if err == nil {
        fmt.Printf("PUT 响应: %s\n", resp.String())
    }
    
    // DELETE 请求
    resp, err = zhttp.Delete("https://httpbin.org/delete")
    if err == nil {
        fmt.Printf("DELETE 响应: %s\n", resp.String())
    }
    
    // HEAD 请求
    resp, err = zhttp.Head("https://httpbin.org/headers")
    if err == nil {
        fmt.Printf("HEAD 响应状态码: %d\n", resp.StatusCode())
        fmt.Printf("响应头: %v\n", resp.Response().Header)
    }
    
    // OPTIONS 请求
    resp, err = zhttp.Options("https://httpbin.org/options")
    if err == nil {
        fmt.Printf("OPTIONS 响应: %s\n", resp.String())
    }
    
    // PATCH 请求
    resp, err = zhttp.Patch("https://httpbin.org/patch", "部分更新")
    if err == nil {
        fmt.Printf("PATCH 响应: %s\n", resp.String())
    }
    
    // CONNECT 请求
    resp, err = zhttp.Connect("https://httpbin.org/connect")
    if err == nil {
        fmt.Printf("CONNECT 响应: %s\n", resp.String())
    }
    
    // TRACE 请求
    resp, err = zhttp.Trace("https://httpbin.org/trace")
    if err == nil {
        fmt.Printf("TRACE 响应: %s\n", resp.String())
    }
    
    // 自定义方法
    resp, err = zhttp.Do("CUSTOM", "https://httpbin.org/anything", "自定义数据")
    if err == nil {
        fmt.Printf("自定义方法响应: %s\n", resp.String())
    }
    
    // 重试机制
    resp, err = zhttp.DoRetry(3, time.Second, func() (*zhttp.Res, error) {
        return zhttp.Get("https://httpbin.org/delay/2")
    })
    if err == nil {
        fmt.Printf("重试后响应: %s\n", resp.String())
    }
    
    // 响应处理
    resp, err = zhttp.Get("https://httpbin.org/json")
    if err == nil {
        // 获取字节数据
        bytes := resp.Bytes()
        fmt.Printf("字节数据长度: %d\n", len(bytes))
        
        // 获取字符串
        body := resp.String()
        fmt.Printf("响应体: %s\n", body)
        
        // 获取 Cookie
        cookies := resp.GetCookie()
        fmt.Printf("Cookie 数量: %d\n", len(cookies))
        
        // 转换为 JSON
        var data map[string]interface{}
        err = resp.ToJSON(&data)
        if err == nil {
            fmt.Printf("JSON 数据: %+v\n", data)
        }
        
        // 下载文件
        err = resp.ToFile("response.json")
        if err == nil {
            fmt.Println("文件下载成功")
        }
    }
    
    // HTML 解析示例
    htmlContent := `
    <html>
        <head><title>测试页面</title></head>
        <body>
            <div class="container">
                <h1 id="title">欢迎</h1>
                <p class="content">这是一个测试页面</p>
                <ul>
                    <li>项目 1</li>
                    <li>项目 2</li>
                    <li>项目 3</li>
                </ul>
            </div>
        </body>
    </html>
    `
    
    doc, err := zhttp.HTMLParse([]byte(htmlContent))
    if err == nil {
        // 选择元素
        title := doc.Select("#title")
        if title.Exist() {
            fmt.Printf("标题: %s\n", title.Text())
        }
        
        // 选择多个元素
        items := doc.SelectAll("li")
        items.ForEach(func(index int, el zhttp.QueryHTML) bool {
            fmt.Printf("项目 %d: %s\n", index+1, el.Text())
            return true
        })
        
        // 选择子元素
        content := doc.SelectChild("p", map[string]string{"class": "content"})
        if content.Exist() {
            fmt.Printf("内容: %s\n", content.Text())
        }
        
        // 选择所有子元素
        allDivs := doc.SelectAllChild("div")
        allDivs.ForEach(func(index int, el zhttp.QueryHTML) bool {
            fmt.Printf("DIV %d: %s\n", index+1, el.Name())
            return true
        })
        
        // 遍历子元素
        doc.ForEachChild(func(index int, child zhttp.QueryHTML) bool {
            fmt.Printf("子元素 %d: %s\n", index+1, child.Name())
            return true
        })
        
        // 获取第 N 个子元素
        firstChild := doc.NthChild(0)
        if firstChild.Exist() {
            fmt.Printf("第一个子元素: %s\n", firstChild.Name())
        }
        
        // 选择兄弟元素
        sibling := doc.SelectBrother("p")
        if sibling.Exist() {
            fmt.Printf("兄弟元素: %s\n", sibling.Text())
        }
        
        // 选择父元素
        parent := doc.SelectParent("div")
        if parent.Exist() {
            fmt.Printf("父元素: %s\n", parent.Name())
        }
        
        // 查找文本
        found := doc.Find("欢迎")
        if found.Exist() {
            fmt.Printf("找到文本: %s\n", found.Text())
        }
        
        // 过滤元素
        filtered := doc.Filter(doc.Select("h1"), doc.Select("p"))
        filtered.ForEach(func(index int, el zhttp.QueryHTML) bool {
            fmt.Printf("过滤后元素 %d: %s - %s\n", index+1, el.Name(), el.Text())
            return true
        })
        
        // 获取属性
        container := doc.Select(".container")
        if container.Exist() {
            fmt.Printf("容器类名: %s\n", container.Attr("class"))
            fmt.Printf("所有属性: %v\n", container.Attrs())
        }
        
        // 获取完整 HTML
        fullHTML := doc.HTML()
        fmt.Printf("完整 HTML 长度: %d\n", len(fullHTML))
    }
    
    // SSE 示例
    sseEngine, err := engine.SSE("https://example.com/events", func(opt *zhttp.SSEOption) {
        // 配置 SSE 选项
    })
    if err == nil {
        // 监听事件
        go func() {
            for event := range sseEngine.Event() {
                fmt.Printf("SSE 事件: %+v\n", event)
            }
        }()
        
        // 监听错误
        go func() {
            for err := range sseEngine.Error() {
                fmt.Printf("SSE 错误: %v\n", err)
            }
        }()
        
        // 等待完成
        <-sseEngine.Done()
        
        // 关闭 SSE
        sseEngine.Close()
    }
    
    // 流式处理
    resp, err = zhttp.Get("https://httpbin.org/stream/5")
    if err == nil {
        err = resp.Stream(func(line []byte, eof bool) error {
            if eof {
                fmt.Println("流处理完成")
                return nil
            }
            fmt.Printf("流数据: %s\n", string(line))
            return nil
        })
        if err != nil {
            fmt.Printf("流处理错误: %v\n", err)
        }
    }
    
    // 实际应用示例
    // 网页爬虫
    func crawlWebPage(url string) error {
        resp, err := zhttp.Get(url)
        if err != nil {
            return err
        }
        
        doc, err := resp.HTML()
        if err != nil {
            return err
        }
        
        // 提取标题
        title := doc.Select("title")
        if title.Exist() {
            fmt.Printf("页面标题: %s\n", title.Text())
        }
        
        // 提取链接
        links := doc.SelectAll("a")
        links.ForEach(func(index int, el zhttp.QueryHTML) bool {
            href := el.Attr("href")
            if href != "" {
                fmt.Printf("链接 %d: %s - %s\n", index+1, href, el.Text())
            }
            return true
        })
        
        return nil
    }
    
    // 使用爬虫
    err = crawlWebPage("https://example.com")
    if err != nil {
        fmt.Printf("爬取失败: %v\n", err)
    }
    
    // API 客户端
    type User struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
        Email string `json:"email"`
    }
    
    // 获取用户列表
    resp, err = zhttp.Get("https://jsonplaceholder.typicode.com/users")
    if err == nil {
        var users []User
        err = resp.ToJSON(&users)
        if err == nil {
            fmt.Printf("用户数量: %d\n", len(users))
            for _, user := range users {
                fmt.Printf("用户: %+v\n", user)
            }
        }
    }
    
    // 创建用户
    newUser := User{
        Name:  "新用户",
        Email: "newuser@example.com",
    }
    
    resp, err = zhttp.Post("https://jsonplaceholder.typicode.com/users", newUser)
    if err == nil {
        var createdUser User
        err = resp.ToJSON(&createdUser)
        if err == nil {
            fmt.Printf("创建的用户: %+v\n", createdUser)
        }
    }
    
    fmt.Println("HTTP 客户端示例完成")
}
```

## 类型定义

### SSE 选项
```go
type SSEOption struct {
    // SSE 配置选项
}
```

### SSE 事件
```go
type SSEEvent struct {
    // SSE 事件数据
}
```

### SSE 引擎
```go
type SSEEngine struct {
    // SSE 引擎
}
```

### 查询 HTML
```go
type QueryHTML struct {
    // HTML 查询对象
}
```

### 元素集合
```go
type Els []QueryHTML
```

## 最佳实践

1. 正确处理响应错误
2. 实现适当的重试机制
3. 缓存常用响应
4. 合理设置超时时间
