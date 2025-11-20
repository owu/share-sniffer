package core

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestAdapter(t *testing.T) {

	var urls []string

	// 夸克网盘测试用例
	urls = append(urls, readTestUrls("quark.txt")...)

	// 天翼云盘测试用例
	urls = append(urls, readTestUrls("telecom.txt")...)

	// 百度网盘测试用例
	urls = append(urls, readTestUrls("baidu.txt")...)

	// 阿里云盘测试用例
	urls = append(urls, readTestUrls("alipan.txt")...)

	// 115网盘测试用例
	urls = append(urls, readTestUrls("yyw.txt")...)

	// 123网盘测试用例
	urls = append(urls, readTestUrls("yes.txt")...)

	// UC网盘测试用例
	urls = append(urls, readTestUrls("uc.txt")...)

	// 迅雷网盘测试用例
	urls = append(urls, readTestUrls("xunlei.txt")...)

	// 移动云盘测试用例
	urls = append(urls, readTestUrls("yd.txt")...)

	for _, url := range urls {
		result := Adapter(context.Background(), url)
		t.Logf("  错误码: %d\n", result.Error)
		t.Logf("  信息: %s\n", result.Msg)
		t.Logf("  网址: %s\n", result.Data.URL)
		t.Logf("  文件名: %s\n", result.Data.Name)
		t.Logf("  耗时: %d 毫秒\n", result.Data.Elapsed)
		t.Log("   ------------------------------------\n")
	}
}

// readTestUrls reads URLs from a test case file
func readTestUrls(filename string) []string {
	// Construct the full path to the test case file
	// First try from project root (where tests are usually run from)
	currentDir, _ := os.Getwd()

	// List of possible paths to try
	paths := []string{
		// Path from project root
		filepath.Join(currentDir, "build", "testcases", filename),
		// Path from internal/core directory
		filepath.Join(currentDir, "..", "..", "build", "testcases", filename),
		// Path from internal directory
		filepath.Join(currentDir, "..", "build", "testcases", filename),
	}

	var filePath string
	found := false

	// Try all paths until we find the file
	for _, p := range paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			filePath = p
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Error: Could not find test case file %s in any of the tried paths. Current dir: %s\n", filename, currentDir)
		return []string{}
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening test case file %s: %v\n", filePath, err)
		return []string{}
	}
	defer file.Close()

	var urls []string
	// Read lines from the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		if url != "" { // Skip empty lines
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading test case file %s: %v\n", filePath, err)
		return []string{}
	}

	return urls
}
