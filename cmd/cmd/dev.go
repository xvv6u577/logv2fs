/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var nodeArray = []*CurrentNode{
			{
				Status: "active",
				Domain: "www.google.com",
				NodeAtCurrentYear: NodeAtPeriod{
					Period: "2021",
					Amount: 100,
					UserTrafficAtPeriod: map[string]int64{
						"john": 100,
					},
				},
				NodeAtCurrentMonth: NodeAtPeriod{
					Period: "202101",
					Amount: 100,
					UserTrafficAtPeriod: map[string]int64{
						"ley": 100,
					},
				},
				NodeAtCurrentDay: NodeAtPeriod{
					Period: "20210101",
					Amount: 100,
					UserTrafficAtPeriod: map[string]int64{
						"ley": 100,
					},
				},
				NodeByYear: []NodeAtPeriod{
					{
						Period: "2020",
						Amount: 100,
						UserTrafficAtPeriod: map[string]int64{
							"john": 100,
						},
					},
				},
				NodeByMonth: []NodeAtPeriod{
					{
						Period: "202012",
						Amount: 100,
						UserTrafficAtPeriod: map[string]int64{
							"john": 100,
						},
					},
				},
				NodeByDay: []NodeAtPeriod{
					{
						Period: "20201231",
						Amount: 100,
						UserTrafficAtPeriod: map[string]int64{
							"john": 100,
						},
					},
				},
			},
			{
				Status: "active",
				Domain: "www.facebook.com",
				NodeAtCurrentYear: NodeAtPeriod{
					Period: "2021",
					Amount: 100,
					UserTrafficAtPeriod: map[string]int64{
						"john": 100,
					},
				},
				NodeAtCurrentMonth: NodeAtPeriod{
					Period: "202101",
					Amount: 100,
					UserTrafficAtPeriod: map[string]int64{
						"ley": 100,
					},
				},
				NodeAtCurrentDay: NodeAtPeriod{
					Period: "20210101",
					Amount: 100,
					UserTrafficAtPeriod: map[string]int64{
						"ley": 100,
					},
				},
				NodeByYear: []NodeAtPeriod{
					{
						Period: "2020",
						Amount: 100,
						UserTrafficAtPeriod: map[string]int64{
							"john": 100,
						},
					},
				},
				NodeByMonth: []NodeAtPeriod{
					{
						Period: "202012",
						Amount: 100,
						UserTrafficAtPeriod: map[string]int64{
							"john": 100,
						},
					},
				},
				NodeByDay: []NodeAtPeriod{
					{
						Period: "20201231",
						Amount: 100,
						UserTrafficAtPeriod: map[string]int64{
							"john": 100,
						},
					},
				},
			},
		}

		var foundNode *CurrentNode
		// print foundNode.NodeAtCurrentYear.Period
		for _, node := range nodeArray {
			if node.Domain == "www.google.com" {
				foundNode = node
			}
			fmt.Printf("node.domain: %s\n", node.Domain)
		}

		foundNode.Domain = "www.twitter.com"

		for _, node := range nodeArray {
			fmt.Printf("node.domain: %s\n", node.Domain)
		}

	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// devCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// devCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
