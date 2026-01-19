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

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	apphttp "share-sniffer/internal/http"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// YywChecker 115网盘链接检查器
type YywChecker struct{}

// Check 实现LinkChecker接口
func (q *YywChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkYyw(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口
func (q *YywChecker) GetPrefix() []string {
	return config.GetSupportedYyw()
}

func (q *YywChecker) checkYyw(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("YywChecker:开始检测115网盘链接: %s", urlStr)

	shareCode, receiveCode, err := extractParamsYyw(urlStr)
	if err != nil || shareCode == "" {
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	response, err := yywRequest(ctx, shareCode, receiveCode)
	if err != nil {
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		return utils.ErrorFatal("检测失败")
	}

	if !(response.State && response.Errno == 0) {
		return utils.ErrorInvalid("分享链接失效")
	}

	name := response.Data.Shareinfo.ShareTitle
	if name == "" && len(response.Data.List) > 0 {
		name = response.Data.List[0].N
	}

	return utils.ErrorValid(unicodeToChinese(name))
}

func extractParamsYyw(urlStr string) (shareCode, receiveCode string, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) > 0 {
		shareCode = parts[len(parts)-1]
	}
	receiveCode = u.Query().Get("password")
	if receiveCode == "" && u.Fragment != "" {
		fragment := strings.TrimPrefix(u.Fragment, "?")
		if params, err := url.ParseQuery(fragment); err == nil {
			receiveCode = params.Get("password")
		}
	}
	return
}

type yywResp struct {
	State bool `json:"state"`
	Errno int  `json:"errno"`
	Data  struct {
		Shareinfo struct {
			ShareTitle string `json:"share_title"`
		} `json:"shareinfo"`
		List []struct {
			N string `json:"n"`
		} `json:"list"`
	} `json:"data"`
}

func yywRequest(ctx context.Context, shareCode, receiveCode string) (*yywResp, error) {
	apiURL := fmt.Sprintf("https://115cdn.com/webapi/share/snap?share_code=%s&receive_code=%s&offset=0&limit=1", shareCode, receiveCode)

	req, err := apphttp.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Referer", fmt.Sprintf("https://115cdn.com/s/%s?password=%s", shareCode, receiveCode))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

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

	var response yywResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	return &response, nil
}

func unicodeToChinese(text string) string {
	re := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	return re.ReplaceAllStringFunc(text, func(s string) string {
		var r rune
		fmt.Sscanf(s[2:], "%04x", &r)
		return string(r)
	})
}
