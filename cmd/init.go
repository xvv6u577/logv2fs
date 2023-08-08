/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init GLOBAL, NODES collection",
	Long:  `init GLOBAL, NODES collection`,
	Run: func(cmd *cobra.Command, args []string) {

		// hello world
		fmt.Println("dev called.")

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
