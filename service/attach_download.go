package service

import (
	"context"
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type AttachDownloadService struct {
	ctx context.Context
}

func NewAttachDownloadService(ctx context.Context) *AttachDownloadService {
	return &AttachDownloadService{
		ctx: ctx,
	}
}

func (a *AttachDownloadService) DownloadImage(req dto.AttachDownloadRequest) ([]byte, string, string, error) {
	respo := repository.NewMessageRepo(a.ctx, vars.DB)
	message := respo.GetByID(req.MessageID)
	if message == nil {
		return nil, "", "", errors.New("消息不存在")
	}
	if message.Type != model.MsgTypeImage {
		return nil, "", "", errors.New("消息类型错误")
	}
	return vars.RobotRuntime.DownloadImage(*message)
}

func (a *AttachDownloadService) DownloadVoice(req dto.AttachDownloadRequest) ([]byte, string, string, error) {
	respo := repository.NewMessageRepo(a.ctx, vars.DB)
	message := respo.GetByID(req.MessageID)
	if message == nil {
		return nil, "", "", errors.New("消息不存在")
	}
	if message.Type != model.MsgTypeVoice {
		return nil, "", "", errors.New("消息类型错误")
	}
	return vars.RobotRuntime.DownloadVoice(a.ctx, *message)
}
