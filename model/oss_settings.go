package model

import "gorm.io/datatypes"

type AutoUploadMode string

const (
	AutoUploadModeAll    AutoUploadMode = "all" // 全部上传
	AutoUploadModeAIOnly AutoUploadMode = "ai_only"
)

type OSSProvider string

const (
	OSSProviderAliyun       OSSProvider = "aliyun"        // 阿里云 OSS
	OSSProviderTencentCloud OSSProvider = "tencent_cloud" // 腾讯云 COS
	OSSProviderCloudflare   OSSProvider = "cloudflare"    // Cloudflare R2
)

type OSSSettings struct {
	ID                      uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:表主键ID" json:"id"`
	AutoUploadImage         *bool          `gorm:"column:auto_upload_image;default:false;comment:启用自动上传图片" json:"auto_upload_image"`
	AutoUploadImageMode     AutoUploadMode `gorm:"column:auto_upload_image_mode;type:enum('all','ai_only');default:'ai_only';not null;comment:自动上传图片模式" json:"auto_upload_image_mode"`
	AutoUploadVideo         *bool          `gorm:"column:auto_upload_video;default:false;comment:启用自动上传视频" json:"auto_upload_video"`
	AutoUploadVideoMode     AutoUploadMode `gorm:"column:auto_upload_video_mode;type:enum('all','ai_only');default:'ai_only';not null;comment:自动上传视频模式" json:"auto_upload_video_mode"`
	AutoUploadFile          *bool          `gorm:"column:auto_upload_file;default:false;comment:启用自动上传文件" json:"auto_upload_file"`
	AutoUploadFileMode      AutoUploadMode `gorm:"column:auto_upload_file_mode;type:enum('all','ai_only');default:'ai_only';not null;comment:自动上传文件模式" json:"auto_upload_file_mode"`
	OSSProvider             OSSProvider    `gorm:"column:oss_provider;type:enum('aliyun','tencent_cloud','cloudflare');default:'aliyun';not null;comment:对象存储服务商" json:"oss_provider"`
	AliyunOSSSettings       datatypes.JSON `gorm:"column:aliyun_oss_settings;type:json;comment:阿里云OSS配置项" json:"aliyun_oss_settings"`
	TencentCloudOSSSettings datatypes.JSON `gorm:"column:tencent_cloud_oss_settings;type:json;comment:腾讯云OSS配置项" json:"tencent_cloud_oss_settings"`
	CloudflareR2Settings    datatypes.JSON `gorm:"column:cloudflare_r2_settings;type:json;comment:Cloudflare R2配置项" json:"cloudflare_r2_settings"`
	CreatedAt               int64          `gorm:"column:created_at;autoCreateTime;not null;comment:创建时间" json:"created_at"`
	UpdatedAt               int64          `gorm:"column:updated_at;autoUpdateTime;not null;comment:更新时间" json:"updated_at"`
}

func (OSSSettings) TableName() string {
	return "oss_settings"
}

// AliyunOSSConfig 阿里云OSS配置
type AliyunOSSConfig struct {
	Endpoint        string `json:"endpoint"`          // 访问域名
	AccessKeyID     string `json:"access_key_id"`     // AccessKey ID
	AccessKeySecret string `json:"access_key_secret"` // AccessKey Secret
	BucketName      string `json:"bucket_name"`       // 存储空间名称
	BasePath        string `json:"base_path"`         // 基础路径（可选）
	CustomDomain    string `json:"custom_domain"`     // 自定义域名（可选）
}

// TencentCloudCOSConfig 腾讯云COS配置
type TencentCloudCOSConfig struct {
	Region       string `json:"region"`        // 地域
	SecretID     string `json:"secret_id"`     // SecretId
	SecretKey    string `json:"secret_key"`    // SecretKey
	BucketURL    string `json:"bucket_url"`    // 存储桶URL (格式: https://bucket-appid.cos.region.myqcloud.com)
	BasePath     string `json:"base_path"`     // 基础路径（可选）
	CustomDomain string `json:"custom_domain"` // 自定义域名（可选）
}

// CloudflareR2Config Cloudflare R2配置
type CloudflareR2Config struct {
	AccountID       string `json:"account_id"`        // 账户ID
	AccessKeyID     string `json:"access_key_id"`     // Access Key ID
	SecretAccessKey string `json:"secret_access_key"` // Secret Access Key
	BucketName      string `json:"bucket_name"`       // 存储桶名称
	BasePath        string `json:"base_path"`         // 基础路径（可选）
	CustomDomain    string `json:"custom_domain"`     // 自定义域名（可选）
}
