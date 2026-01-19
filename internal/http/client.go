package http

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"share-sniffer/internal/config"
	"share-sniffer/internal/errors"
	"share-sniffer/internal/logger"
)

var (
	// client 单例HTTP客户端
	client *http.Client
	// once 确保只初始化一次
	once sync.Once
)

// GetClient 获取HTTP客户端单例
func GetClient() *http.Client {
	once.Do(func() {
		cfg := config.GetConfig()
		transport := &http.Transport{
			MaxIdleConns:        cfg.HTTPClientConfig.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.HTTPClientConfig.MaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.HTTPClientConfig.IdleConnTimeout,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		}

		client = &http.Client{
			Transport: transport,
			Timeout:   cfg.HTTPClientConfig.Timeout,
		}

		logger.Debug("HTTP客户端初始化完成: MaxIdleConns=%d, MaxIdleConnsPerHost=%d, Timeout=%v",
			cfg.HTTPClientConfig.MaxIdleConns,
			cfg.HTTPClientConfig.MaxIdleConnsPerHost,
			cfg.HTTPClientConfig.Timeout,
		)
	})
	return client
}

// DoWithRetry 执行HTTP请求并支持重试
func DoWithRetry(ctx context.Context, req *http.Request, maxRetries int) (*http.Response, error) {
	if maxRetries <= 0 {
		maxRetries = config.GetRetryCount()
	}

	client := GetClient()
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 计算退避时间（指数退避 + 随机因子）
			retryInterval := config.GetRetryInterval() * time.Duration(attempt)
			logger.Debug("请求重试 %d/%d, 等待 %v", attempt, maxRetries, retryInterval)

			// 等待退避时间，同时监听上下文取消
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryInterval):
				// 继续重试
			}
		}

		// 发送请求
		resp, err := client.Do(req.WithContext(ctx))
		if err == nil {
			// 检查响应状态码
			if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				// 服务器错误，需要重试
				logger.Warn("服务器错误 %d, 准备重试", resp.StatusCode)
				resp.Body.Close()
				lastErr = errors.NewResponseErrorWithStatus("服务器错误", resp.StatusCode, nil)
				continue
			}
			// 成功，返回响应
			return resp, nil
		}

		// 记录错误
		lastErr = err
		logger.Warn("请求失败: %v, 准备重试 %d/%d", err, attempt+1, maxRetries)

		// 检查是否可以重试
		if ctx.Err() != nil {
			// 上下文取消，不再重试
			return nil, ctx.Err()
		}
	}

	// 所有重试都失败
	logger.Error("所有重试都失败: %v", lastErr)
	// 根据错误类型包装为更具体的错误
	if netErr, ok := lastErr.(net.Error); ok {
		if netErr.Timeout() {
			return nil, errors.NewTimeoutError("请求超时")
		}
		return nil, errors.NewNetworkError("网络错误", lastErr)
	}
	return nil, errors.NewRequestError("请求失败", lastErr)
}

// NewRequestWithContext 创建带上下文的HTTP请求
func NewRequestWithContext(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	// 这里可以根据body类型进行不同处理
	// 简单实现，实际可能需要更复杂的逻辑
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	// 设置默认请求头
	SetDefaultHeaders(req)

	return req, nil
}

// SetDefaultHeaders 设置默认请求头
func SetDefaultHeaders(req *http.Request) {
	req.Header.Set("accept", "application/json;charset=UTF-8")
	req.Header.Set("accept-language", "en,zh-CN;q=0.9,zh;q=0.8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
}

// CloseResponse 安全关闭HTTP响应
func CloseResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		// 确保读取所有内容以重用连接
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// GetWithTimeout 发送GET请求并设置超时
func GetWithTimeout(url string, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	SetDefaultHeaders(req)
	return GetClient().Do(req)
}

// PostWithTimeout 发送POST请求并设置超时
func PostWithTimeout(url string, body interface{}, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 这里简化处理，实际应该根据body类型设置正确的请求体
	var req *http.Request
	var err error

	switch b := body.(type) {
	case []byte:
		req, err = http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	case string:
		req, err = http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(b))
	default:
		req, err = http.NewRequestWithContext(ctx, "POST", url, nil)
	}

	if err != nil {
		return nil, err
	}

	SetDefaultHeaders(req)
	// 设置内容类型
	req.Header.Set("content-type", "application/json")

	return DoWithRetry(ctx, req, config.GetRetryCount())
}
