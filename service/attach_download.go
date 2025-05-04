package service

import (
	"context"
	"wechat-robot-client/dto"
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
	return nil, "", "", nil
}
