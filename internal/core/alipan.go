// Package core Copyright 2025 Share Sniffer
//
// alipan.go 实现了阿里云盘链接检查器，作为策略模式的具体策略实现
// 提供了AliPanChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、参数提取、API调用和结果解析等完整流程
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/errors"
	apphttp "github.com/owu/share-sniffer/internal/http"
	"github.com/owu/share-sniffer/internal/logger"
)

// Result 检测结果结构体
// 包含URL检查的完整信息
//
// 字段:
// - URL: 被检测的URL字符串
// - Name: 资源名称（如果检测成功）
// - Status: 状态码（1表示正常，0表示失败，-1表示超时）
// - Elapsed: 检测耗时（毫秒）

// AliPanChecker 阿里云盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查阿里云盘分享链接的有效性和获取分享内容信息

type AliPanChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkAliPan方法执行具体的检查逻辑
//
// 参数:
// - urlStr: 需要检查的阿里云盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (q *AliPanChecker) Check(urlStr string) Result {
	return q.checkAliPan(urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回阿里云盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 阿里云盘链接的前缀数组，从配置中获取
func (q *AliPanChecker) GetPrefix() []string {
	return config.GetSupportedAliPan()
}

// checkAliPan 检测阿里云盘链接是否有效
// 这是AliPanChecker的核心方法，执行完整的链接检查流程
//
// 参数:
// - urlStr: 需要检查的阿里云盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体，包括URL、资源名称、状态码和耗时
func (q *AliPanChecker) checkAliPan(urlStr string) Result {
	logger.Debug("AliPanChecker:开始检测阿里云盘链接: %s", urlStr)

	// 创建带超时的context - 确保请求不会无限等待
	timeout := config.GetHTTPClientTimeout()
	logger.Debug("AliPanChecker:创建带超时的context，超时时间: %v", timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // 确保在函数退出时取消context，防止资源泄漏

	// 提取资源ID和密码 - 解析URL中的关键参数
	shareID, err := extractParamsAliPan(urlStr)
	if err != nil {
		logger.Warn("AliPanChecker:extractParamsAliPan,%s,错误: %v\n", urlStr, err)
		return Result{
			Name:   "链接格式无效",
			Status: 0,
		}
	}

	// 发送请求并处理错误 - 调用夸克API获取分享信息
	requestStart := time.Now()
	response, err := aliPanRequest(ctx, shareID)
	requestElapsed := time.Since(requestStart).Milliseconds()
	logger.Debug("AliPanChecker:请求完成，请求耗时: %v", requestElapsed)

	if err != nil {
		// 判断错误类型 - 区分超时错误和其他错误
		if errors.IsTimeoutError(err) {
			logger.Warn("AliPanChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return Result{
				Name:   "请求超时",
				Status: -1,
			}
		}
		logger.Warn("AliPanChecker:检测失败: %s, 错误: %v, 请求耗时: %dms", urlStr, err, requestElapsed)
		return Result{
			Name:   "检测失败: " + err.Error(),
			Status: 0,
		}
	}

	logger.Debug("AliPanChecker:检测成功: %s, 文件名: %s, 请求完成: %dms", urlStr, response.ShareTitle, requestElapsed)
	// 返回成功结果 - 包含资源名称和状态信息
	return Result{
		Name:   response.ShareTitle,
		Status: 1,
	}
}

// 从URL中提取share_id
func extractParamsAliPan(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	// 按"/"分割路径，取最后一部分
	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) == 0 {
		return "", fmt.Errorf("URL中未找到share_id")
	}

	shareID := pathParts[len(pathParts)-1]
	if shareID == "" {
		return "", fmt.Errorf("提取的share_id为空")
	}

	return shareID, nil
}

// 定义响应结构体
type aliPanResp struct {
	CreatorID    string `json:"creator_id"`
	CreatorName  string `json:"creator_name"`
	CreatorPhone string `json:"creator_phone"`
	Expiration   string `json:"expiration"`
	UpdatedAt    string `json:"updated_at"`
	Vip          string `json:"vip"`
	Avatar       string `json:"avatar"`
	ShareName    string `json:"share_name"`
	FileCount    int    `json:"file_count"`
	DisplayName  string `json:"display_name"`
	ShareTitle   string `json:"share_title"` // 这是我们要提取的字段
	HasPwd       bool   `json:"has_pwd"`
	SaveButton   struct {
		Text          string `json:"text"`
		SelectAllText string `json:"select_all_text"`
	} `json:"save_button"`
	FileInfos []struct {
		Type     string `json:"type"`
		FileID   string `json:"file_id"`
		FileName string `json:"file_name"`
	} `json:"file_infos"`
}

// 发起API请求并获取分享信息
func aliPanRequest(ctx context.Context, shareID string) (*aliPanResp, error) {
	apiURL := fmt.Sprintf("https://api.aliyundrive.com/adrive/v3/share_link/get_share_by_anonymous?share_id=%s", shareID)
	logger.Debug("准备请求阿里API: %s, shareID: %s", apiURL, shareID)

	// 构造请求体 - 准备API所需的参数
	requestBody := fmt.Sprintf(`{"share_id":"%s"}`, shareID)

	// 创建POST请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(requestBody))
	if err != nil {
		logger.Error("创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头 - 模拟浏览器请求，确保API能够正确响应
	apphttp.SetDefaultHeaders(req)
	req.Header.Set("authorization", "") // 注意这里根据curl命令设置为空
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://www.alipan.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://www.alipan.com/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("x-canary", "client=web,app=share,version=v2.3.1")

	// 发送请求
	resp, err := apphttp.DoWithRetry(ctx, req, config.GetRetryCount())
	if err != nil {
		// 处理超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.NewTimeoutError("请求超时")
		}
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer apphttp.CloseResponse(resp) // 确保响应体被关闭，防止资源泄漏

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewResponseError("读取响应失败", err)
	}
	logger.Debug("响应体读取完成, 大小: %d字节", len(body))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析JSON响应[9,10](@ref)
	var response aliPanResp
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Warn("解析JSON失败: %v, 响应体: %s", err, string(body[:min(100, len(body))]))
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &response, nil
}
