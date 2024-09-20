/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "init TRAFFIC collection in database",
	Run: func(cmd *cobra.Command, args []string) {

		// globalCollection
		filter := bson.M{"name": "GLOBAL"}
		var global model.GlobalVariable
		err := globalCollection.FindOne(context.TODO(), filter).Decode(&global)
		if err != nil {
			log.Printf("error getting global collection: %v\n", err)
		}

		// insert WorkRelatedDomainList into MoniteringDomains collection, ActiveGlobalNodes into global collection
		for _, domain := range global.WorkRelatedDomainList {
			MoniteringDomainsCol.InsertOne(context.TODO(), domain)

		}

		for _, node := range global.ActiveGlobalNodes {
			globalCollection.InsertOne(context.TODO(), node)
		}

		// userCollection
		cursor, err := userCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			log.Printf("error getting traffic collection: %v\n", err)
		}

		for cursor.Next(context.TODO()) {

			var userTrafficLog model.UserTrafficLogs

			var user model.User
			if err := cursor.Decode(&user); err != nil {
				log.Printf("error decoding traffic: %v\n", err)
			}

			userTrafficLog.Email_As_Id = user.Email
			userTrafficLog.ID = primitive.NewObjectID()
			userTrafficLog.Password = user.Password
			userTrafficLog.UUID = user.UUID
			userTrafficLog.Role = user.Role
			userTrafficLog.Status = user.Status
			userTrafficLog.Name = user.Name
			userTrafficLog.Token = user.Token
			userTrafficLog.Refresh_token = user.Refresh_token
			userTrafficLog.User_id = user.User_id
			userTrafficLog.Used = user.Usedtraffic
			userTrafficLog.Credit = user.Credittraffic
			userTrafficLog.CreatedAt = user.CreatedAt
			userTrafficLog.UpdatedAt = user.UpdatedAt
			userTrafficLog.HourlyLogs = []struct {
				Timestamp time.Time `json:"timestamp" bson:"timestamp"`
				Traffic   int64     `json:"traffic" bson:"traffic"`
			}{}

			if len(user.TrafficByYear) > 0 {
				for _, year := range user.TrafficByYear {
					userTrafficLog.YearlyLogs = append(userTrafficLog.YearlyLogs, struct {
						Year    string `json:"year" bson:"year"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{
						Year:    year.Period,
						Traffic: year.Amount,
					})
				}
			}

			if len(user.TrafficByMonth) > 0 {
				for _, month := range user.TrafficByMonth {
					userTrafficLog.MonthlyLogs = append(userTrafficLog.MonthlyLogs, struct {
						Month   string `json:"month" bson:"month"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{
						Month:   month.Period,
						Traffic: month.Amount,
					})
				}
			}

			if len(user.TrafficByDay) > 0 {
				for _, day := range user.TrafficByDay {
					userTrafficLog.DailyLogs = append(userTrafficLog.DailyLogs, struct {
						Date    string `json:"date" bson:"date"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{
						Date:    day.Period,
						Traffic: day.Amount,
					})
				}
			}

			if user.UsedByCurrentYear.Amount > 0 {
				userTrafficLog.YearlyLogs = append(userTrafficLog.YearlyLogs, struct {
					Year    string `json:"year" bson:"year"`
					Traffic int64  `json:"traffic" bson:"traffic"`
				}{
					Year:    user.UsedByCurrentYear.Period,
					Traffic: user.UsedByCurrentYear.Amount,
				})
			}

			if user.UsedByCurrentMonth.Amount > 0 {
				userTrafficLog.MonthlyLogs = append(userTrafficLog.MonthlyLogs, struct {
					Month   string `json:"month" bson:"month"`
					Traffic int64  `json:"traffic" bson:"traffic"`
				}{
					Month:   user.UsedByCurrentMonth.Period,
					Traffic: user.UsedByCurrentMonth.Amount,
				})
			}

			if user.UsedByCurrentDay.Amount > 0 {
				userTrafficLog.DailyLogs = append(userTrafficLog.DailyLogs, struct {
					Date    string `json:"date" bson:"date"`
					Traffic int64  `json:"traffic" bson:"traffic"`
				}{
					Date:    user.UsedByCurrentDay.Period,
					Traffic: user.UsedByCurrentDay.Amount,
				})
			}

			userTrafficLogs.InsertOne(context.Background(), userTrafficLog)

		}

		// nodesCollection
		cursor, err = nodesCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			log.Printf("error getting traffic collection: %v\n", err)
		}

		for cursor.Next(context.TODO()) {

			var nodeTrafficLog model.NodeTrafficLogs

			var node model.CurrentNode
			if err := cursor.Decode(&node); err != nil {
				log.Printf("error decoding traffic: %v\n", err)
			}

			nodeTrafficLog.Domain_As_Id = node.Domain
			nodeTrafficLog.ID = primitive.NewObjectID()
			nodeTrafficLog.CreatedAt = node.CreatedAt
			nodeTrafficLog.UpdatedAt = node.UpdatedAt
			nodeTrafficLog.Remark = node.Remark
			nodeTrafficLog.Status = node.Status
			nodeTrafficLog.HourlyLogs = []struct {
				Timestamp time.Time `json:"timestamp" bson:"timestamp"`
				Traffic   int64     `json:"traffic" bson:"traffic"`
			}{}

			if len(node.NodeByYear) > 0 {
				for _, year := range node.NodeByYear {
					nodeTrafficLog.YearlyLogs = append(nodeTrafficLog.YearlyLogs, struct {
						Year    string `json:"year" bson:"year"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{
						Year:    year.Period,
						Traffic: year.Amount,
					})
				}
			}

			if len(node.NodeByMonth) > 0 {
				for _, month := range node.NodeByMonth {
					nodeTrafficLog.MonthlyLogs = append(nodeTrafficLog.MonthlyLogs, struct {
						Month   string `json:"month" bson:"month"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{
						Month:   month.Period,
						Traffic: month.Amount,
					})
				}
			}

			if len(node.NodeByDay) > 0 {
				for _, day := range node.NodeByDay {
					nodeTrafficLog.DailyLogs = append(nodeTrafficLog.DailyLogs, struct {
						Date    string `json:"date" bson:"date"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{
						Date:    day.Period,
						Traffic: day.Amount,
					})
				}
			}

			if node.NodeAtCurrentYear.Amount > 0 {
				nodeTrafficLog.YearlyLogs = append(nodeTrafficLog.YearlyLogs, struct {
					Year    string `json:"year" bson:"year"`
					Traffic int64  `json:"traffic" bson:"traffic"`
				}{
					Year:    node.NodeAtCurrentYear.Period,
					Traffic: node.NodeAtCurrentDay.Amount,
				})
			}

			if node.NodeAtCurrentMonth.Amount > 0 {
				nodeTrafficLog.MonthlyLogs = append(nodeTrafficLog.MonthlyLogs, struct {
					Month   string `json:"month" bson:"month"`
					Traffic int64  `json:"traffic" bson:"traffic"`
				}{
					Month:   node.NodeAtCurrentMonth.Period,
					Traffic: node.NodeAtCurrentMonth.Amount,
				})
			}

			if node.NodeAtCurrentDay.Amount > 0 {
				nodeTrafficLog.DailyLogs = append(nodeTrafficLog.DailyLogs, struct {
					Date    string `json:"date" bson:"date"`
					Traffic int64  `json:"traffic" bson:"traffic"`
				}{
					Date:    node.NodeAtCurrentDay.Period,
					Traffic: node.NodeAtCurrentDay.Amount,
				})
			}

			nodeTrafficLogs.InsertOne(context.Background(), nodeTrafficLog)
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
