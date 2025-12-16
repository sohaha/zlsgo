# zfile 模块

`zfile` 提供了文件操作、压缩解压、内存文件、文件锁、文件句柄等功能，用于高效的文件系统操作和文件管理。

## 功能概览

- **文件操作**: 文件存在性检查、大小获取、复制、删除等
- **压缩解压**: 文件压缩和解压缩
- **内存文件**: 内存中的文件操作
- **文件锁**: 文件锁定和解锁
- **文件句柄**: 文件句柄管理

## 核心功能

### 文件和目录操作

```go
// 检查路径是否存在（返回类型：0=文件，1=目录，-1=不存在）
func PathExist(path string) (int, error)
// 检查目录是否存在
func DirExist(path string) bool
// 检查文件是否存在
func FileExist(path string) bool
// 获取文件大小（字符串格式）
func FileSize(file string) string
// 获取文件大小（无符号整数）
func FileSizeUint(file string) uint64
// 格式化文件大小
func SizeFormat(s uint64) string
// 获取根路径
func RootPath() string
// 获取临时目录路径
func TmpPath(pattern ...string) string
// 获取安全路径
func SafePath(path string, pathRange ...string) string
// 获取真实路径
func RealPath(path string, addSlash ...bool) string
// 获取真实路径并创建目录
func RealPathMkdir(path string, addSlash ...bool) string
// 检查是否为子路径
func IsSubPath(subPath, path string) bool
// 删除目录
func Rmdir(path string, notIncludeSelf ...bool) bool
// 删除文件或目录
func Remove(path string) error
// 复制文件
func CopyFile(source string, dest string) error
// 获取可执行文件路径
func ExecutablePath() string
// 获取程序路径
func ProgramPath(addSlash ...bool) string
// 获取MIME类型
func GetMimeType(filename string, content []byte) string
// 检查是否有权限
func HasPermission(path string, perm os.FileMode, noUp ...bool) bool
// 检查是否有读写权限
func HasReadWritePermission(path string) bool
// 获取目录统计信息
func StatDir(path string, options ...DirStatOptions) (size, total uint64, err error)
```

### 压缩解压

```go
func GzCompress(currentPath, dest string) error
func GzDeCompress(tarFile, dest string) error
func ZipCompress(currentPath, dest string) error
func ZipDeCompress(zipFile, dest string) error
```

### 内存文件

```go
// 创建新的内存文件
func NewMemoryFile(name string, opt ...MemoryFileOption) *MemoryFile
// 设置自动刷新间隔
func MemoryFileAutoFlush(second int64) func(*MemoryFile)
// 设置刷新前回调
func MemoryFileFlushBefore(fn memoryFileFlushBefore) func(*MemoryFile)
// 设置文件名
func (f *MemoryFile) SetName(name string)
// 获取字节数据
func (f *MemoryFile) Bytes() []byte
// 获取文件信息
func (f *MemoryFile) Stat() (os.FileInfo, error)
// 读取数据
func (f *MemoryFile) Read(buffer []byte) (int, error)
// 关闭文件
func (f *MemoryFile) Close() error
// 同步数据
func (f *MemoryFile) Sync() error
// 写入数据
func (f *MemoryFile) Write(buffer []byte) (int, error)
// 定位文件指针
func (f *MemoryFile) Seek(offset int64, whence int) (int64, error)
// 获取文件名
func (f *MemoryFile) Name() string
// 获取文件大小
func (f *MemoryFile) Size() int64
// 获取文件模式
func (f *MemoryFile) Mode() os.FileMode
// 获取修改时间
func (f *MemoryFile) ModTime() time.Time
// 是否为目录
func (f *MemoryFile) IsDir() bool
// 获取系统信息
func (f *MemoryFile) Sys() interface{}
```

### 文件锁

```go
func NewFileLock(path string) *FileLock
func (l *FileLock) Lock() error
func (l *FileLock) Unlock() error
func (l *FileLock) Clean() error
```

### 文件句柄

```go
func CopyDir(source string, dest string, filterFn ...func(srcFilePath, destFilePath string) bool) error
func ReadFile(path string) ([]byte, error)
func ReadLineFile(path string, handle func(line int, data []byte) error) error
func WriteFile(path string, b []byte, isAppend ...bool) error
func PutOffset(path string, b []byte, offset int64) error
func PutAppend(path string, b []byte) error
```

### 跨平台支持

```go
func MoveFile(source string, dest string, force ...bool) error
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/zfile"
)

func main() {
    // 文件操作示例
    filePath := "test.txt"
    
    // 检查文件是否存在
    if exist, err := zfile.PathExist(filePath); err == nil && exist == 0 {
        fmt.Printf("文件 %s 存在\n", filePath)
        
        // 获取文件大小
        sizeStr := zfile.FileSize(filePath)
        fmt.Printf("文件大小: %s\n", sizeStr)
        
        sizeUint := zfile.FileSizeUint(filePath)
        fmt.Printf("文件大小 (uint): %d 字节\n", sizeUint)
    } else {
        fmt.Printf("文件 %s 不存在\n", filePath)
    }
    
    // 检查目录是否存在
    dirPath := "testdir"
    if zfile.DirExist(dirPath) {
        fmt.Printf("目录 %s 存在\n", dirPath)
    } else {
        fmt.Printf("目录 %s 不存在\n", dirPath)
    }
    
    // 复制文件
    srcFile := "source.txt"
    dstFile := "destination.txt"
    
    err := zfile.CopyFile(srcFile, dstFile)
    if err == nil {
        fmt.Printf("文件复制成功: %s -> %s\n", srcFile, dstFile)
    } else {
        fmt.Printf("文件复制失败: %v\n", err)
    }
    
    // 复制目录
    srcDir := "sourcedir"
    dstDir := "destdir"
    
    err = zfile.CopyDir(srcDir, dstDir)
    if err == nil {
        fmt.Printf("目录复制成功: %s -> %s\n", srcDir, dstDir)
    } else {
        fmt.Printf("目录复制失败: %v\n", err)
    }
    
    // 文件压缩示例
    sourcePath := "source"
    gzDest := "source.tar.gz"
    
    err = zfile.GzCompress(sourcePath, gzDest)
    if err == nil {
        fmt.Printf("Gzip 压缩成功: %s\n", gzDest)
    } else {
        fmt.Printf("Gzip 压缩失败: %v\n", err)
    }
    
    // ZIP 压缩示例
    zipDest := "source.zip"
    err = zfile.ZipCompress(sourcePath, zipDest)
    if err == nil {
        fmt.Printf("ZIP 压缩成功: %s\n", zipDest)
    } else {
        fmt.Printf("ZIP 压缩失败: %v\n", err)
    }
    
    // 内存文件示例
    memFile := zfile.NewMemoryFile("test.txt")
    
    // 写入数据
    data := []byte("Hello, Memory File!")
    _, err = memFile.Write(data)
    if err == nil {
        fmt.Println("数据写入内存文件成功")
    }
    
    // 读取数据
    buffer := make([]byte, len(data))
    _, err = memFile.Read(buffer)
    if err == nil {
        fmt.Printf("从内存文件读取: %s\n", string(buffer))
    }
    
    // 获取文件信息
    info, err := memFile.Stat()
    if err == nil {
        fmt.Printf("内存文件大小: %d 字节\n", info.Size())
        fmt.Printf("内存文件名称: %s\n", info.Name())
    }
    
    // 关闭内存文件
    memFile.Close()
    
    // 文件锁示例
    lockPath := "test.lock"
    fileLock := zfile.NewFileLock(lockPath)
    
    // 获取锁
    err = fileLock.Lock()
    if err == nil {
        fmt.Println("文件锁获取成功")
        
        // 执行需要锁保护的操作
        fmt.Println("执行受保护的操作...")
        
        // 释放锁
        err = fileLock.Unlock()
        if err == nil {
            fmt.Println("文件锁释放成功")
        }
    } else {
        fmt.Printf("文件锁获取失败: %v\n", err)
    }
    
    // 清理锁文件
    fileLock.Clean()
    
    // 文件读写示例
    testFile := "testfile.txt"
    
    // 写入文件
    content := []byte("这是测试文件内容")
    err = zfile.WriteFile(testFile, content)
    if err == nil {
        fmt.Println("文件写入成功")
    }
    
    // 读取文件
    readContent, err := zfile.ReadFile(testFile)
    if err == nil {
        fmt.Printf("文件内容: %s\n", string(readContent))
    }
    
    // 追加内容
    appendContent := []byte("\n这是追加的内容")
    err = zfile.PutAppend(testFile, appendContent)
    if err == nil {
        fmt.Println("内容追加成功")
    }
    
    // 在指定位置写入
    offsetContent := []byte("替换")
    err = zfile.PutOffset(testFile, offsetContent, 0)
    if err == nil {
        fmt.Println("偏移写入成功")
    }
    
    // 逐行读取文件
    err = zfile.ReadLineFile(testFile, func(line int, data []byte) error {
        fmt.Printf("第 %d 行: %s\n", line+1, string(data))
        return nil
    })
    if err != nil {
        fmt.Printf("逐行读取失败: %v\n", err)
    }
    
    // 路径操作示例
    rootPath := zfile.RootPath()
    fmt.Printf("根路径: %s\n", rootPath)
    
    tmpPath := zfile.TmpPath("test_*.txt")
    fmt.Printf("临时路径: %s\n", tmpPath)
    
    realPath := zfile.RealPath("./testdir", true)
    fmt.Printf("真实路径: %s\n", realPath)
    
    safePath := zfile.SafePath("/etc/passwd", "/home")
    fmt.Printf("安全路径: %s\n", safePath)
    
    // 权限检查
    if zfile.HasReadWritePermission(testFile) {
        fmt.Printf("文件 %s 有读写权限\n", testFile)
    } else {
        fmt.Printf("文件 %s 没有读写权限\n", testFile)
    }
    
    // 获取 MIME 类型
    mimeType := zfile.GetMimeType("test.txt", content)
    fmt.Printf("文件 MIME 类型: %s\n", mimeType)
    
    // 移动文件
    newPath := "moved.txt"
    err = zfile.MoveFile(testFile, newPath)
    if err == nil {
        fmt.Printf("文件移动成功: %s -> %s\n", testFile, newPath)
    } else {
        fmt.Printf("文件移动失败: %v\n", err)
    }
    
    // 清理测试文件
    zfile.Remove(newPath)
    zfile.Remove("testfile.txt")
    zfile.Remove("source.tar.gz")
    zfile.Remove("source.zip")
    
    fmt.Println("测试完成")
}
```

## 文件操作模式

### 安全操作
```go
// 检查文件存在性
exist, err := zfile.PathExist(path)
if err == nil && exist == 0 {
    // 文件存在
}

// 安全路径
safePath := zfile.SafePath(path, allowedRange)
```

### 批量操作
```go
// 复制目录
err := zfile.CopyDir(source, dest, func(src, dst string) bool {
    // 过滤条件
    return true
})

// 统计目录
size, total, err := zfile.StatDir(path)
```

## 最佳实践

1. 使用安全的路径操作
2. 实现适当的错误处理
3. 合理使用文件锁
4. 监控文件操作性能