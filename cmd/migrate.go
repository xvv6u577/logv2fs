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

		var current = time.Now().Local()
		var current_year = current.Format("2006")
		var current_month = current.Format("200601")
		var current_day = current.Format("20060102")

		var projections = bson.D{
			{Key: "email", Value: 1},
		}
		users, err := database.GetPartialInfosForAllUsers(projections)
		if err != nil {
			panic(err)
		}

		var nodeArray []*CurrentNode
		// query all in nodeCollection, and put them into nodeArray
		var nodeFilter = bson.D{{}}
		var nodeProjections = bson.D{}
		cursor, err := nodesCollection.Find(ctx, nodeFilter, options.Find().SetProjection(nodeProjections))
		if err != nil {
			panic(err)
		}
		if err = cursor.All(ctx, &nodeArray); err != nil {
			panic(err)
		}

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
						Status: "inactive",
						Domain: traffic.Domain,
						Remark: traffic.Domain,
						NodeAtCurrentDay: NodeAtPeriod{
							Period:              current_day,
							Amount:              0,
							UserTrafficAtPeriod: map[string]int64{},
						},
						NodeAtCurrentMonth: NodeAtPeriod{
							Period:              current_month,
							Amount:              0,
							UserTrafficAtPeriod: map[string]int64{},
						},
						NodeAtCurrentYear: NodeAtPeriod{
							Period:              current_year,
							Amount:              0,
							UserTrafficAtPeriod: map[string]int64{},
						},
						NodeByDay:   []NodeAtPeriod{},
						NodeByMonth: []NodeAtPeriod{},
						NodeByYear:  []NodeAtPeriod{},
						CreatedAt:   time.Now().Local(),
						UpdatedAt:   time.Now().Local(),
					}

					if node.NodeAtCurrentYear.Period == year {
						node.NodeAtCurrentYear.Amount += traffic.Total
						node.NodeAtCurrentYear.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						node.NodeByYear = append(node.NodeByYear, NodeAtPeriod{
							Period: year,
							Amount: traffic.Total,
							UserTrafficAtPeriod: map[string]int64{
								user.Email: traffic.Total,
							},
						})
					}

					if node.NodeAtCurrentMonth.Period == month {
						node.NodeAtCurrentMonth.Amount += traffic.Total
						node.NodeAtCurrentMonth.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						node.NodeByMonth = append(node.NodeByMonth, NodeAtPeriod{
							Period: month,
							Amount: traffic.Total,
							UserTrafficAtPeriod: map[string]int64{
								user.Email: traffic.Total,
							},
						})
					}

					if node.NodeAtCurrentDay.Period == day {
						node.NodeAtCurrentDay.Amount += traffic.Total
						node.NodeAtCurrentDay.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						node.NodeByDay = append(node.NodeByDay, NodeAtPeriod{
							Period: day,
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
					if foundNode.NodeAtCurrentYear.Period == year {
						foundNode.NodeAtCurrentYear.Amount += traffic.Total
						foundNode.NodeAtCurrentYear.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						// chek period in array node.NodeByYear, if found, then add traffic.Total to node.NodeByYear.Amount, else, append it to node.NodeByYear
						var ifFoundNodeByYear = false
						for i := range foundNode.NodeByYear {
							if foundNode.NodeByYear[i].Period == year {
								foundNode.NodeByYear[i].Amount += traffic.Total
								foundNode.NodeByYear[i].UserTrafficAtPeriod[user.Email] += traffic.Total
								ifFoundNodeByYear = true
								break
							}
						}

						if !ifFoundNodeByYear {
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
					if foundNode.NodeAtCurrentMonth.Period == month {
						foundNode.NodeAtCurrentMonth.Amount += traffic.Total
						foundNode.NodeAtCurrentMonth.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						// chek period in array node.NodeByMonth, if found, then add traffic.Total to node.NodeByMonth.Amount, else, append it to node.NodeByMonth
						var ifFoundNodeByMonth = false
						for i := range foundNode.NodeByMonth {
							if foundNode.NodeByMonth[i].Period == month {
								foundNode.NodeByMonth[i].Amount += traffic.Total
								foundNode.NodeByMonth[i].UserTrafficAtPeriod[user.Email] += traffic.Total
								ifFoundNodeByMonth = true
								break
							}
						}

						if !ifFoundNodeByMonth {
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
					if foundNode.NodeAtCurrentDay.Period == day {
						foundNode.NodeAtCurrentDay.Amount += traffic.Total
						foundNode.NodeAtCurrentDay.UserTrafficAtPeriod[user.Email] += traffic.Total
					} else {
						// chek period in array node.NodeByDay, if found, then add traffic.Total to node.NodeByDay.Amount, else, append it to node.NodeByDay
						var ifFoundNodeByDay = false
						for i := range foundNode.NodeByDay {
							if foundNode.NodeByDay[i].Period == day {
								foundNode.NodeByDay[i].Amount += traffic.Total
								foundNode.NodeByDay[i].UserTrafficAtPeriod[user.Email] += traffic.Total
								ifFoundNodeByDay = true
								break
							}
						}

						if !ifFoundNodeByDay {
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

		// upsert nodeArray to NODES collection.
		for _, node := range nodeArray {
			var filter = bson.D{{Key: "domain", Value: node.Domain}}
			var update = bson.D{
				{Key: "$set", Value: node},
			}
			var upsert = true
			if _, err = nodesCollection.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert}); err != nil {
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
