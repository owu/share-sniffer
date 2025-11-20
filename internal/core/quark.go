// Package core Copyright 2025 Share Sniffer
//
// quark.go 实现了夸克网盘链接检查器，作为策略模式的具体策略实现
// 提供了QuarkChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、参数提取、API调用和结果解析等完整流程
package core

import (
	"bytes"
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

// QuarkChecker 夸克网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查夸克网盘分享链接的有效性和获取分享内容信息

type QuarkChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkQuark方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的夸克网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (q *QuarkChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkQuark(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回夸克网盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 夸克网盘链接的前缀数组，从配置中获取
func (q *QuarkChecker) GetPrefix() []string {
	return config.GetSupportedQuark()
}

// quarkResp 夸克API响应结构
// 用于解析夸克网盘API返回的JSON数据
//
// 字段:
// - Status: HTTP状态码
// - Code: 业务状态码，0表示成功
// - Message: 响应消息
// - Data: 包含实际数据的结构体，其中Title字段表示资源名称
type quarkResp struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Title string `json:"title"`
	} `json:"data"`
}

// checkQuark 检测夸克网盘链接是否有效
// 这是QuarkChecker的核心方法，执行完整的链接检查流程
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的夸克网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体，包括URL、资源名称、错误码和耗时
func (q *QuarkChecker) checkQuark(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("QuarkChecker:开始检测夸克网盘链接: %s", urlStr)

	// 使用传入的context - 确保请求受任务池的超时控制
	logger.Debug("使用传入的context进行夸克链接检测")

	// 提取资源ID和密码 - 解析URL中的关键参数
	resourceID, passCode, err := extractParamsQuark(urlStr)
	if err != nil {
		logger.Info("QuarkChecker:extractParamsQuark,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 发送请求并处理错误 - 调用夸克API获取分享信息
	requestStart := time.Now()
	response, err := quarkRequest(ctx, resourceID, passCode)
	requestElapsed := time.Since(requestStart).Milliseconds()
	logger.Debug("QuarkChecker:请求完成，请求耗时: %v", requestElapsed)

	if err != nil {
		// 判断错误类型 - 区分超时错误和其他错误
		if errors.IsTimeoutError(err) {
			logger.Info("QuarkChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorTimeout()
		}

		if errors.IsStatusCodeError(err) {
			return utils.ErrorInvalid("分享链接失效")
		}

		logger.Info("QuarkChecker:检测失败: %s, 错误: %v, 耗时: %dms", urlStr, err, requestElapsed)
		return utils.ErrorFatal("失败: " + err.Error())
	}

	// 检查API响应状态 - 验证业务层面的成功
	if response.Status != http.StatusOK || response.Code != 0 {
		logger.Info("分享链接失效: %s, 状态码: %d, 错误码: %d", urlStr, response.Status, response.Code)
		return utils.ErrorInvalid("分享链接失效或不存在")
	}

	logger.Debug("检测成功: %s, 文件名: %s, 请求耗时: %dms", urlStr, response.Data.Title, requestElapsed)
	// 返回成功结果 - 包含资源名称和状态信息
	return utils.ErrorValid(response.Data.Title)
}

// quarkRequest 获取夸克网盘分享信息
// 调用夸克网盘API，获取分享链接的详细信息
//
// 参数:
// - ctx: 上下文，用于控制请求超时和取消
// - resourceID: 资源ID，从URL中提取
// - passCode: 分享密码，如果URL中有提供的话
//
// 返回值:
// - *quarkResp: 夸克API响应的解析结果，包含资源信息
// - error: 发生的错误，如果有
func quarkRequest(ctx context.Context, resourceID string, passCode string) (*quarkResp, error) {
	apiURL := "https://drive-h.quark.cn/1/clouddrive/share/sharepage/token"
	logger.Debug("准备请求夸克API: %s, resourceID: %s, passCode: %s", apiURL, resourceID, passCode)

	// 构造请求体 - 准备API所需的参数
	requestBody := map[string]interface{}{
		"pwd_id":                            resourceID, // 资源ID
		"passcode":                          passCode,   // 分享密码
		"support_visit_limit_private_share": true,       // 支持访问限制的私有分享
	}

	// 序列化请求体为JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Warn("构造请求体失败: %v", err)
		return nil, errors.NewRequestError("构造请求体失败", err)
	}

	// 创建HTTP请求 - 使用WithContext确保请求可以被超时控制
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Warn("创建请求失败: %v", err)
		return nil, errors.NewRequestError("创建请求失败", err)
	}

	// 设置请求头 - 模拟浏览器请求，确保API能够正确响应
	apphttp.SetDefaultHeaders(req)
	req.Header.Set("content-type", "application/json") // 设置内容类型为JSON
	req.Header.Set("origin", "https://pan.quark.cn")   // 设置请求来源
	req.Header.Set("referer", "https://pan.quark.cn/") // 设置Referer头

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

	// 检查HTTP状态码 400  404（分享链接失效）
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewStatusCodeError(fmt.Sprintf("状态码: %d, 响应: %s", resp.StatusCode, string(body)))
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析JSON响应
	var response quarkResp
	if err = json.Unmarshal(body, &response); err != nil {
		// 仅打印部分响应体用于调试，避免日志过大
		logger.Info("解析JSON失败: %v, 响应体: %s", err, string(body[:min(100, len(body))]))
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	logger.Debug("获取夸克API响应成功: Error=%d, Code=%d", response.Status, response.Code)
	return &response, nil
}

// min 返回两个整数中较小的一个
// 用于限制响应体日志输出的长度
//
// 参数:
// - a, b: 需要比较的两个整数
//
// 返回值:
// - int: 较小的整数值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 验证URL格式的正则表达式（基于搜索结果中的最佳实践）
// 匹配夸克网盘分享链接的标准格式：https://pan.quark.cn/s/[资源ID]?pwd=[密码]
var urlRegex = regexp.MustCompile(`^https://pan\.quark\.cn/s/[a-zA-Z0-9]+(?:\?pwd=[a-zA-Z0-9]*)?$`)

// isValidURL 验证URL是否合法
// 使用正则表达式快速验证URL的基本格式
//
// 参数:
// - rawURL: 需要验证的URL字符串
//
// 返回值:
// - bool: URL是否符合夸克网盘分享链接的格式
func isValidURL(rawURL string) bool {
	return urlRegex.MatchString(rawURL)
}

// extractParamsQuark 提取参数的增强函数，包含URL验证
// 从夸克网盘链接中提取资源ID和密码，并进行全面的URL验证
//
// 参数:
// - rawURL: 需要解析的夸克网盘分享链接
//
// 返回值:
// - resId: 提取的资源ID
// - pwd: 提取的密码（如果没有则为空字符串）
// - err: 发生的错误，如果有
func extractParamsQuark(rawURL string) (resId, pwd string, err error) {
	// 第一步：使用正则表达式快速验证URL基本格式
	if !isValidURL(rawURL) {
		return "", "", fmt.Errorf("无效的URL格式: %s", rawURL)
	}

	// 第二步：使用标准库解析URL，提取各部分信息
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("URL解析失败: %v", err)
	}

	// 第三步：验证特定的域名和路径格式
	// 确保域名是pan.quark.cn
	if parsedURL.Host != "pan.quark.cn" {
		return "", "", fmt.Errorf("不支持的域名: %s，期望 pan.quark.cn", parsedURL.Host)
	}

	// 确保路径以/s/开头，这是夸克网盘分享链接的标准格式
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

	// 第五步：验证资源ID的格式和长度
	// 确保资源ID长度合理（避免异常值）
	if len(resId) < 8 || len(resId) > 100 {
		return "", "", fmt.Errorf("resId长度无效: %d，应在8-100字符之间", len(resId))
	}

	// 第六步：从查询参数中提取密码（如果有）
	queryParams := parsedURL.Query()
	pwd = strings.TrimSpace(queryParams.Get("pwd"))

	// 第七步：如果存在密码，验证其格式
	if pwd != "" && (len(pwd) < 2 || len(pwd) > 50) {
		return "", "", fmt.Errorf("pwd参数长度无效: %d，应在2-50字符之间", len(pwd))
	}

	// 所有验证通过，返回提取的参数
	return resId, pwd, nil
}
