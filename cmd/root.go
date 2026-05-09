/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	dropdailyallocations "github.com/xvv6u577/logv2fs/cmd/dropdailyallocations"
	httpserver "github.com/xvv6u577/logv2fs/cmd/httpserver"
	migratetrafficlogs "github.com/xvv6u577/logv2fs/cmd/migratetrafficlogs"
	singbox "github.com/xvv6u577/logv2fs/cmd/singbox"
	"github.com/xvv6u577/logv2fs/model"
)

type (
	NodeAtPeriod    = model.NodeAtPeriod
	Domain          = model.SubscriptionNode
	Traffic         = model.Traffic
	TrafficAtPeriod = model.TrafficAtPeriod
	UserTrafficLogs = model.UserTrafficLogs
	NodeTrafficLogs = model.NodeTrafficLogs
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cmd",
	Short: "root command",
	Long:  `root command, which is the entry of this program.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 在任何子命令执行之前，先加载 .env 文件
	// 这样确保所有环境变量都能被正确读取
	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v", err)
	} else {
		// 尝试加载 .env 文件，如果文件不存在也不会报错
		if err := godotenv.Load(pwd + "/.env"); err != nil {
			log.Printf("提示: 未找到 .env 文件或加载失败: %v", err)
		}
	}

	// 添加子命令
	rootCmd.AddCommand(singbox.NewSingboxCmd())
	rootCmd.AddCommand(httpserver.NewHTTPServerCmd())
	rootCmd.AddCommand(migratetrafficlogs.NewMigrateTrafficLogsCmd())
	rootCmd.AddCommand(dropdailyallocations.NewDropDailyAllocationsCmd())

}
