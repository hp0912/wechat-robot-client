package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type LoginService struct {
	ctx context.Context
}

func NewLoginService(ctx context.Context) *LoginService {
	return &LoginService{
		ctx: ctx,
	}
}

func (s *LoginService) Online() error {
	vars.RobotRuntime.Status = model.RobotStatusOnline
	// 启动定时任务
	vars.CronManager.Clear()
	vars.CronManager.Start()
	// 更新机器人状态
	respo := repository.NewRobotAdminRepo(s.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOnline,
	}
	return respo.Update(&robot)
}

func (s *LoginService) Offline() error {
	vars.RobotRuntime.Status = model.RobotStatusOffline
	if vars.RobotRuntime.HeartbeatCancel != nil {
		vars.RobotRuntime.HeartbeatCancel()
	}
	if vars.RobotRuntime.SyncMessageCancel != nil {
		vars.RobotRuntime.SyncMessageCancel()
	}
	// 清空定时任务
	vars.CronManager.Clear()
	// 更新状态
	respo := repository.NewRobotAdminRepo(s.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOffline,
	}
	return respo.Update(&robot)
}

func (s *LoginService) IsRunning() (result bool) {
	result = vars.RobotRuntime.IsRunning()
	if !result && vars.RobotRuntime.Status != model.RobotStatusOffline {
		s.Offline()
	}
	return
}

func (s *LoginService) IsLoggedIn() (result bool) {
	result = vars.RobotRuntime.IsLoggedIn()
	if !result && vars.RobotRuntime.Status != model.RobotStatusOffline {
		s.Offline()
	}
	return
}

func (s *LoginService) HeartbeatStart() {
	ctx := context.Background()
	vars.RobotRuntime.HeartbeatContext, vars.RobotRuntime.HeartbeatCancel = context.WithCancel(ctx)
	var errCount int
	for {
		select {
		case <-vars.RobotRuntime.HeartbeatContext.Done():
			return
		case <-time.After(25 * time.Second):
			mode := os.Getenv("GIN_MODE")
			err := vars.RobotRuntime.Heartbeat()
			log.Println(mode, " 心跳: ", err)
			if err != nil {
				errCount++
				if mode == "release" && errCount%3 == 0 {
					log.Println("检测到机器人掉线，尝试重新登陆...")
					err := vars.RobotRuntime.LoginTwiceAutoAuth()
					if err != nil {
						log.Println("尝试重新登陆失败: ", err)
					}
				}
				if errCount > 10 {
					// 10次心跳失败，认为机器人离线
					s.Offline()
					return
				}
			} else {
				errCount = 0
			}
		}
	}
}

func (s *LoginService) Login() (uuid string, awkenLogin, autoLogin bool, err error) {
	if vars.RobotRuntime.Status == model.RobotStatusOnline {
		err = errors.New("您已经登陆，可以尝试刷新机器人状态")
		return
	}
	uuid, awkenLogin, autoLogin, err = vars.RobotRuntime.Login()
	return
}

func (s *LoginService) LoginCheck(uuid string) (resp robot.CheckUuid, err error) {
	resp, err = vars.RobotRuntime.CheckLoginUuid(uuid)
	if err != nil {
		return
	}
	respo := repository.NewRobotAdminRepo(s.ctx, vars.AdminDB)
	if resp.AcctSectResp.Username != "" {
		// 登陆成功
		vars.RobotRuntime.WxID = resp.AcctSectResp.Username
		vars.RobotRuntime.Status = model.RobotStatusOnline
		// 更新下全局配置的所有人, 全局配置在创建机器人实例的时候创建，这个时候的所有人是空，因为获取不到微信ID
		globalSettingsSvc := NewGlobalSettingsService(s.ctx)
		err = globalSettingsSvc.UpdateGlobalSettings(resp.AcctSectResp.Username)
		if err != nil {
			panic(fmt.Errorf("更新全局配置失败，请联系管理员: %w", err))
		}
		// 开启心跳
		go s.HeartbeatStart()
		// 开启消息同步
		msgService := NewMessageService(context.Background())
		go msgService.SyncMessageStart()
		// 开启定时任务
		vars.CronManager.Clear()
		vars.CronManager.Start()
		// 更新登陆状态
		var profile robot.GetProfileResponse
		profile, err = vars.RobotRuntime.GetProfile(resp.AcctSectResp.Username)
		if err != nil {
			return
		}
		if profile.UserInfo.UserName.String == nil {
			err = errors.New("获取用户信息失败")
			return
		}
		bytes, _ := json.Marshal(profile.UserInfo)
		bytesExt, _ := json.Marshal(profile.UserInfoExt)
		robot := model.RobotAdmin{
			ID:          vars.RobotRuntime.RobotID,
			WeChatID:    *profile.UserInfo.UserName.String,
			Alias:       profile.UserInfo.Alias,
			BindMobile:  profile.UserInfo.BindMobile.String,
			Nickname:    profile.UserInfo.NickName.String,
			Avatar:      profile.UserInfoExt.BigHeadImgUrl, // 从 resp.AcctSectResp.FsUrl 获取的不太靠谱
			Status:      model.RobotStatusOnline,
			Profile:     bytes,
			ProfileExt:  bytesExt,
			LastLoginAt: time.Now().Unix(),
		}
		err = respo.Update(&robot)
	}
	return
}

func (r *LoginService) Logout() (err error) {
	r.Offline()
	err = vars.RobotRuntime.Logout()
	return
}
