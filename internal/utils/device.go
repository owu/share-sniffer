package utils

import (
	"os"
	"runtime"
)

func IsDesktop() bool {
	//docker build,  0 表示屏蔽依赖chromedp的类型
	shield := os.Getenv("IS_DESKTOP")
	if shield == "0" {
		return false
	}

	return runtime.GOOS != "android"
}
