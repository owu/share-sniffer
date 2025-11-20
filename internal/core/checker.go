// Package core Copyright 2025 Share Sniffer
//
// checker.go 实现了链接检查的核心逻辑，采用策略模式和工厂模式的组合
// 提供了LinkChecker接口、检查器注册机制和适配器函数
package core

import (
	"context"
	"strings"
	"sync"

	"github.com/owu/share-sniffer/internal/logger"
	"github.com/owu/share-sniffer/internal/utils"
)

// 链接检查器注册器
var (
	// checkers 存储所有已注册的链接检查器
	// 键为URL前缀，值为对应的检查器实例
	checkers = make(map[string]LinkChecker)

	// once 确保初始化只执行一次
	// 用于保证registerCheckers函数在并发环境下的线程安全
	once sync.Once
)

// LinkChecker 链接检查器接口
// 定义了链接检查器必须实现的两个方法
// 1. Check: 检查给定URL并返回检查结果
// 2. GetPrefix: 获取该检查器支持的URL前缀列表
type LinkChecker interface {
	// Check 检查链接有效性
	//
	// 参数:
	// - ctx: 上下文，用于控制超时和取消
	// - urlStr: 需要检查的URL字符串
	//
	// 返回值:
	// - Result: 包含检查结果的结构体
	Check(ctx context.Context, urlStr string) utils.Result

	// GetPrefix 获取支持的链接前缀列表
	//
	// 返回值:
	// - []string: URL前缀列表，用于在GetChecker中匹配对应的检查器
	GetPrefix() []string
}

// RegisterChecker 注册链接检查器
// 实现了工厂模式，允许动态添加新的检查器
//
// 参数:
// - checker: 实现了LinkChecker接口的检查器实例
func RegisterChecker(checker LinkChecker) {
	prefixes := checker.GetPrefix()
	for _, prefix := range prefixes {
		checkers[prefix] = checker
		logger.Debug("LinkChecker:注册检查器,%s", prefix)
	}
}

// GetChecker 根据URL获取对应的检查器
// 使用策略模式，根据URL特征选择合适的检查器
//
// 参数:
// - urlStr: 需要检查的URL字符串
//
// 返回值:
// - LinkChecker: 匹配的检查器实例，如果没有找到则返回nil
func GetChecker(urlStr string) LinkChecker {
	for prefix, checker := range checkers {
		if strings.HasPrefix(urlStr, prefix) {
			return checker
		}
	}
	return nil
}
