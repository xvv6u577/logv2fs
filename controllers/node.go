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
// 关联键为节点响应文档的 domain_as_id == node_traffic_periods.domain_as_id。
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

// subscriptionNodeKey 订阅节点业务唯一键：(remark, ip)。
func subscriptionNodeKey(node SubscriptionNode) bson.M {
	return bson.M{"remark": node.Remark, "ip": node.IP}
}

func UpsertNodes() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var current = time.Now().Local()
		var rawFormData []SubscriptionNode

		if err := c.BindJSON(&rawFormData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		nodeCol := database.GetCollection(model.SubscriptionNode{})
		submittedKeys := make([]bson.M, 0, len(rawFormData))

		for _, domain := range rawFormData {
			if domain.Remark == "" {
				continue
			}
			if domain.Type == "reality" {
				domain.PUBLIC_KEY = getPublicKey()
				domain.SHORT_ID = getShortID()
			}
			filter := subscriptionNodeKey(domain)
			update := bson.M{
				"$set": bson.M{
					"type":          domain.Type,
					"remark":        domain.Remark,
					"domain":        domain.Domain,
					"ip":            domain.IP,
					"sni":           domain.SNI,
					"uuid":          domain.UUID,
					"path":          domain.PATH,
					"server_port":   domain.SERVER_PORT,
					"password":      domain.PASSWORD,
					"public_key":    domain.PUBLIC_KEY,
					"short_id":      domain.SHORT_ID,
					"enable_openai": domain.EnableOpenai,
					"weight":        domain.Weight,
					"status":        "active",
					"updated_at":    current,
				},
				"$setOnInsert": bson.M{
					"_id":        primitive.NewObjectID(),
					"created_at": current,
				},
			}
			opts := options.Update().SetUpsert(true)
			_, err := nodeCol.UpdateOne(context.TODO(), filter, update, opts)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("UpdateOne in subscription_nodes error: %v", err)
				return
			}
			submittedKeys = append(submittedKeys, subscriptionNodeKey(domain))
		}

		inactiveFilter := bson.M{"type": bson.M{"$in": []string{"reality", "hysteria2", "vlessCDN"}}}
		if len(submittedKeys) > 0 {
			inactiveFilter["$nor"] = submittedKeys
		}
		inactiveUpdate := bson.M{"$set": bson.M{"status": "inactive", "updated_at": current}}
		_, err := nodeCol.UpdateMany(context.TODO(), inactiveFilter, inactiveUpdate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("UpdateMany in subscription_nodes error: %v", err)
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
		var filter = bson.D{
			{Key: "type", Value: bson.D{{Key: "$ne", Value: "work"}}},
			{Key: "status", Value: bson.D{{Key: "$ne", Value: "inactive"}}},
		}
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

		// subscription_nodes 是节点主集合；监控页只展示可统计类型，并按 domain 合并。
		// 通过聚合管道把 node_traffic_periods 里对应粒度的数据注入回 daily/monthly/yearly_logs，
		// 保持给前端的 JSON 结构与拆表前兼容。
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{
				{Key: "status", Value: "active"},
				{Key: "type", Value: bson.D{{Key: "$in", Value: bson.A{"reality", "hysteria2"}}}},
				{Key: "domain", Value: bson.D{{Key: "$ne", Value: ""}}},
			}}},
			{{Key: "$sort", Value: bson.D{{Key: "weight", Value: 1}, {Key: "updated_at", Value: -1}}}},
			{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$domain"},
				{Key: "remark", Value: bson.D{{Key: "$first", Value: "$remark"}}},
				{Key: "status", Value: bson.D{{Key: "$first", Value: "$status"}}},
				{Key: "created_at", Value: bson.D{{Key: "$min", Value: "$created_at"}}},
				{Key: "updated_at", Value: bson.D{{Key: "$max", Value: "$updated_at"}}},
			}}},
			{{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "domain_as_id", Value: "$_id"},
				{Key: "remark", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
			}}},
			lookupNodePeriodStage("daily", "date", "daily_logs", 0),
			lookupNodePeriodStage("monthly", "month", "monthly_logs", 0),
			lookupNodePeriodStage("yearly", "year", "yearly_logs", 0),
		}

		cur, err := database.GetCollection(model.SubscriptionNode{}).Aggregate(context.TODO(), pipeline)
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
