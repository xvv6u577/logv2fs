/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "dev command",
	Long:  `dev command`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Println("dev called")

	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	devCmd.PersistentFlags().String("foo", "", "A help for foo")

	devCmd.Flags().BoolP("toggle", "", false, "Help message for toggle")
}
