// Package core Copyright 2025 Share Sniffer
//
// telecom.go 实现了电信云盘链接检查器，作为策略模式的具体策略实现
// 提供了Telecom1Checker结构体，实现了LinkChecker接口的Check和GetPrefix方法
// 包含链接验证、参数提取、API调用和结果解析等完整流程
package core

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/owu/share-sniffer/internal/config"
)

// Telecom1Checker 电信云盘链接检查器
// 实现了LinkChecker接口，是策略模式的具体策略之一
// 负责检查电信云盘分享链接的有效性和获取分享内容信息

type Telecom1Checker struct{}

// Check 实现LinkChecker接口的Check方法
// 调用内部的checkTelecom方法执行具体的检查逻辑
//
// 参数:
// - urlStr: 需要检查的电信云盘分享链接
//
// 返回值:
// - Result: 包含检查结果的结构体
func (q *Telecom1Checker) Check(urlStr string) Result {

	// 检查code值是否包含编码特征且包含特殊字符，可能需要进一步解码
	if isURLEncoded(urlStr) && containsSpecialChars(urlStr) {
		var err error
		urlStr, err = url.QueryUnescape(urlStr)
		if err != nil {
			return Result{
				Name:   "链接格式无效",
				Status: 0,
			}
		}
	}

	//两种分享链接格式转换
	//config.GetSupportedTelecom1() -> config.GetSupportedTelecom()

	code := strings.Replace(urlStr, config.GetSupportedTelecom1(), "", 1)
	urlStr = fmt.Sprintf("%scode=%s", config.GetSupportedTelecom(), code)

	return checkTelecom(urlStr)
}

// GetPrefix 实现LinkChecker接口的GetPrefix方法
// 返回电信云盘链接的前缀，用于在注册时识别
//
// 返回值:
// - string: 电信云盘链接的前缀，从配置中获取
func (q *Telecom1Checker) GetPrefix() string {
	return config.GetSupportedTelecom1()
}
