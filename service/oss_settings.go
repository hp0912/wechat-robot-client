package service

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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
	tos "github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"
)

const maxVideoSize int64 = 25 * 1024 * 1024 // 25MB

type OSSSettingService struct {
	ctx             context.Context
	messageRepo     *repository.Message
	ossSettingsRepo *repository.OSSSettings
}

func NewOSSSettingService(ctx context.Context) *OSSSettingService {
	return &OSSSettingService{
		ctx:             ctx,
		messageRepo:     repository.NewMessageRepo(ctx, vars.DB),
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

func (s *OSSSettingService) UploadImageToOSSFromEncryptUrl(settings *model.OSSSettings, message *model.Message, encryptUrl string) error {
	data, contentType, extension, err := s.downloadFromUrl(encryptUrl, 0)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}
	return s.uploadMediaToOSS(settings, message, data, contentType, extension, "images")
}

func (s *OSSSettingService) UploadImageToOSS(settings *model.OSSSettings, message *model.Message) error {
	attachDownloadService := NewAttachDownloadService(s.ctx)
	data, contentType, extension, err := attachDownloadService.DownloadImage(message.ID)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}
	return s.uploadMediaToOSS(settings, message, data, contentType, extension, "images")
}

func (s *OSSSettingService) UploadVideoToOSS(settings *model.OSSSettings, message *model.Message) error {
	// 解析视频消息XML，检查视频大小
	videoSize, err := s.getVideoSizeFromXml(message.Content)
	if err != nil {
		return fmt.Errorf("解析视频消息XML失败: %w", err)
	}
	if videoSize > maxVideoSize {
		return fmt.Errorf("视频大小 %dMB 超过限制 25MB", videoSize/(1024*1024))
	}

	// 下载视频
	attachDownloadService := NewAttachDownloadService(s.ctx)
	reader, _, err := attachDownloadService.DownloadVideo(dto.AttachDownloadRequest{MessageID: message.ID})
	if err != nil {
		return fmt.Errorf("下载视频失败: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(io.LimitReader(reader, maxVideoSize+1))
	if err != nil {
		return fmt.Errorf("读取视频数据失败: %w", err)
	}
	if int64(len(data)) > maxVideoSize {
		return fmt.Errorf("视频大小超过限制 25MB")
	}

	contentType := "video/mp4"
	extension := utils.DetectMediaFormat(data)

	return s.uploadMediaToOSS(settings, message, data, contentType, extension, "videos")
}

func (s *OSSSettingService) UploadVideoToOSSFromUrl(settings *model.OSSSettings, message *model.Message, videoUrl string) error {
	data, contentType, extension, err := s.downloadFromUrl(videoUrl, maxVideoSize)
	if err != nil {
		return fmt.Errorf("下载视频失败: %w", err)
	}
	return s.uploadMediaToOSS(settings, message, data, contentType, extension, "videos")
}

// getVideoSizeFromXml 从视频消息XML中解析视频大小
func (s *OSSSettingService) getVideoSizeFromXml(xmlContent string) (int64, error) {
	type videoMsg struct {
		Length int64 `xml:"length,attr"`
	}
	type videoXml struct {
		XMLName  xml.Name `xml:"msg"`
		VideoMsg videoMsg `xml:"videomsg"`
	}
	var v videoXml
	if err := xml.NewDecoder(strings.NewReader(xmlContent)).Decode(&v); err != nil {
		return 0, err
	}
	return v.VideoMsg.Length, nil
}

// downloadFromUrl 从URL下载文件，maxSize为最大允许大小（0表示不限制）
func (s *OSSSettingService) downloadFromUrl(mediaUrl string, maxSize int64) ([]byte, string, string, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Get(mediaUrl)
	if err != nil {
		return nil, "", "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 如果设置了大小限制，先检查Content-Length
	if maxSize > 0 && resp.ContentLength > maxSize {
		return nil, "", "", fmt.Errorf("文件大小 %dMB 超过限制 %dMB", resp.ContentLength/(1024*1024), maxSize/(1024*1024))
	}

	var reader io.Reader = resp.Body
	if maxSize > 0 {
		reader = io.LimitReader(resp.Body, maxSize+1)
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, "", "", fmt.Errorf("读取内容失败: %w", err)
	}

	if maxSize > 0 && int64(buf.Len()) > maxSize {
		return nil, "", "", fmt.Errorf("文件大小超过限制 %dMB", maxSize/(1024*1024))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	data := buf.Bytes()
	extension := utils.DetectMediaFormat(data)

	return data, contentType, extension, nil
}

// uploadMediaToOSS 校验OSS配置并分发到对应的云厂商上传
func (s *OSSSettingService) uploadMediaToOSS(settings *model.OSSSettings, message *model.Message, data []byte, contentType, extension, mediaType string) error {
	if settings.OSSProvider == "" {
		return errors.New("OSS服务商未配置")
	}
	switch settings.OSSProvider {
	case model.OSSProviderAliyun:
		if settings.AliyunOSSSettings == nil {
			return errors.New("阿里云OSS配置项未配置")
		}
		return s.uploadToAliyun(settings, message, data, contentType, extension, mediaType)
	case model.OSSProviderTencentCloud:
		if settings.TencentCloudOSSSettings == nil {
			return errors.New("腾讯云COS配置项未配置")
		}
		return s.uploadToTencentCloud(settings, message, data, contentType, extension, mediaType)
	case model.OSSProviderCloudflare:
		if settings.CloudflareR2Settings == nil {
			return errors.New("cloudflare r2配置项未配置")
		}
		return s.uploadToCloudflareR2(settings, message, data, contentType, extension, mediaType)
	case model.OSSProviderVolcengine:
		if settings.VolcengineTOSSettings == nil {
			return errors.New("火山引擎TOS配置项未配置")
		}
		return s.uploadToVolcengineTOS(settings, message, data, contentType, extension, mediaType)
	default:
		return fmt.Errorf("不支持的OSS服务商: %s", settings.OSSProvider)
	}
}

// updateMessageAttachmentUrl 更新消息附件URL
func (s *OSSSettingService) updateMessageAttachmentUrl(message *model.Message, fileURL, mediaType string) error {
	message.AttachmentUrl = fileURL
	err := s.messageRepo.Update(&model.Message{
		ID:            message.ID,
		AttachmentUrl: fileURL,
	})
	if err != nil {
		return fmt.Errorf("更新消息附件URL失败: %w", err)
	}
	log.Printf("%s上传成功: %s", mediaType, fileURL)
	return nil
}

func (s *OSSSettingService) uploadToAliyun(settings *model.OSSSettings, message *model.Message, data []byte, contentType, extension, mediaType string) error {
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
	objectKey := s.buildObjectKey(config.BasePath, fileName, mediaType)
	reader := bytes.NewReader(data)
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

	return s.updateMessageAttachmentUrl(message, fileURL, mediaType)
}

func (s *OSSSettingService) uploadToTencentCloud(settings *model.OSSSettings, message *model.Message, data []byte, contentType, extension, mediaType string) error {
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
	objectKey := s.buildObjectKey(config.BasePath, fileName, mediaType)
	reader := bytes.NewReader(data)
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

	return s.updateMessageAttachmentUrl(message, fileURL, mediaType)
}

func (s *OSSSettingService) uploadToCloudflareR2(settings *model.OSSSettings, message *model.Message, data []byte, contentType, extension, mediaType string) error {
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

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	fileName := s.generateFileName(extension)
	objectKey := s.buildObjectKey(config.BasePath, fileName, mediaType)

	reader := bytes.NewReader(data)
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

	return s.updateMessageAttachmentUrl(message, fileURL, mediaType)
}

func (s *OSSSettingService) uploadToVolcengineTOS(settings *model.OSSSettings, message *model.Message, data []byte, contentType, extension, mediaType string) error {
	var config model.VolcengineTOSConfig
	if err := json.Unmarshal(settings.VolcengineTOSSettings, &config); err != nil {
		return fmt.Errorf("解析火山引擎TOS配置失败: %w", err)
	}

	if config.Endpoint == "" || config.Region == "" || config.AccessKey == "" || config.SecretKey == "" || config.BucketName == "" {
		return errors.New("火山引擎TOS配置不完整")
	}

	client, err := tos.NewClientV2(config.Endpoint,
		tos.WithRegion(config.Region),
		tos.WithCredentials(tos.NewStaticCredentials(config.AccessKey, config.SecretKey)),
	)
	if err != nil {
		return fmt.Errorf("创建火山引擎TOS客户端失败: %w", err)
	}

	fileName := s.generateFileName(extension)
	objectKey := s.buildObjectKey(config.BasePath, fileName, mediaType)
	reader := bytes.NewReader(data)

	_, err = client.PutObjectV2(s.ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:      config.BucketName,
			Key:         objectKey,
			ContentType: contentType,
		},
		Content: reader,
	})
	if err != nil {
		return fmt.Errorf("上传到火山引擎TOS失败: %w", err)
	}

	var fileURL string
	if config.CustomDomain != "" {
		fileURL = fmt.Sprintf("%s/%s", strings.TrimRight(config.CustomDomain, "/"), objectKey)
	} else {
		endpoint := strings.TrimPrefix(config.Endpoint, "https://")
		endpoint = strings.TrimPrefix(endpoint, "http://")
		fileURL = fmt.Sprintf("https://%s.%s/%s", config.BucketName, endpoint, objectKey)
	}

	return s.updateMessageAttachmentUrl(message, fileURL, mediaType)
}

// generateFileName 生成唯一的文件名
func (s *OSSSettingService) generateFileName(extension string) string {
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, extension)
	return fileName
}

// buildObjectKey 构建对象存储的完整路径
func (s *OSSSettingService) buildObjectKey(basePath, fileName, mediaType string) string {
	now := time.Now()
	datePath := fmt.Sprintf("%s/%d/%02d/%02d", mediaType, now.Year(), now.Month(), now.Day())

	if basePath != "" {
		basePath = strings.Trim(basePath, "/")
		return path.Join(basePath, datePath, fileName)
	}

	return path.Join(datePath, fileName)
}
