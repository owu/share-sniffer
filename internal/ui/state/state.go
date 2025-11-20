package state

import "time"

type AppState struct {
	//检测的文件
	FilePath string

	//服务端标准时间 毫秒
	StandardTime int64
}

func NewAppState() *AppState {
	return &AppState{
		FilePath:     "",
		StandardTime: time.Now().UnixMilli(),
	}
}
