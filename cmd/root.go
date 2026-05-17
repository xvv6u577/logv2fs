/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	httpserver "github.com/xvv6u577/logv2fs/cmd/httpserver"
	migratenodes "github.com/xvv6u577/logv2fs/cmd/migratenodes"
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

var envFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cmd",
	Short: "root command",
	Long:  `root command, which is the entry of this program.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		loadEnvFile()
	},
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
	rootCmd.PersistentFlags().StringVarP(&envFile, "env", "e", "", "path to .env file; defaults to .env in the current directory")

	// 添加子命令
	rootCmd.AddCommand(singbox.NewSingboxCmd())
	rootCmd.AddCommand(httpserver.NewHTTPServerCmd())
	rootCmd.AddCommand(migratenodes.NewMigrateNodesCmd())

}

func loadEnvFile() {
	envPath := envFile
	if envPath == "" {
		envPath = defaultEnvFile()
	}

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("提示: 未找到 .env 文件或加载失败 path=%s err=%v", envPath, err)
	}
}

func defaultEnvFile() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v", err)
		return ".env"
	}
	return filepath.Join(pwd, ".env")
}
