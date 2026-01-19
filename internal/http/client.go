package http

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	"share-sniffer/internal/logger"
)

var (
	// client 单例HTTP客户端
	client *http.Client
	// noRedirectClient 不跟随重定向的HTTP客户端单例
	noRedirectClient *http.Client
	// once 确保只初始化一次
	once sync.Once
)

// initClients 初始化所有客户端单例
func initClients() {
	once.Do(func() {
		cfg := config.GetConfig()
		transport := &http.Transport{
			MaxIdleConns:        cfg.HTTPClientConfig.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.HTTPClientConfig.MaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.HTTPClientConfig.IdleConnTimeout,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		}

		// 默认客户端
		client = &http.Client{
			Transport: transport,
			Timeout:   cfg.HTTPClientConfig.Timeout,
		}

		// 不重定向客户端
		noRedirectClient = &http.Client{
			Transport: transport,
			Timeout:   cfg.HTTPClientConfig.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		logger.Debug("HTTP客户端初始化完成: MaxIdleConns=%d, MaxIdleConnsPerHost=%d, Timeout=%v",
			cfg.HTTPClientConfig.MaxIdleConns,
			cfg.HTTPClientConfig.MaxIdleConnsPerHost,
			cfg.HTTPClientConfig.Timeout,
		)
	})
}

// GetClient 获取HTTP客户端单例
func GetClient() *http.Client {
	initClients()
	return client
}

// GetNoRedirectClient 获取不随重定向的HTTP客户端单例
func GetNoRedirectClient() *http.Client {
	initClients()
	return noRedirectClient
}

// DoWithRetry 执行HTTP请求并支持重试
func DoWithRetry(ctx context.Context, req *http.Request, maxRetries int) (*http.Response, error) {
	return DoWithClient(ctx, GetClient(), req, maxRetries)
}

// DoWithClient 使用指定客户端执行请求并支持重试
func DoWithClient(ctx context.Context, hclient *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	if maxRetries <= 0 {
		maxRetries = config.GetRetryCount()
	}

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 计算退避时间
			retryInterval := config.GetRetryInterval() * time.Duration(attempt)
			logger.Debug("请求重试 %d/%d, 等待 %v", attempt, maxRetries, retryInterval)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryInterval):
			}
		}

		// 发送请求
		resp, err := hclient.Do(req.WithContext(ctx))
		if err == nil {
			// 检查响应状态码 (5xx 通常可以重试)
			if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				logger.Warn("服务器错误 %d, 准备重试", resp.StatusCode)
				CloseResponse(resp)
				lastErr = errors.NewResponseErrorWithStatus("服务器错误", resp.StatusCode, nil)
				continue
			}
			return resp, nil
		}

		lastErr = err
		logger.Warn("请求失败: %v, 准备重试 %d/%d", err, attempt+1, maxRetries)

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	logger.Error("所有重试都失败: %v", lastErr)
	if netErr, ok := lastErr.(net.Error); ok && netErr.Timeout() {
		return nil, errors.NewTimeoutError("请求超时")
	}
	return nil, errors.NewRequestError("请求失败", lastErr)
}

// NewRequestWithContext 创建带上下文和默认Header的HTTP请求
func NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	SetDefaultHeaders(req)
	return req, nil
}

// SetDefaultHeaders 设置通用请求头
func SetDefaultHeaders(req *http.Request) {
	headers := map[string]string{
		"accept":          "application/json, text/javascript, */*; q=0.01",
		"accept-language": "zh-CN,zh;q=0.9,en;q=0.8",
		"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		"cache-control":   "no-cache",
		"pragma":          "no-cache",
		"connection":      "keep-alive",
	}
	for k, v := range headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
}

// CloseResponse 安全关闭并排空响应体
func CloseResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
