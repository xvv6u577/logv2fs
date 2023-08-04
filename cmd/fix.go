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

		// query userCollection by email in emailArray, get used_by_current_day, used_by_current_month, used_by_current_year, then print them
		var emailArray = []string{
			"wangju",
			"m6v377",
			"2ym8g926p",
		}

		var projection = bson.M{
			"used_by_current_day":   1,
			"used_by_current_month": 1,
			"used_by_current_year":  1,
			"name":                  1,
			"email":                 1,
		}

		var findOptions = options.Find()
		findOptions.SetProjection(projection)

		var filter = bson.M{
			"email": bson.M{
				"$in": emailArray,
			},
		}

		var users []User
		cusor, err := userCollection.Find(ctx, filter, findOptions)
		if err != nil {
			panic(err)
		}

		if err = cusor.All(ctx, &users); err != nil {
			panic(err)
		}

		for _, user := range users {

			var update = bson.M{
				"$set": bson.M{
					"used_by_current_day": TrafficAtPeriod{
						Period:       user.UsedByCurrentDay.Period,
						Amount:       user.UsedByCurrentDay.Amount,
						UsedByDomain: map[string]int64{},
					},
					"used_by_current_month": TrafficAtPeriod{
						Period:       user.UsedByCurrentMonth.Period,
						Amount:       user.UsedByCurrentMonth.Amount,
						UsedByDomain: map[string]int64{},
					},
					"used_by_current_year": TrafficAtPeriod{
						Period:       user.UsedByCurrentYear.Period,
						Amount:       user.UsedByCurrentYear.Amount,
						UsedByDomain: map[string]int64{},
					},
				},
			}

			var filter = bson.M{
				"email": user.Email,
			}

			var updateOptions = options.Update()
			updateOptions.SetUpsert(true)

			var result, err = userCollection.UpdateOne(ctx, filter, update, updateOptions)
			if err != nil {
				panic(err)
			}

			if result.MatchedCount != 0 {
				println("update user: ", user.Email, " successfully")
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
