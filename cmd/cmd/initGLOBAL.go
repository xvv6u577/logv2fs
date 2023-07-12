/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// initGLOBALCmd represents the initGLOBAL command
var initGLOBALCmd = &cobra.Command{
	Use:   "initGLOBAL",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		// check if GLOBAL collection exists, if not, create one
		var filter = bson.D{{}}
		var projections = bson.D{{
			Key:   "name",
			Value: "GLOBAL",
		}}
		var options = options.FindOne().SetProjection(projections)
		var result = globalCOLLECTIONS.FindOne(ctx, filter, options)
		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				// GLOBAL collection not exists, create one
				var global = GlobalVariable{
					Name:       "GLOBAL",
					DomainList: map[string]string{},
				}
				_, err := globalCOLLECTIONS.InsertOne(ctx, global)
				if err != nil {
					panic(err)
				}
			} else {
				panic(result.Err())
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(initGLOBALCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initGLOBALCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initGLOBALCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
