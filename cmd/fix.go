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

		// query nodesCollection, upinsert ip string field to every document.
		nodesCollection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"ip": ""}}, options.Update().SetUpsert(true))

		// query globalCollection with "GLOBAL" name, upinsert ip string field to every object in work_related_domain_list array and active_global_nodes array.
		globalCollection.UpdateMany(ctx, bson.M{"name": "GLOBAL"}, bson.M{"$set": bson.M{"work_related_domain_list.$[].ip": ""}}, options.Update().SetUpsert(true))
		globalCollection.UpdateMany(ctx, bson.M{"name": "GLOBAL"}, bson.M{"$set": bson.M{"active_global_nodes.$[].ip": ""}}, options.Update().SetUpsert(true))

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
