package model

type NotificationType string

const (
	NotificationTypePushPlus NotificationType = "push_plus"
	NotificationTypeEmail    NotificationType = "email"
)

type SystemSettings struct {
	ID                         int64            `gorm:"primaryKey;autoIncrement;comment:表主键ID" json:"id"`
	OfflineNotificationEnabled bool             `gorm:"default:false;comment:启用离线通知" json:"offline_notification_enabled"`
	NotificationType           NotificationType `gorm:"type:enum('push_plus','email');default:'push_plus';not null;comment:通知方式：push_plus-推送加，email-邮件" json:"notification_type"`
	PushPlusURL                string           `gorm:"type:varchar(255);default:'';comment:Push Plus的URL" json:"push_plus_url"`
	PushPlusToken              string           `gorm:"type:varchar(255);default:'';comment:Push Plus的Token" json:"push_plus_token"`
	AutoVerifyUser             bool             `gorm:"default:false;comment:自动通过好友验证" json:"auto_verify_user"`
	VerifyUserDelay            int              `gorm:"default:60;comment:自动通过好友验证延迟时间(秒)" json:"verify_user_delay"`
	AutoChatroomInvite         bool             `gorm:"default:false;comment:自动邀请进群" json:"auto_chatroom_invite"`
	CreatedAt                  int64            `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt                  int64            `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (SystemSettings) TableName() string {
	return "system_settings"
}
