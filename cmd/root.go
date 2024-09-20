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
)

var (
	CURRENT_DOMAIN                         = os.Getenv("CURRENT_DOMAIN")
	SERVER_ADDRESS                         = os.Getenv("SERVER_ADDRESS")
	SERVER_PORT                            = os.Getenv("SERVER_PORT")
	GIN_MODE                               = os.Getenv("GIN_MODE")
	trafficCollection    *mongo.Collection = database.OpenCollection(database.Client, "TRAFFIC")
	nodesCollection      *mongo.Collection = database.OpenCollection(database.Client, "NODES")
	userCollection       *mongo.Collection = database.OpenCollection(database.Client, "USERS")
	globalCollection     *mongo.Collection = database.OpenCollection(database.Client, "GLOBAL")
	nodeTrafficLogs                        = database.OpenCollection(database.Client, "NODE_TRAFFIC_LOGS")
	userTrafficLogs                        = database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")
	MoniteringDomainsCol *mongo.Collection = database.OpenCollection(database.Client, "Monitering_Domains")
	address              string
	tlsStatus            bool
	authrRequired        bool
	cronInstance         *cron.Cron
)

type (
	CurrentNode     = model.CurrentNode
	TrafficInDB     = model.TrafficInDB
	NodeAtPeriod    = model.NodeAtPeriod
	GlobalVariable  = model.GlobalVariable
	User            = model.User
	Domain          = model.Domain
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
}
