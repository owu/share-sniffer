package core

import (
	"testing"
)

func TestAdapter(t *testing.T) {
	var urls []string

	// urls = append(urls, testQuark()...)
	urls = append(urls, testTelecom()...)
	// urls = append(urls, testBaidu()...)
	// urls = append(urls, testAliPan()...)
	// urls = append(urls, testYyw()...)

	for _, url := range urls {
		result := Adapter(url)
		t.Logf("  网址: %s\n", result.URL)
		t.Logf("  文件名: %s\n", result.Name)
		t.Logf("  状态: %d\n", result.Status)
		t.Logf("  耗时: %d 毫秒\n", result.Elapsed)
		t.Log("  ------------------------------------\n")
	}
}

// 夸克网盘测试用例
func testQuark() []string {
	urls := []string{
		"https://pan.quark.cn/s/0592e1dbe475",
		"https://pan.quark.cn/s/45c6cd59a7f9?pwd=D3eM",
		"https://pan.quark.cn/s/058cedce8adb?pwd=y5T3",
		"https://pan.quark.cn/s/8815ecf5298b",
		"https://pan.quark.cn/s/8c156054435f",
	}
	return urls
}

// 天翼云盘测试用例
func testTelecom() []string {
	urls := []string{
		"https://cloud.189.cn/web/share?code=bm2iuqZZj632%EF%BC%88%E8%AE%BF%E9%97%AE%E7%A0%81%EF%BC%9Atts9%EF%BC%89",
		"https://cloud.189.cn/web/share?code=eMJZVvUnUbaq",
		"https://cloud.189.cn/t/6FjeIfQvMRba%EF%BC%88%E8%AE%BF%E9%97%AE%E7%A0%81%EF%BC%9A2jio%EF%BC%89",
		"https://cloud.189.cn/t/bm2iuqZZj632%EF%BC%88%E8%AE%BF%E9%97%AE%E7%A0%81%EF%BC%9Atts9%EF%BC%89",
		"https://cloud.189.cn/t/6FjeIfQvMRba（访问码：2jio）",
	}
	return urls
}

// 百度网盘测试用例
func testBaidu() []string {
	urls := []string{
		"https://pan.baidu.com/s/1oq-lgc1uqwwuCAA3PmVRWQ?pwd=MCPH",
		"https://pan.baidu.com/s/1wj6Y-RquDLEUUTLHTWnjAQ?pwd=3wi7",
		"https://pan.baidu.com/s/1zumS333182ampFtimza0FA?pwd=km6y",
		"https://pan.baidu.com/s/1mEH70fRDUSdPeb7f9pNc0w?pwd=e69e",
		"https://pan.baidu.com/s/1uKkS4VV0n2fZoZBAT7xGzA?pwd=86vg",
	}
	return urls
}

// 阿里云盘测试用例
func testAliPan() []string {
	urls := []string{
		"https://www.alipan.com/s/hGz3eqGzXH3x",
		"https://www.alipan.com/s/Aujd5Vxr4i2",
		"https://www.alipan.com/s/Xd4HxfpMdVk",
		"https://www.alipan.com/s/Xd4HxfpMdVkss",
		"https://www.alipan.com/s/dSpD9NqhgN4",
	}
	return urls
}

// 115网盘测试用例
func testYyw() []string {
	urls := []string{
		"https://115cdn.com/s/swwc9o33zh5?password=8848#",
		"https://115cdn.com/s/sww1kjp3zv8?password=xae0&#",
		"https://115cdn.com/s/swwcv883zv8?password=e402&#",
		"https://115cdn.com/s/swwcvss3ffc?password=q685#",
		"https://115cdn.com/s/sww0nyf3zv8?password=n865&#",
	}
	return urls
}
