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

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	apphttp "share-sniffer/internal/http"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// YesChecker 123网盘链接检查器
type YesChecker struct{}

// Check 实现LinkChecker接口
func (y *YesChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return y.checkYes(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口
func (y *YesChecker) GetPrefix() []string {
	return config.GetSupportedYes()
}

func (y *YesChecker) checkYes(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("YesChecker:开始检测123网盘链接: %s", urlStr)

	resourceID, passCode, err := extractParamsYes(urlStr)
	if err != nil {
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	response, err := yesRequest(ctx, urlStr, resourceID, passCode)
	if err != nil {
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		return utils.ErrorFatal("检测失败")
	}

	if response.Info.Code != 0 {
		return utils.ErrorInvalid("分享链接失效")
	}

	return utils.ErrorValid(response.Info.Data.ShareName)
}

type yesResp struct {
	Info struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ShareName string `json:"ShareName"`
		} `json:"data"`
	} `json:"info"`
}

func yesRequest(ctx context.Context, originalURL string, resourceID string, passCode string) (*yesResp, error) {
	// 1. 获取 Cookie
	cookie, err := getCookieFromOriginalURL(ctx, originalURL)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://www.123684.com/gsb/s/%s", resourceID)
	req, err := apphttp.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Referer", originalURL)
	req.Header.Set("Cookie", cookie)

	resp, err := apphttp.DoWithRetry(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	var response yesResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	return &response, nil
}

func getCookieFromOriginalURL(ctx context.Context, originalURL string) (string, error) {
	req, err := apphttp.NewRequestWithContext(ctx, "GET", originalURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := apphttp.DoWithRetry(ctx, req, 1)
	if err != nil {
		return "", err
	}
	defer apphttp.CloseResponse(resp)

	var cookies []string
	for _, c := range resp.Cookies() {
		cookies = append(cookies, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}

	if len(cookies) == 0 {
		return "", fmt.Errorf("未获取到cookie")
	}

	return strings.Join(cookies, "; "), nil
}

var yesUrlRegex = regexp.MustCompile(`^https://www\.(123684|123865)\.com/s/[a-zA-Z0-9\-]+`)

func extractParamsYes(rawURL string) (resId, pwd string, err error) {
	if !yesUrlRegex.MatchString(rawURL) {
		return "", "", fmt.Errorf("URL格式不支持")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}

	resId = path.Base(u.Path)
	if resId == "" || resId == "s" {
		return "", "", fmt.Errorf("提取资源ID失败")
	}

	pwd = u.Query().Get("pwd")
	return resId, pwd, nil
}
