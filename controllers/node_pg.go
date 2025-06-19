package controllers

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
)

// PostgreSQL版本的节点操作函数

// AddNodePG 添加节点 - PostgreSQL版本
func AddNodePG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var current = time.Now().Local()
		var nodeFromWebForm []Domain
		var dataCollectableNodes []Domain
		db := database.GetPostgresDB()

		if err := c.BindJSON(&nodeFromWebForm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// 移除重复的域名
		dataCollectableNodes = removeDuplicateDomains(nodeFromWebForm)

		// 处理每个节点
		for _, domain := range nodeFromWebForm {
			// 如果是reality类型，重新分配public_key和short_id
			if domain.Type == "reality" {
				domain.PUBLIC_KEY = PUBLIC_KEY
				domain.SHORT_ID = SHORT_ID
			}

			// 转换为PostgreSQL模型
			pgDomain := model.SubscriptionNodePG{
				Type:         domain.Type,
				Remark:       domain.Remark,
				Domain:       domain.Domain,
				IP:           domain.IP,
				SNI:          domain.SNI,
				UUID:         domain.UUID,
				Path:         domain.PATH,
				ServerPort:   domain.SERVER_PORT,
				Password:     domain.PASSWORD,
				PublicKey:    domain.PUBLIC_KEY,
				ShortID:      domain.SHORT_ID,
				EnableOpenai: domain.EnableOpenai,
				CreatedAt:    current,
				UpdatedAt:    current,
			}

			// 使用Remark作为过滤条件，检查节点是否存在
			var existingDomain model.SubscriptionNodePG
			result := db.Where("remark = ?", domain.Remark).First(&existingDomain)

			if result.Error != nil {
				// 如果不存在，则创建新节点
				pgDomain.ID = uuid.New()
				if err := db.Create(&pgDomain).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Create domain error: %v", err)
					return
				}
			} else {
				// 如果存在，则更新节点
				if err := db.Model(&existingDomain).Updates(map[string]interface{}{
					"type":          pgDomain.Type,
					"domain":        pgDomain.Domain,
					"ip":            pgDomain.IP,
					"sni":           pgDomain.SNI,
					"uuid":          pgDomain.UUID,
					"path":          pgDomain.Path,
					"server_port":   pgDomain.ServerPort,
					"password":      pgDomain.Password,
					"public_key":    pgDomain.PublicKey,
					"short_id":      pgDomain.ShortID,
					"enable_openai": pgDomain.EnableOpenai,
					"updated_at":    current,
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Update domain error: %v", err)
					return
				}
			}
		}

		// 获取所有节点的Remark列表
		remarks := make([]string, len(nodeFromWebForm))
		for i, domain := range nodeFromWebForm {
			remarks[i] = domain.Remark
		}

		// 删除不在提交列表中的节点
		if err := db.Where("remark NOT IN ?", remarks).Delete(&model.SubscriptionNodePG{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Delete domains error: %v", err)
			return
		}

		// 处理可收集数据的节点
		for _, domain := range dataCollectableNodes {
			// 查找或创建NodeTrafficLogs记录
			var nodeTrafficLog model.NodeTrafficLogsPG
			result := db.Where("domain_as_id = ?", domain.Domain).First(&nodeTrafficLog)

			if result.Error != nil {
				// 如果不存在，则创建新记录
				// 查找对应的Domain记录
				var domainRecord model.SubscriptionNodePG
				if err := db.Where("domain = ?", domain.Domain).First(&domainRecord).Error; err == nil {
					// 创建新的NodeTrafficLog记录
					emptyHourlyLogs, _ := json.Marshal([]model.TrafficLogEntry{})
					emptyDailyLogs, _ := json.Marshal([]model.DailyLogEntry{})
					emptyMonthlyLogs, _ := json.Marshal([]model.MonthlyLogEntry{})
					emptyYearlyLogs, _ := json.Marshal([]model.YearlyLogEntry{})

					newNodeTrafficLog := model.NodeTrafficLogsPG{
						ID:          uuid.New(),
						DomainAsId:  domain.Domain,
						Remark:      domain.Remark,
						Status:      "active",
						CreatedAt:   current,
						UpdatedAt:   current,
						HourlyLogs:  emptyHourlyLogs,
						DailyLogs:   emptyDailyLogs,
						MonthlyLogs: emptyMonthlyLogs,
						YearlyLogs:  emptyYearlyLogs,
					}

					if err := db.Create(&newNodeTrafficLog).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						log.Printf("Create node traffic log error: %v", err)
						return
					}
				}
			} else {
				// 如果存在，则更新记录
				if err := db.Model(&nodeTrafficLog).Updates(map[string]interface{}{
					"remark":     domain.Remark,
					"status":     "active",
					"updated_at": current,
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Update node traffic log error: %v", err)
					return
				}
			}
		}

		// 将不在dataCollectableNodes中的NodeTrafficLogs状态设置为"inactive"
		domainAsIds := make([]string, len(dataCollectableNodes))
		for i, domain := range dataCollectableNodes {
			domainAsIds[i] = domain.Domain
		}

		if err := db.Model(&model.NodeTrafficLogsPG{}).
			Where("domain_as_id NOT IN ?", domainAsIds).
			Update("status", "inactive").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Update inactive nodes error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Congrats! Nodes updated in success!"})
	}
}

// GetActiveGlobalNodesPG 获取活跃的全局节点 - PostgreSQL版本
func GetActiveGlobalNodesPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var pgDomains []model.SubscriptionNodePG

		// 查询非work类型的所有域名
		if err := db.Where("type != ?", "work").Find(&pgDomains).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find domains error: %v", err)
			return
		}

		// 转换为API响应格式
		var activeNodes []Domain
		for _, pgDomain := range pgDomains {
			activeNodes = append(activeNodes, Domain{
				Type:         pgDomain.Type,
				Remark:       pgDomain.Remark,
				Domain:       pgDomain.Domain,
				IP:           pgDomain.IP,
				SNI:          pgDomain.SNI,
				UUID:         pgDomain.UUID,
				PATH:         pgDomain.Path,
				SERVER_PORT:  pgDomain.ServerPort,
				PASSWORD:     pgDomain.Password,
				PUBLIC_KEY:   pgDomain.PublicKey,
				SHORT_ID:     pgDomain.ShortID,
				EnableOpenai: pgDomain.EnableOpenai,
			})
		}

		c.JSON(http.StatusOK, activeNodes)
	}
}

// GetWorkDomainInfoPG 获取工作域名信息 - PostgreSQL版本
func GetWorkDomainInfoPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var pgDomains []model.SubscriptionNodePG

		// 查询类型为work的域名
		if err := db.Where("type = ?", "work").Find(&pgDomains).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find work domains error: %v", err)
			return
		}

		// 准备结果数组和域名分类映射
		var domainInfos []ExpiryCheckDomainInfo
		normalDomains := make(map[string]string)
		unreachableDomains := make(map[string]string)

		// 并行处理域名可达性检查
		var wg sync.WaitGroup
		for _, pgDomain := range pgDomains {
			if pgDomain.Domain == "localhost" {
				continue
			}
			wg.Add(1)
			go func(d model.SubscriptionNodePG) {
				defer wg.Done()
				if helper.IsDomainReachable(d.Domain) {
					normalDomains[d.Domain] = d.Remark
				} else {
					unreachableDomains[d.Domain] = d.Remark
				}
			}(pgDomain)
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

// UpdateExpiryCheckDomainsInfoPG 更新域名信息 - PostgreSQL版本
func UpdateExpiryCheckDomainsInfoPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var domainOfWebForm []ExpiryCheckDomainInfo

		if err := c.BindJSON(&domainOfWebForm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// 跟踪要保留的域名
		domainsToKeep := []string{}

		// 更新或创建域名
		for _, domainInfo := range domainOfWebForm {
			var expiryDomain model.ExpiryCheckDomainInfoPG
			result := db.Where("domain = ?", domainInfo.Domain).First(&expiryDomain)

			if result.Error != nil {
				// 如果不存在，则创建新记录
				expiryDomain = model.ExpiryCheckDomainInfoPG{
					ID:           uuid.New(),
					Domain:       domainInfo.Domain,
					Remark:       domainInfo.Remark,
					ExpiredDate:  domainInfo.ExpiredDate,
					DaysToExpire: domainInfo.DaysToExpire,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}

				if err := db.Create(&expiryDomain).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Create domain error: %v", err)
					return
				}
			} else {
				// 如果存在，则更新记录
				if err := db.Model(&expiryDomain).Updates(map[string]interface{}{
					"remark":         domainInfo.Remark,
					"expired_date":   domainInfo.ExpiredDate,
					"days_to_expire": domainInfo.DaysToExpire,
					"updated_at":     time.Now(),
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Update domain error: %v", err)
					return
				}
			}

			domainsToKeep = append(domainsToKeep, domainInfo.Domain)
		}

		// 删除不在domainOfWebForm中的域名
		if err := db.Where("domain NOT IN ?", domainsToKeep).Delete(&model.ExpiryCheckDomainInfoPG{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Delete domains error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Update expiry check domains list successfully!"})
	}
}

// GetSingboxNodesPG 获取Singbox节点 - PostgreSQL版本
func GetSingboxNodesPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var pgNodeTrafficLogs []model.NodeTrafficLogsPG

		// 查询状态为active的节点
		if err := db.Where("status = ?", "active").Find(&pgNodeTrafficLogs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find active nodes error: %v", err)
			return
		}

		c.JSON(http.StatusOK, pgNodeTrafficLogs)
	}
}

// GetDomainsExpiryInfoPG 获取域名证书过期信息 - PostgreSQL版本
func GetDomainsExpiryInfoPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var expiryDomains []model.ExpiryCheckDomainInfoPG

		// 查询所有需要检查过期的域名
		if err := db.Find(&expiryDomains).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find expiry domains error: %v", err)
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
			go func(d model.ExpiryCheckDomainInfoPG) {
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
