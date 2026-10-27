package migrateemailid

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	singboxctl "github.com/xvv6u577/logv2fs/singbox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMigrateEmailIDCmd 将非 UUID 形态的 email_as_id 换为 UUID，并把旧 ID 写入别名。
func NewMigrateEmailIDCmd() *cobra.Command {
	var dryRun bool
	var broadcast bool
	var backupDir string

	cmd := &cobra.Command{
		Use:   "migrate-email-as-id",
		Short: "为非 UUID 的 email_as_id 生成 UUID，旧 ID 保留为别名",
		Long: `一次性数据迁移：
  - 跳过已是 32 位 hex（UUID 去横杠）的用户
  - 其余用户：email_as_id ← 新 UUID；旧值写入 email_as_id_aliases
  - 同步把 user_traffic_periods.email_as_id 从旧值改到新值
  - 默认 --dry-run=true；真正执行加 --dry-run=false
  - --broadcast=true 时对每个用户 SafeRemoveUser(旧)+SafeAddUser(新)（需节点在线）
  - 执行前会把受影响文档备份为 JSON

注意：生产 httpserver 必须已部署支持别名查找的版本后，旧订阅 URL 才会在改主 ID 后继续可用。`,
		Run: func(cmd *cobra.Command, args []string) {
			run(dryRun, broadcast, backupDir)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "仅列出将迁移的用户与新旧映射，不写库")
	cmd.Flags().BoolVar(&broadcast, "broadcast", false, "写库后向各节点广播 RemoveUser(旧)+AddUser(新)")
	cmd.Flags().StringVar(&backupDir, "backup-dir", "", "备份 JSON 目录；默认 ./migrate-email-as-id-backup-<timestamp>")
	return cmd
}

type mapping struct {
	Name    string `json:"name"`
	OldID   string `json:"old_email_as_id"`
	NewID   string `json:"new_email_as_id"`
	UserID  string `json:"user_id"`
	UUID    string `json:"uuid"`
	Status  string `json:"status"`
	Aliases []string `json:"existing_aliases,omitempty"`
}

func run(dryRun, broadcast bool, backupDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	col := database.GetCollection(model.UserTrafficLogs{})
	cur, err := col.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{
		"email_as_id":         1,
		"email_as_id_aliases": 1,
		"name":                1,
		"user_id":             1,
		"uuid":                1,
		"status":              1,
	}))
	if err != nil {
		log.Fatalf("列出用户失败: %v", err)
	}
	defer cur.Close(ctx)

	var all []model.UserTrafficLogs
	if err := cur.All(ctx, &all); err != nil {
		log.Fatalf("读取用户失败: %v", err)
	}

	used := map[string]struct{}{}
	for _, u := range all {
		used[u.Email_As_Id] = struct{}{}
		for _, a := range u.EmailAsIdAliases {
			used[a] = struct{}{}
		}
	}

	var maps []mapping
	for _, u := range all {
		if helper.IsUUIDEmailAsId(u.Email_As_Id) {
			continue
		}
		newID, err := allocUniqueID(used)
		if err != nil {
			log.Fatalf("生成 UUID 失败: %v", err)
		}
		used[newID] = struct{}{}
		maps = append(maps, mapping{
			Name:    u.Name,
			OldID:   u.Email_As_Id,
			NewID:   newID,
			UserID:  u.User_id,
			UUID:    u.UUID,
			Status:  u.Status,
			Aliases: append([]string{}, u.EmailAsIdAliases...),
		})
	}

	mode := "DRY-RUN"
	if !dryRun {
		mode = "REAL-RUN"
	}
	log.Printf("=== migrate-email-as-id [%s] 非 UUID 用户: %d / 总用户 %d ===", mode, len(maps), len(all))
	for _, m := range maps {
		log.Printf("  %s  %s  ->  %s", m.Name, m.OldID, m.NewID)
	}

	if len(maps) == 0 {
		log.Printf("无需迁移")
		return
	}

	if backupDir == "" {
		backupDir = fmt.Sprintf("migrate-email-as-id-backup-%s", time.Now().Format("20060102-150405"))
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		log.Fatalf("创建备份目录失败: %v", err)
	}
	backupPath := filepath.Join(backupDir, "mappings.json")
	raw, _ := json.MarshalIndent(maps, "", "  ")
	if err := os.WriteFile(backupPath, raw, 0o644); err != nil {
		log.Fatalf("写备份失败: %v", err)
	}
	log.Printf("映射已写入 %s", backupPath)

	// 额外备份完整受影响文档
	var docs []bson.M
	oldIDs := make([]string, 0, len(maps))
	for _, m := range maps {
		oldIDs = append(oldIDs, m.OldID)
	}
	dcur, err := col.Find(ctx, bson.M{"email_as_id": bson.M{"$in": oldIDs}})
	if err == nil {
		_ = dcur.All(ctx, &docs)
		dcur.Close(ctx)
		docPath := filepath.Join(backupDir, "USER_TRAFFIC_LOGS.affected.json")
		raw, _ = json.MarshalIndent(docs, "", "  ")
		_ = os.WriteFile(docPath, raw, 0o644)
		log.Printf("受影响文档备份: %s (%d)", docPath, len(docs))
	}

	if dryRun {
		log.Printf("提示：要真正执行请加 --dry-run=false（务必先部署支持别名查找的 httpserver）")
		return
	}

	periodCol := database.GetCollection(model.UserTrafficPeriod{})
	now := time.Now()
	ok, fail := 0, 0
	for _, m := range maps {
		if err := migrateOne(ctx, col, periodCol, m, now, broadcast); err != nil {
			log.Printf("FAIL %s (%s): %v", m.Name, m.OldID, err)
			fail++
			continue
		}
		log.Printf("OK   %s  %s -> %s", m.Name, m.OldID, m.NewID)
		ok++
	}
	log.Printf("迁移完成: ok=%d fail=%d backup=%s", ok, fail, backupDir)
}

func allocUniqueID(used map[string]struct{}) (string, error) {
	for i := 0; i < 20; i++ {
		id, err := helper.NewUUIDEmailAsId()
		if err != nil {
			return "", err
		}
		if _, exists := used[id]; exists {
			continue
		}
		// 再查库防竞态
		n, err := database.GetCollection(model.UserTrafficLogs{}).CountDocuments(
			context.Background(), helper.UserIDFilter(id))
		if err != nil {
			return "", err
		}
		if n == 0 {
			return id, nil
		}
	}
	return "", fmt.Errorf("unable to allocate unique uuid email_as_id")
}

func migrateOne(ctx context.Context, col, periodCol *mongo.Collection, m mapping, now time.Time, broadcast bool) error {
	// 1) 改主文档：主 ID + 把旧 ID 放进别名（保留已有别名）
	res, err := col.UpdateOne(ctx,
		bson.M{"email_as_id": m.OldID},
		bson.M{
			"$set": bson.M{
				"email_as_id": m.NewID,
				"updated_at":  now,
			},
			"$addToSet": bson.M{"email_as_id_aliases": m.OldID},
		},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	// 2) 周期流量键同步改名
	if _, err := periodCol.UpdateMany(ctx,
		bson.M{"email_as_id": m.OldID},
		bson.M{"$set": bson.M{"email_as_id": m.NewID, "updated_at": now}},
	); err != nil {
		return fmt.Errorf("remap periods: %w", err)
	}

	// 3) 可选：节点用户名切到新主 ID（协议 UUID/user_id 不变，连接不受影响）
	if broadcast {
		singboxctl.SafeRemoveUser(m.OldID)
		singboxctl.SafeAddUser(singboxctl.AddUserRequest{
			EmailAsId: m.NewID,
			UUID:      m.UUID,
			UserId:    m.UserID,
		})
	}
	return nil
}
