/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 去重统计信息
type DeduplicationStats struct {
	UsersProcessed    int
	TotalDuplicates   int
	DailyDuplicates   int
	MonthlyDuplicates int
	YearlyDuplicates  int
}

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "开发命令 - 用于数据维护和调试",
	Long:  `开发命令提供各种数据维护功能，包括用户流量日志去重等操作`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		confirm, _ := cmd.Flags().GetBool("confirm")

		// 参数验证
		if !dryRun && !confirm {
			fmt.Println("❌ 错误: 必须指定 --dry-run 或 --confirm 参数")
			fmt.Println("使用示例:")
			fmt.Println("  预览模式: go run main.go dev --dry-run")
			fmt.Println("  执行模式: go run main.go dev --confirm")
			return
		}

		if dryRun && confirm {
			fmt.Println("❌ 错误: --dry-run 和 --confirm 不能同时使用")
			return
		}

		// 执行去重操作
		if dryRun {
			fmt.Println("🔍 开始预览模式 - 分析重复数据...")
			performDeduplication(true)
		} else {
			fmt.Println("⚠️  警告: 即将执行实际删除操作！")
			fmt.Println("首先预览将要删除的数据:")
			fmt.Println(strings.Repeat("=", 50))

			// 先运行预览
			stats := performDeduplication(true)

			fmt.Println(strings.Repeat("=", 50))
			fmt.Printf("📊 预览结果汇总:\n")
			fmt.Printf("  - 待处理用户: %d\n", stats.UsersProcessed)
			fmt.Printf("  - 总重复记录: %d\n", stats.TotalDuplicates)
			fmt.Printf("  - 日志重复: %d\n", stats.DailyDuplicates)
			fmt.Printf("  - 月志重复: %d\n", stats.MonthlyDuplicates)
			fmt.Printf("  - 年志重复: %d\n", stats.YearlyDuplicates)
			fmt.Println(strings.Repeat("=", 50))

			if stats.TotalDuplicates == 0 {
				fmt.Println("✅ 没有发现重复数据，无需执行删除操作")
				return
			}

			// 确认操作
			if askForConfirmation() {
				fmt.Println("🚀 开始执行实际删除操作...")
				performDeduplication(false)
				fmt.Println("✅ 去重操作完成！")
			} else {
				fmt.Println("❌ 操作已取消")
			}
		}
	},
}

// 执行去重操作
func performDeduplication(dryRun bool) DeduplicationStats {
	stats := DeduplicationStats{}

	// 连接数据库
	userTrafficLogsCol := database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")

	// 查询所有用户数据
	cursor, err := userTrafficLogsCol.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return stats
	}
	defer cursor.Close(context.Background())

	// 遍历每个用户
	for cursor.Next(context.Background()) {
		var userTrafficLogs model.UserTrafficLogs
		if err := cursor.Decode(&userTrafficLogs); err != nil {
			fmt.Printf("❌ 解码用户数据失败: %v\n", err)
			continue
		}

		stats.UsersProcessed++

		// 处理该用户的重复数据
		userStats := processUserDeduplication(userTrafficLogs, dryRun, userTrafficLogsCol)

		// 累计统计
		stats.TotalDuplicates += userStats.TotalDuplicates
		stats.DailyDuplicates += userStats.DailyDuplicates
		stats.MonthlyDuplicates += userStats.MonthlyDuplicates
		stats.YearlyDuplicates += userStats.YearlyDuplicates
	}

	// 输出总体统计
	mode := "预览"
	if !dryRun {
		mode = "执行"
	}

	fmt.Printf("\n📈 %s模式统计结果:\n", mode)
	fmt.Printf("  - 处理用户数: %d\n", stats.UsersProcessed)
	fmt.Printf("  - 总删除记录: %d\n", stats.TotalDuplicates)
	fmt.Printf("  - 日志删除: %d\n", stats.DailyDuplicates)
	fmt.Printf("  - 月志删除: %d\n", stats.MonthlyDuplicates)
	fmt.Printf("  - 年志删除: %d\n", stats.YearlyDuplicates)

	return stats
}

// 处理单个用户的去重
func processUserDeduplication(user model.UserTrafficLogs, dryRun bool, collection *mongo.Collection) DeduplicationStats {
	stats := DeduplicationStats{}
	hasUpdates := false

	fmt.Printf("\n👤 处理用户: %s (ID: %s)\n", user.Email_As_Id, user.ID.Hex())

	// 处理 Daily Logs
	newDailyLogs, dailyRemoved := deduplicateDailyLogs(user.DailyLogs)
	if dailyRemoved > 0 {
		stats.DailyDuplicates = dailyRemoved
		stats.TotalDuplicates += dailyRemoved
		hasUpdates = true
		fmt.Printf("  📅 Daily Logs: 删除 %d 条重复记录\n", dailyRemoved)
	}

	// 处理 Monthly Logs
	newMonthlyLogs, monthlyRemoved := deduplicateMonthlyLogs(user.MonthlyLogs)
	if monthlyRemoved > 0 {
		stats.MonthlyDuplicates = monthlyRemoved
		stats.TotalDuplicates += monthlyRemoved
		hasUpdates = true
		fmt.Printf("  📊 Monthly Logs: 删除 %d 条重复记录\n", monthlyRemoved)
	}

	// 处理 Yearly Logs
	newYearlyLogs, yearlyRemoved := deduplicateYearlyLogs(user.YearlyLogs)
	if yearlyRemoved > 0 {
		stats.YearlyDuplicates = yearlyRemoved
		stats.TotalDuplicates += yearlyRemoved
		hasUpdates = true
		fmt.Printf("  📈 Yearly Logs: 删除 %d 条重复记录\n", yearlyRemoved)
	}

	// 如果没有重复数据
	if !hasUpdates {
		fmt.Printf("  ✅ 该用户无重复数据\n")
		return stats
	}

	// 如果不是 dry-run 模式，执行实际更新
	if !dryRun {
		updateFilter := bson.M{"_id": user.ID}
		updateData := bson.M{
			"$set": bson.M{
				"daily_logs":   newDailyLogs,
				"monthly_logs": newMonthlyLogs,
				"yearly_logs":  newYearlyLogs,
			},
		}

		_, err := collection.UpdateOne(context.Background(), updateFilter, updateData)
		if err != nil {
			fmt.Printf("  ❌ 更新失败: %v\n", err)
			return DeduplicationStats{} // 返回空统计，表示失败
		}
		fmt.Printf("  ✅ 数据库更新成功\n")
	}

	return stats
}

// 去重 Daily Logs
func deduplicateDailyLogs(logs []struct {
	Date    string `json:"date" bson:"date"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}) ([]struct {
	Date    string `json:"date" bson:"date"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}, int) {

	if len(logs) <= 1 {
		return logs, 0
	}

	// 按日期分组
	groups := make(map[string][]struct {
		Date    string `json:"date" bson:"date"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	})

	for _, log := range logs {
		groups[log.Date] = append(groups[log.Date], log)
	}

	var result []struct {
		Date    string `json:"date" bson:"date"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	}
	removedCount := 0

	// 处理每个分组
	for date, group := range groups {
		if len(group) == 1 {
			result = append(result, group[0])
		} else {
			// 找到流量最大的记录
			maxTraffic := group[0]
			for _, log := range group[1:] {
				if log.Traffic > maxTraffic.Traffic {
					maxTraffic = log
				}
			}
			result = append(result, maxTraffic)
			removedCount += len(group) - 1

			fmt.Printf("    📅 日期 %s: 保留最大流量 %d，删除 %d 条记录\n",
				date, maxTraffic.Traffic, len(group)-1)
		}
	}

	// 按日期排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, removedCount
}

// 去重 Monthly Logs
func deduplicateMonthlyLogs(logs []struct {
	Month   string `json:"month" bson:"month"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}) ([]struct {
	Month   string `json:"month" bson:"month"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}, int) {

	if len(logs) <= 1 {
		return logs, 0
	}

	// 按月份分组
	groups := make(map[string][]struct {
		Month   string `json:"month" bson:"month"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	})

	for _, log := range logs {
		groups[log.Month] = append(groups[log.Month], log)
	}

	var result []struct {
		Month   string `json:"month" bson:"month"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	}
	removedCount := 0

	// 处理每个分组
	for month, group := range groups {
		if len(group) == 1 {
			result = append(result, group[0])
		} else {
			// 找到流量最大的记录
			maxTraffic := group[0]
			for _, log := range group[1:] {
				if log.Traffic > maxTraffic.Traffic {
					maxTraffic = log
				}
			}
			result = append(result, maxTraffic)
			removedCount += len(group) - 1

			fmt.Printf("    📊 月份 %s: 保留最大流量 %d，删除 %d 条记录\n",
				month, maxTraffic.Traffic, len(group)-1)
		}
	}

	// 按月份排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Month < result[j].Month
	})

	return result, removedCount
}

// 去重 Yearly Logs
func deduplicateYearlyLogs(logs []struct {
	Year    string `json:"year" bson:"year"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}) ([]struct {
	Year    string `json:"year" bson:"year"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}, int) {

	if len(logs) <= 1 {
		return logs, 0
	}

	// 按年份分组
	groups := make(map[string][]struct {
		Year    string `json:"year" bson:"year"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	})

	for _, log := range logs {
		groups[log.Year] = append(groups[log.Year], log)
	}

	var result []struct {
		Year    string `json:"year" bson:"year"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	}
	removedCount := 0

	// 处理每个分组
	for year, group := range groups {
		if len(group) == 1 {
			result = append(result, group[0])
		} else {
			// 找到流量最大的记录
			maxTraffic := group[0]
			for _, log := range group[1:] {
				if log.Traffic > maxTraffic.Traffic {
					maxTraffic = log
				}
			}
			result = append(result, maxTraffic)
			removedCount += len(group) - 1

			fmt.Printf("    📈 年份 %s: 保留最大流量 %d，删除 %d 条记录\n",
				year, maxTraffic.Traffic, len(group)-1)
		}
	}

	// 按年份排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Year < result[j].Year
	})

	return result, removedCount
}

// 询问用户确认
func askForConfirmation() bool {
	fmt.Print("\n❓ 确认执行删除操作？(输入 'yes' 确认): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.TrimSpace(strings.ToLower(scanner.Text()))

	return response == "yes"
}

func init() {
	rootCmd.AddCommand(devCmd)

	// 添加命令行参数
	devCmd.Flags().Bool("dry-run", false, "预览模式 - 分析重复数据但不执行删除")
	devCmd.Flags().Bool("confirm", false, "执行模式 - 实际删除重复数据")

	// 设置参数说明
	devCmd.Long = `开发命令提供各种数据维护功能，包括用户流量日志去重等操作

使用示例:
  预览模式: go run main.go dev --dry-run
  执行模式: go run main.go dev --confirm

去重逻辑:
  - Daily Logs: 相同日期(YYYYMMDD)保留最大流量记录
  - Monthly Logs: 相同月份(YYYYMM)保留最大流量记录  
  - Yearly Logs: 相同年份(YYYY)保留最大流量记录`
}
