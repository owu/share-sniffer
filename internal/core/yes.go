// Package core Copyright 2025 Share Sniffer
//
// yes.go 实现了123网盘链接检查器，作为策略模式的具体策略实现
// 提供了YesChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、参数提取、API调用和结果解析等完整流程
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/errors"
	apphttp "github.com/owu/share-sniffer/internal/http"
	"github.com/owu/share-sniffer/internal/logger"
	"github.com/owu/share-sniffer/internal/utils"
)

// YesChecker 123网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查123网盘分享链接的有效性和获取分享内容信息
type YesChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkYes方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的123网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (y *YesChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return y.checkYes(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回123网盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 123网盘链接的前缀数组，从配置中获取
func (y *YesChecker) GetPrefix() []string {
	return config.GetSupportedYes()
}

// yesResp 123 API响应结构
// 用于解析123网盘API返回的JSON数据
type yesResp struct {
	Info struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ShareName string `json:"ShareName"`
		} `json:"data"`
	} `json:"info"`
}

// checkYes 检测123网盘链接是否有效
// 这是YesChecker的核心方法，执行完整的链接检查流程
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的123网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体，包括URL、资源名称、错误码和耗时
func (y *YesChecker) checkYes(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("YesChecker:开始检测123网盘链接: %s", urlStr)

	// 提取资源ID和密码 - 解析URL中的关键参数
	resourceID, passCode, err := extractParamsYes(urlStr)
	if err != nil {
		logger.Info("YesChecker:extractParamsYes,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 发送请求并处理错误 - 调用123 API获取分享信息
	requestStart := time.Now()
	response, err := yesRequest(ctx, urlStr, resourceID, passCode)
	requestElapsed := time.Since(requestStart).Milliseconds()
	logger.Debug("YesChecker:请求完成，请求耗时: %v", requestElapsed)

	if err != nil {
		// 判断错误类型 - 区分超时错误和其他错误
		if errors.IsTimeoutError(err) {
			logger.Info("YesChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorTimeout()
		}
		logger.Info("YesChecker:检测失败: %s, 错误: %v, 耗时: %dms", urlStr, err, requestElapsed)
		return utils.ErrorFatal("失败: " + err.Error())
	}

	// 检查API响应状态 - 验证业务层面的成功
	if response.Info.Code != 0 {
		logger.Info("分享链接失效: %s, 状态码: %d, 错误信息: %s", urlStr, response.Info.Code, response.Info.Message)

		return utils.ErrorInvalid("失败: " + response.Info.Message)
	}

	logger.Debug("检测成功: %s, 文件名: %s, 请求耗时: %dms", urlStr, response.Info.Data.ShareName, requestElapsed)
	// 返回成功结果 - 包含资源名称和状态信息
	return utils.ErrorValid(response.Info.Data.ShareName)
}

// yesRequest 获取123网盘分享信息
// 调用123网盘API，获取分享链接的详细信息
//
// 参数:
// - ctx: 上下文，用于控制请求超时和取消
// - originalURL: 原始分享URL，用于设置Referer头
// - resourceID: 资源ID，从URL中提取
// - passCode: 分享密码，如果URL中有提供的话
//
// 返回值:
// - *yesResp: 123 API响应的解析结果，包含资源信息
// - error: 发生的错误，如果有
func yesRequest(ctx context.Context, originalURL string, resourceID string, passCode string) (*yesResp, error) {
	logger.Debug("准备请求123 API: resourceID: %s, passCode: %s", resourceID, passCode)

	// 第一步：请求原始URL获取cookie
	cookie, err := getCookieFromOriginalURL(ctx, originalURL)
	if err != nil {
		return nil, errors.NewRequestError("获取cookie失败", err)
	}

	// 第二步：构造API请求URL
	apiURL := fmt.Sprintf("https://www.123684.com/gsb/s/%s", resourceID)
	logger.Debug("准备请求123 API: %s", apiURL)

	// 创建HTTP请求 - 使用WithContext确保请求可以被超时控制
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		logger.Warn("创建请求失败: %v", err)
		return nil, errors.NewRequestError("创建请求失败", err)
	}

	// 设置请求头 - 模拟浏览器请求，确保API能够正确响应
	apphttp.SetDefaultHeaders(req)
	req.Header.Set("content-type", "application/json") // 设置内容类型为JSON
	req.Header.Set("Referer", originalURL)             // 设置Referer头
	req.Header.Set("Cookie", cookie)                   // 设置从第一步获取的cookie

	// 发送请求
	resp, err := apphttp.DoWithRetry(ctx, req, config.GetRetryCount())
	if err != nil {
		// 处理超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.NewTimeoutError("请求超时")
		}
		return nil, errors.NewRequestError("发送请求失败", err)
	}
	defer apphttp.CloseResponse(resp) // 确保响应体被关闭，防止资源泄漏

	// 读取响应体内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewResponseError("读取响应失败", err)
	}
	logger.Debug("响应体读取完成, 大小: %d字节", len(body))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析JSON响应
	var response yesResp
	if err = json.Unmarshal(body, &response); err != nil {
		// 仅打印部分响应体用于调试，避免日志过大
		logger.Info("解析JSON失败: %v, 响应体: %s", err, string(body[:min(100, len(body))]))
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	logger.Debug("获取123 API响应成功: Code=%d, Message=%s", response.Info.Code, response.Info.Message)
	return &response, nil
}

// getCookieFromOriginalURL 从原始URL获取cookie
// 第一步请求，用于获取API请求所需的cookie
//
// 参数:
// - ctx: 上下文，用于控制请求超时和取消
// - originalURL: 原始分享URL
//
// 返回值:
// - string: 获取到的cookie字符串
// - error: 发生的错误，如果有
func getCookieFromOriginalURL(ctx context.Context, originalURL string) (string, error) {
	logger.Debug("准备请求原始URL获取cookie: %s", originalURL)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", originalURL, nil)
	if err != nil {
		logger.Warn("创建请求失败: %v", err)
		return "", errors.NewRequestError("创建请求失败", err)
	}

	// 设置请求头 - 模拟浏览器请求
	apphttp.SetDefaultHeaders(req)

	// 发送请求
	resp, err := apphttp.DoWithRetry(ctx, req, 1) // 只重试一次
	if err != nil {
		// 处理超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.NewTimeoutError("请求超时")
		}
		return "", errors.NewRequestError("发送请求失败", err)
	}
	defer apphttp.CloseResponse(resp) // 确保响应体被关闭，防止资源泄漏

	// 读取并处理cookie
	var cookie string
	for _, c := range resp.Cookies() {
		cookie += fmt.Sprintf("%s=%s; ", c.Name, c.Value)
	}

	if cookie == "" {
		return "", fmt.Errorf("未获取到cookie")
	}

	// 移除最后一个分号和空格
	cookie = cookie[:len(cookie)-2]
	logger.Debug("成功获取cookie: %s", cookie)

	return cookie, nil
}

// 验证URL格式的正则表达式
// 匹配123网盘分享链接的标准格式：https://www.123684.com/s/[资源ID]?pwd=[密码]
var yesUrlRegex = regexp.MustCompile(`^https://www\.(123684|123865)\.com/s/[a-zA-Z0-9\-]+(?:\?pwd=[a-zA-Z0-9]+)?(?:#)?$`)

// isValidYesURL 验证URL是否合法
// 使用正则表达式快速验证URL的基本格式
//
// 参数:
// - rawURL: 需要验证的URL字符串
//
// 返回值:
// - bool: URL是否符合123网盘分享链接的格式
func isValidYesURL(rawURL string) bool {
	return yesUrlRegex.MatchString(rawURL)
}

// extractParamsYes 提取参数的增强函数，包含URL验证
// 从123网盘链接中提取资源ID和密码，并进行全面的URL验证
//
// 参数:
// - rawURL: 需要解析的123网盘分享链接
//
// 返回值:
// - resId: 提取的资源ID
// - pwd: 提取的密码（如果没有则为空字符串）
// - err: 发生的错误，如果有
func extractParamsYes(rawURL string) (resId, pwd string, err error) {
	// 第一步：使用正则表达式快速验证URL基本格式
	if !isValidYesURL(rawURL) {
		return "", "", fmt.Errorf("无效的URL格式: %s", rawURL)
	}

	// 第二步：使用标准库解析URL，提取各部分信息
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("URL解析失败: %v", err)
	}

	// 第三步：验证特定的域名格式
	// 确保域名是www.123.com或www.123865.com
	host := parsedURL.Host
	if host != "www.123684.com" && host != "www.123865.com" {
		return "", "", fmt.Errorf("不支持的域名: %s，期望 www.123684.com 或 www.123865.com", host)
	}

	// 确保路径以/s/开头，这是123网盘分享链接的标准格式
	if !strings.HasPrefix(parsedURL.Path, "/s/") {
		return "", "", fmt.Errorf("无效的路径格式: %s，期望以 /s/ 开头", parsedURL.Path)
	}

	// 第四步：从路径中提取资源ID
	// 提取路径的最后一部分作为资源ID
	resId = strings.TrimSpace(path.Base(parsedURL.Path))
	// 验证提取的资源ID是否有效
	if resId == "" || resId == "/" || resId == "." || resId == "s" {
		return "", "", fmt.Errorf("无法从URL路径中提取有效的resId")
	}

	// 第五步：从查询参数中提取密码（如果有）
	queryParams := parsedURL.Query()
	pwd = strings.TrimSpace(queryParams.Get("pwd"))

	// 第六步：如果存在密码，验证其格式
	if pwd != "" && (len(pwd) < 2 || len(pwd) > 50) {
		return "", "", fmt.Errorf("pwd参数长度无效: %d，应在2-50字符之间", len(pwd))
	}

	// 所有验证通过，返回提取的参数
	return resId, pwd, nil
}
