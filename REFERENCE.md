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
- ✅ 移动云盘(139云盘)

### 1.2 技术栈

- **开发语言**：Go 1.25
- **GUI框架**：[fyne.io/fyne/v2](https://fyne.io/) - 跨平台GUI框架，支持Android平台
- **CLI框架**：[github.com/spf13/cobra](https://github.com/spf13/cobra) - 命令行框架
- **HTTP客户端**：自定义封装的HTTP客户端，支持重试和超时控制
- **并发模型**：工作池（WorkerPool）模式，支持并发任务处理
- **浏览器自动化**：[chromedp](https://github.com/chromedp/chromedp) - 用于动态页面内容提取
- **对话框库**：[sqweek/dialog](https://github.com/sqweek/dialog) - 桌面端文件选择对话框
- **日志系统**：自定义日志框架，支持多级别日志输出

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

### 3.1 应用程序架构（app）

应用程序模块采用分层架构设计，实现了GUI和CLI的统一启动入口。主要包含以下组件：

#### 3.1.1 ShareSnifferApp结构体
```go
type ShareSnifferApp struct {
    app    fyne.App      // Fyne应用实例
    window fyne.Window   // 主窗口
    state  *state.AppState // 应用状态
}
```

#### 3.1.2 应用启动流程
1. **配置初始化**：从全局配置单例获取应用配置
2. **Fyne应用创建**：使用`app.NewWithID()`创建应用实例
3. **窗口配置**：设置窗口大小、位置和标题
4. **UI内容创建**：构建标签页界面（检查页和关于页）
5. **后台任务启动**：时间同步和版本检查
6. **事件循环启动**：显示窗口并进入主事件循环

#### 3.1.3 平台适配
- **桌面平台**：使用sqweek/dialog提供原生文件选择对话框
- **Android平台**：使用Fyne原生对话框，避免调用桌面专用API

### 3.2 配置管理（config）

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

#### 6.1.1 基于HTTP API的检查器实现示例（以115网盘为例）

```go
// YywChecker 115网盘链接检查器
type YywChecker struct{}

// Check 实现LinkChecker接口的Check方法
func (y *YywChecker) Check(ctx context.Context, urlStr string) utils.Result {
    return y.checkYyw(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
func (y *YywChecker) GetPrefix() []string {
    return config.GetSupportedYyw()
}

// checkYyw 115网盘核心检测逻辑
func (y *YywChecker) checkYyw(ctx context.Context, urlStr string) utils.Result {
    // 1. URL格式验证和参数提取
    parsedURL, err := url.ParseRequestURI(urlStr)
    if err != nil {
        return utils.ErrorMalformed(urlStr, "链接格式无效")
    }
    
    // 2. 提取分享码和访问密码
    shareCode := extractShareCode(parsedURL.Path)
    
    // 3. 构建API请求
    apiURL := "https://webapi.115.com/share/snap"
    params := url.Values{
        "share_code": {shareCode},
        "user_id":    {"0"},
    }
    
    // 4. 发送HTTP请求
    client := apphttp.NewClient()
    resp, err := client.Get(ctx, apiURL+"?"+params.Encode())
    if err != nil {
        return utils.ErrorFatal("网络请求失败: " + err.Error())
    }
    defer resp.Body.Close()
    
    // 5. 解析API响应
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return utils.ErrorFatal("响应解析失败: " + err.Error())
    }
    
    // 6. 处理检测结果
    if code, ok := result["state"].(float64); ok && code == 0 {
        return utils.Success(urlStr, "分享有效", time.Since(requestStart))
    } else {
        return utils.ErrorExpired(urlStr, "分享已失效")
    }
}
```

#### 6.1.2 基于浏览器自动化的检查器实现示例（以移动云盘为例）

```go
// YdChecker 移动云盘(139云盘)链接检查器
type YdChecker struct{}

// Check 实现LinkChecker接口的Check方法
func (y *YdChecker) Check(ctx context.Context, urlStr string) utils.Result {
    return y.checkYd(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
func (y *YdChecker) GetPrefix() []string {
    return config.GetSupportedYd()
}

// checkYd 移动云盘核心检测逻辑（使用chromedp）
func (y *YdChecker) checkYd(ctx context.Context, urlStr string) utils.Result {
    // 1. 配置Chrome浏览器选项（性能优化）
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("blink-settings", "imagesEnabled=false,cssEnabled=false"),
        chromedp.Flag("disable-plugins", true),
        chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
    )
    
    // 2. 创建浏览器上下文
    execCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
    defer cancel()
    
    browserCtx, cancel := chromedp.NewContext(execCtx)
    defer cancel()
    
    // 3. 导航到页面并提取内容
    var pageContent string
    err := chromedp.Run(browserCtx,
        chromedp.Navigate(urlStr),
        chromedp.WaitVisible("body", chromedp.ByQuery),
        chromedp.OuterHTML("html", &pageContent, chromedp.ByQuery),
    )
    
    // 4. 解析页面内容判断有效性
    if err != nil {
        return utils.ErrorFatal("页面访问失败: " + err.Error())
    }
    
    if strings.Contains(pageContent, "文件不存在") || 
       strings.Contains(pageContent, "分享已失效") {
        return utils.ErrorExpired(urlStr, "分享已失效")
    }
    
    return utils.Success(urlStr, "分享有效", time.Since(requestStart))
}
```

#### 6.1.3 注册新的检查器

在`internal/core/register.go`中注册新的检查器：

```go
func registerCheckers() {
    once.Do(func() {
        // 注册夸克网盘检查器
        RegisterChecker(&QuarkChecker{})
        // 注册电信云盘检查器
        RegisterChecker(&TelecomChecker{})
        // 注册百度网盘检查器
        RegisterChecker(&BaiduChecker{})
        // 注册阿里云盘检查器
        RegisterChecker(&AliPanChecker{})
        // 注册115网盘检查器
        RegisterChecker(&YywChecker{})
        // 注册123网盘检查器
        RegisterChecker(&YesChecker{})
        // 注册UC网盘检查器
        RegisterChecker(&UcChecker{})
        // 注册移动云盘检查器
        RegisterChecker(&YdChecker{})

        if utils.IsDesktop() {
            // 注册迅雷网盘检查器
            RegisterChecker(&XunleiChecker{})
        }
    })
}
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
    content := container.NewVBox(
        widget.NewLabel("新功能界面"),
        widget.NewButton("新功能", func() {
            // 实现新功能逻辑
        }),
    )
    
    return container.NewTabItem("新功能", content)
}
```

2. 在应用启动时添加新标签页

```go
func (q *ShareSnifferApp) createContent() fyne.CanvasObject {
    tabs := container.NewAppTabs(
        check.NewCheckTab(q.window, q.state),
        about.NewAboutTab(q.window),
        ui.NewNewTab(q.window), // 添加新标签页
    )
    tabs.SetTabLocation(container.TabLocationLeading)
    return tabs
}
```

## 7. 构建与测试流程

### 7.1 构建脚本

项目提供了多种构建脚本，支持跨平台编译：

#### 7.1.1 Windows平台构建
```powershell
# 构建GUI版本
.\build\scripts\build-gui-windows.ps1

# 构建CLI版本  
.\build\scripts\build-cli-windows.ps1

# 构建Android版本
.\build\scripts\build-android.ps1

# 批量构建所有版本
.\build\scripts\build-all.ps1
```

#### 7.1.2 Linux平台构建
```bash
# 构建GUI版本
./build/scripts/build-gui-linux.sh

# 构建CLI版本
./build/scripts/build-cli-linux.sh

# 构建Android版本
./build/scripts/build-android.sh

# 批量构建所有版本
./build/scripts/build-all.sh
```

### 7.2 测试用例管理

项目使用测试用例文件进行功能验证：

#### 7.2.1 测试用例文件结构
```
build/testcases/
├── alipan.txt      # 阿里云盘测试用例
├── baidu.txt       # 百度网盘测试用例
├── quark.txt       # 夸克网盘测试用例
├── telecom.txt     # 天翼云盘测试用例
├── yd.txt          # 移动云盘测试用例
├── yes.txt         # 123云盘测试用例
├── yyw.txt         # 115网盘测试用例
├── xunlei.txt      # 迅雷云盘测试用例
├── uc.txt          # UC网盘测试用例
└── all.txt         # 所有网盘测试用例
```

#### 7.2.2 测试用例合并脚本
```powershell
# 合并所有测试用例到all.txt
.\build\testcases\merge.ps1
```

### 7.3 开发运行

#### 7.3.1 GUI模式开发运行
```bash
# 初始化依赖
go mod tidy

# 运行GUI应用
go run ./launcher/gui/main.go

# 开发模式（详细编译信息）
go clean -cache && go clean -modcache && go run -x ./launcher/gui/main.go
```

#### 7.3.2 CLI模式开发运行
```bash
# 运行CLI应用
go run ./launcher/cli/main.go

# 带参数运行
go run ./launcher/cli/main.go check --file "test.txt"
```

### 7.4 平台适配说明

#### 7.4.1 Android平台支持
- 使用Fyne框架原生支持Android
- 通过构建标签区分平台特定代码
- 自动使用Fyne原生对话框替代桌面专用API

#### 7.4.2 桌面平台优化
- 使用sqweek/dialog提供原生文件选择体验
- 支持Windows和Linux桌面环境
- 优化Chrome浏览器自动化性能

### 7.5 性能优化建议

1. **HTTP客户端优化**
   - 合理设置连接池大小
   - 配置适当的超时时间
   - 启用连接复用

2. **浏览器自动化优化**
   - 禁用不必要的资源加载
   - 配置合适的超时时间
   - 使用信号量控制并发数

3. **内存管理优化**
   - 及时释放HTTP响应体
   - 合理使用goroutine
   - 避免内存泄漏

## 8. 总结

Share Sniffer项目采用现代化的Go语言开发，结合了多种设计模式和最佳实践，提供了稳定可靠的网盘链接检测功能。通过模块化的架构设计，项目具有良好的可扩展性和维护性，便于二次开发和新功能添加。

项目的主要特点包括：
- 跨平台支持（Windows、Linux、Android）
- 多种网盘检测策略（HTTP API、浏览器自动化）
- 高性能并发处理
- 友好的用户界面
- 完善的测试和构建流程

开发者可以根据实际需求，参考本文档进行功能扩展和定制开发。

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