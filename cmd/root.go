/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

var (
	CURRENT_DOMAIN = os.Getenv("CURRENT_DOMAIN")
	SERVER_ADDRESS = os.Getenv("SERVER_ADDRESS")
	SERVER_PORT    = os.Getenv("SERVER_PORT")
	GIN_MODE       = os.Getenv("GIN_MODE")
	// trafficCollection *mongo.Collection = database.OpenCollection(database.Client, "TRAFFIC")
	MoniteringDomainsCol *mongo.Collection

	// PostgreSQL数据库
	PostgresDB *gorm.DB

	cronInstance *cron.Cron
)

type (
	NodeAtPeriod    = model.NodeAtPeriod
	Domain          = model.Domain
	Traffic         = model.Traffic
	TrafficAtPeriod = model.TrafficAtPeriod
	UserTrafficLogs = model.UserTrafficLogs
	NodeTrafficLogs = model.NodeTrafficLogs

	// PostgreSQL模型
	DomainPG          = model.DomainPG
	UserTrafficLogsPG = model.UserTrafficLogsPG
	NodeTrafficLogsPG = model.NodeTrafficLogsPG
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
	// 根据环境变量决定使用哪个数据库
	if database.IsUsingPostgres() {
		// 使用PostgreSQL
		PostgresDB = database.GetPostgresDB()
	} else {
		// 使用MongoDB
		MoniteringDomainsCol = database.OpenCollection(database.Client, "Monitering_Domains")
	}
}
