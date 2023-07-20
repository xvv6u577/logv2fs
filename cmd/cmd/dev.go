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

		type Person struct {
			Name string
			Age  int
		}

		people := []Person{
			{Name: "Alice", Age: 25},
			{Name: "Bob", Age: 30},
			{Name: "Charlie", Age: 35},
		}

		// Updating the Age of each Person by adding 1
		for i := range people {
			people[i].Age++
		}

		// Printing the updated values
		for _, person := range people {
			fmt.Println(person.Name, person.Age)
		}

		var test = []*NodeAtPeriod{
			{
				Period: "202101",
				Amount: 100,
				UserTrafficAtPeriod: map[string]int64{
					"email1": 10,
				},
			},
			{
				Period: "202102",
				Amount: 200,
				UserTrafficAtPeriod: map[string]int64{
					"email1": 20,
				},
			},
		}
		var foundNode *NodeAtPeriod

		for _, v := range test {
			if v.Period == "202101" {
				foundNode = v
				break
			}
		}

		foundNode.Amount = 300

		for _, v := range test {
			fmt.Println(v)
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
