package model

type ChatRoomMember struct {
	ID                   int64  `gorm:"column:id;type:bigint;primaryKey;autoIncrement;comment:主键ID" json:"id"`                                      // 主键ID
	ChatRoomID           string `gorm:"column:chat_room_id;type:varchar(64);not null;index:idx_chat_room_id;comment:群ID" json:"chat_room_id"`       // 群ID
	WechatID             string `gorm:"column:wechat_id;type:varchar(64);not null;index:idx_wechat_id;comment:微信ID" json:"wechat_id"`               // 微信ID
	Alias                string `gorm:"column:alias;type:varchar(64);comment:微信号" json:"alias"`                                                     // 微信号
	Nickname             string `gorm:"column:nickname;type:varchar(64);comment:昵称" json:"nickname"`                                                // 昵称
	Sex                  *int   `gorm:"column:sex;type:tinyint(1);comment:性别 0未知 1男 2女" json:"sex"`                                                 // 性别																																						// 性别 0未知 1男 2女
	Avatar               string `gorm:"column:avatar;type:varchar(255);comment:头像" json:"avatar"`                                                   // 头像
	InviterWechatID      string `gorm:"column:inviter_wechat_id;type:varchar(64);not null;comment:邀请人微信ID" json:"inviter_wechat_id"`                // 邀请人微信ID
	IsAdmin              *bool  `gorm:"column:is_admin;type:tinyint(1);default:0;comment:是否群管理员" json:"is_admin"`                                   // 是否群管理员
	IsBlacklisted        *bool  `gorm:"column:is_blacklisted;type:tinyint(1);not null;default:0;comment:是否在黑名单" json:"is_blacklisted"`              // 是否在黑名单
	IsLeaved             *bool  `gorm:"column:is_leaved;type:tinyint(1);default:0;comment:是否已经离开群聊" json:"is_leaved"`                               // 是否已经离开群聊
	Score                *int64 `gorm:"column:score;type:bigint;not null;default:0;comment:积分" json:"score"`                                        // 积分
	TemporaryScore       *int64 `gorm:"column:temporary_score;type:bigint;not null;default:0;comment:临时积分" json:"temporary_score"`                  // 临时积分
	TemporaryScoreExpiry *int64 `gorm:"column:temporary_score_expiry;type:bigint;not null;default:0;comment:临时积分有效期" json:"temporary_score_expiry"` // 临时积分有效期
	Remark               string `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`                                                   // 备注
	JoinedAt             int64  `gorm:"column:joined_at;type:bigint;not null;comment:加入时间" json:"joined_at"`                                        // 加入时间
	LastActiveAt         int64  `gorm:"column:last_active_at;type:bigint;not null;comment:最近活跃时间" json:"last_active_at"`                            // 最近活跃时间
	LeavedAt             *int64 `gorm:"column:leaved_at;type:bigint;comment:离开时间" json:"leaved_at"`                                                 // 离开时间
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
