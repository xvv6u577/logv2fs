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

// clear2Cmd represents the clear2 command
var clear2Cmd = &cobra.Command{
	Use:   "clear2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		cursor, err := userTrafficLogs.Find(context.TODO(), bson.M{})
		if err != nil {
			log.Printf("error: %v", err)
			return
		}

		for cursor.Next(context.Background()) {
			var user UserTrafficLogs
			err = cursor.Decode(&user)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}

			if len(user.YearlyLogs) == 0 {
				_, err = userTrafficLogs.UpdateOne(context.TODO(), bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"yearly_logs": []struct {
					Year    string `json:"year" bson:"year"`
					Traffic int64  `json:"traffic" bson:"traffic"`
				}{}}})
				if err != nil {
					log.Printf("error: %v", err)
					return
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(clear2Cmd)
}
