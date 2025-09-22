package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type LoginService struct {
	ctx                 context.Context
	robotAdminRespo     *repository.RobotAdmin
	systemSettingsRespo *repository.SystemSettings
}

func NewLoginService(ctx context.Context) *LoginService {
	return &LoginService{
		ctx:                 ctx,
		robotAdminRespo:     repository.NewRobotAdminRepo(ctx, vars.AdminDB),
		systemSettingsRespo: repository.NewSystemSettingsRepo(ctx, vars.DB),
	}
}

func (s *LoginService) Online() error {
	vars.RobotRuntime.Status = model.RobotStatusOnline
	// 启动定时任务
	if vars.CronManager != nil {
		vars.CronManager.Clear()
		vars.CronManager.Start()
	}
	if vars.RobotRuntime.SyncMomentCancel != nil {
		vars.RobotRuntime.SyncMomentCancel()
	}
	go func() {
		time.Sleep(1 * time.Second)
		NewMomentsService(context.Background()).SyncMomentStart()
	}()
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
	if vars.RobotRuntime.SyncMomentCancel != nil {
		vars.RobotRuntime.SyncMomentCancel()
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

func (s *LoginService) Login(loginType string, isPretender bool) (loginData robot.LoginResponse, err error) {
	if vars.RobotRuntime.Status == model.RobotStatusOnline {
		err := s.Logout()
		if err != nil {
			log.Printf("登出失败: %v\n", err)
		}
	}
	loginData, err = vars.RobotRuntime.Login(loginType, isPretender)
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

func (s *LoginService) LoginYPayVerificationcode(req robot.VerificationCodeRequest) (err error) {
	return vars.RobotRuntime.LoginYPayVerificationcode(req)
}

func (s *LoginService) LoginData62Login(username, password string) (resp robot.UnifyAuthResponse, err error) {
	return vars.RobotRuntime.LoginData62Login(username, password)
}

func (s *LoginService) LoginData62SMSAgain(req robot.LoginData62SMSAgainRequest) (resp string, err error) {
	return vars.RobotRuntime.LoginData62SMSAgain(req)
}

func (s *LoginService) LoginData62SMSVerify(req robot.LoginData62SMSVerifyRequest) (resp string, err error) {
	return vars.RobotRuntime.LoginData62SMSVerify(req)
}

func (s *LoginService) LoginA16Data1(username, password string) (resp robot.UnifyAuthResponse, err error) {
	return vars.RobotRuntime.LoginA16Data1(username, password)
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

func (r *LoginService) LogoutCallback(req dto.LogoutNotificationRequest) (err error) {
	if req.WxID != vars.RobotRuntime.WxID {
		log.Printf("LogoutCallback: wxID(%s) does not match the current robot's wxID", req.WxID)
		return
	}
	defer func() {
		if req.Type != "offline" {
			log.Printf("接收到掉线通知，但类型不是offline，忽略处理: %s", req.Type)
			return
		}
		err := r.Offline()
		if err != nil {
			log.Printf("接收到掉线通知，退出客户端失败: %v", err)
		}
	}()

	systemSettings, err := r.systemSettingsRespo.GetSystemSettings()
	if err != nil {
		log.Printf("接收到掉线通知，获取系统设置失败: %v", err)
		return
	}
	if systemSettings == nil {
		log.Printf("接收到掉线通知，系统设置还未设置")
		return
	}
	if systemSettings.OfflineNotificationEnabled != nil && *systemSettings.OfflineNotificationEnabled {
		// 发送离线通知
		if systemSettings.NotificationType == model.NotificationTypePushPlus {
			var result dto.PushPlusNotificationResponse
			var title string
			var content string
			if req.Type == "offline" {
				title = "机器人掉线通知"
				content = fmt.Sprintf("您的机器人（%s）掉线啦~~~", vars.RobotRuntime.WxID)
			} else {
				title = "机器人发送心跳失败"
				content = fmt.Sprintf("您的机器人（%s）第%d次发送心跳失败了~~~", vars.RobotRuntime.WxID, req.RetryCount)
			}
			httpResp, err1 := resty.New().R().
				SetHeader("Content-Type", "application/json;chartset=utf-8").
				SetBody(dto.PushPlusNotificationRequest{
					Token:   *systemSettings.PushPlusToken,
					Title:   title,
					Content: content,
					Channel: "wechat",
				}).
				SetResult(&result).
				Post(*systemSettings.PushPlusURL)
			if err1 != nil {
				log.Printf("接收到掉线通知，发送离线通知失败: %v", err1)
				return
			}
			if httpResp.StatusCode() != 200 {
				log.Printf("接收到掉线通知，发送离线通知失败，HTTP状态码: %d", httpResp.StatusCode())
				return
			}
			if result.Code != 200 {
				log.Printf("接收到掉线通知，发送离线通知失败，PushPlus返回错误: %s", result.Msg)
				return
			}
			log.Printf("接收到掉线通知，发送离线通知成功")
		}
		return
	}

	log.Printf("接收到掉线通知，系统设置未开启掉线通知")

	return
}
