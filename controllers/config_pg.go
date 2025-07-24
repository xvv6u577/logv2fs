package controllers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"gopkg.in/yaml.v2"
)

// PostgreSQL版本的配置函数

// GetSubscripionURLPG 获取订阅URL - PostgreSQL版本
func GetSubscripionURLPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		var subscription []byte
		var err error
		name := helper.SanitizeStr(c.Param("name"))
		db := database.GetPostgresDB()

		var activeGlobalNodes []model.SubscriptionNodePG

		// 查询活跃的全局节点
		if err := db.Where("type != ?", "work").Find(&activeGlobalNodes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while getting active global nodes"})
			log.Printf("Getting active global nodes error: %s", err.Error())
			return
		}

		// 查询用户状态和UUID
		var pgUser model.UserTrafficLogsPG
		if err := db.Select("status, user_id, uuid").Where("email_as_id = ?", name).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("GetSubscripionURL error: %v", err)
			return
		}

		if pgUser.Status == "plain" {
			var sub string
			for _, node := range activeGlobalNodes {
				// 格式化IP地址以支持IPv6
				formattedIP := helper.FormatIPForURL(node.IP)

				if node.Type == "reality" {
					if len(sub) == 0 {
						sub = "vless://" + pgUser.UUID + "@" + formattedIP + ":" + node.ServerPort + "?encryption=none&flow=xtls-rprx-vision&security=reality&sni=itunes.apple.com&fp=chrome&pbk=" + node.PublicKey + "&sid=" + node.ShortID + "&type=tcp&headerType=none#" + node.Remark
					} else {
						sub = sub + "\n" + "vless://" + pgUser.UUID + "@" + formattedIP + ":" + node.ServerPort + "?encryption=none&flow=xtls-rprx-vision&security=reality&sni=itunes.apple.com&fp=chrome&pbk=" + node.PublicKey + "&sid=" + node.ShortID + "&type=tcp&headerType=none#" + node.Remark
					}
				}

				if node.Type == "hysteria2" {
					if len(sub) == 0 {
						sub = "hysteria2://" + pgUser.UserID + "@" + formattedIP + ":" + node.ServerPort + "?insecure=1&sni=bing.com#" + node.Remark
					} else {
						sub = sub + "\n" + "hysteria2://" + pgUser.UserID + "@" + formattedIP + ":" + node.ServerPort + "?insecure=1&sni=bing.com#" + node.Remark
					}
				}

				if node.Type == "vlessCDN" {
					if len(sub) == 0 {
						sub = "vless://" + node.UUID + "@" + formattedIP + ":" + node.ServerPort + "?encryption=none&security=tls&sni=" + node.Domain + "&fp=randomized&type=ws&host=" + node.Domain + "&path=%2F%3Fed%3D2048#" + node.Remark
					} else {
						sub = sub + "\n" + "vless://" + node.UUID + "@" + formattedIP + ":" + node.ServerPort + "?encryption=none&security=tls&sni=" + node.Domain + "&fp=randomized&type=ws&host=" + node.Domain + "&path=%2F%3Fed%3D2048#" + node.Remark
					}
				}
			}

			subscription = []byte(base64.StdEncoding.EncodeToString([]byte(sub)))
		} else {
			subscription, err = os.ReadFile(helper.CurrentPath() + "/config/error.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("GetSubscripionURL error: %v", err)
				// return
			}
		}

		c.Data(http.StatusOK, "text/plain", subscription)
	}
}

// ReturnSingboxJsonPG 返回Singbox JSON配置 - PostgreSQL版本
func ReturnSingboxJsonPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := helper.SanitizeStr(c.Param("name"))
		db := database.GetPostgresDB()

		var err error
		var jsonFile []byte
		var singboxJSON = SingboxJSON{}
		var pgUser model.UserTrafficLogsPG

		// 查询用户状态和UUID
		if err := db.Select("status, user_id, uuid").Where("email_as_id = ?", name).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("ReturnSingboxJson failed: %s", err.Error())
			return
		}

		var activeGlobalNodes []model.SubscriptionNodePG

		// 查询活跃的全局节点
		if err := db.Where("type != ?", "work").Find(&activeGlobalNodes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while getting active global nodes"})
			log.Printf("Getting active global nodes error: %s", err.Error())
			return
		}

		if pgUser.Status == "plain" {
			jsonFile, err = os.ReadFile(helper.CurrentPath() + "/config/template_singbox.json")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = json.Unmarshal(jsonFile, &singboxJSON)
			if err != nil {
				log.Printf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// 添加reality和hysteria2节点到outbounds
			for _, node := range activeGlobalNodes {
				server_port, _ := strconv.Atoi(node.ServerPort)
				var outboundTags = []string{
					"manual-select",
					"auto",
					"WeChat",
					"Apple",
					"Microsoft",
				}

				if node.Type == "reality" {
					for i, outbound := range singboxJSON.Outbounds {
						if outboundMap, ok := outbound.(map[string]interface{}); ok {
							if Contains(outboundTags, outboundMap["tag"].(string)) || (node.EnableOpenai) && outboundMap["tag"] == "Openai" {
								if outbounds, ok := singboxJSON.Outbounds[i].(map[string]interface{}); ok {
									if outboundsList, ok := outbounds["outbounds"].([]interface{}); ok {
										singboxJSON.Outbounds[i].(map[string]interface{})["outbounds"] = append(outboundsList, node.Remark)
									}
								}
							}
						}
					}

					singboxJSON.Outbounds = append(singboxJSON.Outbounds, RealityJSON{
						Tag:            node.Remark,
						Type:           "vless",
						UUID:           pgUser.UUID,
						ServerPort:     server_port,
						Flow:           "xtls-rprx-vision",
						PacketEncoding: "xudp",
						Server:         helper.FormatIPForURL(node.IP),
						TLS: struct {
							Enabled    bool   `json:"enabled"`
							ServerName string `json:"server_name"`
							Utls       struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							} `json:"utls"`
							Reality struct {
								Enabled   bool   `json:"enabled"`
								PublicKey string `json:"public_key"`
								ShortID   string `json:"short_id"`
							} `json:"reality"`
						}{
							Enabled:    true,
							ServerName: "itunes.apple.com",
							Utls: struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							}{
								Enabled:     true,
								Fingerprint: "chrome",
							},
							Reality: struct {
								Enabled   bool   `json:"enabled"`
								PublicKey string `json:"public_key"`
								ShortID   string `json:"short_id"`
							}{
								Enabled:   true,
								PublicKey: node.PublicKey,
								ShortID:   node.ShortID,
							},
						},
					})
				}

				if node.Type == "hysteria2" {
					for i, outbound := range singboxJSON.Outbounds {
						if outboundMap, ok := outbound.(map[string]interface{}); ok {
							if Contains(outboundTags, outboundMap["tag"].(string)) || (node.EnableOpenai) && outboundMap["tag"] == "Openai" {
								if outbounds, ok := singboxJSON.Outbounds[i].(map[string]interface{}); ok {
									if outboundsList, ok := outbounds["outbounds"].([]interface{}); ok {
										singboxJSON.Outbounds[i].(map[string]interface{})["outbounds"] = append(outboundsList, node.Remark)
									}
								}
							}
						}
					}

					singboxJSON.Outbounds = append(singboxJSON.Outbounds, Hysteria2JSON{
						Tag:        node.Remark,
						Type:       "hysteria2",
						Server:     helper.FormatIPForURL(node.IP),
						ServerPort: server_port,
						UpMbps:     100,
						DownMbps:   100,
						Password:   pgUser.UserID,
						TLS: struct {
							Enabled    bool     `json:"enabled"`
							ServerName string   `json:"server_name"`
							Insecure   bool     `json:"insecure"`
							Alpn       []string `json:"alpn"`
						}{
							Enabled:    true,
							ServerName: "bing.com",
							Insecure:   true,
							Alpn:       []string{"h3"},
						},
					})
				}

				if node.Type == "vlessCDN" {
					for i, outbound := range singboxJSON.Outbounds {
						if outboundMap, ok := outbound.(map[string]interface{}); ok {
							if Contains(outboundTags, outboundMap["tag"].(string)) || (node.EnableOpenai) && outboundMap["tag"] == "Openai" {
								if outbounds, ok := singboxJSON.Outbounds[i].(map[string]interface{}); ok {
									if outboundsList, ok := outbounds["outbounds"].([]interface{}); ok {
										singboxJSON.Outbounds[i].(map[string]interface{})["outbounds"] = append(outboundsList, node.Remark)
									}
								}
							}
						}
					}

					singboxJSON.Outbounds = append(singboxJSON.Outbounds, CFVlessJSON{
						Tag:        node.Remark,
						Type:       "vless",
						Server:     helper.FormatIPForURL(node.IP),
						ServerPort: server_port,
						UUID:       node.UUID,
						Flow:       "",
						TLS: struct {
							Enabled    bool   `json:"enabled"`
							ServerName string `json:"server_name"`
							Insecure   bool   `json:"insecure"`
							Utls       struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							} `json:"utls"`
						}{
							Enabled:    true,
							ServerName: node.Domain,
							Insecure:   false,
							Utls: struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							}{
								Enabled:     true,
								Fingerprint: "chrome",
							},
						},
						Multiplex: struct {
							Enabled    bool   `json:"enabled"`
							Protocol   string `json:"protocol"`
							MaxStreams int    `json:"max_streams"`
						}{
							Enabled:    false,
							Protocol:   "smux",
							MaxStreams: 32,
						},
						PacketEncoding: "xudp",
						Transport: struct {
							Type    string `json:"type"`
							Path    string `json:"path"`
							Headers struct {
								Host string `json:"Host"`
							} `json:"headers"`
						}{
							Type: "ws",
							Path: "/?ed=2048",
							Headers: struct {
								Host string `json:"Host"`
							}{
								Host: node.Domain,
							},
						},
					})
				}
			}
		} else {
			jsonFile, err = os.ReadFile(helper.CurrentPath() + "/config/error.json")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = json.Unmarshal(jsonFile, &singboxJSON)
			if err != nil {
				log.Printf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, singboxJSON)
	}
}

// ReturnVergeYAMLPG 返回Verge YAML配置 - PostgreSQL版本
func ReturnVergeYAMLPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := helper.SanitizeStr(c.Param("name"))
		db := database.GetPostgresDB()

		var err error
		var yamlFile []byte
		var singboxYAML = SingboxYAML{}

		// 查询用户状态和UUID
		var pgUser model.UserTrafficLogsPG
		if err := db.Select("status, user_id, uuid").Where("email_as_id = ?", name).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("ReturnVergeYAML failed: %s", err.Error())
			return
		}

		var activeGlobalNodes []model.SubscriptionNodePG

		// 查询活跃的全局节点
		if err := db.Where("type != ?", "work").Find(&activeGlobalNodes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while getting active global nodes"})
			log.Printf("Getting active global nodes error: %s", err.Error())
			return
		}

		if pgUser.Status == "plain" {
			yamlFile, err = os.ReadFile(helper.CurrentPath() + "/config/template_verge.yaml")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = yaml.Unmarshal(yamlFile, &singboxYAML)
			if err != nil {
				log.Printf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// 添加reality和hysteria2节点到outbounds
			for _, node := range activeGlobalNodes {
				server_port, _ := strconv.Atoi(node.ServerPort)
				if node.Type == "reality" {
					for i, proxy := range singboxYAML.ProxyGroups {
						if proxy.Type == "select" || proxy.Type == "url-test" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, node.Remark)
						}
					}

					singboxYAML.Proxies = append(singboxYAML.Proxies, RealityYAML{
						Name:              node.Remark,
						Type:              "vless",
						Server:            helper.FormatIPForURL(node.IP),
						Port:              server_port,
						UUID:              pgUser.UUID,
						Network:           "tcp",
						UDP:               true,
						TLS:               true,
						Flow:              "xtls-rprx-vision",
						Servername:        "itunes.apple.com",
						ClientFingerprint: "chrome",
						RealityOpts: struct {
							PublicKey string `yaml:"public-key"`
							ShortID   string `yaml:"short-id"`
						}{
							PublicKey: node.PublicKey,
							ShortID:   node.ShortID,
						},
					})
				}

				if node.Type == "hysteria2" {
					for i, proxy := range singboxYAML.ProxyGroups {
						if proxy.Type == "select" || proxy.Type == "url-test" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, node.Remark)
						}
					}

					singboxYAML.Proxies = append(singboxYAML.Proxies, Hysteria2YAML{
						Name:           node.Remark,
						Type:           "hysteria2",
						Server:         helper.FormatIPForURL(node.IP),
						Port:           server_port,
						Password:       pgUser.UserID,
						Sni:            "bing.com",
						SkipCertVerify: true,
						Alpn:           []string{"h3"},
					})
				}

				if node.Type == "vlessCDN" {
					for i, proxy := range singboxYAML.ProxyGroups {
						if proxy.Type == "select" || proxy.Type == "url-test" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, node.Remark)
						}
					}

					singboxYAML.Proxies = append(singboxYAML.Proxies, CFVlessYAML{
						Name:              node.Remark,
						Type:              "vless",
						Server:            helper.FormatIPForURL(node.IP),
						Port:              server_port,
						UUID:              node.UUID,
						Network:           "ws",
						TLS:               true,
						UDP:               false,
						Servername:        node.Domain,
						ClientFingerprint: "chrome",
						WsOpts: struct {
							Path    string `yaml:"path"`
							Headers struct {
								Host string `yaml:"Host"`
							} `yaml:"headers"`
						}{
							Path: node.Path,
							Headers: struct {
								Host string `yaml:"Host"`
							}{
								Host: node.Domain,
							},
						},
					})
				}
			}

			// 如果DIRECT类型不在ProxyGroups的select类型的末尾，将其移到末尾
			for i, proxy := range singboxYAML.ProxyGroups {
				if proxy.Type == "select" {
					for j, p := range proxy.Proxies {
						if p == "DIRECT" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies[:j], singboxYAML.ProxyGroups[i].Proxies[j+1:]...)
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, "DIRECT")
						}
					}
				}
			}
		} else {
			yamlFile, err = os.ReadFile(helper.CurrentPath() + "/config/error.yaml")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = yaml.Unmarshal(yamlFile, &singboxYAML)
			if err != nil {
				log.Printf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.YAML(http.StatusOK, singboxYAML)
	}
}
