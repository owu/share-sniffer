package core

import "time"

// Adapter 适配器函数，根据URL前缀调用对应的检查器
// 提供统一的链接检查入口，隐藏了具体检查器的实现细节
//
// 参数:
// - urlStr: 用户输入的链接字符串
//
// 返回值:
// - Result: 检查结果，包含URL、名称、状态码和耗时等信息
func Adapter(urlStr string) Result {
	// 输入验证
	if urlStr == "" {
		return Result{
			URL:     urlStr,
			Name:    "链接不能为空",
			Status:  0,
			Elapsed: 0,
		}
	}

	// 获取对应的检查器
	checker := GetChecker(urlStr)
	if checker != nil {
		startTime := time.Now()
		result := checker.Check(urlStr)
		result.URL = urlStr
		result.Elapsed = time.Since(startTime).Milliseconds()
		return result
	}

	return Result{
		URL:     urlStr,
		Name:    "链接尚未支持",
		Status:  0,
		Elapsed: 0,
	}
}
