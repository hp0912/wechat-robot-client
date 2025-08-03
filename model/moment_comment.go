package model

type MomentComment struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	WechatID  string `gorm:"column:wechat_id;type:varchar(64);index:idx_wechat_id" json:"wechat_id"`
	MomentID  uint64 `gorm:"column:moment_id;not null;uniqueIndex:uniq_moment_id" json:"moment_id"`
	Comment   string `gorm:"column:comment;type:text" json:"comment"`
	CreatedAt int64  `gorm:"column:created_at;not null;index:idx_created_at" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (MomentComment) TableName() string {
	return "moment_comments"
}
