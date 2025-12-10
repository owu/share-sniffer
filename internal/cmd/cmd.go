package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/core"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "share-sniffer-cli [URL]",
		Short: "Share Sniffer CLI - A tool to detect and analyze shared links",
		Long:  `Share Sniffer CLI is a command-line tool that helps you detect and analyze shared links from various platforms.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// 如果没有提供参数，显示帮助信息
			if len(args) == 0 {
				cmd.Help()
				return
			}

			// 直接传入URL进行检测
			url := args[0]
			if !strings.Contains(url, "https") || len(url) <= 20 {
				cmd.Help()
				return
			}

			result := core.Adapter(url)

			// 构建JSON响应
			response := &CResult{
				Error: getStatusError(result.Status),
				Msg:   getStatusMessage(result.Status),
				Data: CData{
					URL:     result.URL,
					Name:    result.Name,
					Elapsed: result.Elapsed,
				},
			}

			// 输出JSON结果
			jsonBytes, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(jsonBytes))
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Show the version information for Share Sniffer CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Version())
		},
	}

	supportCmd = &cobra.Command{
		Use:   "support",
		Short: "Show supported link types",
		Long:  `Show all supported link types.`,
		Run: func(cmd *cobra.Command, args []string) {
			supportedLinks := config.GetSupportedLinks()
			for _, link := range supportedLinks {
				fmt.Println(link)
			}
		},
	}

	homeCmd = &cobra.Command{
		Use:   "home",
		Short: "Show project homepage",
		Long:  `Show the project homepage URL.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("https://github.com/owu/share-sniffer")
		},
	}
)

// init 初始化命令行
func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(supportCmd)
	rootCmd.AddCommand(homeCmd)
}

// Execute 执行命令行
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// getStatusMessage 根据状态码获取消息
func getStatusMessage(status int) string {
	switch status {
	case 1:
		return "success"
	case -1:
		return "timeout"
	default:
		return "failed"
	}
}

// getStatusError 根据状态码获取响应
func getStatusError(status int) int {
	switch status {
	case 1:
		return 0
	default:
		return status + 10
	}
}

type CResult struct {
	Error int    `json:"error"` //成功0 ，失败 100 +
	Msg   string `json:"msg"`
	Data  CData  `json:"data"`
}

type CData struct {
	URL     string `json:"url"`     // 检测的URL
	Name    string `json:"name"`    // 资源名称
	Elapsed int64  `json:"elapsed"` // 耗时（毫秒）
}
