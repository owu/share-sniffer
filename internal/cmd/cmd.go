package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"share-sniffer/internal/config"
	"share-sniffer/internal/core"
	"share-sniffer/internal/logger"
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

			response := core.Adapter(context.Background(), url)

			// 输出JSON结果
			//jsonBytes, _ := json.MarshalIndent(response, "", "  ")
			jsonBytes, _ := json.Marshal(response)
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
			fmt.Println(config.HomePage())
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
	// 设置日志级别为Fatal，这样只有致命错误会被记录（但会导致程序退出）
	// 这样CLI模式下不会输出任何多余日志，只返回JSON结果
	logger.SetLogLevel(logger.LevelFatal + 1)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
