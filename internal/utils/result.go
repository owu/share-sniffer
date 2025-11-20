package utils

type ErrorType uint32

const MsgMaxLen int = 48

// Result 检测结果结构体
// 包含URL检查的完整信息
//
// 字段:
// - Error: 错误码
// - Msg: 错误信息
// - Data.URL: 被检测的URL字符串
// - Data.Name: 资源名称（如果检测成功）
// - Data.Elapsed: 检测耗时（毫秒）

type Result struct {
	Error ErrorType  `json:"error"` // 错误码
	Msg   string     `json:"msg"`
	Data  ResultData `json:"data"`
}

type ResultData struct {
	URL     string `json:"url"`     // 检测的URL
	Name    string `json:"name"`    // 资源名称
	Elapsed int64  `json:"elapsed"` // 耗时（毫秒）
}

const (
	//Valid 没有错误的，即链接有效
	Valid ErrorType = iota //0

	//Unknown 未知错误
	Unknown = 10

	//Invalid 失效的 链接过期的
	Invalid = 11

	//Malformed 错误的 参数错误等
	Malformed = 12

	//Timeout 超时的
	Timeout = 13

	// Fatal 请求过程中报错
	Fatal = 14

	// Stop 停止 (任务池)
	Stop = 15

	// Done 完成 (任务池)
	Done = 16
)

const (
	//ValidTxt 没有错误的，即链接有效
	ValidTxt = "有效" //0

	//UnknownTxt 未知错误
	UnknownTxt = "未知"

	//InvalidTxt 失效的 链接过期的
	InvalidTxt = "失效"

	//MalformedTxt 错误的 参数错误等
	MalformedTxt = "错误"

	// TimeoutTxt 超时的
	TimeoutTxt = "超时"

	// FatalTxt 请求过程中报错
	FatalTxt = "异常"

	// StopTxt  GUI
	StopTxt = "已停止"

	// DoingTxt  GUI
	DoingTxt = "检测中"
)

func ErrorToMsg(error ErrorType) string {
	msg := ""
	switch error {
	case Valid:
		msg = "valid"
	case Unknown:
		msg = "unknown"
	case Invalid:
		msg = "invalid"
	case Malformed:
		msg = "malformed"
	case Timeout:
		msg = "timeout"
	default:
		msg = "self defined"
	}
	return msg
}

// ErrorMalformed 参数错误
func ErrorMalformed(url string, msg string) Result {
	return Result{
		Error: Malformed,
		Msg: func() string {
			if msg == "" {
				return ErrorToMsg(Malformed)
			}
			return Substr(msg, MsgMaxLen, "")
		}(),
		Data: ResultData{
			URL:     url,
			Name:    "",
			Elapsed: 0,
		},
	}
}

func ErrorTimeout() Result {
	return Result{
		Error: Timeout,
		Msg:   ErrorToMsg(Timeout),
		Data: ResultData{
			URL:     "",
			Name:    "",
			Elapsed: 0,
		},
	}
}

func ErrorUnknown(msg string) Result {
	return Result{
		Error: Unknown,
		Msg: func() string {
			if msg == "" {
				return ErrorToMsg(Unknown)
			}
			return Substr(msg, MsgMaxLen, "")
		}(),
		Data: ResultData{
			URL:     "",
			Name:    "",
			Elapsed: 0,
		},
	}
}

func ErrorValid(name string) Result {
	return Result{
		Error: Valid,
		Msg:   ErrorToMsg(Valid),
		Data: ResultData{
			URL:     "",
			Name:    name,
			Elapsed: 0,
		},
	}
}

func ErrorInvalid(msg string) Result {
	return Result{
		Error: Invalid,
		Msg: func() string {
			if msg == "" {
				return ErrorToMsg(Invalid)
			}
			return Substr(msg, MsgMaxLen, "")
		}(),
		Data: ResultData{
			URL:     "",
			Name:    "",
			Elapsed: 0,
		},
	}
}

func ErrorFatal(msg string) Result {
	return Result{
		Error: Fatal,
		Msg: func() string {
			if msg == "" {
				return ErrorToMsg(Unknown)
			}
			return Substr(msg, MsgMaxLen, "")
		}(),
		Data: ResultData{
			URL:     "",
			Name:    "",
			Elapsed: 0,
		},
	}
}
