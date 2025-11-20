// Package core Copyright 2025 Share Sniffer
//
// uc.go 实现了UC网盘链接检查器，作为策略模式的具体策略实现
// 提供了UcChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
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

// UcChecker UC网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查UC网盘分享链接的有效性和获取分享内容信息
type UcChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkUc方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的UC网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (u *UcChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return u.checkUc(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回UC网盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: UC网盘链接的前缀数组，从配置中获取
func (u *UcChecker) GetPrefix() []string {
	return config.GetSupportedUc()
}

// ucResp UC网盘API响应结构
// 用于解析UC网盘API返回的JSON数据
type ucResp struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		DetailInfo struct {
			IsOwner int `json:"is_owner"`
			Share   struct {
				Title     string `json:"title"`
				ShareType int    `json:"share_type"`
				ShareID   string `json:"share_id"`
				PwdID     string `json:"pwd_id"`
				ShareURL  string `json:"share_url"`
				URLType   int    `json:"url_type"`
				Passcode  string `json:"passcode"`
			} `json:"share"`
		} `json:"detail_info"`
	} `json:"data"`
}

// checkUc 检测UC网盘链接是否有效
// 这是UcChecker的核心方法，执行完整的链接检查流程
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的UC网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体，包括URL、资源名称、错误码和耗时
func (u *UcChecker) checkUc(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("UcChecker:开始检测UC网盘链接: %s", urlStr)

	// 提取资源ID - 解析URL中的关键参数
	code, err := extractParamsUc(urlStr)
	if err != nil {
		logger.Info("UcChecker:extractParamsUc,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 发送请求并处理错误 - 调用UC网盘API获取分享信息
	requestStart := time.Now()
	response, err := ucRequest(ctx, code)
	requestElapsed := time.Since(requestStart).Milliseconds()
	logger.Debug("UcChecker:请求完成，请求耗时: %v", requestElapsed)

	if err != nil {
		// 判断错误类型 - 区分超时错误和其他错误
		if errors.IsTimeoutError(err) {
			logger.Info("UcChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorTimeout()
		}
		logger.Info("UcChecker:检测失败: %s, 错误: %v, 耗时: %dms", urlStr, err, requestElapsed)
		return utils.ErrorFatal("失败: " + err.Error())
	}

	// 检查API响应状态 - 验证业务层面的成功
	if response.Status == http.StatusOK && response.Code == 0 {
		logger.Debug("检测成功: %s, 文件名: %s, 请求耗时: %dms", urlStr, response.Data.DetailInfo.Share.Title, requestElapsed)
		return utils.ErrorValid(response.Data.DetailInfo.Share.Title)
	} else {
		// 链接失效的情况
		logger.Info("分享链接失效: %s, 状态码: %d, 错误码: %d, 错误信息: %s", urlStr, response.Status, response.Code, response.Message)
		return utils.ErrorInvalid(response.Message)
	}
}

// ucRequest 获取UC网盘分享信息
// 调用UC网盘API，获取分享链接的详细信息
//
// 参数:
// - ctx: 上下文，用于控制请求超时和取消
// - code: 从URL中提取的code参数
//
// 返回值:
// - *ucResp: UC网盘API响应的解析结果，包含资源信息
// - error: 发生的错误，如果有
func ucRequest(ctx context.Context, code string) (*ucResp, error) {
	logger.Debug("准备请求UC网盘API: code: %s", code)

	// 构造API请求URL
	apiURL := "https://pc-api.uc.cn/1/clouddrive/share/sharepage/v2/detail?pr=UCBrowser&fr=pc"
	logger.Debug("准备请求UC网盘API: %s", apiURL)

	// 创建请求体数据
	requestBody := fmt.Sprintf(`{"pwd_id":"%s","passcode":"","force":0,"page":1,"size":50,"fetch_banner":1,"fetch_share":1,"fetch_total":1,"sort":"file_type:asc,file_name:asc","banner_platform":"other","web_platform":"windows","fetch_error_background":1}`, code)
	logger.Debug("请求体: %s", requestBody)

	// 创建HTTP请求 - 使用WithContext确保请求可以被超时控制
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(requestBody))
	if err != nil {
		logger.Warn("创建请求失败: %v", err)
		return nil, errors.NewRequestError("创建请求失败", err)
	}

	// 设置请求头 - 模拟浏览器请求，确保API能够正确响应
	apphttp.SetDefaultHeaders(req)
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("origin", "https://drive.uc.cn")
	req.Header.Set("referer", "https://drive.uc.cn/")

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

	// 解析JSON响应
	var response ucResp
	if err = json.Unmarshal(body, &response); err != nil {
		// 仅打印部分响应体用于调试，避免日志过大
		logger.Info("解析JSON失败: %v, 响应体: %s", err, string(body[:min(100, len(body))]))
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	logger.Debug("获取UC网盘API响应成功: Error=%d, Code=%d, Message=%s", response.Status, response.Code, response.Message)
	return &response, nil
}

// 验证URL格式的正则表达式
// 匹配UC网盘分享链接的标准格式：https://drive.uc.cn/s/[code]?public=1
var ucUrlRegex = regexp.MustCompile(`^https://drive\.uc\.cn/s/[a-zA-Z0-9]+(?:\?[a-zA-Z0-9=&]+)?(?:#[a-zA-Z0-9_/]+)?$`)

// isValidUcURL 验证URL是否合法
// 使用正则表达式快速验证URL的基本格式
//
// 参数:
// - rawURL: 需要验证的URL字符串
//
// 返回值:
// - bool: URL是否符合UC网盘分享链接的格式
func isValidUcURL(rawURL string) bool {
	return ucUrlRegex.MatchString(rawURL)
}

// extractParamsUc 从UC网盘链接中提取code参数
// 解析UC网盘链接，提取其中的code参数，并进行URL验证
//
// 参数:
// - rawURL: 需要解析的UC网盘分享链接
//
// 返回值:
// - code: 提取的code参数
// - error: 发生的错误，如果有
func extractParamsUc(rawURL string) (code string, err error) {
	// 第一步：使用正则表达式快速验证URL基本格式
	if !isValidUcURL(rawURL) {
		return "", fmt.Errorf("无效的URL格式: %s", rawURL)
	}

	// 第二步：使用标准库解析URL，提取各部分信息
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("URL解析失败: %v", err)
	}

	// 第三步：验证特定的域名格式
	// 确保域名是drive.uc.cn
	host := parsedURL.Host
	if host != "drive.uc.cn" {
		return "", fmt.Errorf("不支持的域名: %s，期望 drive.uc.cn", host)
	}

	// 确保路径以/s/开头，这是UC网盘分享链接的标准格式
	if !strings.HasPrefix(parsedURL.Path, "/s/") {
		return "", fmt.Errorf("无效的路径格式: %s，期望以 /s/ 开头", parsedURL.Path)
	}

	// 第四步：从路径中提取code参数
	// 提取路径的最后一部分作为code
	code = strings.TrimSpace(path.Base(parsedURL.Path))
	// 验证提取的code是否有效
	if code == "" || code == "/" || code == "." || code == "s" {
		return "", fmt.Errorf("无法从URL路径中提取有效的code")
	}

	// 所有验证通过，返回提取的code
	return code, nil
}
