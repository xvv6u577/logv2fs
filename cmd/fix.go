/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

// fixCmd represents the fix command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "tools - fix the data in the database",
	Long:  `tools - fix the data in the database`,
	Run: func(cmd *cobra.Command, args []string) {

		filter := bson.D{}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "path", Value: "ray"}}}}
		_, err := userCollection.UpdateMany(context.Background(), filter, update)
		if err != nil {
			log.Fatal(err)
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
