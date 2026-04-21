package model

import "gorm.io/datatypes"

type NotificationType string

const (
	NotificationTypePushPlus      NotificationType = "push_plus"
	NotificationTypeEmail         NotificationType = "email"
	NotificationTypeWechatWorkApp NotificationType = "wechat_work_app"
)

type SystemSettings struct {
	ID                         int64             `gorm:"column:id;primaryKey;autoIncrement;comment:表主键ID" json:"id"`
	WebhookURL                 *string           `gorm:"column:webhook_url;type:varchar(255);default:'';comment:Webhook地址" json:"webhook_url"`
	WebhookHeaders             datatypes.JSONMap `gorm:"column:webhook_headers;type:json;comment:Webhook请求头(键值对)" json:"webhook_headers"`
	APITokenEnabled            *bool             `gorm:"column:api_token_enabled;default:false;comment:启用API Token" json:"api_token_enabled"`
	OfflineNotificationEnabled *bool             `gorm:"column:offline_notification_enabled;default:false;comment:启用离线通知" json:"offline_notification_enabled"`
	NotificationType           NotificationType  `gorm:"column:notification_type;type:enum('push_plus','email','wechat_work_app');default:'push_plus';not null;comment:通知方式：push_plus-推送加，email-邮件，wechat_work_app-企业微信应用" json:"notification_type"`
	PushPlusURL                *string           `gorm:"column:push_plus_url;type:varchar(255);default:'';comment:Push Plus的URL" json:"push_plus_url"`
	PushPlusToken              *string           `gorm:"column:push_plus_token;type:varchar(255);default:'';comment:Push Plus的Token" json:"push_plus_token"`
	WechatWorkCorpID           *string           `gorm:"column:wechat_work_corp_id;type:varchar(255);default:'';comment:企业微信企业ID" json:"wechat_work_corp_id"`
	WechatWorkAgentID          *string           `gorm:"column:wechat_work_agent_id;type:varchar(255);default:'';comment:企业微信应用AgentId" json:"wechat_work_agent_id"`
	WechatWorkSecret           *string           `gorm:"column:wechat_work_secret;type:varchar(255);default:'';comment:企业微信应用Secret" json:"wechat_work_secret"`
	WechatWorkProxyURL         *string           `gorm:"column:wechat_work_proxy_url;type:varchar(255);default:'';comment:企业微信代理地址" json:"wechat_work_proxy_url"`
	WechatWorkToUser           *string           `gorm:"column:wechat_work_to_user;type:varchar(512);default:'';comment:企业微信推送用户ID，留空默认ALL" json:"wechat_work_to_user"`
	AutoVerifyUser             *bool             `gorm:"column:auto_verify_user;default:false;comment:自动通过好友验证" json:"auto_verify_user"`
	VerifyUserDelay            *int              `gorm:"column:verify_user_delay;default:60;comment:自动通过好友验证延迟时间(秒)" json:"verify_user_delay"`
	AutoChatroomInvite         *bool             `gorm:"column:auto_chatroom_invite;default:false;comment:自动邀请进群" json:"auto_chatroom_invite"`
	CreatedAt                  int64             `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt                  int64             `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (SystemSettings) TableName() string {
	return "system_settings"
}
