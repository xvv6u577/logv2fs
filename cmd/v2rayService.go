/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
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
		fmt.Println("V2ray process runs at 8070, 10000, 10001, 10002")
		var myCmd = exec.Command(V2RAY, "-config", V2RAY_CONFIG)
		if err := myCmd.Run(); err != nil {
			log.Panic("Panic: ", err)
		}

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
