/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"go.mongodb.org/mongo-driver/bson"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "dev command",
	Long:  `dev command`,
	Run: func(cmd *cobra.Command, args []string) {

		userTrafficLogsCol := database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")

		// 使用聚合管道更新每个用户的 used 字段
		// 新值为 daily_logs 数组中所有 traffic 的和
		pipeline := []bson.M{
			{
				"$set": bson.M{
					"used": bson.M{
						"$sum": "$daily_logs.traffic",
					},
				},
			},
		}

		// 执行更新操作
		result, err := userTrafficLogsCol.UpdateMany(context.Background(), bson.M{}, pipeline)
		if err != nil {
			fmt.Printf("更新失败: %v\n", err)
			return
		}

		fmt.Printf("成功更新了 %d 个用户的 used 字段\n", result.ModifiedCount)
	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	devCmd.PersistentFlags().String("foo", "", "A help for foo")

	devCmd.Flags().BoolP("toggle", "", false, "Help message for toggle")
}
