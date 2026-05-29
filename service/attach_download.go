package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type AttachDownloadService struct {
	ctx     context.Context
	msgRepo *repository.Message
}

func NewAttachDownloadService(ctx context.Context) *AttachDownloadService {
	return &AttachDownloadService{
		ctx:     ctx,
		msgRepo: repository.NewMessageRepo(ctx, vars.DB),
	}
}

func (a *AttachDownloadService) DownloadImage(messageID int64) ([]byte, string, string, error) {
	message, err := a.msgRepo.GetByID(messageID)
	if err != nil {
		return nil, "", "", err
	}
	if message == nil {
		return nil, "", "", errors.New("消息不存在")
	}
	if message.Type != model.MsgTypeImage {
		return nil, "", "", errors.New("消息类型错误")
	}
	return vars.RobotRuntime.DownloadImage(*message)
}

func (a *AttachDownloadService) DownloadVoice(req dto.AttachDownloadRequest) ([]byte, string, string, error) {
	message, err := a.msgRepo.GetByID(req.MessageID)
	if err != nil {
		return nil, "", "", err
	}
	if message == nil {
		return nil, "", "", errors.New("消息不存在")
	}
	if message.Type != model.MsgTypeVoice {
		return nil, "", "", errors.New("消息类型错误")
	}
	return vars.RobotRuntime.DownloadVoice(a.ctx, *message)
}

func (a *AttachDownloadService) DownloadFile(messageID int64) (io.ReadCloser, string, error) {
	message, err := a.msgRepo.GetByID(messageID)
	if err != nil {
		return nil, "", err
	}
	if message == nil {
		return nil, "", errors.New("消息不存在")
	}
	if message.Type != model.MsgTypeApp || message.AppMsgType != model.AppMsgTypeAttach {
		return nil, "", errors.New("消息类型错误")
	}
	return vars.RobotRuntime.DownloadFile(a.ctx, *message)
}

func (a *AttachDownloadService) DownloadVideo(req dto.AttachDownloadRequest) (io.ReadCloser, string, error) {
	message, err := a.msgRepo.GetByID(req.MessageID)
	if err != nil {
		return nil, "", err
	}
	if message == nil {
		return nil, "", errors.New("消息不存在")
	}
	if message.Type != model.MsgTypeVideo {
		return nil, "", errors.New("消息类型错误")
	}
	return vars.RobotRuntime.DownloadVideo(a.ctx, *message)
}

func (a *AttachDownloadService) UploadDownloadedMedia(messageID int64, mediaType string, data []byte, contentType, extension string) (string, error) {
	message, err := a.msgRepo.GetByID(messageID)
	if err != nil {
		return "", err
	}
	if message == nil {
		return "", errors.New("消息不存在")
	}

	ossMediaType, err := resolveOSSMediaType(message, mediaType)
	if err != nil {
		return "", err
	}

	ossService := NewOSSSettingService(a.ctx)
	settings, err := ossService.GetOSSSettingService()
	if err != nil {
		return "", err
	}
	if settings == nil || settings.OSSProvider == "" {
		return "", errors.New("OSS服务商未配置")
	}

	if err := ossService.UploadDownloadedMediaToOSS(settings, message, data, contentType, extension, ossMediaType); err != nil {
		return "", err
	}
	return message.AttachmentUrl, nil
}

func resolveOSSMediaType(message *model.Message, requested string) (string, error) {
	switch requested {
	case "image":
		if message.Type != model.MsgTypeImage {
			return "", errors.New("消息类型不是图片")
		}
		return "images", nil
	case "video":
		if message.Type != model.MsgTypeVideo {
			return "", errors.New("消息类型不是视频")
		}
		return "videos", nil
	case "voice":
		if message.Type != model.MsgTypeVoice {
			return "", errors.New("消息类型不是语音")
		}
		return "voices", nil
	default:
		return "", fmt.Errorf("不支持的媒体类型: %s", requested)
	}
}
