package model

type Moment struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	WechatID      string `gorm:"column:wechat_id;type:varchar(64);index:idx_wechat_id" json:"wechat_id"`
	MomentID      uint64 `gorm:"column:moment_id;not null;uniqueIndex:uniq_moment_id" json:"moment_id"`
	Type          int    `gorm:"column:type;not null;index:idx_type" json:"type"`
	AppMsgType    *int   `gorm:"column:app_msg_type" json:"app_msg_type"`
	Content       string `gorm:"column:content;type:text" json:"content"`
	MessageSource string `gorm:"column:message_source;type:text" json:"message_source"`
	ImgBuf        string `gorm:"column:img_buf;type:text" json:"img_buf"`
	Status        int    `gorm:"column:status;not null;default:0" json:"status"`
	ImgStatus     int    `gorm:"column:img_status;not null;default:0" json:"img_status"`
	PushContent   string `gorm:"column:push_content;type:text" json:"push_content"`
	MessageSeq    int    `gorm:"column:message_seq;not null;default:0" json:"message_seq"`
	CreatedAt     int64  `gorm:"column:created_at;not null;index:idx_created_at" json:"created_at"`
	UpdatedAt     int64  `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (Moment) TableName() string {
	return "moments"
}
