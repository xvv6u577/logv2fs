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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DomainInfo struct {
	Domain       string `json:"domain"`
	Remark       string `json:"remark"`
	ExpiredDate  string `json:"expired_date"`
	DaysToExpire int    `json:"days_to_expire"`
}

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

			// set remark as filter, check if node is in globalCollection. if no, insert it. if yes, update it.
			filter := bson.M{"remark": domain.Remark}
			update := bson.M{"$set": nodeFromWebForm[i]}
			opts := options.Update().SetUpsert(true)
			_, err := globalCollection.UpdateOne(context.TODO(), filter, update, opts)
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
		_, err := globalCollection.DeleteMany(context.TODO(), filter)
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
		cur, err := globalCollection.Find(context.TODO(), filter)
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

func GetWorkDomainInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var domainInfos []DomainInfo
		var workDoamins []Domain

		// get all domains from MoniteringDomainsCol
		cur, err := MoniteringDomainsCol.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find error: %v", err)
			return
		}
		err = cur.All(ctx, &workDoamins)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("All error: %v", err)
			return
		}

		domainInfos, err = getDomainExpiredStatus(workDoamins)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while parsing domain info: %v", err)
			return
		}

		c.JSON(http.StatusOK, domainInfos)
	}
}

func getDomainExpiredStatus(domains []Domain) ([]DomainInfo, error) {

	var minorDomainInfos []DomainInfo
	var port = "443"
	var conf = &tls.Config{
		// InsecureSkipVerify: true,
	}
	var normalDomains = make(map[string]string)
	var unreachableDomains = make(map[string]string)

	// split domains into normalDomains and unreachableDomains by parallel processing. if domain is reachable, append it to normalDomains, else append it to unreachableDomains.
	// if domain is localhost, skip it.
	var wg sync.WaitGroup
	for _, domain := range domains {
		if domain.Domain == "localhost" {
			continue
		}
		wg.Add(1)
		go func(d Domain) {
			defer wg.Done()
			if helper.IsDomainReachable(d.Domain) {
				normalDomains[d.Domain] = d.Remark
			} else {
				unreachableDomains[d.Domain] = d.Remark
			}
		}(domain)
	}
	wg.Wait()

	for domain, remark := range normalDomains {
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 20 * time.Second}, "tcp", domain+":"+port, conf)
		if err != nil {
			log.Printf("tls.DialWithDialer Error: %v", err)
			return nil, err
		}
		err = conn.VerifyHostname(domain)
		if err != nil {
			log.Printf("conn.VerifyHostname Error: %v", err)
			return nil, err
		}
		expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
		defer conn.Close()

		minorDomainInfos = append(minorDomainInfos, DomainInfo{
			Domain:       domain,
			Remark:       remark,
			ExpiredDate:  expiry.Local().Format("2006-01-02 15:04:05"),
			DaysToExpire: int(time.Until(expiry).Hours() / 24),
		})
	}

	for domain, remark := range unreachableDomains {
		minorDomainInfos = append(minorDomainInfos, DomainInfo{
			Domain:       domain,
			Remark:       remark,
			ExpiredDate:  "unreachable",
			DaysToExpire: -1,
		})
	}

	return minorDomainInfos, nil
}

func UpdateDomainInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var domainOfWebForm []DomainInfo
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
				"type":   "work",
				"domain": domain.Domain,
				"remark": domain.Remark,
			}}
			opts := options.Update().SetUpsert(true)
			_, err := MoniteringDomainsCol.UpdateOne(context.TODO(), filter, update, opts)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("UpdateOne error: %v", err)
				return
			}

			domainsToKeep = append(domainsToKeep, domain.Domain)
		}

		// Remove domains not in domainOfWebForm
		filter := bson.M{"domain": bson.M{"$nin": domainsToKeep}}
		_, err = MoniteringDomainsCol.DeleteMany(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("DeleteMany error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Update GLOBAL domain list successfully!"})
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
