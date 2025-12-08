package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/owu/share-sniffer/internal/logger"
)

// 网站URL列表
var webUrls = []struct {
	name string
	url  string
}{
	{"2345", "http://www.2345.com"},
	{"网易", "http://www.163.com"},
	{"知乎", "http://www.zhihu.com"},
	{"豆瓣", "http://www.douban.com"},
	{"百度", "http://www.baidu.com"},
	{"国家授时中心", "http://www.ntsc.ac.cn"},
	{"360安全卫士", "http://www.360.cn"},
	{"beijing-time", "http://www.beijing-time.org"},
	{"腾讯", "http://www.qq.com"},
}

// 随机打乱URL切片
func shuffleUrls(urls []struct {
	name string
	url  string
}) []struct {
	name string
	url  string
} {
	shuffled := make([]struct {
		name string
		url  string
	}, len(urls))
	copy(shuffled, urls)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}

// 获取单个网站的时间戳（毫秒）
func getWebsiteTimestamp(webUrl string) (int64, error) {
	client := &http.Client{
		Timeout: 2 * time.Second, // 减少超时时间到1秒
	}

	req, err := http.NewRequest("HEAD", webUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头，模拟浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 从响应头获取Date字段
	dateHeader := resp.Header.Get("Date")
	if dateHeader == "" {
		return 0, fmt.Errorf("响应头中没有Date字段")
	}

	// 解析HTTP日期格式
	var parsedTime time.Time
	parsedTime, err = time.Parse(time.RFC1123, dateHeader)
	if err != nil {
		// 尝试其他日期格式
		parsedTime, err = time.Parse(time.RFC1123Z, dateHeader)
		if err != nil {
			return 0, fmt.Errorf("时间解析失败: %v", err)
		}
	}

	// 转换为北京时间 (UTC+8) 并返回Unix时间戳（毫秒）
	beijingTime := parsedTime.In(time.FixedZone("CST", 8*3600))
	return beijingTime.UnixNano() / 1e6, nil
}

// 随机选择URL获取服务器时间，最多尝试2个不同的URL
func getRandomTimestamp() (int64, string, error) {
	// 打乱URL顺序
	shuffledUrls := shuffleUrls(webUrls)

	var lastErr error
	triedUrls := make(map[string]bool) // 记录已尝试的URL

	// 最多尝试2个不同的URL，减少尝试次数
	for i := 0; i < 2 && i < len(shuffledUrls); i++ {
		site := shuffledUrls[i]

		// 如果这个URL已经尝试过，跳过
		if triedUrls[site.url] {
			continue
		}
		triedUrls[site.url] = true

		timestamp, err := getWebsiteTimestamp(site.url)
		if err == nil {
			// 获取成功，返回时间戳和网站名称
			return timestamp, site.name, nil
		}

		lastErr = err

		// 如果不是最后一个尝试，等待一小段时间再试下一个
		if i < 1 && i < len(shuffledUrls)-1 {
			time.Sleep(100 * time.Millisecond) // 减少等待时间
		}
	}

	// 所有尝试都失败，快速返回本地时间
	localTimestamp := time.Now().UnixNano() / 1e6
	logger.Warn("所有服务器尝试失败，使用本地时间: %d", localTimestamp)
	return localTimestamp, "本地时间", lastErr
}

// StandardTime 获取时间戳的主要函数
func StandardTime() int64 {
	startTime := time.Now()
	timestamp, siteName, err := getRandomTimestamp()
	elapsed := time.Since(startTime)

	if err != nil {
		logger.Warn("最终结果: 使用%s, 时间戳: %d, 耗时: %v", siteName, timestamp, elapsed)
	} else {
		logger.Info("成功从 [%s] 获取时间, 时间戳: %d, 耗时: %v", siteName, timestamp, elapsed)
	}
	return timestamp
}
