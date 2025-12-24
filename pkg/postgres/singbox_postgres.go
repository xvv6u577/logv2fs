package postgres_pkg

import (
	"fmt"
	"log"
	"sync"

	"github.com/sagernet/sing-box/option"
	"github.com/xvv6u577/logv2fs/database/postgres"
	"github.com/xvv6u577/logv2fs/model"
)

type (
	UserTrafficLogsPG = model.UserTrafficLogsPG
)

// UpdateOptionsFromPostgreSQL 从 PostgreSQL 数据库更新配置（公共函数）
func UpdateOptionsFromPostgreSQL(opt option.Options) (option.Options, error) {
	log.Println("使用 PostgreSQL 从数据库更新 sing-box 配置...")
	db := postgres.GetPostgresDB()
	if db == nil {
		return opt, fmt.Errorf("PostgreSQL 数据库连接不可用")
	}

	// 查询活跃用户的关键信息
	var pgUsers []UserTrafficLogsPG
	if err := db.Select("email_as_id, status, uuid, user_id").
		Where("status = ?", "plain").
		Find(&pgUsers).Error; err != nil {
		log.Printf("查询 PostgreSQL 用户信息时出错: %v\n", err)
		return opt, err
	}

	if len(pgUsers) > 0 {
		var wg sync.WaitGroup

		for _, user := range pgUsers {
			wg.Add(1)
			go func(user UserTrafficLogsPG) {
				defer wg.Done()

				// 为每个用户添加 VlessUser 和 Hysteria2User 到 opt.Inbounds
				for inbound := range opt.Inbounds {

					var usersToAppend = []string{user.EmailAsId + "-reality", user.EmailAsId + "-hysteria2"}
					opt.Experimental.V2RayAPI.Stats.Users = append(opt.Experimental.V2RayAPI.Stats.Users, usersToAppend...)

					if opt.Inbounds[inbound].Type == "vless" {
						opt.Inbounds[inbound].VLESSOptions.Users = append(opt.Inbounds[inbound].VLESSOptions.Users, option.VLESSUser{
							Name: user.EmailAsId + "-reality",
							UUID: user.UUID,
							Flow: "xtls-rprx-vision",
						})
					}

					if opt.Inbounds[inbound].Type == "hysteria2" {
						opt.Inbounds[inbound].Hysteria2Options.Users = append(opt.Inbounds[inbound].Hysteria2Options.Users, option.Hysteria2User{
							Name:     user.EmailAsId + "-hysteria2",
							Password: user.UserID,
						})
					}
				}

			}(user)
		}

		wg.Wait()
		log.Printf("成功从 PostgreSQL 加载了 %d 个用户的配置", len(pgUsers))
	} else {
		log.Println("PostgreSQL 中没有找到活跃用户")
	}

	return opt, nil
}
