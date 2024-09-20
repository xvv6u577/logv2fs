/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"

	"go.mongodb.org/mongo-driver/bson"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear all collections named by email in database",
	Run: func(cmd *cobra.Command, args []string) {

		// drop user collection
		err := userCollection.Drop(context.TODO())
		if err != nil {
			log.Printf("error dropping user collection: %v\n", err)
		}

		// delete node collection
		err = nodesCollection.Drop(context.TODO())
		if err != nil {
			log.Printf("error dropping node collection: %v\n", err)
		}

		// delete "name":"GLOBAL" document in global collection
		_, err = globalCollection.DeleteOne(context.TODO(), bson.M{"name": "GLOBAL"})
		if err != nil {
			log.Printf("error dropping global collection: %v\n", err)
		}

		// loop through traffic collection, delete created_at date older than 3 months
		threeMonthsAgo := time.Now().AddDate(0, -3, 0)
		filter := bson.M{"created_at": bson.M{"$lt": threeMonthsAgo}}

		result, err := trafficCollection.DeleteMany(context.TODO(), filter)
		if err != nil {
			log.Fatalf("Failed to delete documents: %v", err)
		}
		log.Printf("Deleted %v documents in the traffic collection", result.DeletedCount)

	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
