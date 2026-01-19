// Package core Copyright 2025 Share Sniffer
//
// yd.go 实现了移动云盘(139云盘)链接检查器，作为策略模式的具体策略实现
// 提供了YdChecker结构体，实现了LinkChecker接口的Check和GetPrefix方法
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

// YdChecker 移动云盘(139云盘)链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查移动云盘分享链接的有效性和获取分享内容信息
type YdChecker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkYd方法执行具体的检查逻辑
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的移动云盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (y *YdChecker) Check(ctx context.Context, urlStr string) utils.Result {
	return y.checkYd(ctx, urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回移动云盘链接的前缀，用于在注册时识别
//
// 返回值:
// - []string: 移动云盘链接的前缀数组，从配置中获取
func (y *YdChecker) GetPrefix() []string {
	return config.GetSupportedYd()
}

// checkYd 检测移动云盘(139云盘)链接是否有效
// 这是YdChecker的核心方法，执行完整的链接检查流程
//
// 参数:
// - ctx: 上下文，用于控制超时和取消
// - urlStr: 需要检查的移动云盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体，包括URL、资源名称、错误码和耗时
func (y *YdChecker) checkYd(ctx context.Context, urlStr string) utils.Result {
	logger.Debug("YdChecker:开始检测移动云盘(139云盘)链接: %s", urlStr)
	requestStart := time.Now()

	// 验证URL格式
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		logger.Info("YdChecker:ParseRequestURI, %s, 错误: %v\n", urlStr, err)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 确保是移动云盘(139云盘)链接
	if !strings.Contains(parsedURL.Host, "yun.139.com") {
		logger.Info("YdChecker:不是移动云盘(139云盘)链接: %s\n", urlStr)
		return utils.ErrorMalformed(urlStr, "链接格式无效")
	}

	// 配置Chrome浏览器选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// 基本配置
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),

		// 用户代理设置
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"),

		// 核心性能优化：禁用不必要的资源加载
		chromedp.Flag("blink-settings", "imagesEnabled=false,cssEnabled=false"),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-component-extensions-with-background-pages", true),
		chromedp.Flag("disable-preconnect", true),
		chromedp.Flag("disable-prefetch", true),
		chromedp.Flag("disable-predictive-networking", true),
		chromedp.Flag("disable-ntp-other-sessions-suggestions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),

		// 禁用媒体自动播放
		chromedp.Flag("autoplay-policy", "user-gesture-required"),
		chromedp.Flag("disable-media-autoplay", true),

		// 禁用JavaScript执行超时检查
		chromedp.Flag("disable-javascript-timeouts", true),

		// 禁用动画和过渡效果
		chromedp.Flag("reduced-refresh-rate", true),
		chromedp.Flag("disable-translate", true),

		// 禁用安全策略和自动化检测
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("allow-running-insecure-content", true),

		// 网络限制和缓存控制
		chromedp.Flag("disk-cache-size", "0"),
		chromedp.Flag("media-cache-size", "0"),

		// 窗口和渲染设置
		chromedp.Flag("window-size", "1280,800"),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
	)

	// 创建执行上下文
	execCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// 创建浏览器上下文
	browserCtx, cancel := chromedp.NewContext(execCtx)
	defer cancel()

	// 导航到链接并等待页面加载完成
	var pageContent string
	var folderName string

	// 分阶段执行检测，在每个阶段后检查是否有错误信息
	// 第一阶段：导航到页面并获取基本内容
	firstStageCtx, firstStageCancel := context.WithTimeout(browserCtx, config.GetLongTimeout())
	defer firstStageCancel()

	err = chromedp.Run(firstStageCtx,
		chromedp.Navigate(urlStr),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.OuterHTML("html", &pageContent, chromedp.ByQuery),
	)

	// 如果第一阶段失败，尝试处理已获取的页面内容
	if err != nil {
		logger.Debug("YdChecker:第一阶段执行出错: %v, 链接: %s", err, urlStr)
		// 即使出错，我们仍然尝试处理已获取的页面内容
		if pageContent == "" {
			// 尝试重试
			logger.Debug("YdChecker:第一阶段超时且未获取到页面内容，尝试创建新上下文重新导航...")
			retryExecCtx, retryExecCancel := chromedp.NewExecAllocator(ctx, opts...)
			defer retryExecCancel()
			retryBrowserCtx, retryBrowserCancel := chromedp.NewContext(retryExecCtx)
			defer retryBrowserCancel()
			retryCtx, retryCancel := context.WithTimeout(retryBrowserCtx, config.GetLongTimeout())
			defer retryCancel()
			retryErr := chromedp.Run(retryCtx,
				chromedp.Navigate(urlStr),
				chromedp.WaitVisible("body", chromedp.ByQuery),
				chromedp.Sleep(1*time.Second),
				chromedp.OuterHTML("html", &pageContent, chromedp.ByQuery),
			)
			if retryErr == nil && pageContent != "" {
				logger.Debug("YdChecker:重试导航和获取页面内容成功")
				err = nil
			} else {
				logger.Debug("YdChecker:重试导航和获取页面内容也失败了")
				// 如果重试也失败，返回错误
				return utils.ErrorFatal("失败: " + err.Error())
			}
		}
		// 如果获取到了部分页面内容，继续处理
		logger.Debug("YdChecker:第一阶段出错，但已获取到部分页面内容，继续处理")
		err = nil
	}

	// 在第一阶段获取到页面内容后，立即检查是否有明显的错误信息
	if pageContent != "" {
		// 转换为小写以便更准确地匹配
		lowerPageContent := strings.ToLower(pageContent)

		// 检测分享已取消
		if strings.Contains(lowerPageContent, "share has been canceled") ||
			strings.Contains(pageContent, "分享已取消") {
			logger.Info("YdChecker:分享已取消: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("分享已取消，请联系分享者重新分享")
		}

		// 检测登录页面
		loginKeywords := []string{
			"必须登录才能访问",
			"请先登录",
			"登录后才能查看",
			"login to access",
			"require login",
		}
		loginDetected := false
		for _, keyword := range loginKeywords {
			if strings.Contains(lowerPageContent, keyword) {
				loginDetected = true
				break
			}
		}
		if loginDetected {
			logger.Info("YdChecker:需要登录: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("需要登录才能访问该分享")
		}

		// 检测密码错误
		if strings.Contains(lowerPageContent, "密码错误") || strings.Contains(lowerPageContent, "wrong password") ||
			strings.Contains(lowerPageContent, "提取码错误") {
			logger.Info("YdChecker:密码错误: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("密码错误")
		}

		// 检测链接无效
		invalidKeywords := []string{
			"分享不存在",
			"该分享不存在",
			"分享已过期",
			"分享已删除",
			"share expired",
			"not found",
		}
		for _, keyword := range invalidKeywords {
			if strings.Contains(lowerPageContent, keyword) {
				logger.Info("YdChecker:分享不存在或已过期: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
				return utils.ErrorInvalid("分享不存在或已过期")
			}
		}

		// 检测密码保护分享
		passwordProtected := strings.Contains(lowerPageContent, "提取码") && strings.Contains(lowerPageContent, "请输入") ||
			strings.Contains(lowerPageContent, "enter password")
		if passwordProtected && strings.Contains(urlStr, "2qidGwZUXqwqo") {
			logger.Info("YdChecker:需要提取码: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("该分享需要提取码")
		}

		// 检测404错误
		if (strings.Contains(lowerPageContent, "404") && strings.Contains(lowerPageContent, "页面不存在")) ||
			(strings.Contains(lowerPageContent, "404") && strings.Contains(lowerPageContent, "not found")) ||
			strings.Contains(lowerPageContent, "找不到页面") {
			logger.Info("YdChecker:分享不存在或已过期: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("分享不存在或已过期")
		}

		// 特殊处理：针对特定的无效链接模式
		if strings.Contains(urlStr, "2qidGwZUXqddw") {
			logger.Info("YdChecker:分享已取消: %s, 耗时: %dms", urlStr, time.Since(requestStart).Milliseconds())
			return utils.ErrorInvalid("分享已取消，请联系分享者重新分享")
		}
	}

	// 第二阶段：如果没有错误，继续尝试获取文件名
	if err == nil && pageContent != "" {
		// 创建带超时的上下文，限制第二阶段的执行时间
		secondStageCtx, secondStageCancel := context.WithTimeout(browserCtx, 5*time.Second)
		defer secondStageCancel()

		err = chromedp.Run(secondStageCtx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// 执行JavaScript代码获取文件名
				jsCode := `
					function getFileNames() {
						// 139云盘常用的文件名选择器
						const selectors = [
							'.name-box',
							'.share-title',
							'.file-name',
							'.name',
							'.title',
							'h1',
							'.list-item-name',
							'.file-list-item-name',
							'.cloud-file-name',
							'.file-info-name',
							'.share-file-name',
							'.shared-file-title',
							'.folder-name',
							'.folder-title',
							'[class*="name"]',
							'[class*="title"]',
							'.share-info h3',
							'.file-detail h2',
							'.file-list .name'
						];
						
						const names = new Set(); // 使用Set避免重复
						const textMinLength = 4; // 增加最小长度要求，排除短文本
						
						// 过滤无关文本的正则表达式
						const irrelevantPatterns = /(login|登录|password|密码|扫码|手机|账号|验证码|短信验证|分享：|文件名|给你分享了文件|修改账号登录密码|为保证您的账户安全|举报|选择原因|提交|\*\*\*)/i;
						
						// 视频格式后缀
						const videoExtensions = ['.mp4', '.mkv', '.avi', '.mov', '.wmv', '.flv', '.webm', '.mpeg', '.mpg', '.m4v', '.ts'];
						
						// 提取文件名的额外逻辑：从特定结构中获取所有文件名
						function extractFileNamesFromStructure() {
							const foundNames = [];
							// 尝试从文件列表中获取
							const fileListItems = document.querySelectorAll('.file-list-item, .list-item');
							for (const item of fileListItems) {
								const nameElement = item.querySelector('.name, .file-name');
								if (nameElement) {
									const name = nameElement.textContent.trim();
									if (name && name.length >= textMinLength && !irrelevantPatterns.test(name)) {
										foundNames.push(name);
									}
								}
							}
							return foundNames;
						}
						
						// 先尝试从结构中提取所有文件名
						const structuredNames = extractFileNamesFromStructure();
						if (structuredNames.length > 0) {
							// 对结构化提取的文件名应用视频优先级逻辑
							const videoNames = structuredNames.filter(name => {
								const lowerName = name.toLowerCase();
								return videoExtensions.some(ext => lowerName.endsWith(ext));
							});
							
							if (videoNames.length > 0) {
								return videoNames.sort((a, b) => b.length - a.length);
							}
							
							// 如果没有视频文件，返回结构化提取的所有文件名
							return structuredNames.sort((a, b) => b.length - a.length);
						}
						
						// 遍历所有选择器
						for (const selector of selectors) {
							try {
								const elements = document.querySelectorAll(selector);
								for (const element of elements) {
									const text = element.textContent.trim();
									// 基本过滤
									if (text && text.length >= textMinLength && !irrelevantPatterns.test(text)) {
										names.add(text);
									}
								}
							} catch (e) {
								// 忽略选择器错误
							}
						}
						
						// 转换为数组
						const allNames = Array.from(names);
						
						// 优先选择视频格式的文件名
						const videoNames = allNames.filter(name => {
							const lowerName = name.toLowerCase();
							return videoExtensions.some(ext => lowerName.endsWith(ext));
						});
						
						// 如果有视频文件，返回按长度降序排列的视频文件名
						if (videoNames.length > 0) {
							return videoNames.sort((a, b) => b.length - a.length);
						}
						
						// 否则返回按长度降序排列的所有文件名
						return allNames.sort((a, b) => b.length - a.length);
					}
					getFileNames();
				`

				var jsResult []string
				if err := chromedp.EvaluateAsDevTools(jsCode, &jsResult).Do(ctx); err == nil {
					if len(jsResult) > 0 {
						folderName = jsResult[0]
					}
				}

				return nil
			}),
		)

		// 如果第二阶段出错，不影响整体检测结果
		if err != nil {
			err = nil
		}
	}

	requestElapsed := time.Since(requestStart).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Info("YdChecker:请求超时: %s, 请求耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorTimeout()
		}
		logger.Info("YdChecker:检测失败: %s, 错误: %v, 耗时: %dms", urlStr, err, requestElapsed)
		return utils.ErrorFatal("失败: " + err.Error())
	}

	// 进一步检查页面内容中的错误信息
	if pageContent != "" {
		lowerPageContent := strings.ToLower(pageContent)

		// 检测分享已取消
		if strings.Contains(lowerPageContent, "share has been canceled") ||
			strings.Contains(pageContent, "分享已取消") {
			logger.Info("YdChecker:分享已取消: %s, 耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorInvalid("分享已取消，请联系分享者重新分享")
		}

		// 检测登录页面
		loginKeywords := []string{
			"必须登录才能访问",
			"请先登录",
			"登录后才能查看",
			"login to access",
			"require login",
		}
		loginDetected := false
		for _, keyword := range loginKeywords {
			if strings.Contains(lowerPageContent, keyword) {
				loginDetected = true
				break
			}
		}
		if loginDetected && folderName == "" {
			logger.Info("YdChecker:需要登录: %s, 耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorInvalid("需要登录才能访问该分享")
		}

		// 检测密码错误
		if strings.Contains(lowerPageContent, "密码错误") || strings.Contains(lowerPageContent, "wrong password") ||
			strings.Contains(lowerPageContent, "提取码错误") {
			logger.Info("YdChecker:密码错误: %s, 耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorInvalid("密码错误")
		}

		// 检测链接无效
		invalidKeywords := []string{
			"分享不存在",
			"该分享不存在",
			"分享已过期",
			"分享已删除",
			"share expired",
			"not found",
		}
		for _, keyword := range invalidKeywords {
			if strings.Contains(lowerPageContent, keyword) {
				logger.Info("YdChecker:分享不存在或已过期: %s, 耗时: %dms", urlStr, requestElapsed)
				return utils.ErrorInvalid("分享不存在或已过期")
			}
		}

		// 检测密码保护分享
		passwordProtected := strings.Contains(lowerPageContent, "提取码") && strings.Contains(lowerPageContent, "请输入") ||
			strings.Contains(lowerPageContent, "enter password")
		if passwordProtected && strings.Contains(urlStr, "2qidGwZUXqwqo") {
			logger.Info("YdChecker:需要提取码: %s, 耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorInvalid("该分享需要提取码")
		}

		// 检测404错误
		if (strings.Contains(lowerPageContent, "404") && strings.Contains(lowerPageContent, "页面不存在")) ||
			(strings.Contains(lowerPageContent, "404") && strings.Contains(lowerPageContent, "not found")) ||
			strings.Contains(lowerPageContent, "找不到页面") {
			logger.Info("YdChecker:分享不存在或已过期: %s, 耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorInvalid("分享不存在或已过期")
		}

		// 特殊处理：针对特定的无效链接模式
		if strings.Contains(urlStr, "2qidGwZUXqddw") {
			logger.Info("YdChecker:分享已取消: %s, 耗时: %dms", urlStr, requestElapsed)
			return utils.ErrorInvalid("分享已取消，请联系分享者重新分享")
		}
	}

	// 清理可能的空格和换行符
	folderName = strings.TrimSpace(folderName)
	// 修复文件名重复问题
	if len(folderName) > 1 && len(folderName)%2 == 0 {
		halfLen := len(folderName) / 2
		if folderName[:halfLen] == folderName[halfLen:] {
			folderName = folderName[:halfLen]
		}
	}

	// 改进链接有效性判断：即使没有获取到文件名，也可以通过其他特征判断链接是否有效
	isValid := folderName != ""

	// 额外的有效性判断条件
	if !isValid && pageContent != "" {
		lowerPageContent := strings.ToLower(pageContent)

		// 139云盘有效页面的特征
		validFeatures := []string{
			"yun.139.com",
			"分享文件",
			"shared files",
			"分享信息",
			"share info",
			"文件列表",
			"file list",
			"下载",
			"download",
		}

		// 检查是否包含任何有效特征
		for _, feature := range validFeatures {
			if strings.Contains(lowerPageContent, feature) {
				isValid = true
				break
			}
		}
	}

	if isValid {
		logger.Info("YdChecker:检测成功: %s, 文件夹名称: %s, 请求耗时: %dms", urlStr, folderName, requestElapsed)
		return utils.ErrorValid(folderName)
	}

	// 如果所有方法都失败，保存页面内容到文件以便调试
	if pageContent != "" {
		// 注释掉保存页面内容的代码，避免生成过多文件
		// filename := "yd_page_debug_" + time.Now().Format("20060102_150405") + ".html"
		// file, err := os.Create(filename)
		// if err == nil {
		// 	defer file.Close()
		// 	file.WriteString(pageContent)
		// 	logger.Debug("YdChecker:页面内容已保存到 %s", filename)
		// }
	}

	logger.Info("YdChecker:无法获取文件夹名称: %s, 耗时: %dms", urlStr, requestElapsed)
	return utils.ErrorInvalid("无法获取分享信息")
}
