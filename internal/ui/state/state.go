package state

import (
	"fyne.io/fyne/v2"
	"time"
)

type AppState struct {
	//检测的文件路径
	FilePath string
	//检测的文件URI
	FileURI fyne.URI

	//服务端标准时间 毫秒
	StandardTime int64
}

func NewAppState() *AppState {
	return &AppState{
		FilePath:     "",
		StandardTime: time.Now().UnixMilli(),
	}
}
