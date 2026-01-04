package mongodb

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xvv6u577/logv2fs/database/mongodb"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
		mongodb.GetCollection(model.SubscriptionNode{}).DeleteMany(context.TODO(), bson.M{})
		for i, domain := range rawFormData {
			if domain.Type == "reality" {
				rawFormData[i].PUBLIC_KEY = getPublicKey()
				rawFormData[i].SHORT_ID = getShortID()
			}
			mongodb.GetCollection(model.SubscriptionNode{}).InsertOne(context.TODO(), domain)
		}

		// check if domain is in nodeTrafficLogsCol. if no, insert it. if yes, update it.
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
					"hourly_logs": []struct {
						Timestamp time.Time `json:"timestamp" bson:"timestamp"`
						Traffic   int64     `json:"traffic" bson:"traffic"`
					}{},
					"daily_logs": []struct {
						Date    string `json:"date" bson:"date"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{},
					"monthly_logs": []struct {
						Month   string `json:"month" bson:"month"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{},
					"yearly_logs": []struct {
						Year    string `json:"year" bson:"year"`
						Traffic int64  `json:"traffic" bson:"traffic"`
					}{},
				},
			}
			opts := options.Update().SetUpsert(true)
			_, err := mongodb.GetCollection(model.NodeTrafficLogs{}).UpdateOne(context.TODO(), filter, update, opts)
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
		_, err := mongodb.GetCollection(model.NodeTrafficLogs{}).UpdateMany(context.TODO(), inactiveFilter, inactiveUpdate)
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
		cur, err := mongodb.GetCollection(model.SubscriptionNode{}).Find(context.TODO(), filter)
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

		var activeNodes []NodeTrafficLogs
		var filter = bson.D{primitive.E{Key: "status", Value: "active"}}
		cur, err := mongodb.GetCollection(model.NodeTrafficLogs{}).Find(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find error: %v", err)
			return
		}

		for cur.Next(context.TODO()) {
			var node NodeTrafficLogs
			if err := cur.Decode(&node); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("Decode error: %v", err)
				return
			}
			activeNodes = append(activeNodes, node)
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

		_, err := mongodb.GetCollection(model.CustomDate{}).UpdateOne(context.TODO(), filter, update, opts)
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

		cur, err := mongodb.GetCollection(model.CustomDate{}).Find(context.TODO(), bson.M{})
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
