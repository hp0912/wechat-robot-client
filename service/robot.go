package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
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

func (r *RobotService) Online() {
	vars.RobotRuntime.Status = model.RobotStatusOnline
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOnline,
	}
	respo.Update(&robot)
}

func (r *RobotService) Offline() {
	vars.RobotRuntime.Status = model.RobotStatusOffline
	if vars.RobotRuntime.HeartbeatCancel != nil {
		vars.RobotRuntime.HeartbeatCancel()
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
	if vars.RobotRuntime.Status == model.RobotStatusOnline {
		err = errors.New("您已经登陆，可以尝试刷新机器人状态")
		return
	}
	uuid, awken, err = vars.RobotRuntime.Login()
	return
}

func (r *RobotService) HeartbeatStart() {
	ctx := context.Background()
	vars.RobotRuntime.HeartbeatContext, vars.RobotRuntime.HeartbeatCancel = context.WithCancel(ctx)
	var errCount int
	for {
		select {
		case <-vars.RobotRuntime.HeartbeatContext.Done():
			return
		case <-time.After(3 * time.Second):
			err := vars.RobotRuntime.Heartbeat()
			log.Println("心跳", err)
			if err != nil {
				if !strings.Contains(err.Error(), "在运行") {
					errCount++
					if errCount > 10 {
						// 10次心跳失败，认为机器人离线
						r.Offline()
						return
					}
				} else {
					errCount = 0
				}
			} else {
				errCount = 0
			}
		}
	}
}

func (r *RobotService) LoginCheck(uuid string) (resp robot.CheckUuid, err error) {
	resp, err = vars.RobotRuntime.CheckLoginUuid(uuid)
	if err != nil {
		return
	}
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	if resp.AcctSectResp.Username != "" {
		// 登陆成功
		vars.RobotRuntime.WxID = resp.AcctSectResp.Username
		vars.RobotRuntime.Status = model.RobotStatusOnline
		// 开启心跳
		go r.HeartbeatStart()
		// 更新登陆状态
		var profile robot.UserProfile
		profile, err = vars.RobotRuntime.GetProfile(resp.AcctSectResp.Username)
		if err != nil {
			return
		}
		bytes, _ := json.Marshal(profile.UserInfo)
		bytesExt, _ := json.Marshal(profile.UserInfoExt)
		robot := model.RobotAdmin{
			ID:          vars.RobotRuntime.RobotID,
			WeChatID:    profile.UserInfo.UserName.String,
			Alias:       profile.UserInfo.Alias,
			BindMobile:  profile.UserInfo.BindMobile.String,
			Nickname:    profile.UserInfo.NickName.String,
			Avatar:      profile.UserInfoExt.BigHeadImgUrl, // 从 resp.AcctSectResp.FsUrl 获取的不太靠谱
			Status:      model.RobotStatusOnline,
			Profile:     bytes,
			ProfileExt:  bytesExt,
			LastLoginAt: time.Now().Unix(),
		}
		respo.Update(&robot)
	}
	return
}

func (r *RobotService) Logout() (err error) {
	r.Offline()
	err = vars.RobotRuntime.Logout()
	return
}
