package utils

import (
	"strings"
)

func Substr(str string, length int) string {
	var builder strings.Builder
	if length >= len(str) {
		builder.WriteString(str[:])
	} else {
		builder.WriteString(str[:length])
		builder.WriteString("...") // 添加省略号
	}
	result := builder.String() // 获取最终结果
	return result
}
