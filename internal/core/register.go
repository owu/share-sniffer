package core

import (
	"github.com/owu/share-sniffer/internal/utils"
)

// 初始化检查器
func init() {
	registerCheckers()
}

// registerCheckers 注册所有链接检查器
// 使用sync.Once确保只执行一次初始化
func registerCheckers() {
	once.Do(func() {
		// 注册夸克网盘检查器
		RegisterChecker(&QuarkChecker{})
		// 注册电信云盘检查器
		RegisterChecker(&TelecomChecker{})
		// 注册百度网盘检查器
		RegisterChecker(&BaiduChecker{})
		// 注册阿里云盘检查器
		RegisterChecker(&AliPanChecker{})
		// 注册115网盘检查器
		RegisterChecker(&YywChecker{})
		// 注册123网盘检查器
		RegisterChecker(&YesChecker{})
		// 注册UC网盘检查器
		RegisterChecker(&UcChecker{})

		if utils.IsDesktop() {
			// 注册迅雷网盘检查器
			RegisterChecker(&XunleiChecker{})
			// 注册移动云盘(139云盘)检查器
			RegisterChecker(&YdChecker{})
		}
	})
}
