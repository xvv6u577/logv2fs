package helper

import (
	"context"
	"regexp"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UUID 形态的 email_as_id：与前端 adduser.js 一致，crypto.randomUUID() 去掉横杠后的 32 位 hex。
var uuidEmailAsIdRe = regexp.MustCompile(`(?i)^[0-9a-f]{32}$`)

// IsUUIDEmailAsId 判断是否为 UUID 生成的邮箱 ID（32 位十六进制，无横杠）。
func IsUUIDEmailAsId(id string) bool {
	return uuidEmailAsIdRe.MatchString(strings.TrimSpace(id))
}

// NewUUIDEmailAsId 生成与一键添加用户相同格式的 email_as_id。
func NewUUIDEmailAsId() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(u.String(), "-", ""), nil
}

// UserIDFilter 按主 ID 或别名匹配用户（订阅 / 登录 / 管理接口共用）。
func UserIDFilter(id string) bson.M {
	id = strings.TrimSpace(id)
	return bson.M{
		"$or": []bson.M{
			{"email_as_id": id},
			{"email_as_id_aliases": id},
		},
	}
}

// ResolveCanonicalEmailAsId 将主 ID 或别名解析为当前主 email_as_id。
// 找不到时返回原 id 与 err。
func ResolveCanonicalEmailAsId(ctx context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	var doc struct {
		EmailAsId string `bson:"email_as_id"`
	}
	err := database.GetCollection(model.UserTrafficLogs{}).
		FindOne(ctx, UserIDFilter(id), options.FindOne().SetProjection(bson.M{"email_as_id": 1})).
		Decode(&doc)
	if err != nil {
		return id, err
	}
	return doc.EmailAsId, nil
}

// SameUserID 判断两个标识（主 ID 或别名）是否指向同一用户。
func SameUserID(ctx context.Context, a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	var doc struct {
		EmailAsId string   `bson:"email_as_id"`
		Aliases   []string `bson:"email_as_id_aliases"`
	}
	err := database.GetCollection(model.UserTrafficLogs{}).
		FindOne(ctx, UserIDFilter(a), options.FindOne().SetProjection(bson.M{
			"email_as_id":         1,
			"email_as_id_aliases": 1,
		})).Decode(&doc)
	if err != nil {
		return false
	}
	if doc.EmailAsId == b {
		return true
	}
	for _, al := range doc.Aliases {
		if al == b {
			return true
		}
	}
	return false
}
