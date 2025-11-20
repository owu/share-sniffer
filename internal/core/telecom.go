// Package core Copyright 2025 Share Sniffer
//
// telecom.go 实现了电信云盘链接检查器，作为策略模式的具体策略实现
// 提供了TelecomChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、参数提取、API调用和结果解析等完整流程
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/errors"
	apphttp "github.com/owu/share-sniffer/internal/http"
	"github.com/owu/share-sniffer/internal/logger"
	"github.com/owu/share-sniffer/internal/utils"
)

// TelecomChecker 电信云盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查电信云盘分享链接的有效性和获取分享内容信息

type TelecomChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkTelecom方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的电信云盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (q *TelecomChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return checkTelecom(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回电信云盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 电信云盘链接的前缀数组，从配置中获取
func (q *TelecomChecker) GetPrefix() []string {
	return config.GetSupportedTelecom()
}

// checkTelecom 检查电信云盘链接
// 记录开始时间，调用具体的检查方法，并计算耗时
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的电信云盘分享链接
//
// 返回值:
// - Result: 包含检查结果和耗时的结构体
func checkTelecom(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("TelecomChecker:开始检测电信云盘链接: %s", urlStr)

	// 使用传入的context - 确保请求受任务池的超时控制
	logger.Debug("TelecomChecker:使用传入的context进行检测")

	// 1. 提取code参数 - 这是访问电信云盘API的关键参数
	codeValue, refererValue, err := extractParamsTelecom(urlStr)
	if err != nil {
		logger.Info("TelecomChecker:extractParamsTelecom,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}
	// 发送请求并处理错误 - 调用夸克API获取分享信息
	requestStart := time.Now()
	response, err := telecomRequest(ctx, codeValue, refererValue)
	requestElapsed := time.Since(requestStart).Milliseconds()
	logger.Debug("TelecomChecker:请求完成，请求耗时: %v", requestElapsed)

	if err != nil {
		// 判断错误类型 - 区分超时错误和其他错误
		if errors.IsTimeoutError(err) {
			logger.Info("TelecomChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorTimeout()
		}

		if errors.IsStatusCodeError(err) {
			return utils.ErrorInvalid("分享链接失效")
		}

		logger.Info("TelecomChecker:检测失败: %s, 错误: %v, 请求耗时: %dms", urlStr, err, requestElapsed)
		return utils.ErrorFatal("失败: " + err.Error())
	}

	logger.Debug("TelecomChecker:检测成功: %s, 文件名: %s, 请求完成: %dms", urlStr, response.FileName, requestElapsed)

	// 10. 根据接口返回状态设置结果 - 检查API返回的业务状态码
	if response.ResCode == 0 && response.ResMessage == "成功" {
		return utils.ErrorValid(response.FileName)
	} else {
		logger.Debug("接口返回错误: res_code=%d, res_message=%s\n", response.ResCode, response.ResMessage)
		return utils.ErrorInvalid("")
	}
}

func telecomRequest(ctx context.Context, codeValue string, refererValue string) (*TelecomResp, error) {

	// 2. 生成随机noCache参数 - 避免API返回缓存结果
	rand.Seed(time.Now().UnixNano())
	noCacheValue := rand.Float64()

	// 3. 对shareCode进行URL编码 - 确保特殊字符被正确处理
	shareCodeEncoded := url.QueryEscape(codeValue)

	// 4. 构建目标URL - 使用电信云盘开放API接口
	baseURL := "https://cloud.189.cn/api/open/share/getShareInfoByCodeV2.action"
	targetURL, err := url.Parse(baseURL)
	if err != nil {
		logger.Warn("解析基础URL失败: %v\n", err)
		return nil, fmt.Errorf("解析基础URL失败: %v", err)
	}

	// 添加查询参数
	query := targetURL.Query()
	query.Set("noCache", fmt.Sprintf("%f", noCacheValue))
	query.Set("shareCode", shareCodeEncoded)
	targetURL.RawQuery = query.Encode()

	// 6. 创建HTTP请求 - 准备发送到电信云盘API
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL.String(), nil)
	if err != nil {
		logger.Warn("创建HTTP请求失败: %v\n", err)
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 7. 设置请求头 - 模拟浏览器行为以避免被识别为爬虫
	apphttp.SetDefaultHeaders(req)
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", refererValue)
	req.Header.Set("sec-ch-ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sign-type", "1")

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

	// 检查HTTP状态码 400  404（分享链接失效）
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewStatusCodeError(fmt.Sprintf("状态码: %d, 响应: %s", resp.StatusCode, string(body)))
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 9. 读取响应体 - 将JSON响应解析为TelecomResponse结构体
	var response TelecomResp
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Info("解析JSON失败: %v, 响应体: %s", err, string(body[:min(100, len(body))]))
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &response, nil
}

// TelecomResp 对应电信云盘API返回的数据结构
// 用于解析API响应的JSON数据，获取分享链接的详细信息
//
// 关键字段说明:
// - ResCode: 响应状态码，0表示成功
// - ResMessage: 响应消息
// - FileName: 文件名，这是我们主要关心的字段
// - FileId: 文件ID
// - FileSize: 文件大小
// - ExpireTime: 过期时间

type TelecomResp struct {
	ResCode    int    `json:"res_code"`    // 响应状态码，0表示成功
	ResMessage string `json:"res_message"` // 响应消息
	AccessCode string `json:"accessCode"`  // 访问码
	Creator    struct {
		IconURL      string `json:"iconURL"`      // 创建者头像URL
		NickName     string `json:"nickName"`     // 创建者昵称
		Oper         bool   `json:"oper"`         // 是否为运营商用户
		OwnerAccount string `json:"ownerAccount"` // 所有者账号
		SuperVip     int    `json:"superVip"`     // 超级VIP状态
		Vip          int    `json:"vip"`          // VIP状态
	} `json:"creator"` // 创建者信息
	ExpireTime     int    `json:"expireTime"`     // 过期时间（单位：秒）
	ExpireType     int    `json:"expireType"`     // 过期类型
	FileCreateDate string `json:"fileCreateDate"` // 文件创建日期
	FileId         string `json:"fileId"`         // 文件ID
	FileLastOpTime string `json:"fileLastOpTime"` // 文件最后操作时间
	FileName       string `json:"fileName"`       // 文件名（主要信息）
	FileSize       int    `json:"fileSize"`       // 文件大小（字节）
	FileType       string `json:"fileType"`       // 文件类型
	IsFolder       bool   `json:"isFolder"`       // 是否为文件夹
	NeedAccessCode int    `json:"needAccessCode"` // 是否需要访问码
	ReviewStatus   int    `json:"reviewStatus"`   // 审核状态
	ShareDate      int64  `json:"shareDate"`      // 分享日期（时间戳）
	ShareId        int64  `json:"shareId"`        // 分享ID
	ShareMode      int    `json:"shareMode"`      // 分享模式
	ShareType      int    `json:"shareType"`      // 分享类型
}

// isURLEncoded 检查字符串是否包含URL编码特征
// 通过检测百分号后跟两位十六进制字符的模式来判断是否为URL编码
//
// 参数:
// - s: 需要检查的字符串
//
// 返回值:
// - bool: 如果字符串包含有效的URL编码模式，则返回true
func isURLEncoded(s string) bool {
	for i := 0; i < len(s); i++ {
		// 检查是否存在百分号(%)并且后面至少有两个字符
		if s[i] == '%' && i+2 < len(s) {
			// 检查百分号后面的两个字符是否都是十六进制数字
			if isHex(s[i+1]) && isHex(s[i+2]) {
				return true
			}
		}
	}
	return false
}

// isHex 检查字符是否为十六进制数字
//
// 参数:
// - c: 需要检查的字节字符
//
// 返回值:
// - bool: 如果字符是0-9、a-f或A-F，则返回true
func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// extractParamsTelecom 从URL中提取code参数，自动处理编码状态
// 这是电信云盘链接检查的核心辅助函数，负责提取API所需的关键参数
//
// 参数:
// - urlStr: 原始电信云盘分享链接
//
// 返回值:
// - string: 提取到的code参数值，可能已经解码
// - error: 如果解析失败或找不到code参数，则返回错误
func extractParamsTelecom(urlStr string) (string, string, error) {
	// 定义支持的前缀
	const webSharePrefix = "https://cloud.189.cn/web/share?code="
	const tPrefix = "https://cloud.189.cn/t/"

	var codeValue string

	// 检查URL属于哪种类型
	if strings.HasPrefix(urlStr, webSharePrefix) {
		// 类型1: https://cloud.189.cn/web/share?code=xxx
		// 提取code值
		codeValue = strings.TrimPrefix(urlStr, webSharePrefix)

		// 解析URL以获取查询参数
		parsedURL, err := url.Parse(urlStr)
		if err == nil {
			queryParams := parsedURL.Query()
			codeParam := queryParams.Get("code")
			if codeParam != "" {
				codeValue = codeParam
			}
		}
	} else if strings.HasPrefix(urlStr, tPrefix) {
		// 类型2: https://cloud.189.cn/t/xxx
		// 提取t/后面的部分
		codeValue = strings.TrimPrefix(urlStr, tPrefix)
	} else {
		return "", "", fmt.Errorf("不支持的电信云盘URL格式")
	}

	// 检查是否找到code值
	if codeValue == "" {
		return "", "", fmt.Errorf("输入URL中未找到有效的code参数")
	}

	// 处理可能的访问码后缀，例如：xxx（访问码：yyy）或 xxx%EF%BC%88%E8%AE%BF%E9%97%AE%E7%A0%81%EF%BC%9Ayyy%EF%BC%89
	// 这些访问码后缀在API调用中不需要，需要去除
	if idx := strings.IndexAny(codeValue, "（%"); idx != -1 {
		codeValue = codeValue[:idx]
	}

	// 对code值进行URL解码
	decodedCode, err := url.QueryUnescape(codeValue)
	if err == nil {
		codeValue = decodedCode
	}

	// 设置Referer值
	refererValue := urlStr

	return codeValue, refererValue, nil
}

// containsSpecialChars 检查字符串是否包含需要URL编码的特殊字符
// 用于判断code参数是否可能需要进一步解码
//
// 参数:
// - s: 需要检查的字符串
//
// 返回值:
// - bool: 如果字符串包含任何特殊字符，则返回true
func containsSpecialChars(s string) bool {
	// 定义需要URL编码的特殊字符集合
	specialChars := " ()[]{}<>!@#$%^&*+=|\\:;\"',?/~"
	for _, char := range s {
		// 检查字符是否在特殊字符集合中
		if strings.ContainsRune(specialChars, char) {
			return true
		}
	}
	return false
}
