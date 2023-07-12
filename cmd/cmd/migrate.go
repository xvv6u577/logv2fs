/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/caster8013/logv2rayfullstack/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "init TRAFFIC collection in database",
	Run: func(cmd *cobra.Command, args []string) {
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		var projections = bson.D{
			{Key: "email", Value: 1},
		}
		users, err := database.GetPartialInfosForAllUsers(projections)
		if err != nil {
			panic(err)
		}

		var nodeArray []*CurrentNode
		var current = time.Now().Local()
		var current_year = current.Format("2006")
		var current_month = current.Format("200601")
		var current_day = current.Format("20060102")

		for _, user := range users {
			var myTraffic *mongo.Collection = database.OpenCollection(database.Client, user.Email)
			var TrafficInDBArray []*TrafficInDB

			var filter = bson.D{{}}
			var myProjections = bson.D{}

			cursor, err := myTraffic.Find(ctx, filter, options.Find().SetProjection(myProjections))
			if err != nil {
				panic(err)
			}

			if err = cursor.All(ctx, &TrafficInDBArray); err != nil {
				panic(err)
			}

			for _, traffic := range TrafficInDBArray {
				// compare with time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), if traffic.CreatedAt is before this time, then skip
				if traffic.CreatedAt.Before(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)) {
					continue
				}

				var trafficInfo = TrafficInDB{
					Email:     user.Email,
					Domain:    traffic.Domain,
					CreatedAt: traffic.CreatedAt,
					Total:     traffic.Total,
				}

				if _, err = trafficCollection.InsertOne(ctx, trafficInfo); err != nil {
					panic(err)
				}

				// check if this domain is in nodeArray by traffic.Domain, if not, add it, else, merge the traffic info to the node
				var found = false
				var foundNode *CurrentNode
				var year = traffic.CreatedAt.Format("2006")
				var month = traffic.CreatedAt.Format("200601")
				var day = traffic.CreatedAt.Format("20060102")

				for _, node := range nodeArray {
					if node.Domain == traffic.Domain {
						found = true
						foundNode = node
						break
					}
				}

				if !found {
					var node = &CurrentNode{
						Status:    "inactive",
						Domain:    traffic.Domain,
						CreatedAt: time.Now().Local(),
						UpdatedAt: time.Now().Local(),
					}
					// if node_at_current_year equals to current_year, then add traffic.Total to node_at_current_year.Amount
					// else, add it to node.NodeByYear.
					if node.NodeAtCurrentYear.Period == current_year {
						node.NodeAtCurrentYear.Amount += traffic.Total
						node.NodeAtCurrentYear.Period = current_year
						node.NodeAtCurrentYear.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						node.NodeByYear = append(node.NodeByYear, NodeAtPeriod{
							Period: current_year,
							Amount: traffic.Total,
							UserTrafficAtPeriod: map[string]int64{
								user.Email: traffic.Total,
							},
						})
					}

					// if node_at_current_month equals to current_month, then add traffic.Total to node_at_current_month.Amount
					// else, add it to node.NodeByMonth.
					if node.NodeAtCurrentMonth.Period == current_month {
						node.NodeAtCurrentMonth.Amount += traffic.Total
						node.NodeAtCurrentMonth.Period = current_month
						node.NodeAtCurrentMonth.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						node.NodeByMonth = append(node.NodeByMonth, NodeAtPeriod{
							Period: current_month,
							Amount: traffic.Total,
							UserTrafficAtPeriod: map[string]int64{
								user.Email: traffic.Total,
							},
						})
					}

					// if node_at_current_day equals to current_day, then add traffic.Total to node_at_current_day.Amount
					// else, add it to node.NodeByDay.
					if node.NodeAtCurrentDay.Period == current_day {
						node.NodeAtCurrentDay.Amount += traffic.Total
						node.NodeAtCurrentDay.Period = current_day
						node.NodeAtCurrentDay.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						node.NodeByDay = append(node.NodeByDay, NodeAtPeriod{
							Period: current_day,
							Amount: traffic.Total,
							UserTrafficAtPeriod: map[string]int64{
								user.Email: traffic.Total,
							},
						})
					}
					nodeArray = append(nodeArray, node)
				} else {
					// if node_at_current_year equals to current_year, then add traffic.Total to node_at_current_year.Amount
					// else, add it to node.NodeByYear.
					if foundNode.NodeAtCurrentYear.Period == current_year {
						foundNode.NodeAtCurrentYear.Amount += traffic.Total
						foundNode.NodeAtCurrentYear.Period = current_year
						foundNode.NodeAtCurrentYear.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						// chek period in array node.NodeByYear, if found, then add traffic.Total to node.NodeByYear.Amount, else, append it to node.NodeByYear
						var ifFoundNodeByYear = false
						var foundNodeByYear *NodeAtPeriod
						for _, node := range foundNode.NodeByYear {
							if node.Period == year {
								ifFoundNodeByYear = true
								foundNodeByYear = &node
								break
							}
						}

						if ifFoundNodeByYear {
							foundNodeByYear.Amount += traffic.Total
							foundNodeByYear.UserTrafficAtPeriod[user.Email] += traffic.Total
						} else {
							foundNode.NodeByYear = append(foundNode.NodeByYear, NodeAtPeriod{
								Period: year,
								Amount: traffic.Total,
								UserTrafficAtPeriod: map[string]int64{
									user.Email: traffic.Total,
								},
							})
						}
					}

					// if node_at_current_month equals to current_month, then add traffic.Total to node_at_current_month.Amount
					// else, add it to node.NodeByMonth.
					if foundNode.NodeAtCurrentMonth.Period == current_month {
						foundNode.NodeAtCurrentMonth.Amount += traffic.Total
						foundNode.NodeAtCurrentMonth.Period = current_month
						foundNode.NodeAtCurrentMonth.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						// chek period in array node.NodeByMonth, if found, then add traffic.Total to node.NodeByMonth.Amount, else, append it to node.NodeByMonth
						var ifFoundNodeByMonth = false
						var foundNodeByMonth *NodeAtPeriod
						for _, node := range foundNode.NodeByMonth {
							if node.Period == month {
								ifFoundNodeByMonth = true
								foundNodeByMonth = &node
								break
							}
						}

						if ifFoundNodeByMonth {
							foundNodeByMonth.Amount += traffic.Total
							foundNodeByMonth.UserTrafficAtPeriod[user.Email] += traffic.Total
						} else {
							foundNode.NodeByMonth = append(foundNode.NodeByMonth, NodeAtPeriod{
								Period: month,
								Amount: traffic.Total,
								UserTrafficAtPeriod: map[string]int64{
									user.Email: traffic.Total,
								},
							})
						}
					}

					// if node_at_current_day equals to current_day, then add traffic.Total to node_at_current_day.Amount
					// else, add it to node.NodeByDay.
					if foundNode.NodeAtCurrentDay.Period == current_day {
						foundNode.NodeAtCurrentDay.Amount += traffic.Total
						foundNode.NodeAtCurrentDay.Period = current_day
						foundNode.NodeAtCurrentDay.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						// chek period in array node.NodeByDay, if found, then add traffic.Total to node.NodeByDay.Amount, else, append it to node.NodeByDay
						var ifFoundNodeByDay = false
						var foundNodeByDay *NodeAtPeriod
						for _, node := range foundNode.NodeByDay {
							if node.Period == day {
								ifFoundNodeByDay = true
								foundNodeByDay = &node
								break
							}
						}

						if ifFoundNodeByDay {
							foundNodeByDay.Amount += traffic.Total
							foundNodeByDay.UserTrafficAtPeriod[user.Email] += traffic.Total
						} else {
							foundNode.NodeByDay = append(foundNode.NodeByDay, NodeAtPeriod{
								Period: day,
								Amount: traffic.Total,
								UserTrafficAtPeriod: map[string]int64{
									user.Email: traffic.Total,
								},
							})
						}
					}
				}

			}

		}

		// insert nodeArray to NODES collection.
		// at first, check domain if it is in adminUser.NodeGlobalList, if itis , set status to "active", else, set status to "inactive"
		// then, insert it.
		var adminUser *User
		var filter = bson.D{
			{Key: "role", Value: "admin"},
		}
		var adminProjections = bson.D{
			{Key: "email", Value: 1},
			{Key: "node_global_list", Value: 1},
		}

		if err := userCollection.FindOne(ctx, filter, options.FindOne().SetProjection(adminProjections)).Decode(&adminUser); err != nil {
			panic(err)
		}

		for _, node := range nodeArray {
			var found = false
			for _, domain := range adminUser.NodeGlobalList {
				if node.Domain == domain {
					found = true
					break
				}
			}

			if found {
				node.Status = "active"
			} else {
				node.Status = "inactive"
			}

			if _, err := nodesCollection.InsertOne(ctx, node); err != nil {
				panic(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
