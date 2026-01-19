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

// UcChecker UC网盘链接检查器
type UcChecker struct{}

// Check 实现LinkChecker接口
func (u *UcChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return u.checkUc(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口
func (u *UcChecker) GetPrefix() []string {
	return config.GetSupportedUc()
}

func (u *UcChecker) checkUc(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("UcChecker:开始检测UC网盘链接: %s", urlStr)

	code, err := extractParamsUc(urlStr)
	if err != nil {
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	response, err := ucRequest(ctx, code)
	if err != nil {
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		return utils.ErrorFatal("检测失败")
	}

	if response.Status == http.StatusOK && response.Code == 0 {
		return utils.ErrorValid(response.Data.DetailInfo.Share.Title)
	}

	return utils.ErrorInvalid("分享链接失效")
}

type ucResp struct {
	Status int `json:"status"`
	Code   int `json:"code"`
	Data   struct {
		DetailInfo struct {
			Share struct {
				Title string `json:"title"`
			} `json:"share"`
		} `json:"detail_info"`
	} `json:"data"`
}

func ucRequest(ctx context.Context, code string) (*ucResp, error) {
	apiURL := "https://pc-api.uc.cn/1/clouddrive/share/sharepage/v2/detail?pr=UCBrowser&fr=pc"
	requestBody := fmt.Sprintf(`{"pwd_id":"%s","passcode":"","force":0,"page":1,"size":50,"fetch_banner":1,"fetch_share":1,"fetch_total":1,"sort":"file_type:asc,file_name:asc","banner_platform":"other","web_platform":"windows","fetch_error_background":1}`, code)

	req, err := apphttp.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("origin", "https://drive.uc.cn")
	req.Header.Set("referer", "https://drive.uc.cn/")

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

	var response ucResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	return &response, nil
}

var ucUrlRegex = regexp.MustCompile(`^https://drive\.uc\.cn/s/[a-zA-Z0-9]+`)

func extractParamsUc(rawURL string) (string, error) {
	if !ucUrlRegex.MatchString(rawURL) {
		return "", fmt.Errorf("URL格式不支持")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	code := path.Base(u.Path)
	if code == "" || code == "s" {
		return "", fmt.Errorf("提取code失败")
	}

	return code, nil
}
