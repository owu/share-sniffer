package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	apphttp "share-sniffer/internal/http"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// AliPanChecker 阿里云盘链接检查器
type AliPanChecker struct{}

// Check 实现LinkChecker接口
func (q *AliPanChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkAliPan(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口
func (q *AliPanChecker) GetPrefix() []string {
	return config.GetSupportedAliPan()
}

func (q *AliPanChecker) checkAliPan(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("AliPanChecker:开始检测阿里云盘链接: %s", urlStr)

	shareID, err := extractParamsAliPan(urlStr)
	if err != nil {
		logger.Info("AliPanChecker:extractParamsAliPan,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	response, err := aliPanRequest(ctx, shareID)
	if err != nil {
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		if errors.IsStatusCodeError(err) {
			return utils.ErrorInvalid("分享链接失效")
		}
		return utils.ErrorFatal("检测失败")
	}

	return utils.ErrorValid(response.ShareTitle)
}

func extractParamsAliPan(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("未找到share_id")
	}
	return parts[len(parts)-1], nil
}

type aliPanResp struct {
	ShareTitle string `json:"share_title"`
}

func aliPanRequest(ctx context.Context, shareID string) (*aliPanResp, error) {
	apiURL := "https://api.aliyundrive.com/adrive/v3/share_link/get_share_by_anonymous"
	requestBody := fmt.Sprintf(`{"share_id":"%s"}`, shareID)

	req, err := apphttp.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://www.alipan.com")
	req.Header.Set("referer", "https://www.alipan.com/")
	req.Header.Set("x-canary", "client=web,app=share,version=v2.3.1")

	resp, err := apphttp.DoWithRetry(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewStatusCodeError("链接已失效")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	var response aliPanResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	return &response, nil
}
