package config

import (
	"os"
	"sync"
	"time"

	"github.com/owu/share-sniffer/internal/utils"
)

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
		HomePage       string
		StaticApi      string
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
		Yd       []string
	}
}

var (
	instance *Config
	once     sync.Once
)

// GetConfig 获取配置单例
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.initDefault()
		instance.loadFromEnv()
	})
	return instance
}

// initDefault 初始化默认配置
func (q *Config) initDefault() {
	// HTTP客户端默认配置
	q.HTTPClientConfig.Timeout = 5 * time.Second // 减少超时时间，避免长时间阻塞
	q.HTTPClientConfig.MaxIdleConns = 100
	q.HTTPClientConfig.MaxIdleConnsPerHost = 20
	q.HTTPClientConfig.IdleConnTimeout = 90 * time.Second
	q.HTTPClientConfig.RetryCount = 1 // 减少重试次数，加快失败处理

	// 检测默认配置
	q.CheckConfig.MaxConcurrentTasks = 8 // 增加并发数，提高处理能力
	q.CheckConfig.DefaultTimeout = 5 * time.Second
	q.CheckConfig.RetryInterval = 1 * time.Second // 减少重试间隔
	// 长耗时任务配置
	q.CheckConfig.LongTimeout = 10 * time.Second // 长耗时检测需要更长时间
	q.CheckConfig.LongMaxConcurrent = 2          // 限制长耗时任务并发数，避免资源消耗过高

	// 应用信息默认配置
	q.AppInfo.Version = "0.1.3"
	q.AppInfo.AppName = "Share Sniffer"
	q.AppInfo.AppNameCN = "分享嗅探器"
	q.AppInfo.HomePage = "https://github.com/owu/share-sniffer"
	q.AppInfo.StaticApi = "https://owu.github.io/api/open-source/share-sniffer/base.json"
	q.AppInfo.ExpirationDate = 1798732799000 // 2026-12-31 23:59:59的时间戳 毫秒

	q.SupportedLinkTypes.Quark = []string{"https://pan.quark.cn/s/"}
	q.SupportedLinkTypes.Telecom = []string{"https://cloud.189.cn/web/share?", "https://cloud.189.cn/t/"}
	q.SupportedLinkTypes.Baidu = []string{"https://pan.baidu.com/s/"}
	q.SupportedLinkTypes.AliPan = []string{"https://www.alipan.com/s/"}
	q.SupportedLinkTypes.Yyw = []string{"https://115cdn.com/s/"}
	q.SupportedLinkTypes.Yes = []string{"https://www.123684.com/s/", "https://www.123865.com/s/"}
	q.SupportedLinkTypes.Uc = []string{"https://drive.uc.cn/s/"}
	q.SupportedLinkTypes.Xunlei = []string{"https://pan.xunlei.com/s/"}
	q.SupportedLinkTypes.Yd = []string{"https://yun.139.com/shareweb/"}

	// 收集所有支持的链接前缀
	q.SupportedLinkTypes.AllLinks = []string{}
	q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Quark...)
	q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Telecom...)
	q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Baidu...)
	q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.AliPan...)
	q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Yyw...)
	q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Yes...)
	q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Uc...)
	if utils.IsDesktop() {
		q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Xunlei...)
		q.SupportedLinkTypes.AllLinks = append(q.SupportedLinkTypes.AllLinks, q.SupportedLinkTypes.Yd...)
	}
}

// loadFromEnv 从环境变量加载配置
func (q *Config) loadFromEnv() {
	// 从环境变量读取配置，如果有设置
	if maxConcurrent := os.Getenv("MAX_CONCURRENT_TASKS"); maxConcurrent != "" {
		// 这里可以添加字符串到int的转换逻辑
	}

	// 其他环境变量加载逻辑...
}

// Name 返回应用名称（英文）和中文名称
func Name() (string, string) {
	config := GetConfig()
	return config.AppInfo.AppName, config.AppInfo.AppNameCN
}

// HomePage 返回应用首页URL
func HomePage() string {
	return GetConfig().AppInfo.HomePage
}

// StaticApi 返回应用静态接口URL
func StaticApi() string {
	return GetConfig().AppInfo.StaticApi
}

// Version 返回应用版本号
func Version() string {
	return GetConfig().AppInfo.Version
}

// ExpirationDate 返回应用过期日期时间戳（毫秒）
// 这里设置为未来的一个日期，实际项目中可能会从服务器获取
func ExpirationDate() int64 {
	return GetConfig().AppInfo.ExpirationDate
}

// GetHTTPClientTimeout 获取HTTP客户端超时时间
func GetHTTPClientTimeout() time.Duration {
	return GetConfig().HTTPClientConfig.Timeout
}

// GetMaxConcurrentTasks 获取最大并发任务数
func GetMaxConcurrentTasks() int {
	return GetConfig().CheckConfig.MaxConcurrentTasks
}

// GetRetryCount 获取重试次数
func GetRetryCount() int {
	return GetConfig().HTTPClientConfig.RetryCount
}

// GetRetryInterval 获取重试间隔
func GetRetryInterval() time.Duration {
	return GetConfig().CheckConfig.RetryInterval
}

// GetLongTimeout 获取长耗时检测超时时间
func GetLongTimeout() time.Duration {
	return GetConfig().CheckConfig.LongTimeout
}

// GetLongMaxConcurrent 获取长耗时任务最大并发数
func GetLongMaxConcurrent() int {
	return GetConfig().CheckConfig.LongMaxConcurrent
}

// GetSupportedLinks 获取支持的链接类型列表
func GetSupportedLinks() []string {
	return GetConfig().SupportedLinkTypes.AllLinks
}

// GetSupportedQuark 获取支持的夸克网盘链接前缀
func GetSupportedQuark() []string {
	return GetConfig().SupportedLinkTypes.Quark
}

// GetSupportedTelecom 获取支持的电信云盘链接前缀
func GetSupportedTelecom() []string {
	return GetConfig().SupportedLinkTypes.Telecom
}

// GetSupportedBaidu 获取支持的百度网盘链接前缀
func GetSupportedBaidu() []string {
	return GetConfig().SupportedLinkTypes.Baidu
}

// GetSupportedAliPan 获取支持的阿里云盘链接前缀
func GetSupportedAliPan() []string {
	return GetConfig().SupportedLinkTypes.AliPan
}

// GetSupportedYyw 获取支持的115网盘链接前缀
func GetSupportedYyw() []string {
	return GetConfig().SupportedLinkTypes.Yyw
}

// GetSupportedYes 获取支持的123网盘链接前缀
func GetSupportedYes() []string {
	return GetConfig().SupportedLinkTypes.Yes
}

// GetSupportedUc 获取支持的UC网盘链接前缀
func GetSupportedUc() []string {
	return GetConfig().SupportedLinkTypes.Uc
}

// GetSupportedXunlei 获取支持的迅雷网盘链接前缀
func GetSupportedXunlei() []string {
	return GetConfig().SupportedLinkTypes.Xunlei
}

// GetSupportedYd 获取支持的移动云盘(139云盘)链接前缀
func GetSupportedYd() []string {
	return GetConfig().SupportedLinkTypes.Yd
}
