package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// lookupNodePeriodStage 同 lookupUserPeriodStage，但作用于 node_traffic_periods 集合，
// 关联键为 NODE_TRAFFIC_LOGS.domain_as_id == node_traffic_periods.domain_as_id。
//
// 参数语义与 lookupUserPeriodStage 一致：
//   - kind: "daily" | "monthly" | "yearly"
//   - periodAlias: 输出数组中表示周期的字段名（"date" / "month" / "year"）
//   - asField: 注入字段名（"daily_logs" / "monthly_logs" / "yearly_logs"）
//   - limit: <=0 表示不限
func lookupNodePeriodStage(kind, periodAlias, asField string, limit int) bson.D {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{{Key: "$and", Value: bson.A{
				bson.D{{Key: "$eq", Value: bson.A{"$domain_as_id", "$$domain"}}},
				bson.D{{Key: "$eq", Value: bson.A{"$kind", kind}}},
			}}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "period", Value: -1}}}},
	}
	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: periodAlias, Value: "$period"},
		{Key: "traffic", Value: 1},
	}}})

	return bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: model.NodeTrafficPeriod{}.CollectionName()},
		{Key: "let", Value: bson.D{{Key: "domain", Value: "$domain_as_id"}}},
		{Key: "pipeline", Value: pipeline},
		{Key: "as", Value: asField},
	}}}
}

// check if a domain is in a domain object list
func IsDomainInDomainList(domain string, domainList []SubscriptionNode) bool {
	for _, domainObj := range domainList {
		if domainObj.Domain == domain {
			return true
		}
	}
	return false
}

// check if domain's remark is in a domain object list
func IsRemarkInDomainList(remark string, domainList []SubscriptionNode) bool {
	for _, domainObj := range domainList {
		if domainObj.Remark == remark {
			return true
		}
	}
	return false
}

// Function to remove duplicated domains, also remove vlessCDN nodes.
func sanitizeNodes(domains []SubscriptionNode) []SubscriptionNode {
	seen := make(map[string]bool)
	var result []SubscriptionNode
	for _, domain := range domains {
		if domain.Type == "vlessCDN" {
			continue
		}
		if _, exists := seen[domain.Domain]; !exists {
			seen[domain.Domain] = true
			result = append(result, domain)
		}
	}
	return result
}

func UpsertNodes() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var current = time.Now().Local()
		var rawFormData, dataCollectableNodes []SubscriptionNode

		if err := c.BindJSON(&rawFormData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// remove duplicated domains, also remove vlessCDN nodes.
		dataCollectableNodes = sanitizeNodes(rawFormData)

		// types: reality, hysteria2, vlessCDN! if type is reality, reassgin public_key and short_id.
		// then, empty subscription_nodes collection, and insert rawFormData into it.
		database.GetCollection(model.SubscriptionNode{}).DeleteMany(context.TODO(), bson.M{})
		for i, domain := range rawFormData {
			if domain.Type == "reality" {
				rawFormData[i].PUBLIC_KEY = getPublicKey()
				rawFormData[i].SHORT_ID = getShortID()
			}
			database.GetCollection(model.SubscriptionNode{}).InsertOne(context.TODO(), domain)
		}

		// check if domain is in nodeTrafficLogsCol. if no, insert it. if yes, update it.
		// 周期级流量字段已下沉到 node_traffic_periods 集合，主文档不再预置任何空数组。
		for _, domain := range dataCollectableNodes {
			filter := bson.M{"domain_as_id": domain.Domain}
			update := bson.M{
				"$set": bson.M{
					"remark":     domain.Remark,
					"status":     "active",
					"updated_at": current,
				},
				"$setOnInsert": bson.M{
					"_id":          primitive.NewObjectID(),
					"domain_as_id": domain.Domain,
					"created_at":   current,
				},
			}
			opts := options.Update().SetUpsert(true)
			_, err := database.GetCollection(model.NodeTrafficLogs{}).UpdateOne(context.TODO(), filter, update, opts)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("UpdateOne in nodeTrafficLogsCol error: %v", err)
				return
			}

		}

		// Set status to "inactive" for NodeTrafficLogs entries not in dataCollectableNodes
		domainAsIds := make([]string, len(dataCollectableNodes))
		for i, domain := range dataCollectableNodes {
			domainAsIds[i] = domain.Domain
		}
		inactiveFilter := bson.M{"domain_as_id": bson.M{"$nin": domainAsIds}}
		inactiveUpdate := bson.M{"$set": bson.M{"status": "inactive"}}
		_, err := database.GetCollection(model.NodeTrafficLogs{}).UpdateMany(context.TODO(), inactiveFilter, inactiveUpdate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("UpdateMany in nodeTrafficLogsCol error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Congrats! Nodes updated in success!"})
	}
}

func GetSubscriptionNodes() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var activeNodes []SubscriptionNode
		// type is not "work"
		var filter = bson.D{{Key: "type", Value: bson.D{{Key: "$ne", Value: "work"}}}}
		cur, err := database.GetCollection(model.SubscriptionNode{}).Find(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find error: %v", err)
			return
		}
		err = cur.All(context.TODO(), &activeNodes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("All error: %v", err)
			return
		}

		c.JSON(http.StatusOK, activeNodes)
	}
}

func GetSingboxNodes() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 通过聚合管道把 node_traffic_periods 里对应粒度的数据注入回 daily/monthly/yearly_logs，
		// 保持给前端的 JSON 结构与拆表前完全一致。前端 nodes.js 仍按原字段名访问。
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{{Key: "status", Value: "active"}}}},
			lookupNodePeriodStage("daily", "date", "daily_logs", 0),
			lookupNodePeriodStage("monthly", "month", "monthly_logs", 0),
			lookupNodePeriodStage("yearly", "year", "yearly_logs", 0),
		}

		cur, err := database.GetCollection(model.NodeTrafficLogs{}).Aggregate(context.TODO(), pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetSingboxNodes aggregate error: %v", err)
			return
		}
		defer cur.Close(context.TODO())

		var activeNodes []NodeTrafficLogs
		if err := cur.All(context.TODO(), &activeNodes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetSingboxNodes cursor.All error: %v", err)
			return
		}

		c.JSON(http.StatusOK, activeNodes)
	}
}

// SaveCustomDate 保存节点自定义日期 - MongoDB版本
func SaveCustomDate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var request struct {
			DomainAsId string `json:"domain_as_id" binding:"required"`
			CustomDate string `json:"custom_date" binding:"required"`
		}

		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 使用upsert操作更新或插入自定义日期
		filter := bson.M{"domain_as_id": request.DomainAsId}
		update := bson.M{
			"$set": bson.M{
				"domain_as_id": request.DomainAsId,
				"custom_date":  request.CustomDate,
				"updated_at":   time.Now(),
			},
			"$setOnInsert": bson.M{
				"created_at": time.Now(),
			},
		}
		opts := options.Update().SetUpsert(true)

		_, err := database.GetCollection(model.CustomDate{}).UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("保存自定义日期失败: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "自定义日期保存成功"})
	}
}

// GetCustomDates 获取所有节点自定义日期 - MongoDB版本
func GetCustomDates() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cur, err := database.GetCollection(model.CustomDate{}).Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("查询自定义日期失败: %v", err)
			return
		}
		defer cur.Close(context.Background())

		customDates := make(map[string]string)
		for cur.Next(context.TODO()) {
			var doc struct {
				DomainAsId string `bson:"domain_as_id"`
				CustomDate string `bson:"custom_date"`
			}
			if err := cur.Decode(&doc); err != nil {
				log.Printf("解码自定义日期失败: %v", err)
				continue
			}
			customDates[doc.DomainAsId] = doc.CustomDate
		}

		c.JSON(http.StatusOK, customDates)
	}
}
