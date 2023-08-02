/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	trafficCollection *mongo.Collection = database.OpenCollection(database.Client, "TRAFFIC")
	nodesCollection   *mongo.Collection = database.OpenCollection(database.Client, "NODES")
	userCollection    *mongo.Collection = database.OpenCollection(database.Client, "USERS")
	globalCOLLECTIONS *mongo.Collection = database.OpenCollection(database.Client, "GLOBAL")
)

type (
	CurrentNode    = model.CurrentNode
	TrafficInDB    = model.TrafficInDB
	NodeAtPeriod   = model.NodeAtPeriod
	GlobalVariable = model.GlobalVariable
	User           = model.User
	Domain         = model.Domain
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cmd.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
