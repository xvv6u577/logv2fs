package controllers

import (
	"context"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CertTarget 是证书到期监控需要的最小信息。
// 刻意不包含 password / uuid / public_key 等敏感字段。
type CertTarget struct {
	SNI    string `json:"sni"`
	IP     string `json:"ip"`
	Port   string `json:"port"`
	Remark string `json:"remark"`
}

// buildCertTargets 把 AnyTLS 订阅节点转换为证书检查目标，并按 (SNI, Port) 去重。
//   - SNI 采用与客户端一致的规则：SNI 非空用 SNI，否则回退到 Domain；
//   - 端口为空时按 443 处理；
//   - 同一 (SNI, Port) 的多个节点（如 IPv4/IPv6 双栈）只保留第一条，输入应已按 weight 排序；
//   - 输出按 SNI、Port 排序，便于对比与测试。
func buildCertTargets(nodes []model.SubscriptionNode) []CertTarget {
	seen := make(map[string]struct{}, len(nodes))
	targets := make([]CertTarget, 0, len(nodes))
	for _, n := range nodes {
		sni := strings.ToLower(strings.TrimSpace(helper.TLSServerName(n.SNI, n.Domain)))
		if sni == "" {
			continue
		}
		port := strings.TrimSpace(n.SERVER_PORT)
		if port == "" {
			port = "443"
		}
		key := sni + "|" + port
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, CertTarget{
			SNI:    sni,
			IP:     strings.TrimSpace(n.IP),
			Port:   port,
			Remark: n.Remark,
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].SNI != targets[j].SNI {
			return targets[i].SNI < targets[j].SNI
		}
		return targets[i].Port < targets[j].Port
	})
	return targets
}

// GetCertTargets 返回所有 active 的 AnyTLS 节点的证书检查目标。
// 供 cloudflare-logsbfs Worker 定时同步；由 middleware.CertTargetsToken 保护。
func GetCertTargets() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		filter := bson.M{"type": "anytls", "status": "active"}
		opts := options.Find().
			SetProjection(bson.M{"domain": 1, "sni": 1, "ip": 1, "server_port": 1, "remark": 1, "weight": 1}).
			SetSort(bson.D{{Key: "weight", Value: 1}, {Key: "updated_at", Value: -1}})

		cur, err := database.GetCollection(model.SubscriptionNode{}).Find(ctx, filter, opts)
		if err != nil {
			log.Printf("GetCertTargets find error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query nodes"})
			return
		}
		defer cur.Close(ctx)

		var nodes []model.SubscriptionNode
		if err := cur.All(ctx, &nodes); err != nil {
			log.Printf("GetCertTargets cursor error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read nodes"})
			return
		}

		c.JSON(http.StatusOK, buildCertTargets(nodes))
	}
}
