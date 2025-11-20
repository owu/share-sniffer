# Share Sniffer 技术实现文档

## 1. 项目概述

Share Sniffer（分享嗅探器）是一款跨平台的网盘分享链接检测工具，支持多种主流网盘的分享链接有效性检测。该工具提供了直观的图形界面（GUI）和便捷的命令行界面（CLI），方便用户根据需求选择使用方式。

### 1.1 支持的网盘类型

- ✅ 夸克网盘
- ✅ 天翼云盘
- ✅ 百度网盘
- ✅ 阿里云盘
- ✅ 115网盘
- ✅ 123云盘
- ✅ UC网盘
- ✅ 迅雷云盘

### 1.2 技术栈

- **开发语言**：Go 1.25
- **GUI框架**：[fyne.io/fyne/v2](https://fyne.io/) - 跨平台GUI框架
- **CLI框架**：[github.com/spf13/cobra](https://github.com/spf13/cobra) - 命令行框架
- **HTTP客户端**：自定义封装的HTTP客户端，支持重试和超时控制
- **并发模型**：工作池（WorkerPool）模式，支持并发任务处理

## 2. 项目结构

```
share-sniffer/
├── build/                  # 构建脚本和测试用例
│   ├── scripts/           # 自动化构建脚本
│   └── testcases/         # 测试用例
├── internal/              # 内部包
│   ├── app/               # 应用程序核心逻辑
│   ├── cmd/               # CLI命令行实现
│   ├── config/            # 配置管理
│   ├── core/              # 核心检测逻辑
│   ├── errors/            # 错误处理
│   ├── http/              # HTTP客户端封装
│   ├── logger/            # 日志记录
│   ├── ui/                # GUI界面实现
│   ├── utils/             # 工具函数
│   └── workerpool/        # 工作池实现
├── launcher/              # 启动器
│   ├── cli/               # CLI入口
│   └── gui/               # GUI入口
├── screenshot/            # 截图
├── static/                # 静态资源
├── .gitignore             # Git忽略文件
├── LICENSE                # 许可证
├── README.md              # 项目说明
├── README_en.md           # 英文说明
├── README_jp.md           # 日文说明
├── go.mod                 # Go模块依赖
└── go.sum                 # 依赖校验和
```

## 3. 核心模块设计

### 3.1 配置管理（config）

配置管理模块采用单例模式实现，提供全局配置访问。配置包含HTTP客户端配置、检测配置、应用信息和支持的链接类型等。

```go
// Config 应用配置结构
type Config struct {
    // HTTP客户端配置
    HTTPClientConfig struct {
        Timeout             time.Duration
        MaxIdleConns        int
        MaxIdleConnsPerHost int
        IdleConnTimeout     time.Duration
        RetryCount          int
    }

    // 检测配置
    CheckConfig struct {
        MaxConcurrentTasks int
        DefaultTimeout     time.Duration
        RetryInterval      time.Duration
        // 长耗时任务配置
        LongTimeout       time.Duration
        LongMaxConcurrent int
    }

    // 应用信息
    AppInfo struct {
        Version        string
        AppName        string
        AppNameCN      string
        ExpirationDate int64
    }

    // 支持的链接类型
    SupportedLinkTypes struct {
        AllLinks []string
        Quark    []string
        Telecom  []string
        Baidu    []string
        AliPan   []string
        Yyw      []string
        Yes      []string
        Uc       []string
        Xunlei   []string
    }
}
```

### 3.2 核心检测逻辑（core）

核心检测逻辑采用策略模式和工厂模式的组合设计，支持动态添加新的网盘检查器。

#### 3.2.1 接口设计

```go
// LinkChecker 链接检查器接口
type LinkChecker interface {
    // Check 检查链接有效性
    Check(ctx context.Context, urlStr string) utils.Result

    // GetPrefix 获取支持的链接前缀列表
    GetPrefix() []string
}
```

#### 3.2.2 注册机制

```go
// RegisterChecker 注册链接检查器
func RegisterChecker(checker LinkChecker) {
    prefixes := checker.GetPrefix()
    for _, prefix := range prefixes {
        checkers[prefix] = checker
        logger.Debug("LinkChecker:注册检查器,%s", prefix)
    }
}
```

#### 3.2.3 适配器模式

```go
// Adapter 适配器函数，根据URL前缀调用对应的检查器
func Adapter(ctx context.Context, urlStr string) utils.Result {
    // 输入验证
    if "" == urlStr {
        return utils.ErrorMalformed(urlStr, "链接不能为空")
    }

    // 获取对应的检查器
    checker := GetChecker(urlStr)
    if nil == checker {
        return utils.ErrorMalformed(urlStr, "链接尚未支持")
    }

    startTime := time.Now()
    result := checker.Check(ctx, urlStr)
    result.Data.URL = urlStr
    result.Data.Elapsed = time.Since(startTime).Milliseconds()
    result.Data.Name = strings.TrimSpace(result.Data.Name)

    return result
}
```

#### 3.2.4 检查器实现示例（夸克网盘）

```go
// QuarkChecker 夸克网盘链接检查器
type QuarkChecker struct{}

// Check 实现LinkChecker接口的Check方法
func (q *QuarkChecker) Check(ctx context.Context, urlStr string) utils.Result {
    return q.checkQuark(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
func (q *QuarkChecker) GetPrefix() []string {
    return config.GetSupportedQuark()
}
```

### 3.3 工作池（workerpool）

工作池模块实现了并发任务处理，支持普通任务和长耗时任务的并发控制，提高检测效率。

```go
// WorkerPool 工作池结构
type WorkerPool struct {
    workers           int
    queueSize         int
    taskQueue         chan Task
    resultChan        chan Result
    wg                sync.WaitGroup
    ctx               context.Context
    cancel            context.CancelFunc
    // 长耗时任务并发控制
    longTaskSemaphore   chan struct{}
    maxLongConcurrent int
}
```

工作池支持以下功能：
- 并发任务处理
- 长耗时任务并发控制
- 任务超时和取消
- 结果收集

### 3.4 HTTP客户端（http）

HTTP客户端模块封装了网络请求处理，支持重试、超时控制和错误处理。

```go
// DoWithRetry 执行HTTP请求并支持重试
func DoWithRetry(ctx context.Context, req *http.Request, maxRetries int) (*http.Response, error) {
    // 实现重试逻辑
}

// SetDefaultHeaders 设置默认请求头
func SetDefaultHeaders(req *http.Request) {
    req.Header.Set("accept", "application/json;charset=UTF-8")
    req.Header.Set("accept-language", "en,zh-CN;q=0.9,zh;q=0.8")
    req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
    req.Header.Set("cache-control", "no-cache")
    req.Header.Set("pragma", "no-cache")
}
```

## 4. 应用程序架构

### 4.1 GUI应用程序

GUI应用程序基于Fyne框架实现，提供直观的用户界面。

```go
// ShareSnifferApp 是应用程序的主结构
type ShareSnifferApp struct {
    app    fyne.App
    window fyne.Window
    state  *state.AppState
}

// NewShareSnifferApp 创建并初始化ShareSnifferApp的新实例
func NewShareSnifferApp() *ShareSnifferApp {
    // 实现应用程序初始化
}

// Run 启动应用程序
func (q *ShareSnifferApp) Run() {
    // 实现应用程序启动
}
```

### 4.2 CLI命令行

CLI命令行基于Cobra框架实现，提供便捷的命令行接口。

```go
var (
    rootCmd = &cobra.Command{
        Use:   "share-sniffer-cli [URL]",
        Short: "Share Sniffer CLI - A tool to detect and analyze shared links",
        Long:  `Share Sniffer CLI is a command-line tool that helps you detect and analyze shared links from various platforms.`,
        Args:  cobra.MaximumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            // 实现命令行处理逻辑
        },
    }
)
```

## 5. 实现流程

### 5.1 链接检测流程

1. 用户输入分享链接
2. 适配器函数根据URL前缀匹配对应的检查器
3. 检查器执行具体的检测逻辑：
   - 验证URL格式
   - 提取必要参数
   - 发送HTTP请求到网盘API
   - 解析响应结果
   - 返回检查结果
4. 适配器函数处理结果并返回

### 5.2 并发处理流程

1. 创建工作池实例
2. 启动工作协程
3. 提交检测任务到任务队列
4. 工作协程从任务队列中获取任务并执行
5. 长耗时任务使用信号量控制并发数
6. 任务执行完成后将结果发送到结果通道
7. 收集结果并返回给用户

## 6. 二次开发指南

### 6.1 添加新的网盘检查器

1. 创建新的检查器结构体，实现LinkChecker接口

```go
// NewPanChecker 新网盘链接检查器
type NewPanChecker struct{}

// Check 实现LinkChecker接口的Check方法
func (c *NewPanChecker) Check(ctx context.Context, urlStr string) utils.Result {
    // 实现新网盘的检查逻辑
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
func (c *NewPanChecker) GetPrefix() []string {
    return []string{"https://new.pan.example.com/s/"}
}
```

2. 在配置中添加新网盘的支持

```go
// 在config.go的SupportedLinkTypes结构体中添加
NewPan []string

// 在initDefault方法中初始化
b.SupportedLinkTypes.NewPan = []string{"https://new.pan.example.com/s/"}

// 在AllLinks中添加
b.SupportedLinkTypes.AllLinks = append(b.SupportedLinkTypes.AllLinks, b.SupportedLinkTypes.NewPan...)
```

3. 在register.go中注册新的检查器

```go
// 在init函数中添加
RegisterChecker(&NewPanChecker{})
```

### 6.2 自定义HTTP客户端配置

```go
// 在应用启动时设置环境变量
export MAX_CONCURRENT_TASKS=16
export HTTP_TIMEOUT=10s

// 或者修改config.go中的默认配置
func (q *Config) initDefault() {
    // 修改HTTP客户端默认配置
    q.HTTPClientConfig.Timeout = 10 * time.Second
    q.HTTPClientConfig.MaxIdleConns = 200
    q.HTTPClientConfig.MaxIdleConnsPerHost = 40
    
    // 修改检测默认配置
    q.CheckConfig.MaxConcurrentTasks = 16
}
```

### 6.3 扩展GUI界面

1. 在ui目录下创建新的界面组件

```go
// new_tab.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// NewNewTab 创建新标签页
func NewNewTab(window fyne.Window) *container.TabItem {
    // 实现新标签页的UI组件
    return container.NewTabItem("新功能", widget.NewLabel("新功能界面"))
}
```

2. 在application.go中添加新标签页

```go
// createContent 创建应用的主界面内容
func (q *ShareSnifferApp) createContent() fyne.CanvasObject {
    // 使用默认的Tabs布局 - 创建标签页容器
    tabs := container.NewAppTabs(
        // 添加检查标签页，用于检查分享链接
        check.NewCheckTab(q.window, q.state),
        // 添加新功能标签页
        ui.NewNewTab(q.window),
        // 添加关于标签页，显示应用信息
        about.NewAboutTab(),
    )
    // 设置标签页位置在窗口左侧
    tabs.SetTabLocation(container.TabLocationLeading)

    // 返回创建的UI内容
    return tabs
}
```

## 7. 性能优化

### 7.1 并发控制

- 使用工作池限制并发数，避免资源耗尽
- 长耗时任务使用信号量控制并发数，提高整体效率

### 7.2 网络请求优化

- 复用HTTP连接，减少连接建立开销
- 合理设置超时时间，避免长时间阻塞
- 实现指数退避重试策略，提高请求成功率

### 7.3 内存优化

- 限制任务队列和结果通道的容量
- 及时释放资源，避免内存泄漏

## 8. 测试

项目提供了测试用例，位于build/testcases目录下，可以使用以下命令运行测试：

```bash
# 运行单元测试
go test ./internal/core...

# 测试用例
cat build/testcases/all.txt

# 测试用例进行集成测试
go test ./internal/core...  -v -run="TestAdapter"


```

## 9. 打包编译

项目提供了自动化打包脚本，位于build/scripts目录下，支持Windows和Linux平台的打包：

```bash
# Windows平台
cd build/scripts
./build-gui-windows.ps1

# Linux平台
cd build/scripts
./build-gui-linux.sh
```

## 10. 许可证

[GNU GPL v3 License](LICENSE)

## 11. 贡献

欢迎提交Issue和Pull Request！

## 12. 联系方式

项目地址：https://github.com/owu/share-sniffer