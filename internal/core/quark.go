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

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	apphttp "share-sniffer/internal/http"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// QuarkChecker 夸克网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查夸克网盘分享链接的有效性和获取分享内容信息

type QuarkChecker struct{}

// Check 实现LinkChecker接口的Check方法
func (q *QuarkChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkQuark(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
func (q *QuarkChecker) GetPrefix() []string {
	return config.GetSupportedQuark()
}

// quarkResp 夸克API响应结构
type quarkResp struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Title string `json:"title"`
	} `json:"data"`
}

// checkQuark 检测夸克网盘链接是否有效
func (q *QuarkChecker) checkQuark(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("QuarkChecker:开始检测夸克网盘链接: %s", urlStr)

	// 提取资源ID和密码
	resourceID, passCode, err := extractParamsQuark(urlStr)
	if err != nil {
		logger.Info("QuarkChecker:extractParamsQuark,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 发送请求
	requestStart := time.Now()
	response, err := quarkRequest(ctx, resourceID, passCode)
	requestElapsed := time.Since(requestStart).Milliseconds()
	logger.Debug("QuarkChecker:请求完成，请求耗时: %v", requestElapsed)

	if err != nil {
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		if errors.IsStatusCodeError(err) {
			return utils.ErrorInvalid("分享链接失效")
		}
		return utils.ErrorFatal("请求失败")
	}

	// 检查API响应状态
	if response.Status != http.StatusOK || response.Code != 0 {
		return utils.ErrorInvalid("分享链接失效或不存在")
	}

	return utils.ErrorValid(response.Data.Title)
}

// quarkRequest 获取夸克网盘分享信息
func quarkRequest(ctx context.Context, resourceID string, passCode string) (*quarkResp, error) {
	apiURL := "https://drive-h.quark.cn/1/clouddrive/share/sharepage/token"

	// 构造请求体
	requestBody := map[string]interface{}{
		"pwd_id":                            resourceID,
		"passcode":                          passCode,
		"support_visit_limit_private_share": true,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, err := apphttp.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://pan.quark.cn")
	req.Header.Set("referer", "https://pan.quark.cn/")

	resp, err := apphttp.DoWithRetry(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查HTTP状态码
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewStatusCodeError("链接已失效")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("状态码: %d", resp.StatusCode)
	}

	var response quarkResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	return &response, nil
}

// 验证URL格式的正则表达式
var urlRegex = regexp.MustCompile(`^https://pan\.quark\.cn/s/[a-zA-Z0-9]+(?:\?pwd=[a-zA-Z0-9]*)?$`)

func isValidURL(rawURL string) bool {
	return urlRegex.MatchString(rawURL)
}

func extractParamsQuark(rawURL string) (resId, pwd string, err error) {
	if !isValidURL(rawURL) {
		return "", "", fmt.Errorf("无效的URL格式")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}

	if parsedURL.Host != "pan.quark.cn" {
		return "", "", fmt.Errorf("不支持的域名")
	}

	resId = path.Base(parsedURL.Path)
	if resId == "" || resId == "s" {
		return "", "", fmt.Errorf("无法寻找资源ID")
	}

	pwd = parsedURL.Query().Get("pwd")
	return resId, pwd, nil
}
