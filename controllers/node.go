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

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var domainsOfWebForm, dataCollectableNodes []Domain
		var current = time.Now().Local()

		if err := c.BindJSON(&domainsOfWebForm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// types: reality, hysteria2, vmesstls, vmessws, vlessCDN! if type is reality, reassgin public_key and short_id
		for i, domain := range domainsOfWebForm {
			if domain.Type == "reality" {
				domainsOfWebForm[i].PUBLIC_KEY = PUBLIC_KEY
				domainsOfWebForm[i].SHORT_ID = SHORT_ID
				if !IsDomainInDomainList(domain.Domain, dataCollectableNodes) {
					dataCollectableNodes = append(dataCollectableNodes, domain)
				}
			}
			if domain.Type == "hysteria2" && !IsDomainInDomainList(domain.Domain, dataCollectableNodes) {
				dataCollectableNodes = append(dataCollectableNodes, domain)
			}

			if (domain.Type == "vmesstls" || domain.Type == "vmessws") && !IsDomainInDomainList(domain.Domain, dataCollectableNodes) {
				dataCollectableNodes = append(dataCollectableNodes, domain)
			}
		}

		// replace ActiveGlobalNodes in globalCollection with domainsOfWebForm
		var replacedDocument GlobalVariable
		err := globalCollection.FindOneAndUpdate(ctx,
			bson.M{"name": "GLOBAL"},
			bson.M{"$set": bson.M{
				"active_global_nodes": domainsOfWebForm,
			}},

			options.FindOneAndUpdate().SetReturnDocument(1),
		).Decode(&replacedDocument)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOneAndUpdate error: %v", err)
			return
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

		// for domain in singboxNodes, if it is not in allNodes, insert new one into nodeCollection; if yes, check if it is inactive, if yes, enable it.
		for _, domain := range dataCollectableNodes {

			// if it is a local mode, only localhost is checked; if it is a remote(main/attached) mode, all remote domain are checked.
			// if NODE_TYPE == "local" && domain.Domain != "localhost" {
			// 	continue
			// }

			if _, ok := allNodeStatus[domain.Domain]; ok {
				// if domain is inactive, enable it. else, keep it.
				if allNodeStatus[domain.Domain] == "inactive" {
					var nodeAtCurrentDay, nodeAtCurrentMonth, nodeAtCurrentYear NodeAtPeriod
					var nodeByDay, nodeByMonth, nodeByYear []NodeAtPeriod
					// fetch one node from nodeCollection by domain, and update status, remark, updated_at, node_at_current_year, node_at_current_month, node_at_current_day
					filter := bson.D{primitive.E{Key: "domain", Value: domain.Domain}}
					toBeFixedNode := CurrentNode{}
					err = nodeCollection.FindOne(ctx, filter).Decode(&toBeFixedNode)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						log.Printf("error occured while querying node: %v", err)
						return
					}

					if toBeFixedNode.NodeAtCurrentYear.Period == current.Format("2006") {
						nodeAtCurrentYear = toBeFixedNode.NodeAtCurrentYear
						nodeByYear = toBeFixedNode.NodeByYear
					} else {
						nodeAtCurrentYear = NodeAtPeriod{
							Period:              current.Format("2006"),
							Amount:              0,
							UserTrafficAtPeriod: map[string]int64{},
						}
						nodeByYear = append(toBeFixedNode.NodeByYear, toBeFixedNode.NodeAtCurrentYear)
					}

					if toBeFixedNode.NodeAtCurrentMonth.Period == current.Format("200601") {
						nodeAtCurrentMonth = toBeFixedNode.NodeAtCurrentMonth
						nodeByMonth = toBeFixedNode.NodeByMonth
					} else {
						nodeAtCurrentMonth = NodeAtPeriod{
							Period:              current.Format("200601"),
							Amount:              0,
							UserTrafficAtPeriod: map[string]int64{},
						}
						nodeByMonth = append(toBeFixedNode.NodeByMonth, toBeFixedNode.NodeAtCurrentMonth)
					}

					if toBeFixedNode.NodeAtCurrentDay.Period == current.Format("20060102") {
						nodeAtCurrentDay = toBeFixedNode.NodeAtCurrentDay
						nodeByDay = toBeFixedNode.NodeByDay
					} else {
						nodeAtCurrentDay = NodeAtPeriod{
							Period:              current.Format("20060102"),
							Amount:              0,
							UserTrafficAtPeriod: map[string]int64{},
						}
						nodeByDay = append(toBeFixedNode.NodeByDay, toBeFixedNode.NodeAtCurrentDay)
					}

					update := bson.M{"$set": bson.M{
						"status":                "active",
						"updated_at":            time.Now().Local(),
						"remark":                domain.Remark,
						"ip":                    domain.IP,
						"node_at_current_year":  nodeAtCurrentYear,
						"node_at_current_month": nodeAtCurrentMonth,
						"node_at_current_day":   nodeAtCurrentDay,
						"node_by_year":          nodeByYear,
						"node_by_month":         nodeByMonth,
						"node_by_day":           nodeByDay,
					}}

					_, err = nodeCollection.UpdateOne(ctx, filter, update)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						log.Printf("update node status error: %v", err)
						return
					}
				}
			} else {
				var node CurrentNode
				node.Domain = domain.Domain
				node.Status = "active"
				node.Remark = domain.Remark
				node.IP = domain.IP
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
		}

		// for nodes in allNodes, if it is not in domains, set it to inactive.
		for domain := range allNodeStatus {
			if !IsDomainInDomainList(domain, dataCollectableNodes) {
				filter := bson.D{primitive.E{Key: "domain", Value: domain}}
				update := bson.M{"$set": bson.M{"status": "inactive", "updated_at": time.Now().Local()}}
				_, err = nodeCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("error occured while updating node: %v", err)
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Congrats! Nodes updated in success!"})
	}
}

func GetActiveGlobalNodesInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var domainInfos []DomainInfo
		var foundGlobal GlobalVariable
		var filter = bson.D{primitive.E{Key: "name", Value: "GLOBAL"}}
		err := globalCollection.FindOne(ctx, filter).Decode(&foundGlobal)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOne error: %v", err)
			return
		}
		domainInfos, err = getDomainExpiredStatus(foundGlobal.WorkRelatedDomainList)
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

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var comingDomainList []DomainInfo
		var workRelatedDomainList []Domain
		err := c.BindJSON(&comingDomainList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		for _, domain := range comingDomainList {
			workRelatedDomainList = append(workRelatedDomainList, Domain{
				Type:   "work",
				Domain: domain.Domain,
				Remark: domain.Remark,
			})
		}

		var replacedDocument GlobalVariable
		err = globalCollection.FindOneAndUpdate(ctx,
			bson.M{"name": "GLOBAL"},
			bson.M{"$set": bson.M{"work_related_domain_list": workRelatedDomainList}},
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

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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
