package core

import (
	"context"
	"strings"
	"time"

	"share-sniffer/internal/utils"
)

// Adapter 适配器函数，根据URL前缀调用对应的检查器
// 提供统一的链接检查入口，隐藏了具体检查器的实现细节
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 用户输入的链接字符串
//
// 返回值:
// - Result: 包含检查结果的结构体
func Adapter(ctx context.Context, urlStr string) utils.Result {
	// 输入验证
	if "" == urlStr {
		return utils.ErrorMalformed(urlStr, "链接不能为空")
	}

	// 获取对应的检查器
	checker := GetChecker(urlStr)
	if nil == checker {
		return utils.ErrorMalformed(urlStr, "链接尚未支持")
	}

	startTime := time.Now()
	result := checker.Check(ctx, urlStr)
	result.Data.URL = urlStr
	result.Data.Elapsed = time.Since(startTime).Milliseconds()
	result.Data.Name = strings.TrimSpace(result.Data.Name)

	return result
}
