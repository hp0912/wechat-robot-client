package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"

	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type OSSSettingService struct {
	ctx             context.Context
	messageRespo    *repository.Message
	ossSettingsRepo *repository.OSSSettings
}

func NewOSSSettingService(ctx context.Context) *OSSSettingService {
	return &OSSSettingService{
		ctx:             ctx,
		messageRespo:    repository.NewMessageRepo(ctx, vars.DB),
		ossSettingsRepo: repository.NewOSSSettingsRepo(ctx, vars.DB),
	}
}

func (s *OSSSettingService) GetOSSSettingService() (*model.OSSSettings, error) {
	ossSettings, err := s.ossSettingsRepo.GetOSSSettings()
	if err != nil {
		return nil, fmt.Errorf("获取OSS设置失败: %w", err)
	}
	if ossSettings == nil {
		return &model.OSSSettings{}, nil
	}
	return ossSettings, nil
}

func (s *OSSSettingService) SaveOSSSettingService(req *model.OSSSettings) error {
	if req.ID == 0 {
		ossSettings, err := s.ossSettingsRepo.GetOSSSettings()
		if err != nil {
			return err
		}
		if ossSettings != nil {
			return fmt.Errorf("OSS设置已存在，不能重复创建")
		}
		return s.ossSettingsRepo.Create(req)
	}
	return s.ossSettingsRepo.Update(req)
}

func (s *OSSSettingService) UploadImageToOSS(settings *model.OSSSettings, message *model.Message) error {
	if settings.OSSProvider == "" {
		return errors.New("OSS服务商未配置")
	}
	switch settings.OSSProvider {
	case model.OSSProviderAliyun:
		if settings.AliyunOSSSettings == nil {
			return errors.New("阿里云OSS配置项未配置")
		}
		err := s.UploadImageToAliyun(settings, message)
		if err != nil {
			return err
		}
	case model.OSSProviderTencentCloud:
		if settings.TencentCloudOSSSettings == nil {
			return errors.New("腾讯云COS配置项未配置")
		}
		err := s.UploadImageToTencentCloud(settings, message)
		if err != nil {
			return err
		}
	case model.OSSProviderCloudflare:
		if settings.CloudflareR2Settings == nil {
			return errors.New("cloudflare r2配置项未配置")
		}
		err := s.UploadImageToCloudflareR2(settings, message)
		if err != nil {
			return err
		}
	default:
		log.Printf("不支持的OSS服务商: %s", settings.OSSProvider)
	}
	return nil
}

func (s *OSSSettingService) UploadImageToAliyun(settings *model.OSSSettings, message *model.Message) error {
	attachDownloadService := NewAttachDownloadService(s.ctx)
	imageBytes, contentType, extension, err := attachDownloadService.DownloadImage(message.ID)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}

	var config model.AliyunOSSConfig
	if err := json.Unmarshal(settings.AliyunOSSSettings, &config); err != nil {
		return fmt.Errorf("解析阿里云OSS配置失败: %w", err)
	}

	if config.Endpoint == "" || config.AccessKeyID == "" || config.AccessKeySecret == "" || config.BucketName == "" {
		return errors.New("阿里云OSS配置不完整")
	}

	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("创建阿里云OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return fmt.Errorf("获取存储空间失败: %w", err)
	}

	fileName := s.generateFileName(extension)
	objectKey := s.buildObjectKey(config.BasePath, fileName)
	reader := bytes.NewReader(imageBytes)
	options := []oss.Option{
		oss.ContentType(contentType),
	}
	if err := bucket.PutObject(objectKey, reader, options...); err != nil {
		return fmt.Errorf("上传到阿里云OSS失败: %w", err)
	}

	var fileURL string
	if config.CustomDomain != "" {
		fileURL = fmt.Sprintf("%s/%s", strings.TrimRight(config.CustomDomain, "/"), objectKey)
	} else {
		endpoint := strings.TrimPrefix(config.Endpoint, "https://")
		endpoint = strings.TrimPrefix(endpoint, "http://")
		fileURL = fmt.Sprintf("https://%s.%s/%s", config.BucketName, endpoint, objectKey)
	}

	message.AttachmentUrl = fileURL
	err = s.messageRespo.Update(&model.Message{
		ID:            message.ID,
		AttachmentUrl: fileURL,
	})
	if err != nil {
		return fmt.Errorf("更新消息附件URL失败: %w", err)
	}
	log.Printf("图片上传成功: %s", fileURL)

	return nil
}

func (s *OSSSettingService) UploadImageToTencentCloud(settings *model.OSSSettings, message *model.Message) error {
	attachDownloadService := NewAttachDownloadService(s.ctx)
	imageBytes, contentType, extension, err := attachDownloadService.DownloadImage(message.ID)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}

	var config model.TencentCloudCOSConfig
	if err := json.Unmarshal(settings.TencentCloudOSSSettings, &config); err != nil {
		return fmt.Errorf("解析腾讯云COS配置失败: %w", err)
	}

	if config.BucketURL == "" || config.SecretID == "" || config.SecretKey == "" {
		return errors.New("腾讯云COS配置不完整")
	}

	bucketURL, err := url.Parse(config.BucketURL)
	if err != nil {
		return fmt.Errorf("解析Bucket URL失败: %w", err)
	}

	// 创建COS客户端
	client := cos.NewClient(
		&cos.BaseURL{BucketURL: bucketURL},
		&http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  config.SecretID,
				SecretKey: config.SecretKey,
			},
		},
	)

	fileName := s.generateFileName(extension)
	objectKey := s.buildObjectKey(config.BasePath, fileName)
	reader := bytes.NewReader(imageBytes)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	_, err = client.Object.Put(s.ctx, objectKey, reader, opt)
	if err != nil {
		return fmt.Errorf("上传到腾讯云COS失败: %w", err)
	}

	var fileURL string
	if config.CustomDomain != "" {
		fileURL = fmt.Sprintf("%s/%s", strings.TrimRight(config.CustomDomain, "/"), objectKey)
	} else {
		fileURL = fmt.Sprintf("%s/%s", strings.TrimRight(config.BucketURL, "/"), objectKey)
	}

	message.AttachmentUrl = fileURL
	err = s.messageRespo.Update(&model.Message{
		ID:            message.ID,
		AttachmentUrl: fileURL,
	})
	if err != nil {
		return fmt.Errorf("更新消息附件URL失败: %w", err)
	}

	log.Printf("图片上传成功: %s", fileURL)

	return nil
}

func (s *OSSSettingService) UploadImageToCloudflareR2(settings *model.OSSSettings, message *model.Message) error {
	attachDownloadService := NewAttachDownloadService(s.ctx)
	imageBytes, contentType, extension, err := attachDownloadService.DownloadImage(message.ID)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}

	var config model.CloudflareR2Config
	if err := json.Unmarshal(settings.CloudflareR2Settings, &config); err != nil {
		return fmt.Errorf("解析Cloudflare R2配置失败: %w", err)
	}

	if config.AccountID == "" || config.AccessKeyID == "" || config.SecretAccessKey == "" || config.BucketName == "" {
		return errors.New("cloudflare R2配置不完整")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", config.AccountID)
	cfg, err := awsconfig.LoadDefaultConfig(s.ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AccessKeyID,
			config.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("创建AWS配置失败: %w", err)
	}

	// 创建S3客户端，指定R2的endpoint
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // R2需要使用path-style访问
	})

	fileName := s.generateFileName(extension)
	objectKey := s.buildObjectKey(config.BasePath, fileName)

	reader := bytes.NewReader(imageBytes)
	_, err = s3Client.PutObject(s.ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.BucketName),
		Key:         aws.String(objectKey),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("上传到Cloudflare R2失败: %w", err)
	}

	var fileURL string
	if config.CustomDomain != "" {
		fileURL = fmt.Sprintf("%s/%s", strings.TrimRight(config.CustomDomain, "/"), objectKey)
	} else {
		fileURL = fmt.Sprintf("https://pub-%s.r2.dev/%s", config.BucketName, objectKey)
	}

	message.AttachmentUrl = fileURL
	err = s.messageRespo.Update(&model.Message{
		ID:            message.ID,
		AttachmentUrl: fileURL,
	})
	if err != nil {
		return fmt.Errorf("更新消息附件URL失败: %w", err)
	}

	log.Printf("图片上传成功: %s", fileURL)

	return nil
}

// generateFileName 生成唯一的文件名
func (s *OSSSettingService) generateFileName(extension string) string {
	// 使用UUID生成唯一标识
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")

	// 构造文件名: 日期_时间戳_uuid.ext
	fileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, extension)
	return fileName
}

// buildObjectKey 构建对象存储的完整路径
func (s *OSSSettingService) buildObjectKey(basePath, fileName string) string {
	// 按日期分组存储: images/2024/01/02/filename.jpg
	now := time.Now()
	datePath := fmt.Sprintf("images/%d/%02d/%02d", now.Year(), now.Month(), now.Day())

	if basePath != "" {
		basePath = strings.Trim(basePath, "/")
		return path.Join(basePath, datePath, fileName)
	}

	return path.Join(datePath, fileName)
}
