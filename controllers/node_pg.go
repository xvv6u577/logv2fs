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
	"gorm.io/gorm"
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

		// 第一步：清空整个 subscription_nodes 表格
		log.Printf("清空 subscription_nodes 表格...")
		if err := db.Exec(`DELETE FROM "subscription_nodes"`).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("清空 subscription_nodes 表格失败: %v", err)
			return
		}
		log.Printf("✅ subscription_nodes 表格已清空")

		// 第二步：逐条插入新提交的数据
		log.Printf("开始插入 %d 个新节点...", len(nodeFromWebForm))
		for i, domain := range nodeFromWebForm {
			// 如果是reality类型，重新分配public_key和short_id
			if domain.Type == "reality" {
				domain.PUBLIC_KEY = PUBLIC_KEY
				domain.SHORT_ID = SHORT_ID
			}

			// 转换为PostgreSQL模型并生成新的UUID
			pgDomain := model.SubscriptionNodePG{
				ID:           uuid.New(), // 为每个节点生成新的UUID
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

			// 直接插入新节点（不需要检查是否存在，因为表格已清空）
			insertQuery := `INSERT INTO "subscription_nodes" 
						(id, type, remark, domain, ip, sni, uuid, path, server_port, password, public_key, short_id, enable_openai, created_at, updated_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
			if err := db.Exec(insertQuery,
				pgDomain.ID, pgDomain.Type, pgDomain.Remark, pgDomain.Domain, pgDomain.IP,
				pgDomain.SNI, pgDomain.UUID, pgDomain.Path, pgDomain.ServerPort, pgDomain.Password,
				pgDomain.PublicKey, pgDomain.ShortID, pgDomain.EnableOpenai, pgDomain.CreatedAt, pgDomain.UpdatedAt).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("插入第 %d 个节点失败 (备注: %s): %v", i+1, domain.Remark, err)
				return
			}
			log.Printf("✅ 成功插入节点 [%d/%d]: %s", i+1, len(nodeFromWebForm), domain.Remark)
		}
		log.Printf("✅ 所有 %d 个节点插入完成", len(nodeFromWebForm))

		// 处理可收集数据的节点
		for _, domain := range dataCollectableNodes {
			// 查找或创建NodeTrafficLogs记录 - 使用GORM方式
			var nodeTrafficLog model.NodeTrafficLogsPG
			result := db.Where("domain_as_id = ?", domain.Domain).First(&nodeTrafficLog)

			// 正确处理GORM的ErrRecordNotFound错误
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				log.Printf("查找节点流量记录失败: %v", result.Error)
				return
			}

			isNewRecord := result.Error == gorm.ErrRecordNotFound

			if isNewRecord {
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

				// 使用GORM创建新记录
				if err := db.Create(&newNodeTrafficLog).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("创建节点流量记录失败 (域名: %s): %v", domain.Domain, err)
					return
				}
				log.Printf("✅ 成功创建节点流量记录: %s (%s)", domain.Domain, domain.Remark)
			} else {
				// 如果存在，则更新记录
				if err := db.Model(&nodeTrafficLog).Updates(map[string]interface{}{
					"remark":     domain.Remark,
					"status":     "active",
					"updated_at": current,
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("更新节点流量记录失败 (域名: %s): %v", domain.Domain, err)
					return
				}
				log.Printf("✅ 成功更新节点流量记录: %s (%s)", domain.Domain, domain.Remark)
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

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Congrats! 已成功清空表格并插入 %d 个节点!", len(nodeFromWebForm))})
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

		query := `SELECT * FROM "subscription_nodes" WHERE type != 'work'`
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

		query := `SELECT * FROM "subscription_nodes" WHERE type = 'work'`
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
			var count int64

			// 首先检查域名是否存在
			if err := db.Model(&model.ExpiryCheckDomainInfoPG{}).Where("domain = ?", domainInfo.Domain).Count(&count).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("检查域名是否存在失败: %v", err)
				return
			}

			// 明确判断是否需要创建新记录
			if count == 0 {
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

				// 使用GORM的Create方法插入新记录
				if err := db.Create(&expiryDomain).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("创建域名记录失败 (域名: %s): %v", domainInfo.Domain, err)
					return
				}
				log.Printf("✅ 成功创建新域名记录: %s (%s)", domainInfo.Domain, domainInfo.Remark)
			} else {
				// 如果存在，则使用GORM的Updates方法更新记录
				if err := db.Model(&model.ExpiryCheckDomainInfoPG{}).Where("domain = ?", domainInfo.Domain).Updates(map[string]interface{}{
					"remark":         domainInfo.Remark,
					"expired_date":   domainInfo.ExpiredDate,
					"days_to_expire": domainInfo.DaysToExpire,
					"updated_at":     time.Now(),
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("更新域名记录失败 (域名: %s): %v", domainInfo.Domain, err)
					return
				}
				log.Printf("✅ 成功更新域名记录: %s (%s)", domainInfo.Domain, domainInfo.Remark)
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

		query := `SELECT * FROM "node_traffic_logs" WHERE status = 'active'`
		if err := db.Raw(query).Scan(&pgNodeTrafficLogs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find active nodes error: %v", err)
			return
		}

		// 转换为前端期望的MongoDB格式
		var mongoFormatNodes []map[string]interface{}
		for _, pgNode := range pgNodeTrafficLogs {
			node := map[string]interface{}{
				"domain_as_id": pgNode.DomainAsId,
				"remark":       pgNode.Remark,
				"status":       pgNode.Status,
				"created_at":   pgNode.CreatedAt,
				"updated_at":   pgNode.UpdatedAt,
				"daily_logs":   pgNode.DailyLogs,
				"monthly_logs": pgNode.MonthlyLogs,
				"yearly_logs":  pgNode.YearlyLogs,
				"hourly_logs":  pgNode.HourlyLogs,
			}
			mongoFormatNodes = append(mongoFormatNodes, node)
		}

		c.JSON(http.StatusOK, mongoFormatNodes)
	}
}

// SaveCustomDatePG 保存节点自定义日期 - PostgreSQL版本
func SaveCustomDatePG() gin.HandlerFunc {
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

		db := database.GetPostgresDB()

		// 更新或插入自定义日期
		query := `
			INSERT INTO "node_custom_dates" (domain_as_id, custom_date, created_at, updated_at)
			VALUES (?, ?, NOW(), NOW())
			ON CONFLICT (domain_as_id) 
			DO UPDATE SET custom_date = EXCLUDED.custom_date, updated_at = NOW()
		`

		if err := db.Exec(query, request.DomainAsId, request.CustomDate).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("保存自定义日期失败: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "自定义日期保存成功"})
	}
}

// GetCustomDatesPG 获取所有节点自定义日期 - PostgreSQL版本
func GetCustomDatesPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()

		// 查询所有自定义日期
		rows, err := db.Raw(`SELECT domain_as_id, custom_date FROM "node_custom_dates"`).Rows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("查询自定义日期失败: %v", err)
			return
		}
		defer rows.Close()

		customDates := make(map[string]string)
		for rows.Next() {
			var domainAsId, customDate string
			if err := rows.Scan(&domainAsId, &customDate); err != nil {
				log.Printf("扫描自定义日期失败: %v", err)
				continue
			}
			customDates[domainAsId] = customDate
		}

		c.JSON(http.StatusOK, customDates)
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
