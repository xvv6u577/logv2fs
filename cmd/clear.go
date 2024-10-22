/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear all collections named by email in database",
	Run: func(cmd *cobra.Command, args []string) {

		log.Println("clearing all collections named by email in database")

	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
