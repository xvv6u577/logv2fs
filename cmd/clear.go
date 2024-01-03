/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/xvv6u577/logv2fs/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear all collections named by email in database",
	Run: func(cmd *cobra.Command, args []string) {
		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var projections = bson.D{
			{Key: "email", Value: 1},
		}
		users, err := database.GetAllUsersPartialInfo(projections)
		if err != nil {
			panic(err)
		}

		for _, user := range users {
			var myTraffic *mongo.Collection = database.OpenCollection(database.Client, user.Email)

			// drop the collection
			if err = myTraffic.Drop(ctx); err != nil {
				panic(err)
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clearCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clearCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
