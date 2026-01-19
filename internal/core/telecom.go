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

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	apphttp "share-sniffer/internal/http"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// TelecomChecker 电信云盘链接检查器
type TelecomChecker struct{}

// Check 实现LinkChecker接口
func (q *TelecomChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return q.checkTelecom(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口
func (q *TelecomChecker) GetPrefix() []string {
	return config.GetSupportedTelecom()
}

func (q *TelecomChecker) checkTelecom(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("TelecomChecker:开始检测电信云盘链接: %s", urlStr)

	codeValue, refererValue, err := extractParamsTelecom(urlStr)
	if err != nil {
		logger.Info("TelecomChecker:extractParamsTelecom,%s,错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	response, err := telecomRequest(ctx, codeValue, refererValue)
	if err != nil {
		if errors.IsTimeoutError(err) {
			return utils.ErrorTimeout()
		}
		if errors.IsStatusCodeError(err) {
			return utils.ErrorInvalid("分享链接失效")
		}
		return utils.ErrorFatal("检测失败")
	}

	if response.ResCode == 0 && response.ResMessage == "成功" {
		return utils.ErrorValid(response.FileName)
	}

	logger.Debug("TelecomChecker:接口返回业务错误: res_code=%d, res_message=%s", response.ResCode, response.ResMessage)
	return utils.ErrorInvalid("分享内容无法访问")
}

func telecomRequest(ctx context.Context, codeValue string, refererValue string) (*TelecomResp, error) {
	// 4. 构建目标URL
	baseURL := "https://cloud.189.cn/api/open/share/getShareInfoByCodeV2.action"

	params := url.Values{}
	params.Set("noCache", fmt.Sprintf("%f", rand.New(rand.NewSource(time.Now().UnixNano())).Float64()))
	params.Set("shareCode", codeValue)

	targetURL := baseURL + "?" + params.Encode()

	req, err := apphttp.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("referer", refererValue)
	req.Header.Set("sign-type", "1")
	req.Header.Set("accept", "application/json;charset=UTF-8")

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
		return nil, errors.NewStatusCodeError(fmt.Sprintf("HTTP状态码: %d", resp.StatusCode))
	}

	var response TelecomResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewParseError("解析JSON失败", err)
	}

	return &response, nil
}

type TelecomResp struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`
	FileName   string `json:"fileName"`
}

func extractParamsTelecom(urlStr string) (string, string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}

	var codeValue string
	if strings.Contains(parsedURL.Path, "/t/") {
		codeValue = strings.TrimPrefix(parsedURL.Path, "/t/")
	} else {
		codeValue = parsedURL.Query().Get("code")
	}

	if codeValue == "" {
		return "", "", fmt.Errorf("未找到分享码")
	}

	// 清理多余字符
	if idx := strings.IndexAny(codeValue, "（%?&"); idx != -1 {
		codeValue = codeValue[:idx]
	}

	return codeValue, urlStr, nil
}
