/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
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

		log.Println("clear2 called")

	},
}

func init() {
	rootCmd.AddCommand(clear2Cmd)
}
