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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init GLOBAL, NODES collection",
	Long:  `init GLOBAL, NODES collection`,
	Run: func(cmd *cobra.Command, args []string) {

		var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		// if GLOBAL collection exists, delete it
		if err := globalCOLLECTIONS.Drop(ctx); err != nil {
			panic(err)
		}

		var global = GlobalVariable{
			Name: "GLOBAL",
			WorkRelatedDomainList: []Domain{
				{
					Type:              "work",
					Domain:            "www.baidu.com",
					Remark:            "百度",
					SNI:               "www.baidu.com",
					EnableSubcription: true,
					EnableChatgpt:     false,
					UUID:              "7d2a8695-ee88-484d-8bea-ad86c95e6ff6",
					PATH:              "/",
				},
			},
			ActiveGlobalNodes: []Domain{
				{
					Type:              "vmess",
					Domain:            "sel.undervineyard.com",
					Remark:            "韩国sel",
					SNI:               "",
					EnableSubcription: true,
					EnableChatgpt:     true,
					UUID:              "",
					PATH:              "",
				},
				{
					Type:              "vmessCDN",
					Domain:            "sel.logv2.link",
					Remark:            "韩国selCDN",
					EnableSubcription: true,
					EnableChatgpt:     false,
					SNI:               "",
					UUID:              "",
					PATH:              "",
				},
				{
					Type:              "vlessCDN",
					Domain:            "anycastus.undervineyard.link",
					Remark:            "地区不定CDN",
					SNI:               "anycastus.undervineyard.link",
					EnableSubcription: true,
					EnableChatgpt:     false,
					UUID:              "b66da0cb-342e-482e-a0cf-1d94698a4731",
					PATH:              "/?ed=2048",
				},
			},
		}

		_, err := globalCOLLECTIONS.InsertOne(ctx, global)
		if err != nil {
			panic(err)
		}

		// remove node_global_list key from all users in USERS collection

		_, err = userCollection.UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"node_global_list": ""}})
		if err != nil {
			panic(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
