// Package userstatus 封装管理员手动禁用/启用用户的状态变更逻辑。
package userstatus

import (
	"context"
	"log"
	"time"

	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	singboxctl "github.com/xvv6u577/logv2fs/singbox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	StatusPlain    = "plain"
	StatusDisabled = "disabled"
	StatusOverdue  = "overdue"
	legacyDeleted  = "deleted"
)

// NormalizeStatus 读时兼容：历史 deleted 视为 disabled。
func NormalizeStatus(status string) string {
	if status == legacyDeleted {
		return StatusDisabled
	}
	return status
}

type userRecord struct {
	EmailAsId string `bson:"email_as_id"`
	UUID      string `bson:"uuid"`
	UserId    string `bson:"user_id"`
	Name      string `bson:"name"`
	Role      string `bson:"role"`
	Status    string `bson:"status"`
}

func loadUser(ctx context.Context, emailAsId string) (*userRecord, error) {
	var u userRecord
	err := database.GetCollection(model.UserTrafficLogs{}).
		FindOne(ctx, bson.M{"email_as_id": emailAsId}).
		Decode(&u)
	if err != nil {
		return nil, err
	}
	u.Status = NormalizeStatus(u.Status)
	return &u, nil
}

func updateStatus(ctx context.Context, emailAsId, status string) (*userRecord, error) {
	now := time.Now()
	var updated userRecord
	err := database.GetCollection(model.UserTrafficLogs{}).FindOneAndUpdate(
		ctx,
		bson.M{"email_as_id": emailAsId},
		bson.M{"$set": bson.M{"status": status, "updated_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		return nil, err
	}
	updated.Status = NormalizeStatus(updated.Status)
	return &updated, nil
}

// SetUserDisabled 管理员手动禁用：status=disabled + 广播各节点。
func SetUserDisabled(ctx context.Context, emailAsId string) (*userRecord, error) {
	u, err := loadUser(ctx, emailAsId)
	if err != nil {
		return nil, err
	}
	if u.Role == "admin" {
		return nil, errAdminCannotDisable
	}

	updated, err := updateStatus(ctx, emailAsId, StatusDisabled)
	if err != nil {
		return nil, err
	}
	go singboxctl.SafeDisableUser(updated.EmailAsId)
	log.Printf("[userstatus] 用户 %s 已手动禁用", updated.Name)
	return updated, nil
}

// SetUserEnabled 管理员手动启用：status=plain + 广播各节点。
func SetUserEnabled(ctx context.Context, emailAsId string) (*userRecord, error) {
	updated, err := updateStatus(ctx, emailAsId, StatusPlain)
	if err != nil {
		return nil, err
	}
	go singboxctl.SafeEnableUser(singboxctl.AddUserRequest{
		EmailAsId: updated.EmailAsId,
		UUID:      updated.UUID,
		UserId:    updated.UserId,
	})
	log.Printf("[userstatus] 用户 %s 已手动启用", updated.Name)
	return updated, nil
}

var errAdminCannotDisable = &adminDisableError{}

type adminDisableError struct{}

func (e *adminDisableError) Error() string { return "cannot disable admin user" }

// IsAdminDisableError 判断是否为管理员不可禁用错误。
func IsAdminDisableError(err error) bool {
	_, ok := err.(*adminDisableError)
	return ok
}
