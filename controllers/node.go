package controllers

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/caster8013/logv2rayfullstack/database"
	helper "github.com/caster8013/logv2rayfullstack/helpers"
	"github.com/caster8013/logv2rayfullstack/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func AddNode() gin.HandlerFunc {
	return func(c *gin.Context) {

		ReturnIfNotAdmin(c)

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var comingDomains, vmessDomains []Domain
		var current = time.Now().Local()

		if err := c.BindJSON(&comingDomains); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// replace ActiveGlobalNodes in globalCollection with comingDomains
		var replacedDocument GlobalVariable
		err := globalCollection.FindOneAndUpdate(ctx,
			bson.M{"name": "GLOBAL"},
			bson.M{"$set": bson.M{"active_global_nodes": comingDomains}},
			options.FindOneAndUpdate().SetReturnDocument(1),
		).Decode(&replacedDocument)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOneAndUpdate error: %v", err)
			return
		}

		// separate vmessDomains from comingDomains.
		for _, domain := range comingDomains {
			if domain.Type == "vmess" {
				vmessDomains = append(vmessDomains, domain)
			}
		}

		// query all nodes by projections, combine domain and status into a map
		var allNodeStatus = map[string]string{}
		var nodeProjections = bson.D{
			{Key: "domain", Value: 1},
			{Key: "status", Value: 1},
		}
		cursor, err := nodeCollection.Find(ctx, bson.M{}, options.Find().SetProjection(nodeProjections))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while querying nodeCollection")
			return
		}

		for cursor.Next(ctx) {
			var t CurrentNode
			err := cursor.Decode(&t)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error occured while decoding nodes")
				return
			}
			allNodeStatus[t.Domain] = t.Status
		}

		// for domain in vmessDomains, if it is not in allNodes, insert new one into nodeCollection; if yes, check if it is inactive, if yes, enable it.
		for _, domain := range vmessDomains {
			// if it is a local mode, only localhost is checked; if it is a remote(main/attached) mode, all remote domain are checked.
			if NODE_TYPE == "local" && domain.Domain != "localhost" {
				continue
			}

			if _, ok := allNodeStatus[domain.Domain]; !ok {
				var node CurrentNode
				node.Domain = domain.Domain
				node.Status = "active"
				node.NodeAtCurrentYear = NodeAtPeriod{
					Period:              current.Format("2006"),
					Amount:              0,
					UserTrafficAtPeriod: map[string]int64{},
				}
				node.NodeAtCurrentMonth = NodeAtPeriod{
					Period:              current.Format("200601"),
					Amount:              0,
					UserTrafficAtPeriod: map[string]int64{},
				}
				node.NodeAtCurrentDay = NodeAtPeriod{
					Period:              current.Format("20060102"),
					Amount:              0,
					UserTrafficAtPeriod: map[string]int64{},
				}
				node.NodeByYear = []NodeAtPeriod{}
				node.NodeByMonth = []NodeAtPeriod{}
				node.NodeByDay = []NodeAtPeriod{}
				node.CreatedAt = current
				node.UpdatedAt = current

				_, err = nodeCollection.InsertOne(ctx, node)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("error occured while inserting node: %v", err)
					return
				}

			}

			if allNodeStatus[domain.Domain] == "inactive" {
				filter := bson.D{primitive.E{Key: "domain", Value: domain.Domain}}
				update := bson.M{"$set": bson.M{"status": "active", "updated_at": time.Now().Local()}}
				_, err = nodeCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("update node status error: %v", err)
					return
				}
			}

		}

		// for node in allNodes, if it is not in domains, set it to inactive.
		for node := range allNodeStatus {
			if !IsDomainInDomainList(node, vmessDomains) {
				filter := bson.D{primitive.E{Key: "domain", Value: node}}
				update := bson.M{"$set": bson.M{"status": "inactive", "updated_at": time.Now().Local()}}
				_, err = nodeCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("error occured while updating node: %v", err)
					return
				}
			}
		}

		var projections = bson.D{
			{Key: "node_global_list", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "role", Value: 1},
			{Key: "email", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "name", Value: 1},
			{Key: "path", Value: 1},
		}
		allUsers, err := database.GetAllUsersPartialInfo(projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		for _, user := range allUsers {

			// fmt.Println("Before: ", user)
			// shadowrocket suburl.
			user.UpdateNodeStatusInUse(comingDomains)
			user.ProduceSuburl(comingDomains)

			// use yamTools.GenerateYAML to generate yaml.
			err = user.GenerateYAML(comingDomains)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			// fmt.Println("After: ", user)
			filter := bson.D{primitive.E{Key: "user_id", Value: user.User_id}}
			update := bson.M{"$set": bson.M{"updated_at": time.Now().Local(), "node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}
			_, err = userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("%v", err.Error())
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Congrats! Nodes updated in success!"})
	}
}

func GetActiveGlobalNodesInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ReturnIfNotAdmin(c)

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var activeGlobalNodesInfo GlobalVariable
		var filter = bson.D{bson.E{Key: "name", Value: "GLOBAL"}}
		err := globalCollection.FindOne(ctx, filter).Decode(&activeGlobalNodesInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOne error: %v", err)
			return
		}

		c.JSON(http.StatusOK, activeGlobalNodesInfo.ActiveGlobalNodes)
	}
}

func GetDomainInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		ReturnIfNotAdmin(c)

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// query GlobalVariable
		var err error
		var domainInfos = []DomainInfo{}
		var tempArray []DomainInfo
		// var foundGlobal GlobalVariable
		// var filter = bson.D{primitive.E{Key: "name", Value: "GLOBAL"}}
		// err := globalCollection.FindOne(ctx, filter).Decode(&foundGlobal)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	log.Printf("FindOne error: %v", err)
		// 	return
		// }
		// tempArray, err := buildDomainInfo(foundGlobal.DomainList, false)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	log.Printf("error occured while parsing domain info: %v", err)
		// 	return
		// }
		// domainInfos = append(domainInfos, tempArray...)

		// query user
		var adminUser model.User
		var userId = ADMINUSERID
		err = userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&adminUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOne error: %v", err)
			return
		}
		tempArray, err = buildDomainInfo(adminUser.NodeGlobalList, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while parsing domain info: %v", err)
			return
		}
		domainInfos = append(domainInfos, tempArray...)

		c.JSON(http.StatusOK, domainInfos)
	}
}

func buildDomainInfo(domains map[string]string, isInUvp bool) ([]DomainInfo, error) {

	var minorDomainInfos []DomainInfo
	var port = "443"
	var conf = &tls.Config{
		// InsecureSkipVerify: true,
	}
	var normalDomains []string
	var unreachableDomains []string

	// split domains into normalDomains and unreachableDomains by parallel processing. if domain is reachable, append it to normalDomains, else append it to unreachableDomains.
	// if domain is localhost, skip it.
	var wg sync.WaitGroup
	for _, domain := range domains {
		if domain == "localhost" {
			continue
		}
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			if helper.IsDomainReachable(domain) {
				normalDomains = append(normalDomains, domain)
			} else {
				unreachableDomains = append(unreachableDomains, domain)
			}
		}(domain)
	}
	wg.Wait()

	for _, domain := range normalDomains {
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 20 * time.Second}, "tcp", domain+":"+port, conf)
		if err != nil {
			log.Printf("tls.DialWithDialer Error: %v", err)
		}
		err = conn.VerifyHostname(domain)
		if err != nil {
			log.Printf("conn.VerifyHostname Error: %v", err)
		}
		expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
		defer conn.Close()

		minorDomainInfos = append(minorDomainInfos, DomainInfo{
			IsInUVP:      isInUvp,
			Domain:       domain,
			ExpiredDate:  expiry.Local().Format("2006-01-02 15:04:05"),
			DaysToExpire: int(time.Until(expiry).Hours() / 24),
		})
	}

	for _, domain := range unreachableDomains {
		minorDomainInfos = append(minorDomainInfos, DomainInfo{
			IsInUVP:      isInUvp,
			Domain:       domain,
			ExpiredDate:  "unreachable",
			DaysToExpire: -1,
		})
	}

	return minorDomainInfos, nil
}

func UpdateDomainInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		ReturnIfNotAdmin(c)

		var tempDomainList map[string]string
		err := c.BindJSON(&tempDomainList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// replace domain_list in GlobalVariable in globalCollection with tempDomainList
		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var replacedDocument GlobalVariable
		err = globalCollection.FindOneAndUpdate(ctx,
			bson.M{"name": "GLOBAL"},
			bson.M{"$set": bson.M{"domain_list": tempDomainList}},
			options.FindOneAndUpdate().SetReturnDocument(1),
		).Decode(&replacedDocument)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOneAndUpdate error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Update GLOBAL domain list successfully!"})
	}
}

func GetNodePartial() gin.HandlerFunc {
	return func(c *gin.Context) {

		ReturnIfNotAdmin(c)

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// query nodeCollection by projection, and return nodePartial with json format
		var nodePartial []*CurrentNode
		var projections = bson.D{
			{Key: "domain", Value: 1},
			{Key: "status", Value: 1},
			{Key: "remark", Value: 1},
			{Key: "node_at_current_day", Value: 1},
			{Key: "node_at_current_month", Value: 1},
			{Key: "node_at_current_year", Value: 1},
			{Key: "node_by_day", Value: 1},
			{Key: "node_by_month", Value: 1},
			{Key: "node_by_year", Value: 1},
			{Key: "created_at", Value: 1},
		}
		cur, err := nodeCollection.Find(ctx, bson.D{}, options.Find().SetProjection(projections))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find error: %v", err)
			return
		}
		for cur.Next(ctx) {
			var node CurrentNode
			err := cur.Decode(&node)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("Decode error: %v", err)
				return
			}
			nodePartial = append(nodePartial, &node)
		}

		c.JSON(http.StatusOK, nodePartial)
	}
}
