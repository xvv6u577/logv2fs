/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "dev command",
	Long:  `dev command`,
	Run: func(cmd *cobra.Command, args []string) {

		update := bson.D{primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "updated_at", Value: time.Now().Unix()},
		}}}

		fmt.Println(update)

		update[0].Value = append(update[0].Value.(bson.D), primitive.E{Key: "used_by_current_day", Value: 0})

		fmt.Println(update)

	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	devCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	devCmd.Flags().BoolP("toggle", "", false, "Help message for toggle")
}
