package model

type ChatRoomMember struct {
	ID                   int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                            // 主键ID
	ChatRoomID           string `gorm:"column:chat_room_id;not null;index:idx_chat_room_id" json:"chat_room_id"` // 群ID
	WechatID             string `gorm:"column:wechat_id;not null;index:idx_wechat_id" json:"wechat_id"`          // 微信ID
	Alias                string `gorm:"column:alias" json:"alias"`                                               // 微信号
	Nickname             string `gorm:"column:nickname" json:"nickname"`                                         // 昵称
	Avatar               string `gorm:"column:avatar" json:"avatar"`                                             // 头像
	InviterWechatID      string `gorm:"column:inviter_wechat_id;not null" json:"inviter_wechat_id"`              // 邀请人微信ID
	IsAdmin              *bool  `gorm:"column:is_admin;default:false" json:"is_admin"`                           // 是否群管理员
	IsBlacklisted        *bool  `gorm:"column:is_blacklisted;default:false" json:"is_blacklisted"`               // 是否在黑名单
	IsLeaved             *bool  `gorm:"column:is_leaved;default:false" json:"is_leaved"`                         // 是否已经离开群聊
	Score                *int64 `gorm:"column:score;default:0" json:"score"`                                     // 积分
	TemporaryScore       *int64 `gorm:"column:temporary_score;default:0" json:"temporary_score"`                 // 临时积分
	TemporaryScoreExpiry *int64 `gorm:"column:temporary_score_expiry;default:0" json:"temporary_score_expiry"`   // 临时积分有效期
	Remark               string `gorm:"column:remark" json:"remark"`                                             // 备注
	JoinedAt             int64  `gorm:"column:joined_at;not null" json:"joined_at"`                              // 加入时间
	LastActiveAt         int64  `gorm:"column:last_active_at;not null" json:"last_active_at"`                    // 最近活跃时间
	LeavedAt             *int64 `gorm:"column:leaved_at" json:"leaved_at"`                                       // 离开时间
}

// TableName 设置表名
func (ChatRoomMember) TableName() string {
	return "chat_room_members"
}

type ScoreAction string

const (
	ScoreActionIncrease ScoreAction = "increase"
	ScoreActionReduce   ScoreAction = "reduce"
)

type UpdateChatRoomMember struct {
	ChatRoomID           string       `form:"chat_room_id" json:"chat_room_id" binding:"required"`
	WechatID             string       `form:"wechat_id" json:"wechat_id" binding:"required"`
	Batch                bool         `form:"batch" json:"batch"`
	IsAdmin              *bool        `form:"is_admin" json:"is_admin"`
	IsBlacklisted        *bool        `form:"is_blacklisted" json:"is_blacklisted"`
	TemporaryScoreAction *ScoreAction `form:"temporary_score_action" json:"temporary_score_action"`
	TemporaryScore       *int64       `form:"temporary_score" json:"temporary_score"`
	ScoreAction          *ScoreAction `form:"score_action" json:"score_action"`
	Score                *int64       `form:"score" json:"score"`
}
