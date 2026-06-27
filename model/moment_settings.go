package model

type MomentSettings struct {
	ID                      int64   `gorm:"primaryKey;autoIncrement;comment:表主键ID" json:"id"`
	SyncKey                 string  `gorm:"type:text;comment:朋友圈同步Key" json:"sync_key"`
	AutoLike                *bool   `gorm:"default:false;comment:开启自动点赞" json:"auto_like"`
	AutoComment             *bool   `gorm:"default:false;comment:开启自动评论" json:"auto_comment"`
	Whitelist               *string `gorm:"type:text;comment:自动点赞、评论白名单" json:"whitelist"`
	Blacklist               *string `gorm:"type:text;comment:自动点赞、评论黑名单" json:"blacklist"`
	AIBaseURL               string  `gorm:"type:varchar(255);default:'';comment:AI的基础URL地址" json:"ai_base_url"`
	AIAPIKey                string  `gorm:"type:varchar(255);default:'';comment:AI的API密钥" json:"ai_api_key"`
	WorkflowModel           string  `gorm:"type:varchar(100);default:'';comment:工作流模型" json:"workflow_model"`
	CommentModel            string  `gorm:"type:varchar(100);default:'';comment:评论模型" json:"comment_model"`
	ImageUnderstandingModel string  `gorm:"type:varchar(100);default:'doubao-seed-2-0-pro-260215';comment:图片理解模型" json:"image_understanding_model"`
	VideoUnderstandingModel string  `gorm:"type:varchar(100);default:'doubao-seed-2-0-pro-260215';comment:视频理解模型" json:"video_understanding_model"`
	CommentPrompt           string  `gorm:"type:text;comment:评论系统提示词" json:"comment_prompt"`
	MaxCompletionTokens     *int    `gorm:"default:0;comment:评论最大回复" json:"max_completion_tokens"`
	SyncInterval            int     `gorm:"default:120;comment:朋友圈同步间隔(分钟)" json:"sync_interval" binding:"min=10,max=86400"`
}

func (MomentSettings) TableName() string {
	return "moment_settings"
}
