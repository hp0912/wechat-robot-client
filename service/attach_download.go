package service

import (
	"context"
	"errors"
	"io"
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
	message, err := respo.GetByID(req.MessageID)
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
	respo := repository.NewMessageRepo(a.ctx, vars.DB)
	message, err := respo.GetByID(req.MessageID)
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

func (a *AttachDownloadService) DownloadFile(req dto.AttachDownloadRequest) (io.ReadCloser, string, error) {
	respo := repository.NewMessageRepo(a.ctx, vars.DB)
	message, err := respo.GetByID(req.MessageID)
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
	respo := repository.NewMessageRepo(a.ctx, vars.DB)
	message, err := respo.GetByID(req.MessageID)
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
