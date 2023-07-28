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

// fixCmd represents the fix command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "fix traffic data error in nodesCollection",
	Long:  `fix traffic data error in nodesCollection,`,
	Run: func(cmd *cobra.Command, args []string) {

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		var nodeArray []*CurrentNode
		var nodeFilter = bson.D{{}}
		var nodeProjections = bson.D{}
		cursor, err := nodesCollection.Find(ctx, nodeFilter, options.Find().SetProjection(nodeProjections))
		if err != nil {
			panic(err)
		}
		if err = cursor.All(ctx, &nodeArray); err != nil {
			panic(err)
		}

		for _, node := range nodeArray {

			node.NodeAtCurrentDay.Amount = 0
			node.NodeAtCurrentMonth.Amount = 0
			node.NodeAtCurrentYear.Amount = 0

			for _, v := range node.NodeAtCurrentDay.UserTrafficAtPeriod {
				node.NodeAtCurrentDay.Amount += v
			}

			for _, v := range node.NodeAtCurrentMonth.UserTrafficAtPeriod {
				node.NodeAtCurrentMonth.Amount += v
			}

			for _, v := range node.NodeAtCurrentYear.UserTrafficAtPeriod {
				node.NodeAtCurrentYear.Amount += v
			}

			for i := range node.NodeByDay {
				node.NodeByDay[i].Amount = 0
				for _, v := range node.NodeByDay[i].UserTrafficAtPeriod {
					node.NodeByDay[i].Amount += v
				}
			}

			for i := range node.NodeByMonth {
				node.NodeByMonth[i].Amount = 0
				for _, v := range node.NodeByMonth[i].UserTrafficAtPeriod {
					node.NodeByMonth[i].Amount += v
				}
			}

			for i := range node.NodeByYear {
				node.NodeByYear[i].Amount = 0
				for _, v := range node.NodeByYear[i].UserTrafficAtPeriod {
					node.NodeByYear[i].Amount += v
				}
			}

			var filter = bson.D{{Key: "domain", Value: node.Domain}}
			var update = bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "node_at_current_year", Value: node.NodeAtCurrentYear},
					{Key: "node_at_current_month", Value: node.NodeAtCurrentMonth},
					{Key: "node_at_current_day", Value: node.NodeAtCurrentDay},
					{Key: "node_by_year", Value: node.NodeByYear},
					{Key: "node_by_month", Value: node.NodeByMonth},
					{Key: "node_by_day", Value: node.NodeByDay},
				}},
			}
			_, err := nodesCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				panic(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(fixCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fixCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fixCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
