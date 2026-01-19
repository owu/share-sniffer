// Package core Copyright 2025 Share Sniffer
//
// xunlei.go 实现了迅雷网盘链接检查器，作为策略模式的具体策略实现
// 提供了XunleiChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、页面访问和结果解析等完整流程
package core

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"share-sniffer/internal/config"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
)

// XunleiChecker 迅雷网盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查迅雷网盘分享链接的有效性和获取分享内容信息
type XunleiChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkXunlei方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的迅雷网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (x *XunleiChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return x.checkXunlei(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回迅雷网盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 迅雷网盘链接的前缀数组，从配置中获取
func (x *XunleiChecker) GetPrefix() []string {
	return config.GetSupportedXunlei()
}

// checkXunlei 检测迅雷网盘链接是否有效
// 这是XunleiChecker的核心方法，执行完整的链接检查流程
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的迅雷网盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体，包括URL、资源名称、错误码和耗时
func (x *XunleiChecker) checkXunlei(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("XunleiChecker:开始检测迅雷网盘链接: %s", urlStr)
	requestStart := time.Now()

	// 使用传入的context，不再创建新的context
	logger.Debug("使用传入的context进行迅雷链接检测")

	// 验证URL格式
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		logger.Info("XunleiChecker:ParseRequestURI, %s, 错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 确保是迅雷网盘链接
	if !strings.Contains(parsedURL.Host, "pan.xunlei.com") && !strings.Contains(parsedURL.Host, "lixian.vip.xunlei.com") {
		logger.Info("XunleiChecker:不是迅雷网盘链接: %s\n", urlStr)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 配置Chrome浏览器选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// 基本配置
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),

		// 更新用户代理为现代Chrome版本
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"),

		// 核心性能优化：禁用不必要的资源加载
		chromedp.Flag("blink-settings", "imagesEnabled=false,cssEnabled=false"),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-component-extensions-with-background-pages", true),
		chromedp.Flag("disable-preconnect", true),
		chromedp.Flag("disable-prefetch", true),
		chromedp.Flag("disable-predictive-networking", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-javascript-timeouts", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disk-cache-size", "0"),
		chromedp.Flag("media-cache-size", "0"),
		chromedp.Flag("window-size", "1280,800"),
	)

	// 创建执行上下文
	execCtx, execCancel := chromedp.NewExecAllocator(ctx, opts...)
	// 立即定义defer确保资源释放
	defer execCancel()

	// 创建浏览器上下文
	browserCtx, browserCancel := chromedp.NewContext(execCtx)
	// 立即定义defer确保资源释放
	defer browserCancel()

	// 导航到链接并等待页面加载完成
	var pageContent string
	var folderName string

	// 优化策略：分阶段执行检测，在每个阶段后检查是否有错误信息
	// 这样可以在检测到错误后立即返回，而不需要等待整个流程完成

	// 第一阶段：导航到页面并获取基本内容
	// 使用带超时的chromedp.Run调用，超时时间从配置中获取
	firstStageCtx, firstStageCancel := context.WithTimeout(browserCtx, config.GetLongTimeout())
	defer firstStageCancel()

	err = chromedp.Run(firstStageCtx,
		chromedp.Navigate(urlStr),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// 减少睡眠时间，避免不必要的等待
		chromedp.Sleep(500*time.Millisecond),
		chromedp.OuterHTML("html", &pageContent, chromedp.ByQuery),
	)

	// 如果第一阶段失败，尝试直接返回错误信息
	if err != nil {
		logger.Debug("XunleiChecker:第一阶段检测出错: %v", err)
		// 即使出错，我们仍然尝试处理已获取的页面内容
		if pageContent != "" {
			logger.Debug("XunleiChecker:第一阶段检测出错，但已获取部分页面内容")
		} else {
			logger.Debug("XunleiChecker:第一阶段检测出错，未获取到页面内容")
		}
	}

	// 检查是否有错误信息
	if pageContent != "" {
		lowerPageContent := strings.ToLower(pageContent)

		// 检查分享已删除
		if strings.Contains(lowerPageContent, "作者删除") || strings.Contains(lowerPageContent, "分享已删除") {
			logger.Info("XunleiChecker:分享已删除: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("该分享已被作者删除")
		}

		// 检查分享不存在或已过期
		if strings.Contains(lowerPageContent, "分享不存在") || strings.Contains(lowerPageContent, "已过期") || strings.Contains(lowerPageContent, "页面不存在") {
			logger.Info("XunleiChecker:分享不存在或已过期: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("分享不存在或已过期")
		}

		// 检查违规内容（与第三阶段保持一致，需要至少匹配2个关键词）
		violationKeywords := []string{"涉及侵权", "色情", "反动", "低俗", "无法访问"}
		violationCount := 0
		for _, keyword := range violationKeywords {
			if strings.Contains(lowerPageContent, strings.ToLower(keyword)) {
				violationCount++
			}
		}
		if violationCount >= 2 {
			logger.Info("XunleiChecker:分享内容违规: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("该分享内容可能涉及违规信息，无法访问！")
		}
	}

	// 第二阶段：如果没有错误，继续尝试获取文件夹名称
	if err == nil && pageContent != "" {
		// 创建一个带超时的上下文，限制第二阶段的执行时间
		secondStageCtx, cancel := context.WithTimeout(browserCtx, 3*time.Second)

		err = chromedp.Run(secondStageCtx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// 尝试多种DOM选择器来获取文件夹名称，但限制尝试次数和时间
				selectors := []string{
					".SourceListItem__name--y6dVw a span.highlight-text",
					".SourceListItem__name--y6dVw a",
					".SourceListItem__name--y6dVw",
					".highlight-text",
					".SourceListItem__title--fq2DG",
				}

				// 限制尝试的选择器数量，避免耗时过长
				maxSelectors := 5
				if len(selectors) > maxSelectors {
					selectors = selectors[:maxSelectors]
				}

				for _, selector := range selectors {
					// 检查上下文是否已取消
					if ctx.Err() != nil {
						return ctx.Err()
					}

					var tempName string
					// 使用带超时的文本获取操作
					selectorCtx, selectorCancel := context.WithTimeout(ctx, 300*time.Millisecond)
					err := chromedp.Text(selector, &tempName, chromedp.ByQuery).Do(selectorCtx)
					selectorCancel()

					if err == nil && strings.TrimSpace(tempName) != "" {
						folderName = strings.TrimSpace(tempName)
						return nil
					}
				}

				// 不再尝试获取所有文本内容，避免耗时过长
				return nil
			}),
		)

		// 释放第二阶段的上下文
		cancel()

		// 如果第二阶段出错，不影响整体检测结果
		if err != nil {
			logger.Debug("XunleiChecker:第二阶段检测出错: %v", err)
			err = nil // 重置错误，继续后续检测
		}
	}

	requestElapsed := time.Since(requestStart).Milliseconds()

	if err != nil {
		// 判断错误类型
		if ctx.Err() == context.DeadlineExceeded {
			logger.Info("XunleiChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorTimeout()
		}
		logger.Info("XunleiChecker:检测失败: %s, 错误: %v, 耗时: %dms", urlStr, err, requestElapsed)
		return utils.ErrorFatal("失败: " + err.Error())
	}

	// 检查页面内容中的错误信息
	lowerPageContent := strings.ToLower(pageContent)

	// 1. 检测分享已删除
	deletedKeywords := []string{"该分享已被作者删除", "分享已删除", "share has been deleted"}
	deletedCount := 0
	for _, keyword := range deletedKeywords {
		if strings.Contains(lowerPageContent, strings.ToLower(keyword)) {
			deletedCount++
		}
	}
	if deletedCount >= 1 {
		logger.Info("XunleiChecker:分享已删除: %s, 耗时: %dms", urlStr, requestElapsed)
		return utils.ErrorInvalid("该分享已被作者删除")
	}

	// 2. 检测涉及违规内容
	violationKeywords := []string{"涉及侵权", "色情", "反动", "低俗", "无法访问"}
	violationCount := 0
	for _, keyword := range violationKeywords {
		if strings.Contains(lowerPageContent, strings.ToLower(keyword)) {
			violationCount++
		}
	}
	if violationCount >= 2 {
		logger.Info("XunleiChecker:分享内容违规: %s, 耗时: %dms", urlStr, requestElapsed)
		return utils.ErrorInvalid("该分享内容可能涉及违规信息，无法访问！")
	}

	// 3. 检测暂无文件
	if strings.Contains(lowerPageContent, "暂无文件") {
		logger.Info("XunleiChecker:暂无文件: %s, 耗时: %dms", urlStr, requestElapsed)
		return utils.ErrorInvalid("暂无文件")
	}

	// 4. 检测分享不存在或已过期
	expiredKeywords := []string{"分享不存在", "已过期", "页面不存在", "not found", "404"}
	expiredCount := 0
	for _, keyword := range expiredKeywords {
		if strings.Contains(lowerPageContent, strings.ToLower(keyword)) {
			expiredCount++
		}
	}
	if expiredCount >= 2 {
		logger.Info("XunleiChecker:分享不存在或已过期: %s, 耗时: %dms", urlStr, requestElapsed)
		return utils.ErrorInvalid("分享不存在或已过期")
	}

	// 5. 检查是否为404页面
	if strings.Contains(pageContent, "statusCode:404") || strings.Contains(pageContent, "path:\"\\u002Fs") {
		logger.Info("XunleiChecker:分享不存在或已过期: %s, 耗时: %dms", urlStr, requestElapsed)
		return utils.ErrorInvalid("分享不存在或已过期")
	}

	// 如果从DOM中提取到了标题，使用它
	if folderName != "" {
		// 清理可能的空格和换行符
		folderName = strings.TrimSpace(folderName)
		// 修复文件名重复问题
		if len(folderName) > 10 && strings.Contains(folderName, folderName[:len(folderName)/2]) {
			folderName = folderName[:len(folderName)/2]
		}
		logger.Debug("XunleiChecker:检测成功: %s, 文件夹名称: %s, 请求耗时: %dms", urlStr, folderName, requestElapsed)
		return utils.ErrorValid(folderName)
	}

	// 如果所有方法都失败，保存页面内容到文件以便调试
	//if folderName == "" {
	//	filename := "xunlei_page_debug_" + time.Now().Format("20060102_150405") + ".html"
	//	file, err := os.Create(filename)
	//	if err == nil {
	//		defer file.Close()
	//		file.WriteString(pageContent)
	//		logger.Debug("XunleiChecker:页面内容已保存到 %s", filename)
	//	}
	//}

	// 如果所有方法都失败，返回未知错误
	logger.Info("XunleiChecker:无法获取文件夹名称: %s, 耗时: %dms", urlStr, requestElapsed)
	return utils.ErrorInvalid("无法获分享信息")
}
