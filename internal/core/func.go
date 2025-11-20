package core

// Result 检测结果结构体
// 包含URL检查的完整信息
//
// 字段:
// - URL: 被检测的URL字符串
// - Name: 资源名称（如果检测成功）
// - Status: 状态码（1表示正常，0表示失败，-1表示超时）
// - Elapsed: 检测耗时（毫秒）

type Result struct {
	URL     string // 检测的URL
	Name    string // 资源名称
	Status  int    // 状态码: 1表示正常，0表示失败，-1表示超时
	Elapsed int64  // 耗时（毫秒）
}
