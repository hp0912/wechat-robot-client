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
	ctx             context.Context
	robotAdminRespo *repository.RobotAdmin
}

func NewLoginService(ctx context.Context) *LoginService {
	return &LoginService{
		ctx:             ctx,
		robotAdminRespo: repository.NewRobotAdminRepo(ctx, vars.AdminDB),
	}
}

func (s *LoginService) Online() error {
	vars.RobotRuntime.Status = model.RobotStatusOnline
	// 启动定时任务
	if vars.CronManager != nil {
		vars.CronManager.Clear()
		vars.CronManager.Start()
	}
	// 开启自动心跳，包括长连接自动同步消息
	err := s.AutoHeartBeat()
	if err != nil {
		return fmt.Errorf("开启自动心跳失败: %w", err)
	}
	// 更新机器人状态
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOnline,
	}
	return s.robotAdminRespo.Update(&robot)
}

func (s *LoginService) Offline() error {
	vars.RobotRuntime.Status = model.RobotStatusOffline
	err := s.CloseAutoHeartBeat()
	if err != nil {
		log.Printf("关闭自动心跳失败: %v\n", err)
	}
	// 清空定时任务
	if vars.CronManager != nil {
		vars.CronManager.Clear()
	}
	// 更新状态
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: model.RobotStatusOffline,
	}
	return s.robotAdminRespo.Update(&robot)
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

func (s *LoginService) AutoHeartBeat() error {
	return vars.RobotRuntime.AutoHeartBeat()
}

func (s *LoginService) CloseAutoHeartBeat() error {
	return vars.RobotRuntime.CloseAutoHeartBeat()
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
		err := s.Logout()
		if err != nil {
			log.Printf("登出失败: %v\n", err)
		}
	}
	uuid, awkenLogin, autoLogin, err = vars.RobotRuntime.Login()
	return
}

func (s *LoginService) LoginCheck(uuid string) (resp robot.CheckUuid, err error) {
	resp, err = vars.RobotRuntime.CheckLoginUuid(uuid)
	if err != nil {
		return
	}
	if resp.AcctSectResp.Username != "" {
		// 一个机器人实例只能绑定一个微信账号
		var robotAdmin *model.RobotAdmin
		robotAdmin, err = s.robotAdminRespo.GetByRobotID(vars.RobotRuntime.RobotID)
		if err != nil {
			return
		}
		if robotAdmin == nil {
			err = errors.New("查询机器人实例失败，请联系管理员。")
			return
		}
		if robotAdmin.WeChatID != "" && robotAdmin.WeChatID != resp.AcctSectResp.Username {
			err = errors.New("一个机器人实例只能绑定一个微信账号。")
			_ = s.Logout()
			return
		}
		// 登陆成功
		vars.RobotRuntime.WxID = resp.AcctSectResp.Username
		vars.RobotRuntime.Status = model.RobotStatusOnline
		err = s.Online()
		if err != nil {
			return
		}
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
		err = s.robotAdminRespo.Update(&robot)
		if err != nil {
			return
		}
	}
	return
}

func (r *LoginService) Logout() (err error) {
	err = r.Offline()
	if err != nil {
		return
	}
	err = vars.RobotRuntime.Logout()
	return
}

func (r *LoginService) SyncMessageCallback(wxID string, syncMessage robot.SyncMessage) {
	if wxID != vars.RobotRuntime.WxID {
		return
	}
	NewMessageService(r.ctx).ProcessMessage(syncMessage)
}

func (r *LoginService) LogoutCallback(wxID string) (err error) {
	if wxID != vars.RobotRuntime.WxID {
		return
	}
	return r.Offline()
}
