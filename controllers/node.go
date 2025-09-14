package controllers

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	ExpiryCheckDomainInfo = model.ExpiryCheckDomainInfo
)

// check if a domain is in a domain object list
func IsDomainInDomainList(domain string, domainList []Domain) bool {
	for _, domainObj := range domainList {
		if domainObj.Domain == domain {
			return true
		}
	}
	return false
}

// check if domain's remark is in a domain object list
func IsRemarkInDomainList(remark string, domainList []Domain) bool {
	for _, domainObj := range domainList {
		if domainObj.Remark == remark {
			return true
		}
	}
	return false
}

// Function to remove duplicate Domain.Domain in a Domain slice
func removeDuplicateDomains(domains []Domain) []Domain {
	seen := make(map[string]bool)
	var result []Domain
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

func AddNode() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var current = time.Now().Local()
		var nodeFromWebForm, dataCollectableNodes []Domain

		if err := c.BindJSON(&nodeFromWebForm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// remove duplicated domains in nodeFromWebForm
		dataCollectableNodes = removeDuplicateDomains(nodeFromWebForm)

		// types: reality, hysteria2, vlessCDN! if type is reality, reassgin public_key and short_id.
		for i, domain := range nodeFromWebForm {
			if domain.Type == "reality" {
				nodeFromWebForm[i].PUBLIC_KEY = PUBLIC_KEY
				nodeFromWebForm[i].SHORT_ID = SHORT_ID
			}

			// set remark as filter, check if node is in subNodesCol. if no, insert it. if yes, update it.
			filter := bson.M{"remark": domain.Remark}
			update := bson.M{"$set": nodeFromWebForm[i]}
			opts := options.Update().SetUpsert(true)
			_, err := subNodesCol.UpdateOne(context.TODO(), filter, update, opts)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("UpdateOne error: %v", err)
				return
			}

		}

		remarks := make([]string, len(nodeFromWebForm))
		for i, domain := range nodeFromWebForm {
			remarks[i] = domain.Remark
		}
		filter := bson.M{"remark": bson.M{"$nin": remarks}}
		_, err := subNodesCol.DeleteMany(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("DeleteMany error: %v", err)
			return
		}

		for _, domain := range dataCollectableNodes {

			// check if domain is in nodeTrafficLogsCol. if no, insert it. if yes, update it.
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
			_, err = nodeTrafficLogsCol.UpdateOne(context.TODO(), filter, update, opts)
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
		_, err = nodeTrafficLogsCol.UpdateMany(context.TODO(), inactiveFilter, inactiveUpdate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("UpdateMany in nodeTrafficLogsCol error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Congrats! Nodes updated in success!"})
	}
}

func GetActiveGlobalNodes() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var activeNodes []Domain
		// type is not "work"
		var filter = bson.D{{Key: "type", Value: bson.D{{Key: "$ne", Value: "work"}}}}
		cur, err := subNodesCol.Find(context.TODO(), filter)
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

func GetDomainsExpiryInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// 获取所有需要检查的域名
		var expiryDomains []ExpiryCheckDomainInfo
		cur, err := expiryCheckDomainCol.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find error: %v", err)
			return
		}

		if err = cur.All(ctx, &expiryDomains); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("All error: %v", err)
			return
		}

		// 准备结果数组和域名分类映射
		var domainInfos []ExpiryCheckDomainInfo
		normalDomains := make(map[string]string)
		unreachableDomains := make(map[string]string)

		// 并行处理域名可达性检查
		var wg sync.WaitGroup
		for _, domain := range expiryDomains {
			if domain.Domain == "localhost" {
				continue
			}
			wg.Add(1)
			go func(d ExpiryCheckDomainInfo) {
				defer wg.Done()
				if helper.IsDomainReachable(d.Domain) {
					normalDomains[d.Domain] = d.Remark
				} else {
					unreachableDomains[d.Domain] = d.Remark
				}
			}(domain)
		}
		wg.Wait()

		// 处理可达域名的证书信息
		port := "443"
		conf := &tls.Config{}

		for domain, remark := range normalDomains {
			conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 20 * time.Second}, "tcp", domain+":"+port, conf)
			if err != nil {
				log.Printf("tls.DialWithDialer Error for domain %s: %v", domain, err)
				// 如果TLS连接失败，将其标记为不可达
				unreachableDomains[domain] = remark
				continue
			}

			if err = conn.VerifyHostname(domain); err != nil {
				log.Printf("conn.VerifyHostname Error for domain %s: %v", domain, err)
				conn.Close()
				// 如果主机名验证失败，将其标记为不可达
				unreachableDomains[domain] = remark
				continue
			}

			expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
			conn.Close()

			domainInfos = append(domainInfos, ExpiryCheckDomainInfo{
				Domain:       domain,
				Remark:       remark,
				ExpiredDate:  expiry.Local().Format("2006-01-02 15:04:05"),
				DaysToExpire: int(time.Until(expiry).Hours() / 24),
			})
		}

		// 处理不可达域名
		for domain, remark := range unreachableDomains {
			domainInfos = append(domainInfos, ExpiryCheckDomainInfo{
				Domain:       domain,
				Remark:       remark,
				ExpiredDate:  "unreachable",
				DaysToExpire: -1,
			})
		}

		c.JSON(http.StatusOK, domainInfos)
	}
}

func UpdateExpiryCheckDomainsInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var domainOfWebForm []ExpiryCheckDomainInfo
		err := c.BindJSON(&domainOfWebForm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// Track domains to keep
		domainsToKeep := []string{}

		for _, domain := range domainOfWebForm {
			filter := bson.M{"domain": domain.Domain}
			update := bson.M{"$set": bson.M{
				"domain":         domain.Domain,
				"remark":         domain.Remark,
				"expired_date":   domain.ExpiredDate,
				"days_to_expire": domain.DaysToExpire,
			}}
			opts := options.Update().SetUpsert(true)
			_, err := expiryCheckDomainCol.UpdateOne(context.TODO(), filter, update, opts)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("UpdateOne error: %v", err)
				return
			}

			domainsToKeep = append(domainsToKeep, domain.Domain)
		}

		// Remove domains not in domainOfWebForm
		filter := bson.M{"domain": bson.M{"$nin": domainsToKeep}}
		_, err = expiryCheckDomainCol.DeleteMany(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("DeleteMany error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Update expiry check domains list successfully!"})
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
		cur, err := nodeTrafficLogsCol.Find(context.TODO(), filter)
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

		_, err := customDatesCol.UpdateOne(context.TODO(), filter, update, opts)
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

		cur, err := customDatesCol.Find(context.TODO(), bson.M{})
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
