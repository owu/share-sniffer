// Package core Copyright 2025 Share Sniffer
//
// yyw.go 实现了115网盘链接检查器，作为策略模式的具体策略实现
// 提供了YywChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、参数提取、API调用和结果解析等完整流程
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	apphttp "share-sniffer/internal/http"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// YywChecker 115网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查115网盘分享链接的有效性和获取分享内容信息

type YywChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkYyw方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的115网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (q *YywChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkYyw(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回115网盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 115网盘链接的前缀数组，从配置中获取
func (q *YywChecker) GetPrefix() []string {
	return config.GetSupportedYyw()
}

// checkYyw 检测115网盘链接是否有效
// 这是YywChecker的核心方法，执行完整的链接检查流程
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的115网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体，包括URL、资源名称、错误码和耗时
func (q *YywChecker) checkYyw(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("YywChecker:开始检测115网盘链接: %s", urlStr)

	// 提取参数
	shareCode, receiveCode, err := extractParamsYyw(urlStr)
	if err != nil || shareCode == "" || receiveCode == "" {
		logger.Info("YywChecker:extractParamsYyw,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 发送请求并处理错误 - 调用夸克API获取分享信息
	requestStart := time.Now()
	response, err := yywRequest(ctx, shareCode, receiveCode)
	requestElapsed := time.Since(requestStart).Milliseconds()
	logger.Debug("YywChecker:请求完成，请求耗时: %v", requestElapsed)

	if err != nil {
		// 判断错误类型 - 区分超时错误和其他错误
		if errors.IsTimeoutError(err) {
			logger.Info("YywChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorTimeout()
		}
		logger.Info("YywChecker:检测失败: %s, 错误: %v, 请求耗时: %dms", urlStr, err, requestElapsed)
		return utils.ErrorFatal("失败: " + err.Error())
	}

	// 检查API响应状态
	if !(response.State && response.Errno == 0) {
		return utils.ErrorFatal("失败")
	}
	result := utils.ErrorValid("")

	// 获取资源名称，优先使用share_title，如果没有则使用第一个文件的名称
	if response.Data.Shareinfo.ShareTitle != "" {
		result.Data.Name = response.Data.Shareinfo.ShareTitle
	} else if len(response.Data.List) > 0 && response.Data.List[0].N != "" {
		result.Data.Name = response.Data.List[0].N
	}
	// 处理Unicode转义序列（如\u86df\u9f99\u884c\u52a8）
	result.Data.Name = unicodeToChinese(result.Data.Name)

	logger.Debug("YywChecker:检测成功: %s, 文件名: %s, 请求完成: %dms", urlStr, result.Data.Name, requestElapsed)

	return result
}

func yywRequest(ctx context.Context, shareCode, receiveCode string) (*yywResp, error) {
	// 构建API请求URL
	apiURL := fmt.Sprintf("https://115cdn.com/webapi/share/snap?share_code=%s&offset=0&limit=20&receive_code=%s&cid=",
		shareCode, receiveCode)

	logger.Debug("准备请求115 API: %s, shareCode: %s, receiveCode: %s", apiURL, shareCode, receiveCode)

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Warn("创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头 - 模拟浏览器请求，确保API能够正确响应
	apphttp.SetDefaultHeaders(req)
	//req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	//req.Header.Set("Accept-Language", "en")
	//req.Header.Set("Cache-Control", "no-cache")
	//req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", fmt.Sprintf("https://115cdn.com/s/%s?password=%s&", shareCode, receiveCode))
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

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

	// 解析JSON响应
	var response yywResp
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Info("解析JSON失败: %v, 响应体: %s", err, string(body[:min(100, len(body))]))
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &response, nil
}

type yywResp struct {
	State bool   `json:"state"`
	Error string `json:"error"`
	Errno int    `json:"errno"`
	Data  struct {
		Userinfo struct {
			UserId   string `json:"user_id"`
			UserName string `json:"user_name"`
			Face     string `json:"face"`
		} `json:"userinfo"`
		Shareinfo struct {
			SnapId           string `json:"snap_id"`
			FileSize         int64  `json:"file_size"`
			ShareTitle       string `json:"share_title"`
			ShareState       int    `json:"share_state"`
			ForbidReason     string `json:"forbid_reason"`
			CreateTime       int    `json:"create_time"`
			ReceiveCode      string `json:"receive_code"`
			ReceiveCount     int    `json:"receive_count"`
			ExpireTime       int    `json:"expire_time"`
			FileCategory     int    `json:"file_category"`
			AutoRenewal      int    `json:"auto_renewal"`
			ShareDuration    int    `json:"share_duration"`
			AutoFillRecvcode int    `json:"auto_fill_recvcode"`
			CanReport        int    `json:"can_report"`
			CanNotice        int    `json:"can_notice"`
			HaveVioFile      int    `json:"have_vio_file"`
			SkipLoginState   string `json:"skip_login_state"`
		} `json:"shareinfo"`
		Count int `json:"count"`
		List  []struct {
			Fid         string        `json:"fid"`
			Uid         int           `json:"uid"`
			Cid         json.Number   `json:"cid"`
			N           string        `json:"n"`
			Ns          string        `json:"ns"`
			S           int64         `json:"s"`
			Fc          int           `json:"fc"`
			T           string        `json:"t"`
			D           int           `json:"d"`
			C           int           `json:"c"`
			E           string        `json:"e"`
			Ico         string        `json:"ico"`
			Sha         string        `json:"sha"`
			IsSkipLogin int           `json:"is_skip_login"`
			Fl          []interface{} `json:"fl"`
			U           string        `json:"u"`
			Iv          int           `json:"iv"`
			Vdi         int           `json:"vdi"`
			PlayLong    int           `json:"play_long"`
		} `json:"list"`
		ShareState int `json:"share_state"`
		UserAppeal struct {
			CanAppeal       int `json:"can_appeal"`
			CanShareAppeal  int `json:"can_share_appeal"`
			PopupAppealPage int `json:"popup_appeal_page"`
			CanGlobalAppeal int `json:"can_global_appeal"`
		} `json:"user_appeal"`
	} `json:"data"`
}

// extractParamsYyw 从URL中提取share_code和receive_code
func extractParamsYyw(urlStr string) (shareCode, receiveCode string, err error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}

	// 从查询参数中获取share_code（路径中的最后一个部分）
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) > 0 {
		shareCode = pathParts[len(pathParts)-1]
	}

	// 从查询参数中获取password
	receiveCode = parsedURL.Query().Get("password")

	// 如果password为空，尝试从锚点中获取
	if receiveCode == "" && parsedURL.Fragment != "" {
		if strings.Contains(parsedURL.Fragment, "password=") {
			fragmentParams, _ := url.ParseQuery(parsedURL.Fragment)
			receiveCode = fragmentParams.Get("password")
		}
	}

	return shareCode, receiveCode, nil
}

// unicodeToChinese 将Unicode转义序列转换为中文字符
func unicodeToChinese(text string) string {
	if text == "" {
		return text
	}

	// 正则表达式匹配Unicode转义序列
	re := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	result := re.ReplaceAllStringFunc(text, func(s string) string {
		// 去掉\u前缀，转换为rune
		var r rune
		fmt.Sscanf(s[2:], "%04x", &r)
		return string(r)
	})

	return result
}
