/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/v2ray"
	"google.golang.org/grpc"
)

// v2rayServiceCmd represents the v2rayService command
var v2rayServiceCmd = &cobra.Command{
	Use:   "v2rayService",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic(err)
		}
		defer cmdConn.Close()

		var user_vmessws = User{
			Path:  "vmessws",
			UUID:  "b831381d-6324-4d53-ad4f-8cda48b30811",
			Email: "mytestuser",
		}

		NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user_vmessws.Path)
		NHSClient.AddUser(user_vmessws)

		var user_vmess = User{
			Path:  "ray",
			UUID:  "b831381d-6324-4d53-ad4f-8cda48b30811",
			Email: "mytestuser",
		}

		NHSClient = v2ray.NewHandlerServiceClient(cmdConn, user_vmess.Path)
		NHSClient.AddUser(user_vmess)

	},
}

func init() {
	rootCmd.AddCommand(v2rayServiceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// v2rayServiceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// v2rayServiceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
