package service

import (
	"context"
	"wechat-robot-client/vars"
)

type OfficialAccountService struct {
	ctx context.Context
}

func NewOfficialAccountService(ctx context.Context) *OfficialAccountService {
	return &OfficialAccountService{
		ctx: ctx,
	}
}

func (s *OfficialAccountService) GetAppMsgExt(url string) (string, error) {
	return vars.RobotRuntime.GetAppMsgExt(url)
}
