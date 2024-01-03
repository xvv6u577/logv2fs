/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

// fixCmd represents the fix command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "fix traffic data error in nodesCollection",
	Long:  `fix traffic data error in nodesCollection,`,
	Run: func(cmd *cobra.Command, args []string) {

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// query userCollection, update credit field of every document to 10995116277760.
		userCollection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"credit": 10995116277760}})
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
