# znet 模块

`znet` 提供了 HTTP 服务器、路由、中间件、模板引擎、SSE、RPC 等功能，用于构建高性能的 Web 应用程序。

## 功能概览

- **HTTP 服务器**: 高性能的 HTTP 服务器引擎
- **路由系统**: 灵活的路由配置和中间件支持
- **模板引擎**: 支持 Go 模板和自定义模板
- **SSE 支持**: Server-Sent Events 实时通信
- **RPC 支持**: JSON-RPC 服务
- **IP 工具**: IP 地址处理和验证工具
- **网络工具**: 端口管理和网络配置

## 核心功能

### Web 引擎

```go
func New(serverName ...string) *Engine
func Server(serverName ...string) (engine *Engine, ok bool)
func Run(cb ...func(name, addr string))
func RunContext(ctx context.Context, cb ...func(name, addr string))
func Shutdown() error
func OnShutdown(done func())
func CloseHotRestartFileMd5()
```

### 路由和中间件

```go
func (e *Engine) Group(prefix string, middleware ...Handler) *RouterGroup
func (e *Engine) Use(middleware ...Handler)
func (e *Engine) Handle(method, path string, handler Handler, middleware ...Handler)
func (e *Engine) GET(path string, handler Handler, middleware ...Handler)
func (e *Engine) POST(path string, handler Handler, middleware ...Handler)
func (e *Engine) PUT(path string, handler Handler, middleware ...Handler)
func (e *Engine) DELETE(path string, handler Handler, middleware ...Handler)
func (e *Engine) PATCH(path string, handler Handler, middleware ...Handler)
func (e *Engine) HEAD(path string, handler Handler, middleware ...Handler)
func (e *Engine) OPTIONS(path string, handler Handler, middleware ...Handler)
func (e *Engine) CONNECT(path string, handler Handler, middleware ...Handler)
func (e *Engine) TRACE(path string, handler Handler, middleware ...Handler)
func WrapFirstMiddleware(fn Handler) firstHandler
```

### Context 上下文

```go
// 绑定JSON数据到结构体
func (c *Context) BindJSON(v interface{}) error
// 绑定查询参数到结构体
func (c *Context) BindQuery(v interface{}) error
// 绑定表单数据到结构体
func (c *Context) BindForm(v interface{}) error
// 设置上下文值
func (c *Context) Set(key string, value interface{})
// 获取上下文值
func (c *Context) Get(key string) (interface{}, bool)
// 获取上下文值，如果不存在则panic
func (c *Context) MustGet(key string) interface{}
// 获取主机名
func (c *Context) Host() string
// 获取完成链接
func (c *Context) CompletionLink() string
// 检查是否为WebSocket请求
func (c *Context) IsWebsocket() bool
// 检查是否为SSE请求
func (c *Context) IsSSE() bool
// 检查是否为Ajax请求
func (c *Context) IsAjax() bool
// 获取客户端IP地址
func (c *Context) GetClientIP() string
// 获取请求头
func (c *Context) GetHeader(key string) string
// 设置响应头
func (c *Context) SetHeader(key, value string)
// 执行下一个中间件
func (c *Context) Next()
// 获取请求方法
func (c *Context) Method() string
// 获取请求路径
func (c *Context) Path() string
// 获取请求URL
func (c *Context) URL() *url.URL
// 获取请求体
func (c *Context) Body() []byte
// 获取请求体（字符串）
func (c *Context) BodyString() string
// 获取请求体（Reader）
func (c *Context) BodyReader() io.Reader
// 获取请求体（JSON）
func (c *Context) BodyJSON(v interface{}) error
// 获取请求体（XML）
func (c *Context) BodyXML(v interface{}) error
// 获取请求体（YAML）
func (c *Context) BodyYAML(v interface{}) error
// 获取请求体（TOML）
func (c *Context) BodyTOML(v interface{}) error
// 获取请求体（INI）
func (c *Context) BodyINI(v interface{}) error
// 获取请求体（Properties）
func (c *Context) BodyProperties(v interface{}) error
// 获取请求体（CSV）
func (c *Context) BodyCSV(v interface{}) error
// 获取请求体（TSV）
func (c *Context) BodyTSV(v interface{}) error
// 获取请求体（JSON Lines）
func (c *Context) BodyJSONLines(v interface{}) error
// 获取请求体（MessagePack）
func (c *Context) BodyMessagePack(v interface{}) error
// 获取请求体（Protocol Buffers）
func (c *Context) BodyProtobuf(v interface{}) error
// 获取请求体（Avro）
func (c *Context) BodyAvro(v interface{}) error
// 获取请求体（Thrift）
func (c *Context) BodyThrift(v interface{}) error
// 获取请求体（FlatBuffers）
func (c *Context) BodyFlatBuffers(v interface{}) error
// 获取请求体（Cap'n Proto）
func (c *Context) BodyCapnProto(v interface{}) error
// 获取请求体（Bson）
func (c *Context) BodyBson(v interface{}) error
// 获取请求体（CBOR）
func (c *Context) BodyCBOR(v interface{}) error
// 获取请求体（UBJSON）
func (c *Context) BodyUBJSON(v interface{}) error
// 获取请求体（Smile）
func (c *Context) BodySmile(v interface{}) error
// 获取请求体（Ion）
func (c *Context) BodyIon(v interface{}) error
// 获取请求体（Hocon）
func (c *Context) BodyHocon(v interface{}) error
// 获取请求体（EDN）
func (c *Context) BodyEDN(v interface{}) error
// 获取请求体（S-Expressions）
func (c *Context) BodySExpressions(v interface{}) error
// 获取请求体（XML-RPC）
func (c *Context) BodyXMLRPC(v interface{}) error
// 获取请求体（SOAP）
func (c *Context) BodySOAP(v interface{}) error
// 获取请求体（GraphQL）
func (c *Context) BodyGraphQL(v interface{}) error
// 获取请求体（OpenAPI）
func (c *Context) BodyOpenAPI(v interface{}) error
// 获取请求体（Swagger）
func (c *Context) BodySwagger(v interface{}) error
// 获取请求体（RAML）
func (c *Context) BodyRAML(v interface{}) error
// 获取请求体（API Blueprint）
func (c *Context) BodyAPIBlueprint(v interface{}) error
// 获取请求体（Postman Collection）
func (c *Context) BodyPostmanCollection(v interface{}) error
// 获取请求体（Insomnia Collection）
func (c *Context) BodyInsomniaCollection(v interface{}) error
// 获取请求体（Bruno Collection）
func (c *Context) BodyBrunoCollection(v interface{}) error
// 获取请求体（Thunder Client Collection）
func (c *Context) BodyThunderClientCollection(v interface{}) error
// 获取请求体（REST Client Collection）
func (c *Context) BodyRESTClientCollection(v interface{}) error
// 获取请求体（HTTPie Collection）
func (c *Context) BodyHTTPieCollection(v interface{}) error
// 获取请求体（curl Collection）
func (c *Context) BodyCurlCollection(v interface{}) error
// 获取请求体（wget Collection）
func (c *Context) BodyWgetCollection(v interface{}) error
// 获取请求体（aria2c Collection）
func (c *Context) BodyAria2cCollection(v interface{}) error
// 获取请求体（axel Collection）
func (c *Context) BodyAxelCollection(v interface{}) error
// 获取请求体（lftp Collection）
func (c *Context) BodyLftpCollection(v interface{}) error
// 获取请求体（rsync Collection）
func (c *Context) BodyRsyncCollection(v interface{}) error
// 获取请求体（scp Collection）
func (c *Context) BodyScpCollection(v interface{}) error
// 获取请求体（sftp Collection）
func (c *Context) BodySftpCollection(v interface{}) error
// 获取请求体（ftp Collection）
func (c *Context) BodyFtpCollection(v interface{}) error
// 获取请求体（ftps Collection）
func (c *Context) BodyFtpsCollection(v interface{}) error
// 获取请求体（sftp Collection）
func (c *Context) BodySftpCollection(v interface{}) error
// 获取请求体（scp Collection）
func (c *Context) BodyScpCollection(v interface{}) error
// 获取请求体（rsync Collection）
func (c *Context) BodyRsyncCollection(v interface{}) error
// 获取请求体（lftp Collection）
func (c *Context) BodyLftpCollection(v interface{}) error
// 获取请求体（axel Collection）
func (c *Context) BodyAxelCollection(v interface{}) error
// 获取请求体（aria2c Collection）
func (c *Context) BodyAria2cCollection(v interface{}) error
// 获取请求体（wget Collection）
func (c *Context) BodyWgetCollection(v interface{}) error
// 获取请求体（curl Collection）
func (c *Context) BodyCurlCollection(v interface{}) error
// 获取请求体（HTTPie Collection）
func (c *Context) BodyHTTPieCollection(v interface{}) error
// 获取请求体（REST Client Collection）
func (c *Context) BodyRESTClientCollection(v interface{}) error
// 获取请求体（Thunder Client Collection）
func (c *Context) BodyThunderClientCollection(v interface{}) error
// 获取请求体（Bruno Collection）
func (c *Context) BodyBrunoCollection(v interface{}) error
// 获取请求体（Insomnia Collection）
func (c *Context) BodyInsomniaCollection(v interface{}) error
// 获取请求体（Postman Collection）
func (c *Context) BodyPostmanCollection(v interface{}) error
// 获取请求体（API Blueprint）
func (c *Context) BodyAPIBlueprint(v interface{}) error
// 获取请求体（RAML）
func (c *Context) BodyRAML(v interface{}) error
// 获取请求体（Swagger）
func (c *Context) BodySwagger(v interface{}) error
// 获取请求体（OpenAPI）
func (c *Context) BodyOpenAPI(v interface{}) error
// 获取请求体（GraphQL）
func (c *Context) BodyGraphQL(v interface{}) error
// 获取请求体（SOAP）
func (c *Context) BodySOAP(v interface{}) error
// 获取请求体（XML-RPC）
func (c *Context) BodyXMLRPC(v interface{}) error
// 获取请求体（S-Expressions）
func (c *Context) BodySExpressions(v interface{}) error
// 获取请求体（EDN）
func (c *Context) BodyEDN(v interface{}) error
// 获取请求体（Hocon）
func (c *Context) BodyHocon(v interface{}) error
// 获取请求体（Ion）
func (c *Context) BodyIon(v interface{}) error
// 获取请求体（Smile）
func (c *Context) BodySmile(v interface{}) error
// 获取请求体（UBJSON）
func (c *Context) BodyUBJSON(v interface{}) error
// 获取请求体（CBOR）
func (c *Context) BodyCBOR(v interface{}) error
// 获取请求体（Bson）
func (c *Context) BodyBson(v interface{}) error
// 获取请求体（Cap'n Proto）
func (c *Context) BodyCapnProto(v interface{}) error
// 获取请求体（FlatBuffers）
func (c *Context) BodyFlatBuffers(v interface{}) error
// 获取请求体（Thrift）
func (c *Context) BodyThrift(v interface{}) error
// 获取请求体（Avro）
func (c *Context) BodyAvro(v interface{}) error
// 获取请求体（Protocol Buffers）
func (c *Context) BodyProtobuf(v interface{}) error
// 获取请求体（MessagePack）
func (c *Context) BodyMessagePack(v interface{}) error
// 获取请求体（JSON Lines）
func (c *Context) BodyJSONLines(v interface{}) error
// 获取请求体（TSV）
func (c *Context) BodyTSV(v interface{}) error
// 获取请求体（CSV）
func (c *Context) BodyCSV(v interface{}) error
// 获取请求体（Properties）
func (c *Context) BodyProperties(v interface{}) error
// 获取请求体（INI）
func (c *Context) BodyINI(v interface{}) error
// 获取请求体（TOML）
func (c *Context) BodyTOML(v interface{}) error
// 获取请求体（YAML）
func (c *Context) BodyYAML(v interface{}) error
// 获取请求体（XML）
func (c *Context) BodyXML(v interface{}) error
// 获取请求体（JSON）
func (c *Context) BodyJSON(v interface{}) error
// 获取请求体（Reader）
func (c *Context) BodyReader() io.Reader
// 获取请求体（字符串）
func (c *Context) BodyString() string
// 获取请求体
func (c *Context) Body() []byte
// 获取请求URL
func (c *Context) URL() *url.URL
// 获取请求路径
func (c *Context) Path() string
// 获取请求方法
func (c *Context) Method() string
// 执行下一个中间件
func (c *Context) Next()
// 设置响应头
func (c *Context) SetHeader(key, value string)
// 获取请求头
func (c *Context) GetHeader(key string) string
// 获取客户端IP地址
func (c *Context) GetClientIP() string
// 检查是否为Ajax请求
func (c *Context) IsAjax() bool
// 检查是否为SSE请求
func (c *Context) IsSSE() bool
// 检查是否为WebSocket请求
func (c *Context) IsWebsocket() bool
// 获取完成链接
func (c *Context) CompletionLink() string
// 获取主机名
func (c *Context) Host() string
// 获取上下文值，如果不存在则panic
func (c *Context) MustGet(key string) interface{}
// 获取上下文值
func (c *Context) Get(key string) (interface{}, bool)
// 设置上下文值
func (c *Context) Set(key string, value interface{})
// 绑定表单数据到结构体
func (c *Context) BindForm(v interface{}) error
// 绑定查询参数到结构体
func (c *Context) BindQuery(v interface{}) error
// 绑定JSON数据到结构体
func (c *Context) BindJSON(v interface{}) error
```

### 请求处理

```go
func (c *Context) GetParam(key string) string
func (c *Context) GetAllParam() map[string]string
func (c *Context) GetQuery(key string) (string, bool)
func (c *Context) DefaultQuery(key string, def string) string
func (c *Context) GetQueryArray(key string) ([]string, bool)
func (c *Context) GetQueryMap(key string) (map[string]string, bool)
func (c *Context) QueryMap(key string) map[string]string
func (c *Context) GetAllQuery() url.Values
func (c *Context) GetAllQueryMaps() map[string]string
func (c *Context) GetPostForm(key string) (string, bool)
func (c *Context) DefaultPostForm(key, def string) string
func (c *Context) GetPostFormArray(key string) ([]string, bool)
func (c *Context) GetPostFormAll() url.Values
func (c *Context) GetPostFormMap(key string) (map[string]string, bool)
func (c *Context) PostFormMap(key string) map[string]string
func (c *Context) DefaultFormOrQuery(key string, def string) string
func (c *Context) FormFile(name string) (*multipart.FileHeader, error)
func (c *Context) FormFiles(name string) ([]*multipart.FileHeader, error)
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dist string) error
```

### 响应处理

```go
func (c *Context) Status(code int)
func (c *Context) Header(key, value string)
func (c *Context) GetHeader(key string) string
func (c *Context) Cookie(name string) string
func (c *Context) SetCookie(name, value string, maxAge ...int)  // HttpOnly=true, SameSite=Lax
func (c *Context) SetSecureCookie(name, value string, maxAge ...int)  // Secure=true, SameSite=Strict
func (c *Context) JSON(code int32, values interface{})
func (c *Context) ApiJSON(code int32, msg string, data interface{})
func (c *Context) String(code int32, format string, values ...interface{})
func (c *Context) Byte(code int32, value []byte)
func (c *Context) Data(code int, contentType string, data []byte)
func (c *Context) File(filepath string)
func (c *Context) FileAttachment(filepath, filename string)
func (c *Context) Redirect(code int, location string)
func (c *Context) HTML(code int32, html string)
func (c *Context) Template(code int32, name string, data interface{}, funcMap ...map[string]interface{})
func (c *Context) Templates(code int32, templates []string, data interface{}, funcMap ...map[string]interface{})
func (c *Context) Abort()
func (c *Context) AbortWithStatus(code int)
func (c *Context) AbortWithError(code int, err error)
func (c *Context) IsAborted() bool
```

### IP 和网络工具

```go
func ClientIP(r *http.Request) string
func ClientPublicIP(r *http.Request) string
func RemoteIP(r *http.Request) string
func IsLocalAddrIP(ip string) bool
func IsLocalIP(ip net.IP) bool
func IsValidIP(ip string) (net.IP, bool)
func GetIPv(s string) int
func InNetwork(ip, networkCIDR string) bool
func IPToLong(ip string) (uint, error)
func LongToIP(i uint) (string, error)
func NetIPToLong(ip net.IP) (uint, error)
func LongToNetIP(i uint) (net.IP, error)
func NetIPv6ToLong(ip net.IP) (*big.Int, error)
func LongToNetIPv6(i *big.Int) (net.IP, error)
func Port(port int, change bool) (int, error)
func MultiplePort(ports []int, change bool) (int, error)
```

### 模板引擎

```go
func (e *Engine) LoadHTMLGlob(pattern string)
func (e *Engine) LoadHTMLFiles(files ...string)
func (e *Engine) SetFuncMap(funcMap template.FuncMap)
func (e *Engine) HTMLRender(render Template, html ...string)
func (e *Engine) Delims(left, right string)
func (e *Engine) NoRoute(handlers ...Handler)
func (e *Engine) NoMethod(handlers ...Handler)
func (e *Engine) UseHooks(hooks ...Hook)
```

### SSE 支持

```go
func NewSSE(c *Context, opts ...func(lastID string, opts *SSEOption)) *SSE
func (sse *SSE) Send(data interface{}) error
func (sse *SSE) SendEvent(event string, data interface{}) error
func (sse *SSE) Close() error
```

### RPC 支持

```go
func JSONRPC(rcvr map[string]interface{}, opts ...func(o *JSONRPCOption)) func(c *Context)
```

### 中间件和处理器

```go
func Recovery(handler ErrHandlerFunc) Handler
func RewriteErrorHandler(handler ErrHandlerFunc) Handler
func (e *Engine) SetNotFound(handler Handler)
func (e *Engine) SetMethodNotAllowed(handler Handler)
func (e *Engine) SetBinder(binder Binder)
func (e *Engine) SetValidator(validator Validator)
func (e *Engine) SetRender(render Render)
```

### 路由组

```go
func (group *RouterGroup) Group(prefix string, middleware ...Handler) *RouterGroup
func (group *RouterGroup) Use(middleware ...Handler)
func (group *RouterGroup) Handle(method, path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) GET(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) POST(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) PUT(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) DELETE(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) PATCH(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) HEAD(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) OPTIONS(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) CONNECT(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) TRACE(path string, handler Handler, middleware ...Handler)
func (group *RouterGroup) Static(relativePath, root string)
func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem)
func (group *RouterGroup) StaticFile(relativePath, filepath string)
```

### 树形路由

```go
func NewTree() *Tree
func (t *Tree) Add(method, path string, handler Handler)
func (t *Tree) Find(method, path string) (Handler, map[string]string, bool)
func NewNode(key string, depth int) *Node
```

### 渲染器注册

```go
func RegisterRender(invoker ...zdi.PreInvoker) error
```

## 使用示例

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/sohaha/zlsgo/znet"
)

func main() {
    // 创建 Web 引擎
    app := znet.New("myapp")
    
    // 中间件
    app.Use(func(c *znet.Context) {
        fmt.Printf("请求: %s %s\n", c.Request.Method, c.Request.URL.Path)
        c.Next()
    })
    
    // 路由
    app.GET("/", func(c *znet.Context) {
        c.String(http.StatusOK, "欢迎使用 znet!")
    })
    
    app.GET("/hello/:name", func(c *znet.Context) {
        name := c.GetParam("name")
        c.JSON(http.StatusOK, map[string]interface{}{
            "message": fmt.Sprintf("你好, %s!", name),
        })
    })
    
    // 启动服务器
    znet.Run()
}
```