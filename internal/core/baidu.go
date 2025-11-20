// Package core Copyright 2025 Share Sniffer
//
// baidu.go 实现了百度网盘链接检查器，作为策略模式的具体策略实现
// 提供了BaiduChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、参数提取、API调用和结果解析等完整流程
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/errors"
	"github.com/owu/share-sniffer/internal/logger"
	"github.com/owu/share-sniffer/internal/utils"
)

// BaiduChecker 百度网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查百度网盘分享链接的有效性和获取分享内容信息

type BaiduChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkBaidu方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的百度网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (q *BaiduChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkBaidu(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回百度网盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 百度网盘链接的前缀数组，从配置中获取
func (q *BaiduChecker) GetPrefix() []string {
	return config.GetSupportedBaidu()
}

// checkBaidu 检查百度网盘链接
// 记录开始时间，调用具体的检查方法，并计算耗时
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的百度网盘分享链接
//
// 返回值:
// - Result: 包含检查结果和耗时的结构体
func (q *BaiduChecker) checkBaidu(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("BaiduChecker:开始检测百度网盘链接: %s", urlStr)

	// 解析URL字符串
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Info("BaiduChecker:Parse,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 获取查询参数
	queryParams := parsedURL.Query()
	// 访问查询参数
	password := queryParams.Get("pwd") // 使用Get方法获取name的值

	logger.Debug("开始执行完整HTTP请求流程（第一步 → 第二步 → 第三步）...")

	// === 第一步：初始请求 ===
	logger.Debug("\n1. 执行第一步请求...")
	requestStart := time.Now()
	step1Result, err := step1Request(ctx, urlStr)
	requestElapsed1 := time.Since(requestStart).Milliseconds()
	if err != nil {
		logger.Info("BaiduChecker:step1Request,%s,错误: %v\n", urlStr, err)

		if errors.IsTimeoutError(err) {
			logger.Info("BaiduChecker:请求超时1: %s, 请求耗时: %dms", urlStr, requestElapsed1)
			return utils.ErrorTimeout()
		}

		return utils.ErrorFatal("第一步请求失败")
	}

	//过期 200
	if step1Result.StatusCode == http.StatusOK && step1Result.FullRedirectURL == "" {
		return utils.ErrorInvalid("分享文件已过期")
	}

	//正常 302
	if step1Result.StatusCode != http.StatusFound || step1Result.FullRedirectURL == "" || step1Result.SURL == "" {
		return utils.ErrorFatal("第一步302失败")
	}

	// === 第二步：验证请求 ===
	logger.Debug("\n2. 执行第二步验证请求...")
	step2Result, err := step2Request(ctx, step1Result, password)
	requestElapsed2 := time.Since(requestStart).Milliseconds()
	if err != nil {
		logger.Info("BaiduChecker:step2Request,%s,错误: %v\n", urlStr, err)

		if errors.IsTimeoutError(err) {
			logger.Info("BaiduChecker:请求超时2: %s, 请求耗时: %dms", urlStr, requestElapsed2)
			return utils.ErrorTimeout()
		}
		return utils.ErrorFatal("第二步请求失败")
	}

	if "" == step2Result.BDCLND {
		return utils.ErrorFatal("第二步响应未返回BDCLND Cookie")
	}

	// === 第三步：获取文件列表 ===
	logger.Debug("\n3. 执行第三步文件列表请求...")
	step3Result, err := step3Request(ctx, step1Result, step2Result)
	requestElapsed3 := time.Since(requestStart).Milliseconds()
	if err != nil {
		logger.Info("BaiduChecker:step3Request,%s,错误: %v\n", urlStr, err)

		if errors.IsTimeoutError(err) {
			logger.Info("BaiduChecker:请求超时3: %s, 请求耗时: %dms", urlStr, requestElapsed3)
			return utils.ErrorTimeout()
		}

		return utils.ErrorFatal("第三步请求失败")
	}

	logger.Debug("\n=== 流程完成 ===")
	logger.Debug("第一步获取Cookie数量: %d\n", len(step1Result.SetCookies))
	logger.Debug("第二步获取Cookie数量: %d\n", len(step2Result.SetCookies))
	logger.Debug("第三步响应状态码: %d\n", step3Result.StatusCode)

	if step3Result.JSONResponse != nil {
		logger.Debug("文件数量: %d\n", len(step3Result.JSONResponse.List))
	}
	return utils.ErrorValid(step3Result.JSONResponse.Title) // 返回完整的检查结果
}

// Step1Response 第一步响应结构体
type Step1Response struct {
	Status          string
	StatusCode      int
	Location        string
	FullRedirectURL string
	SetCookies      []*http.Cookie
	FlowLevel       string
	RequestID       string
	LogID           string
	ContentType     string
	Date            time.Time
	XReadtime       string
	XPoweredBy      string
	CookiesMap      map[string]string
	SURL            string // 从Location中提取的surl参数
}

// Step2Response 第二步响应结构体
type Step2Response struct {
	StatusCode      int
	Status          string
	CacheControl    string
	Connection      string
	ContentEncoding string
	ContentType     string
	Date            time.Time
	FlowLevel       string
	LogID           string
	Server          string
	SetCookies      []*http.Cookie
	Vary            []string
	XPoweredBy      string
	Yld             string
	Yme             string
	ContentLength   int
	Body            []byte
	JSONResponse    map[string]interface{}
	BDCLND          string // 从Cookie中提取的BDCLND
}

// Step3Response 第三步响应结构体
type Step3Response struct {
	StatusCode   int
	Status       string
	Headers      http.Header
	Body         []byte
	JSONResponse *ShareListResponse
}

// ShareListResponse 第三步JSON响应结构体
type ShareListResponse struct {
	Errno        int         `json:"errno"`
	RequestID    int64       `json:"request_id"`
	ServerTime   int64       `json:"server_time"`
	CfromID      int         `json:"cfrom_id"`
	Hitrisk      int         `json:"hitrisk"`
	AppealStatus int         `json:"appeal_status"`
	IsZombie     int         `json:"is_zombie"`
	VipPoint     int         `json:"vip_point"`
	VipLevel     int         `json:"vip_level"`
	Svip10ID     string      `json:"svip10_id"`
	VipType      int         `json:"vip_type"`
	Sharetype    int         `json:"sharetype"`
	ViewVisited  int         `json:"view_visited"`
	ViewLimit    int         `json:"view_limit"`
	ExpiredType  int         `json:"expired_type"`
	Title        string      `json:"title"`
	List         []*FileInfo `json:"list"`
	ShareID      int64       `json:"share_id"`
	UK           int64       `json:"uk"`
	ShowMsg      string      `json:"show_msg"`
}

// FileInfo 文件信息结构体
type FileInfo struct {
	Category       string     `json:"category"`
	Duration       int        `json:"duration"`
	ExtentInt8     string     `json:"extent_int8"`
	FsID           string     `json:"fs_id"`
	Isdir          string     `json:"isdir"`
	LocalCtime     string     `json:"local_ctime"`
	LocalMtime     string     `json:"local_mtime"`
	MD5            string     `json:"md5"`
	MediaType      string     `json:"mediaType"`
	Path           string     `json:"path"`
	Resolution     string     `json:"resolution"`
	ServerCtime    string     `json:"server_ctime"`
	ServerFilename string     `json:"server_filename"`
	ServerMtime    string     `json:"server_mtime"`
	Size           string     `json:"size"`
	Thumbs         *ThumbInfo `json:"thumbs"`
}

// ThumbInfo 缩略图信息结构体
type ThumbInfo struct {
	URL1 string `json:"url1"`
	URL2 string `json:"url2"`
	URL3 string `json:"url3"`
	Icon string `json:"icon"`
}

// 第一步请求：获取重定向信息和Cookie
func step1Request(ctx context.Context, targetURL string) (*Step1Response, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	setStep1Headers(req)

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.NewTimeoutError("请求超时")
		}

		return nil, fmt.Errorf("第一步请求失败: %v", err)
	}
	defer resp.Body.Close()

	return parseStep1Response(resp, targetURL)
}

// 设置第一步请求头
func setStep1Headers(req *http.Request) {
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language":           "en",
		"Connection":                "keep-alive",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		"sec-ch-ua":                 `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"Windows"`,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// 解析第一步响应
func parseStep1Response(resp *http.Response, originalURL string) (*Step1Response, error) {
	result := &Step1Response{
		Status:      resp.Status,
		StatusCode:  resp.StatusCode,
		Location:    resp.Header.Get("Location"),
		SetCookies:  resp.Cookies(),
		FlowLevel:   resp.Header.Get("Flow-Level"),
		RequestID:   resp.Header.Get("X-Request-Id"),
		LogID:       resp.Header.Get("Logid"),
		ContentType: resp.Header.Get("Content-Type"),
		XReadtime:   resp.Header.Get("X-Readtime"),
		XPoweredBy:  resp.Header.Get("X-Powered-By"),
		CookiesMap:  make(map[string]string),
	}

	if dateStr := resp.Header.Get("Date"); dateStr != "" {
		if date, err := time.Parse(time.RFC1123, dateStr); err == nil {
			result.Date = date
		}
	}

	if result.Location != "" {
		fullURL, err := buildFullRedirectURL(originalURL, result.Location)
		if err != nil {
			return nil, fmt.Errorf("构建重定向URL失败: %v", err)
		}
		result.FullRedirectURL = fullURL
	}

	if result.Location != "" {
		if surl, err := extractSURLFromLocation(result.Location); err == nil {
			result.SURL = surl
		}
	}

	for _, cookie := range result.SetCookies {
		result.CookiesMap[cookie.Name] = cookie.Value
	}

	return result, nil
}

// 第二步请求：验证请求
func step2Request(ctx context.Context, step1Result *Step1Response, password string) (*Step2Response, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建Cookie Jar失败: %v", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	baseURL := "https://pan.baidu.com/share/verify"
	params := url.Values{}
	params.Add("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Add("surl", step1Result.SURL)
	params.Add("channel", "chunlei")
	params.Add("web", "1")
	params.Add("app_id", "250528")
	params.Add("clienttype", "0")

	fullURL := baseURL + "?" + params.Encode()

	postData := url.Values{}
	postData.Add("pwd", password)
	postData.Add("vcode", "")
	postData.Add("vcode_str", "")

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBufferString(postData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建第二步请求失败: %v", err)
	}

	setStep2Headers(req, step1Result.FullRedirectURL)

	u, _ := url.Parse("https://pan.baidu.com")
	jar.SetCookies(u, step1Result.SetCookies)

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.NewTimeoutError("请求超时")
		}

		return nil, fmt.Errorf("第二步请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	return parseStep2Response(resp, body)
}

// 设置第二步请求头
func setStep2Headers(req *http.Request, refererURL string) {
	headers := map[string]string{
		"Accept":             "application/json, text/javascript, */*; q=0.01",
		"Accept-Language":    "en",
		"Connection":         "keep-alive",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
		"Origin":             "https://pan.baidu.com",
		"Referer":            refererURL,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		"X-Requested-With":   "XMLHttpRequest",
		"sec-ch-ua":          `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// 解析第二步响应
func parseStep2Response(resp *http.Response, body []byte) (*Step2Response, error) {
	result := &Step2Response{
		StatusCode:      resp.StatusCode,
		Status:          resp.Status,
		CacheControl:    resp.Header.Get("Cache-Control"),
		Connection:      resp.Header.Get("Connection"),
		ContentEncoding: resp.Header.Get("Content-Encoding"),
		ContentType:     resp.Header.Get("Content-Type"),
		FlowLevel:       resp.Header.Get("Flow-Level"),
		LogID:           resp.Header.Get("Logid"),
		Server:          resp.Header.Get("Server"),
		SetCookies:      resp.Cookies(),
		Vary:            resp.Header.Values("Vary"),
		XPoweredBy:      resp.Header.Get("X-Powered-By"),
		Yld:             resp.Header.Get("Yld"),
		Yme:             resp.Header.Get("Yme"),
		ContentLength:   int(resp.ContentLength),
		Body:            body,
		JSONResponse:    make(map[string]interface{}),
	}

	if dateStr := resp.Header.Get("Date"); dateStr != "" {
		if date, err := time.Parse(time.RFC1123, dateStr); err == nil {
			result.Date = date
		}
	}

	if strings.Contains(result.ContentType, "application/json") {
		if err := json.Unmarshal(body, &result.JSONResponse); err != nil {
			logger.Info("JSON解析失败: %v\n", err)
		}
	}

	// 提取BDCLND Cookie
	for _, cookie := range result.SetCookies {
		if cookie.Name == "BDCLND" {
			result.BDCLND = cookie.Value
			break
		}
	}

	return result, nil
}

// 第三步请求：获取文件列表
func step3Request(ctx context.Context, step1Result *Step1Response, step2Result *Step2Response) (*Step3Response, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建Cookie Jar失败: %v", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	// 构建第三步URL参数
	baseURL := "https://pan.baidu.com/share/list"
	params := url.Values{}
	params.Add("web", "5")
	params.Add("app_id", "250528")
	params.Add("desc", "1")
	params.Add("showempty", "0")
	params.Add("page", "1")
	params.Add("num", "20")
	params.Add("order", "time")
	params.Add("shorturl", step1Result.SURL) // 使用第一步的surl
	params.Add("root", "1")
	params.Add("view_mode", "1")
	params.Add("channel", "chunlei")
	params.Add("web", "1")
	params.Add("bdstoken", "")
	params.Add("clienttype", "0")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建第三步请求失败: %v", err)
	}

	setStep3Headers(req, step1Result)

	// 设置Cookie（包含第一步的Cookie和第二步的BDCLND）
	u, _ := url.Parse("https://pan.baidu.com")

	// 复制第一步的Cookie
	cookies := make([]*http.Cookie, len(step1Result.SetCookies))
	copy(cookies, step1Result.SetCookies)

	// 添加第二步的BDCLND Cookie
	if step2Result.BDCLND != "" {
		bdclndCookie := &http.Cookie{
			Name:  "BDCLND",
			Value: step2Result.BDCLND,
		}
		cookies = append(cookies, bdclndCookie)
	}

	jar.SetCookies(u, cookies)

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.NewTimeoutError("请求超时")
		}

		return nil, fmt.Errorf("第三步请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	return parseStep3Response(resp, body)
}

// 设置第三步请求头
func setStep3Headers(req *http.Request, step1Result *Step1Response) {
	headers := map[string]string{
		"Accept":             "*/*",
		"Accept-Language":    "en",
		"Connection":         "keep-alive",
		"Referer":            step1Result.FullRedirectURL,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		"X-Requested-With":   "XMLHttpRequest",
		"sec-ch-ua":          `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// 解析第三步响应
func parseStep3Response(resp *http.Response, body []byte) (*Step3Response, error) {
	result := &Step3Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       body,
	}

	// 解析JSON响应
	var shareListResp ShareListResponse
	if err := json.Unmarshal(body, &shareListResp); err != nil {
		return nil, fmt.Errorf("解析JSON响应失败: %v", err)
	}
	result.JSONResponse = &shareListResp

	return result, nil
}

// 构建完整的重定向URL
func buildFullRedirectURL(baseURL, location string) (string, error) {
	if location == "" {
		return "", fmt.Errorf("location为空")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	redirect, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(redirect).String(), nil
}

// 从Location中提取surl参数
func extractSURLFromLocation(location string) (string, error) {
	parsedURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	queryParams, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return "", err
	}

	surl := queryParams.Get("surl")
	if surl == "" {
		return "", fmt.Errorf("未找到surl参数")
	}

	return surl, nil
}
