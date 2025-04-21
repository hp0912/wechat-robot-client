package service

import (
	"context"
	"log"
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

func (r *RobotService) Offline() {
	vars.RobotRuntime.Status = model.RobotStatusOffline
	err := vars.RobotRuntime.AutoHeartbeatStop()
	if err != nil {
		log.Println("停止心跳失败:", err)
	}
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOffline,
	}
	respo.Update(&robot)
}

func (r *RobotService) IsRunning() (result bool) {
	result = vars.RobotRuntime.IsRunning()
	if !result && vars.RobotRuntime.Status != model.RobotStatusOffline {
		r.Offline()
	}
	return
}

func (r *RobotService) IsLoggedIn() (result bool) {
	result = vars.RobotRuntime.IsLoggedIn()
	if !result && vars.RobotRuntime.Status != model.RobotStatusOffline {
		r.Offline()
	}
	return
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
		// 扫码登陆成功
		vars.RobotRuntime.WxID = resp.AcctSectResp.Username
		vars.RobotRuntime.Status = model.RobotStatusOnline
		// 开启心跳
		err = vars.RobotRuntime.AutoHeartbeatStart()
		if err != nil {
			return
		}
		// 更新登陆状态
		var profile robot.UserProfile
		profile, err = vars.RobotRuntime.GetProfile(resp.AcctSectResp.Username)
		if err != nil {
			return
		}
		robot := model.RobotAdmin{
			ID:          vars.RobotRuntime.RobotID,
			WeChatID:    profile.UserInfo.UserName.String,
			Alias:       profile.UserInfo.Alias,
			BindMobile:  profile.UserInfo.BindMobile.String,
			Nickname:    profile.UserInfo.NickName.String,
			Avatar:      profile.UserInfoExt.BigHeadImgUrl, // 从 resp.AcctSectResp.FsUrl 获取的不太靠谱
			Status:      model.RobotStatusOnline,
			LastLoginAt: time.Now().Unix(),
		}
		respo.Update(&robot)
	}
	return
}

func (r *RobotService) Logout() (err error) {
	// 停止心跳
	err = vars.RobotRuntime.AutoHeartbeatStop()
	if err != nil {
		return
	}
	err = vars.RobotRuntime.Logout()
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOffline,
	}
	respo.Update(&robot)

	vars.RobotRuntime.Status = model.RobotStatusOffline

	return
}
