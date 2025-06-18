package controllers

import (
	"encoding/json"
	"log"
	"net/http"
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
			pgDomain := model.DomainPG{
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
			var existingDomain model.DomainPG
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
		if err := db.Where("remark NOT IN ?", remarks).Delete(&model.DomainPG{}).Error; err != nil {
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
				var domainRecord model.DomainPG
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
						DomainID:    &domainRecord.ID,
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
		var pgDomains []model.DomainPG

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
		var pgDomains []model.DomainPG

		// 查询类型为work的域名
		if err := db.Where("type = ?", "work").Find(&pgDomains).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Find work domains error: %v", err)
			return
		}

		// 转换为Domain格式
		var workDomains []Domain
		for _, pgDomain := range pgDomains {
			workDomains = append(workDomains, Domain{
				Domain: pgDomain.Domain,
				Remark: pgDomain.Remark,
			})
		}

		// 获取域名过期状态
		domainInfos, err := getDomainExpiredStatus(workDomains)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Get domain expired status error: %v", err)
			return
		}

		c.JSON(http.StatusOK, domainInfos)
	}
}

// UpdateDomainInfoPG 更新域名信息 - PostgreSQL版本
func UpdateDomainInfoPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var domainOfWebForm []DomainInfo

		if err := c.BindJSON(&domainOfWebForm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// 跟踪要保留的域名
		domainsToKeep := []string{}

		// 更新或创建域名
		for _, domainInfo := range domainOfWebForm {
			var pgDomain model.DomainPG
			result := db.Where("domain = ?", domainInfo.Domain).First(&pgDomain)

			if result.Error != nil {
				// 如果不存在，则创建新记录
				pgDomain = model.DomainPG{
					ID:        uuid.New(),
					Type:      "work",
					Domain:    domainInfo.Domain,
					Remark:    domainInfo.Remark,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				if err := db.Create(&pgDomain).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Create domain error: %v", err)
					return
				}
			} else {
				// 如果存在，则更新记录
				if err := db.Model(&pgDomain).Updates(map[string]interface{}{
					"type":       "work",
					"remark":     domainInfo.Remark,
					"updated_at": time.Now(),
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Update domain error: %v", err)
					return
				}
			}

			domainsToKeep = append(domainsToKeep, domainInfo.Domain)
		}

		// 删除不在domainOfWebForm中的work类型域名
		if err := db.Where("domain NOT IN ? AND type = ?", domainsToKeep, "work").Delete(&model.DomainPG{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Delete domains error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Update GLOBAL domain list successfully!"})
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
