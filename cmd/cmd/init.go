/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init GLOBAL, NODES collection",
	Long:  `init GLOBAL, NODES collection`,
	Run: func(cmd *cobra.Command, args []string) {

		var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		// if GLOBAL collection exists, delete it
		if err := globalCOLLECTIONS.Drop(ctx); err != nil {
			panic(err)
		}

		var global = GlobalVariable{
			Name: "GLOBAL",
			DomainList: map[string]string{
				"www.baidu.com": "www.baidu.com",
			},
			NodeGlobalList: map[string]string{
				"localhost": "localhost",
			},
		}

		var adminUser *User
		var filter = bson.D{
			{Key: "role", Value: "admin"},
		}
		var adminProjections = bson.D{
			{Key: "email", Value: 1},
			{Key: "node_global_list", Value: 1},
		}

		if err := userCollection.FindOne(ctx, filter, options.FindOne().SetProjection(adminProjections)).Decode(&adminUser); err != nil {
			panic(err)
		}

		for key, value := range adminUser.NodeGlobalList {
			global.NodeGlobalList[value] = key
		}

		_, err := globalCOLLECTIONS.InsertOne(ctx, global)
		if err != nil {
			panic(err)
		}

		// insert NODES collection for each node in global.NodeGlobalList
		for key, value := range global.NodeGlobalList {
			var node = CurrentNode{
				Status: "active",
				Domain: key,
				Remark: value,
				NodeAtCurrentYear: NodeAtPeriod{
					Period:              time.Now().Format("2006"),
					Amount:              0,
					UserTrafficAtPeriod: map[string]int64{},
				},
				NodeAtCurrentMonth: NodeAtPeriod{
					Period:              time.Now().Format("200601"),
					Amount:              0,
					UserTrafficAtPeriod: map[string]int64{},
				},
				NodeAtCurrentDay: NodeAtPeriod{
					Period:              time.Now().Format("20060102"),
					Amount:              0,
					UserTrafficAtPeriod: map[string]int64{},
				},
				NodeByYear:  []NodeAtPeriod{},
				NodeByMonth: []NodeAtPeriod{},
				NodeByDay:   []NodeAtPeriod{},
				CreatedAt:   time.Now().Local(),
				UpdatedAt:   time.Now().Local(),
			}
			_, err := nodesCollection.InsertOne(ctx, node)
			if err != nil {
				panic(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
