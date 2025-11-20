package utils

import "runtime"

func IsDesktop() bool {
	return runtime.GOOS != "android"
}
