package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	apphttp "share-sniffer/internal/http"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// BaiduChecker 百度网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查百度网盘分享链接的有效性和获取分享内容信息

type BaiduChecker struct{}

// Check 实现LinkChecker接口的Check方法
func (q *BaiduChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkBaidu(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
func (q *BaiduChecker) GetPrefix() []string {
	return config.GetSupportedBaidu()
}

// checkBaidu 检查百度网盘链接
func (q *BaiduChecker) checkBaidu(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("BaiduChecker:开始检测百度网盘链接: %s", urlStr)

	// 解析URL字符串
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Info("BaiduChecker:Parse,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	password := parsedURL.Query().Get("pwd")

	logger.Debug("开始执行完整HTTP请求流程（第一步 → 第二步 → 第三步）...")

	// === 第一步：初始请求 ===
	logger.Debug("\n1. 执行第一步请求...")
	step1Result, err := step1Request(ctx, urlStr)
	if err != nil {
		logger.Info("BaiduChecker:step1Request,%s,错误: %v\n", urlStr, err)
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		return utils.ErrorFatal("第一步请求失败")
	}

	//过期 200 (百度通常在过期时返回200而不是跳转)
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
	if err != nil {
		logger.Info("BaiduChecker:step2Request,%s,错误: %v\n", urlStr, err)
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		return utils.ErrorFatal("验证请求异常")
	}

	if step2Result.BDCLND == "" {
		// 检查业务错误码
		if errno, ok := step2Result.JSONResponse["errno"].(float64); ok && errno != 0 {
			if errno == -9 {
				return utils.ErrorInvalid("提取码错误")
			}
			return utils.ErrorInvalid(fmt.Sprintf("验证失败(errno:%v)", errno))
		}
		return utils.ErrorFatal("验证未通过")
	}

	// === 第三步：获取文件列表 ===
	logger.Debug("\n3. 执行第三步文件列表请求...")
	step3Result, err := step3Request(ctx, step1Result, step2Result)
	if err != nil {
		logger.Info("BaiduChecker:step3Request,%s,错误: %v\n", urlStr, err)
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		return utils.ErrorFatal("获取列表失败")
	}

	logger.Debug("\n=== 流程完成 ===")
	if step3Result.JSONResponse == nil || step3Result.JSONResponse.Errno != 0 {
		return utils.ErrorInvalid("获取分享内容失败")
	}

	return utils.ErrorValid(step3Result.JSONResponse.Title)
}

// Step1Response 第一步响应结构体
type Step1Response struct {
	StatusCode      int
	FullRedirectURL string
	SetCookies      []*http.Cookie
	SURL            string
}

// Step2Response 第二步响应结构体
type Step2Response struct {
	StatusCode   int
	SetCookies   []*http.Cookie
	JSONResponse map[string]interface{}
	BDCLND       string
}

// Step3Response 第三步响应结构体
type Step3Response struct {
	StatusCode   int
	JSONResponse *ShareListResponse
}

// ShareListResponse 第三步JSON响应结构体
type ShareListResponse struct {
	Errno int         `json:"errno"`
	Title string      `json:"title"`
	List  []*FileInfo `json:"list"`
}

// FileInfo 文件信息结构体
type FileInfo struct {
	ServerFilename string `json:"server_filename"`
	Size           string `json:"size"`
	Isdir          string `json:"isdir"`
}

// 第一步请求：获取重定向信息和Cookie
func step1Request(ctx context.Context, targetURL string) (*Step1Response, error) {
	req, err := apphttp.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	// 百度第一步需要特定的浏览行为 Headers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := apphttp.DoWithClient(ctx, apphttp.GetNoRedirectClient(), req, 0)
	if err != nil {
		return nil, err
	}
	defer apphttp.CloseResponse(resp)

	result := &Step1Response{
		StatusCode: resp.StatusCode,
		SetCookies: resp.Cookies(),
	}

	location := resp.Header.Get("Location")
	if location != "" {
		fullURL, _ := buildFullRedirectURL(targetURL, location)
		result.FullRedirectURL = fullURL
		result.SURL, _ = extractSURLFromLocation(location)
	}

	return result, nil
}

// 第二步请求：验证提取码
func step2Request(ctx context.Context, step1Result *Step1Response, password string) (*Step2Response, error) {
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://pan.baidu.com")
	jar.SetCookies(u, step1Result.SetCookies)

	client := &http.Client{Jar: jar, Timeout: config.GetHTTPClientTimeout()}

	apiURL := fmt.Sprintf("https://pan.baidu.com/share/verify?t=%d&surl=%s&channel=chunlei&web=1&app_id=250528&clienttype=0",
		time.Now().UnixMilli(), step1Result.SURL)

	postData := url.Values{}
	postData.Add("pwd", password)
	postData.Add("vcode", "")
	postData.Add("vcode_str", "")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(postData.Encode()))
	if err != nil {
		return nil, err
	}

	apphttp.SetDefaultHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Referer", step1Result.FullRedirectURL)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Step2Response{
		StatusCode:   resp.StatusCode,
		SetCookies:   resp.Cookies(),
		JSONResponse: make(map[string]interface{}),
	}
	json.Unmarshal(body, &result.JSONResponse)

	for _, c := range result.SetCookies {
		if c.Name == "BDCLND" {
			result.BDCLND = c.Value
			break
		}
	}

	return result, nil
}

// 第三步请求：获取内容
func step3Request(ctx context.Context, step1Result *Step1Response, step2Result *Step2Response) (*Step3Response, error) {
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://pan.baidu.com")

	// 组合 Cookie
	cookies := append(step1Result.SetCookies, step2Result.SetCookies...)
	jar.SetCookies(u, cookies)

	client := &http.Client{Jar: jar, Timeout: config.GetHTTPClientTimeout()}

	apiURL := fmt.Sprintf("https://pan.baidu.com/share/list?web=1&app_id=250528&shorturl=%s&root=1&channel=chunlei&clienttype=0",
		step1Result.SURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	apphttp.SetDefaultHeaders(req)
	req.Header.Set("Referer", step1Result.FullRedirectURL)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var listResp ShareListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, err
	}

	return &Step3Response{
		StatusCode:   resp.StatusCode,
		JSONResponse: &listResp,
	}, nil
}

func buildFullRedirectURL(baseURL, location string) (string, error) {
	base, _ := url.Parse(baseURL)
	redirect, _ := url.Parse(location)
	return base.ResolveReference(redirect).String(), nil
}

func extractSURLFromLocation(location string) (string, error) {
	u, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return u.Query().Get("surl"), nil
}
