package about

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/logger"
)

// StaticConfigResponse 远程配置文件响应结构体
type StaticConfigResponse struct {
	Latest string `json:"latest"`
}

// staticConfig 查询远程配置文件，包含超时和重试机制
func staticConfig() (*StaticConfigResponse, error) {
	url := fmt.Sprintf("%s?t=%d", config.StaticApi(), time.Now().Unix())

	// 设置HTTP客户端超时和重试机制
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 重试机制：最多重试3次
	var resp *http.Response
	var err error

	for i := 0; i < 3; i++ {
		resp, err = client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if i < 2 { // 不是最后一次重试
			time.Sleep(2 * time.Second) // 等待2秒后重试
		}
	}

	if err != nil {
		return nil, fmt.Errorf("请求远程配置失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("远程配置请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 解析JSON
	var config StaticConfigResponse
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &config, nil
}

// compareVersion 比较版本号，返回是否有新版本
func compareVersion(currentVersion, latestVersion string) (bool, error) {
	// 将版本号字符串转换为数字进行比较
	currentNum, err := versionToNumber(currentVersion)
	if err != nil {
		return false, err
	}

	latestNum, err := versionToNumber(latestVersion)
	if err != nil {
		return false, err
	}

	return currentNum < latestNum, nil
}

// versionToNumber 将版本号字符串转换为数字
func versionToNumber(version string) (int, error) {
	// 移除可能的v前缀
	version = strings.TrimPrefix(version, "v")

	// 分割版本号
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("版本号格式错误: %s", version)
	}

	// 解析主版本号
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("主版本号解析失败: %v", err)
	}

	// 解析次版本号
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("次版本号解析失败: %v", err)
	}

	// 解析修订版本号
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, fmt.Errorf("修订版本号解析失败: %v", err)
	}

	// 格式化版本号为数字：主版本号3位，次版本号2位，修订版本号2位
	versionNum := major*10000 + minor*100 + patch

	return versionNum, nil
}

// CheckUpdate 检查是否有新版本，并在UI线程中显示弹窗提示
func CheckUpdate(window fyne.Window, clicked bool) {
	// 在协程中执行检查，避免阻塞主界面
	go func() {
		// 查询远程配置
		remoteConfig, err := staticConfig()
		if err != nil {
			// 处理网络错误，在UI线程中显示错误提示
			fyne.Do(func() {
				logger.Error("检查更新失败: %v", err)

				// 根据错误类型显示不同的提示信息
				errorMsg := "检查更新失败，请检查网络连接或稍后重试"
				if strings.Contains(err.Error(), "超时") || strings.Contains(err.Error(), "timeout") {
					errorMsg = "网络连接超时，请检查网络连接后重试"
				} else if strings.Contains(err.Error(), "解析JSON") {
					errorMsg = "服务器响应格式错误，请稍后重试"
				}

				logger.Error("检查更新失败: %v", errorMsg)
				//dialog.ShowError(fmt.Errorf("%s", errorMsg), window)
			})
			return
		}

		// 读取当前版本号
		currentVersion := config.Version()
		
		// 比对版本
		hasUpdate, err := compareVersion(currentVersion, remoteConfig.Latest)
		if err != nil {
			// 处理版本比较错误
			fyne.Do(func() {
				logger.Error("版本比较失败: %v", err)
				//dialog.ShowError(fmt.Errorf("版本比较失败: %v", err), window)
			})
			return
		}

		// 在UI线程中显示弹窗提示
		fyne.Do(func() {
			if hasUpdate {
				// 有新版本，显示自定义更新提示对话框
				showUpdateDialog(window, remoteConfig.Latest, currentVersion)
				logger.Info("发现新版本: %s (当前版本: %s)", remoteConfig.Latest, currentVersion)
			} else {
				// 没有新版本，显示最新版本提示
				if clicked {
					dialog.ShowInformation("版本检查", fmt.Sprintf("当前已是最新版本: v%s", config.Version()), window)
				}
				logger.Info("当前已是最新版本: %s", currentVersion)
			}
		})
	}()
}

// showUpdateDialog 显示包含GitHub发布页超链接的自定义更新对话框
func showUpdateDialog(window fyne.Window, latestVersion string, currentVersion string) {
	// 创建对话框内容
	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("发现新版本: v%s , 当前版本: v%s", latestVersion, currentVersion)),
		widget.NewLabel("\n请前往GitHub下载最新版本:"),
	)

	releases := fmt.Sprintf("%s/releases", config.HomePage())

	// 创建GitHub发布页超链接
	githubURL, _ := url.Parse(releases)
	githubLink := widget.NewHyperlink(releases, githubURL)

	// 设置超链接点击事件
	githubLink.OnTapped = func() {
		if githubLink.URL != nil {
			// 尝试打开GitHub发布页
			fyne.CurrentApp().OpenURL(githubLink.URL)
		}
	}

	// 将超链接添加到内容中
	content.Add(githubLink)

	// 创建自定义对话框
	customDialog := dialog.NewCustom(
		"检查更新",
		"确定",
		content,
		window,
	)

	// 设置对话框大小
	customDialog.Resize(fyne.NewSize(400, 150))

	// 显示对话框
	customDialog.Show()
}
