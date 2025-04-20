package service

import (
	"context"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type RobotService struct {
	ctx context.Context
}

func NewRobotService(ctx context.Context) *RobotService {
	return &RobotService{
		ctx: ctx,
	}
}

func (r *RobotService) Login() (uuid, qrcode string, err error) {
	_, uuid, qrcode, err = vars.RobotRuntime.Login()
	if err != nil {
		return
	}
	if uuid != "" && qrcode != "" {
		// 如果二维码不为空，说明需要扫码登陆
		return
	}
	// 唤醒登陆成功，更新登陆状态
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:          vars.RobotRuntime.RobotID,
		Status:      model.RobotStatusOnline,
		LastLoginAt: time.Now().Unix(),
	}
	respo.Update(&robot)
	return
}

func (r *RobotService) LoginCheck() (loggedIn bool, err error) {
	var resp robot.CheckUuid
	resp, err = vars.RobotRuntime.CheckLoginUuid()
	if err != nil {
		return
	}
	if resp.AcctSectResp.Username != "" {
		loggedIn = true
		// 扫码登陆成功，更新登陆状态
		respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
		robot := model.RobotAdmin{
			ID:          vars.RobotRuntime.RobotID,
			WeChatID:    resp.AcctSectResp.Username,
			BindMobile:  resp.AcctSectResp.BindMobile,
			Nickname:    resp.AcctSectResp.Nickname,
			Avatar:      resp.HeadImgUrl,
			Status:      model.RobotStatusOnline,
			LastLoginAt: time.Now().Unix(),
		}
		respo.Update(&robot)
	}
	return
}

func (r *RobotService) Logout() error {
	err := vars.RobotRuntime.Logout()
	if err != nil {
		return err
	}
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOffline,
	}
	respo.Update(&robot)
	return nil
}
