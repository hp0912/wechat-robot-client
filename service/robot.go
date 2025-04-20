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

func (r *RobotService) Login() (uuid string, awken bool, err error) {
	_, uuid, awken, err = vars.RobotRuntime.Login()
	if err != nil {
		return
	}
	if uuid != "" {
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

func (r *RobotService) LoginCheck(uuid string) (resp robot.CheckUuid, err error) {
	resp, err = vars.RobotRuntime.CheckLoginUuid(uuid)
	if err != nil {
		return
	}
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	if resp.AcctSectResp.Username != "" {
		// 扫码登陆成功，更新登陆状态
		robot := model.RobotAdmin{
			ID:          vars.RobotRuntime.RobotID,
			WeChatID:    resp.AcctSectResp.Username,
			BindMobile:  resp.AcctSectResp.BindMobile,
			Nickname:    resp.AcctSectResp.Nickname,
			Avatar:      resp.AcctSectResp.FsUrl,
			Status:      model.RobotStatusOnline,
			LastLoginAt: time.Now().Unix(),
		}
		respo.Update(&robot)
	}
	return
}

func (r *RobotService) Logout() (err error) {
	err = vars.RobotRuntime.Logout()
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOffline,
	}
	respo.Update(&robot)
	return
}
