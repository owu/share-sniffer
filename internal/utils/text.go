package utils

import (
	"strings"
	"unicode/utf8"
)

// Substr 字符串截取
// 参数:
// - str: 带截断的字符串
// - length: 保留的字符串长度
// - prefix: 要替换为空的字符串，例如 https
//
// 返回值:
// - string: 截断后的字符串
func Substr(str string, length int, prefix string) string {
	if 0 < len(prefix) {
		str = strings.Replace(str, prefix, "", 1)
	}

	// 统计实际字符数
	charCount := 0
	endIndex := 0
	for i := range str {
		if charCount >= length {
			break
		}
		charCount++
		endIndex = i + utf8.RuneLen([]rune(str)[charCount-1])
	}

	if charCount < length {
		return str
	}

	return str[:endIndex] + "..."
}
