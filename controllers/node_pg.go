package controllers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
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

			// 使用原始SQL查询替代GORM的高级API，避免prepared statement缓存问题
			var existingDomain model.SubscriptionNodePG
			query := `SELECT * FROM "subscritption_nodes" WHERE remark = $1 LIMIT 1`
			result := db.Raw(query, domain.Remark).Scan(&existingDomain)

			if result.Error != nil {
				// 如果不存在，则使用原始SQL创建新节点
				pgDomain.ID = uuid.New()
				insertQuery := `INSERT INTO "subscritption_nodes" 
							(id, type, remark, domain, ip, sni, uuid, path, server_port, password, public_key, short_id, enable_openai, created_at, updated_at) 
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
				if err := db.Exec(insertQuery,
					pgDomain.ID, pgDomain.Type, pgDomain.Remark, pgDomain.Domain, pgDomain.IP,
					pgDomain.SNI, pgDomain.UUID, pgDomain.Path, pgDomain.ServerPort, pgDomain.Password,
					pgDomain.PublicKey, pgDomain.ShortID, pgDomain.EnableOpenai, pgDomain.CreatedAt, pgDomain.UpdatedAt).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Create domain error: %v", err)
					return
				}
			} else {
				// 如果存在，则使用原始SQL更新节点
				updateQuery := `UPDATE "subscritption_nodes" 
							   SET type = $1, domain = $2, ip = $3, sni = $4, uuid = $5, 
							   path = $6, server_port = $7, password = $8, public_key = $9, 
							   short_id = $10, enable_openai = $11, updated_at = $12
							   WHERE id = $13`
				if err := db.Exec(updateQuery,
					pgDomain.Type, pgDomain.Domain, pgDomain.IP, pgDomain.SNI, pgDomain.UUID,
					pgDomain.Path, pgDomain.ServerPort, pgDomain.Password, pgDomain.PublicKey,
					pgDomain.ShortID, pgDomain.EnableOpenai, current, existingDomain.ID).Error; err != nil {
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

		// 删除不在提交列表中的节点 - 使用原始SQL
		if len(remarks) > 0 {
			// 构建域名列表的字符串表示
			placeholders := make([]string, len(remarks))
			args := make([]interface{}, len(remarks))
			for i, remark := range remarks {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				args[i] = remark
			}

			// 使用原始SQL删除不在列表中的节点
			query := fmt.Sprintf(`DELETE FROM "subscritption_nodes" WHERE remark NOT IN (%s)`, strings.Join(placeholders, ","))
			if err := db.Exec(query, args...).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("Delete domains error: %v", err)
				return
			}
		}

		// 处理可收集数据的节点
		for _, domain := range dataCollectableNodes {
			// 查找或创建NodeTrafficLogs记录 - 使用原始SQL
			var nodeTrafficLog model.NodeTrafficLogsPG
			query := `SELECT * FROM "node_traffic_logs" WHERE domain_as_id = $1 LIMIT 1`
			result := db.Raw(query, domain.Domain).Scan(&nodeTrafficLog)

			if result.Error != nil {
				// 如果不存在，则创建新记录
				// 查找对应的Domain记录 - 使用原始SQL
				var domainRecord model.SubscriptionNodePG
				domainQuery := `SELECT * FROM "subscritption_nodes" WHERE domain = $1 LIMIT 1`
				if err := db.Raw(domainQuery, domain.Domain).Scan(&domainRecord).Error; err == nil {
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

					// 使用原始SQL创建新记录
					insertQuery := `INSERT INTO "node_traffic_logs" 
								(id, domain_as_id, remark, status, created_at, updated_at, hourly_logs, daily_logs, monthly_logs, yearly_logs) 
								VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
					if err := db.Exec(insertQuery,
						newNodeTrafficLog.ID, newNodeTrafficLog.DomainAsId, newNodeTrafficLog.Remark,
						newNodeTrafficLog.Status, newNodeTrafficLog.CreatedAt, newNodeTrafficLog.UpdatedAt,
						newNodeTrafficLog.HourlyLogs, newNodeTrafficLog.DailyLogs,
						newNodeTrafficLog.MonthlyLogs, newNodeTrafficLog.YearlyLogs).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						log.Printf("Create node traffic log error: %v", err)
						return
					}
				}
			} else {
				// 如果存在，则使用原始SQL更新记录
				updateQuery := `UPDATE "node_traffic_logs" 
							   SET remark = $1, status = $2, updated_at = $3
							   WHERE id = $4`
				if err := db.Exec(updateQuery,
					domain.Remark, "active", current, nodeTrafficLog.ID).Error; err != nil {
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

		// 使用原始SQL更新不在列表中的节点状态
		if len(domainAsIds) > 0 {
			// 构建域名列表的字符串表示
			placeholders := make([]string, len(domainAsIds))
			args := make([]interface{}, len(domainAsIds))
			for i, domain := range domainAsIds {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				args[i] = domain
			}

			// 使用原始SQL更新不在列表中的节点状态
			query := fmt.Sprintf(`UPDATE "node_traffic_logs" SET status = 'inactive', updated_at = $%d WHERE domain_as_id NOT IN (%s)`,
				len(args)+1, strings.Join(placeholders, ","))
			args = append(args, current)

			if err := db.Exec(query, args...).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("Update inactive nodes error: %v", err)
				return
			}
		} else {
			// 如果没有域名要保留，则将所有节点设为inactive
			if err := db.Exec(`UPDATE "node_traffic_logs" SET status = 'inactive', updated_at = $1`, current).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("Update all nodes to inactive error: %v", err)
				return
			}
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

		// 使用原始SQL查询替代GORM的高级API，避免prepared statement缓存问题
		query := `SELECT * FROM "subscritption_nodes" WHERE type != 'work'`
		if err := db.Raw(query).Scan(&pgDomains).Error; err != nil {
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

		// 使用原始SQL查询替代GORM的高级API，避免prepared statement缓存问题
		query := `SELECT * FROM "subscritption_nodes" WHERE type = 'work'`
		if err := db.Raw(query).Scan(&pgDomains).Error; err != nil {
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
			// 使用原始SQL查询替代GORM的高级API，避免prepared statement缓存问题
			query := `SELECT * FROM "expiry_check_domains" WHERE domain = $1 LIMIT 1`
			result := db.Raw(query, domainInfo.Domain).Scan(&expiryDomain)

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

				// 使用原始SQL插入新记录
				query := `INSERT INTO "expiry_check_domains" 
						(id, domain, remark, expired_date, days_to_expire, created_at, updated_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7)`
				if err := db.Exec(query,
					expiryDomain.ID,
					expiryDomain.Domain,
					expiryDomain.Remark,
					expiryDomain.ExpiredDate,
					expiryDomain.DaysToExpire,
					expiryDomain.CreatedAt,
					expiryDomain.UpdatedAt).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Create domain error: %v", err)
					return
				}
			} else {
				// 如果存在，则使用原始SQL更新记录
				updateQuery := `UPDATE "expiry_check_domains" 
							   SET remark = $1, expired_date = $2, days_to_expire = $3, updated_at = $4
							   WHERE domain = $5`
				if err := db.Exec(updateQuery,
					domainInfo.Remark,
					domainInfo.ExpiredDate,
					domainInfo.DaysToExpire,
					time.Now(),
					domainInfo.Domain).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Update domain error: %v", err)
					return
				}
			}

			domainsToKeep = append(domainsToKeep, domainInfo.Domain)
		}

		// 删除不在domainOfWebForm中的域名 - 使用原始SQL
		if len(domainsToKeep) > 0 {
			// 构建域名列表的字符串表示
			placeholders := make([]string, len(domainsToKeep))
			args := make([]interface{}, len(domainsToKeep))
			for i, domain := range domainsToKeep {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				args[i] = domain
			}

			// 使用原始SQL删除不在列表中的域名
			query := fmt.Sprintf(`DELETE FROM "expiry_check_domains" WHERE domain NOT IN (%s)`, strings.Join(placeholders, ","))
			if err := db.Exec(query, args...).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("Delete domains error: %v", err)
				return
			}
		} else {
			// 如果没有域名要保留，则删除所有域名
			if err := db.Exec(`DELETE FROM "expiry_check_domains"`).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("Delete all domains error: %v", err)
				return
			}
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

		// 使用原始SQL查询替代GORM的高级API，避免prepared statement缓存问题
		query := `SELECT * FROM "node_traffic_logs" WHERE status = 'active'`
		if err := db.Raw(query).Scan(&pgNodeTrafficLogs).Error; err != nil {
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

		// 使用原始SQL查询替代GORM的高级API，避免prepared statement缓存问题
		query := `SELECT * FROM "expiry_check_domains"`
		if err := db.Raw(query).Scan(&expiryDomains).Error; err != nil {
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
