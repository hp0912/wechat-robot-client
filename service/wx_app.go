package service

import (
	"context"
	"wechat-robot-client/vars"
)

type WXAppService struct {
	ctx context.Context
}

func NewWXAppService(ctx context.Context) *WXAppService {
	return &WXAppService{
		ctx: ctx,
	}
}

func (s *WXAppService) WxappQrcodeAuthLogin(URL string) error {
	return vars.RobotRuntime.WxappQrcodeAuthLogin(URL)
}
