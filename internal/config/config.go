package config

import (
	"os"
	"sync"
	"time"

	"share-sniffer/internal/utils"
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
		HomePageFree   string
		HomePage       string
		StaticApiFree  string
		StaticApi      string
		ExpirationDate int64
	}

	// 支持的链接类型
	SupportedLinkTypes struct {
		Providers   map[string][]string
		AllPrefixes []string
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
		instance.SupportedLinkTypes.Providers = make(map[string][]string)
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
	q.AppInfo.Version = "0.3.0"
	q.AppInfo.AppName = "Share Sniffer"
	q.AppInfo.AppNameCN = "分享嗅探器"
	q.AppInfo.HomePage = "https://gitee.com/bye/share-sniffer"
	q.AppInfo.StaticApi = "https://gitee.com/bye/oss/raw/main/sharesniffer/api/base.json"
	q.AppInfo.HomePageFree = "https://github.com/owu/share-sniffer"
	q.AppInfo.StaticApiFree = "https://raw.githubusercontent.com/owu/oss/refs/heads/main/sharesniffer/api/base.json"
	q.AppInfo.ExpirationDate = 1798732799000 // 2026-12-31 23:59:59的时间戳 毫秒

	// 初始化网盘前缀
	p := q.SupportedLinkTypes.Providers
	p["quark"] = []string{"https://pan.quark.cn/s/"}
	p["telecom"] = []string{"https://cloud.189.cn/web/share?", "https://cloud.189.cn/t/"}
	p["baidu"] = []string{"https://pan.baidu.com/s/"}
	p["alipan"] = []string{"https://www.alipan.com/s/"}
	p["yyw"] = []string{"https://115cdn.com/s/"}
	p["yes"] = []string{"https://www.123684.com/s/", "https://www.123865.com/s/"}
	p["uc"] = []string{"https://drive.uc.cn/s/"}
	p["xunlei"] = []string{"https://pan.xunlei.com/s/"}
	p["yd"] = []string{"https://yun.139.com/shareweb/"}

	q.refreshAllPrefixes()
}

// refreshAllPrefixes 刷新所有支持的前缀列表
func (q *Config) refreshAllPrefixes() {
	all := []string{}
	desktopOnly := map[string]bool{"xunlei": true, "yd": true}

	for name, prefixes := range q.SupportedLinkTypes.Providers {
		if desktopOnly[name] && !utils.IsDesktop() {
			continue
		}
		all = append(all, prefixes...)
	}
	q.SupportedLinkTypes.AllPrefixes = all
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
	cfg := GetConfig()
	return cfg.AppInfo.AppName, cfg.AppInfo.AppNameCN
}

func inChina() bool {
	// 改进的判断逻辑，可以通过环境变量或时区偏移综合判断
	now := time.Now()
	name, offset := now.Zone()
	return name == "CST" || offset == 8*3600
}

// HomePage 返回应用首页URL
func HomePage() string {
	if inChina() {
		return GetConfig().AppInfo.HomePage
	}
	return GetConfig().AppInfo.HomePageFree
}

// StaticApi 返回应用静态接口URL
func StaticApi() string {
	if inChina() {
		return GetConfig().AppInfo.StaticApi
	}
	return GetConfig().AppInfo.StaticApiFree
}

// Version 返回应用版本号
func Version() string {
	return GetConfig().AppInfo.Version
}

// ExpirationDate 返回应用过期日期时间戳（毫秒）
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

// GetSupported 获取指定网盘的支持前缀
func GetSupported(provider string) []string {
	return GetConfig().SupportedLinkTypes.Providers[provider]
}

// GetSupportedLinks 获取所有支持的链接前缀列表
func GetSupportedLinks() []string {
	return GetConfig().SupportedLinkTypes.AllPrefixes
}

// 以下是兼容旧代码的快捷方法，内部调用 GetSupported
func GetSupportedQuark() []string   { return GetSupported("quark") }
func GetSupportedTelecom() []string { return GetSupported("telecom") }
func GetSupportedBaidu() []string   { return GetSupported("baidu") }
func GetSupportedAliPan() []string  { return GetSupported("alipan") }
func GetSupportedYyw() []string     { return GetSupported("yyw") }
func GetSupportedYes() []string     { return GetSupported("yes") }
func GetSupportedUc() []string      { return GetSupported("uc") }
func GetSupportedXunlei() []string  { return GetSupported("xunlei") }
func GetSupportedYd() []string      { return GetSupported("yd") }
