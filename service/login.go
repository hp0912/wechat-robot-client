package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type LoginService struct {
	ctx                context.Context
	robotAdminRepo     *repository.RobotAdmin
	systemSettingsRepo *repository.SystemSettings
}

func NewLoginService(ctx context.Context) *LoginService {
	return &LoginService{
		ctx:                ctx,
		robotAdminRepo:     repository.NewRobotAdminRepo(ctx, vars.AdminDB),
		systemSettingsRepo: repository.NewSystemSettingsRepo(ctx, vars.DB),
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
	return s.robotAdminRepo.Update(&robot)
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
	return s.robotAdminRepo.Update(&robot)
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

func (s *LoginService) GetCachedInfo() (robot.LoginData, error) {
	return vars.RobotRuntime.GetCachedInfo()
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
		robotAdmin, err = s.robotAdminRepo.GetByRobotID(vars.RobotRuntime.RobotID)
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
		vars.RobotRuntime.LoginTime = time.Now().Unix()
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
		err = s.robotAdminRepo.Update(&robot)
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
	resp, err = vars.RobotRuntime.LoginData62SMSVerify(req)
	if err == nil {
		vars.RobotRuntime.Status = model.RobotStatusOnline
		vars.RobotRuntime.LoginTime = time.Now().Unix()
		_ = s.Online()
	}
	return
}

func (s *LoginService) LoginA16Data(username, password string) (resp robot.UnifyAuthResponse, err error) {
	resp, err = vars.RobotRuntime.LoginA16Data(username, password)
	if err == nil {
		vars.RobotRuntime.Status = model.RobotStatusOnline
		vars.RobotRuntime.LoginTime = time.Now().Unix()
		_ = s.Online()
	}
	return
}

func (s *LoginService) ImportLoginData(loginDataStr string) (err error) {
	var loginData robot.LoginData
	err = json.Unmarshal([]byte(loginDataStr), &loginData)
	if err != nil {
		return
	}
	robot := model.RobotAdmin{
		ID:         vars.RobotRuntime.RobotID,
		DeviceID:   loginData.Deviceid_str,
		DeviceName: loginData.DeviceName,
		WeChatID:   loginData.Wxid,
	}
	err = s.robotAdminRepo.Update(&robot)
	if err != nil {
		return
	}
	return vars.RedisClient.Set(s.ctx, loginData.Wxid, loginDataStr, 0).Err()
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

	systemSettings, err := r.systemSettingsRepo.GetSystemSettings()
	if err != nil {
		log.Printf("接收到掉线通知，获取系统设置失败: %v", err)
		return
	}
	if systemSettings == nil {
		log.Printf("接收到掉线通知，系统设置还未设置")
		return
	}
	if systemSettings.OfflineNotificationEnabled != nil && *systemSettings.OfflineNotificationEnabled {
		if notifyErr := r.sendOfflineNotification(systemSettings, req); notifyErr != nil {
			log.Printf("接收到掉线通知，发送离线通知失败: %v", notifyErr)
			return
		}
		log.Printf("接收到掉线通知，发送离线通知成功")
		return
	}

	log.Printf("接收到掉线通知，系统设置未开启离线通知")
	return
}

func (r *LoginService) sendOfflineNotification(systemSettings *model.SystemSettings, req dto.LogoutNotificationRequest) error {
	title, content := buildOfflineNotificationMessage(req)

	switch systemSettings.NotificationType {
	case model.NotificationTypePushPlus:
		return r.sendPushPlusNotification(systemSettings, title, content)
	case model.NotificationTypeWechatWorkApp:
		return r.sendWechatWorkAppNotification(systemSettings, content)
	case model.NotificationTypeEmail:
		return errors.New("暂不支持邮件通知")
	default:
		return fmt.Errorf("不支持的通知类型: %s", systemSettings.NotificationType)
	}
}

func (r *LoginService) sendPushPlusNotification(systemSettings *model.SystemSettings, title, content string) error {
	pushPlusURL := getTrimmedString(systemSettings.PushPlusURL)
	pushPlusToken := getTrimmedString(systemSettings.PushPlusToken)
	if pushPlusURL == "" {
		return errors.New("PushPlus 地址不能为空")
	}
	if pushPlusToken == "" {
		return errors.New("PushPlus Token 不能为空")
	}

	var result dto.PushPlusNotificationResponse
	httpResp, err := resty.New().R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(dto.PushPlusNotificationRequest{
			Token:   pushPlusToken,
			Title:   title,
			Content: content,
			Channel: "wechat",
		}).
		SetResult(&result).
		Post(pushPlusURL)
	if err != nil {
		return err
	}
	if httpResp.StatusCode() != 200 {
		return fmt.Errorf("PushPlus HTTP 状态码异常: %d", httpResp.StatusCode())
	}
	if result.Code != 200 {
		return fmt.Errorf("PushPlus 返回错误: %s", result.Msg)
	}
	return nil
}

func (r *LoginService) sendWechatWorkAppNotification(systemSettings *model.SystemSettings, content string) error {
	corpID := getTrimmedString(systemSettings.WechatWorkCorpID)
	agentIDText := getTrimmedString(systemSettings.WechatWorkAgentID)
	secret := getTrimmedString(systemSettings.WechatWorkSecret)
	proxyURL := getTrimmedString(systemSettings.WechatWorkProxyURL)
	toUser := getTrimmedString(systemSettings.WechatWorkToUser)

	if corpID == "" {
		return errors.New("企业微信企业ID不能为空")
	}
	if agentIDText == "" {
		return errors.New("企业微信应用AgentId不能为空")
	}
	if secret == "" {
		return errors.New("企业微信应用Secret不能为空")
	}
	if toUser == "" {
		toUser = "ALL"
	}

	agentID, err := strconv.ParseInt(agentIDText, 10, 64)
	if err != nil {
		return fmt.Errorf("企业微信应用AgentId格式错误: %w", err)
	}

	client := resty.New()
	if proxyURL != "" {
		client.SetProxy(proxyURL)
	}

	var tokenResp dto.WechatWorkAccessTokenResponse
	httpResp, err := client.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetQueryParam("corpid", corpID).
		SetQueryParam("corpsecret", secret).
		SetResult(&tokenResp).
		Get("https://qyapi.weixin.qq.com/cgi-bin/gettoken")
	if err != nil {
		return fmt.Errorf("获取企业微信 access_token 失败: %w", err)
	}
	if httpResp.StatusCode() != 200 {
		return fmt.Errorf("获取企业微信 access_token HTTP 状态码异常: %d", httpResp.StatusCode())
	}
	if tokenResp.ErrCode != 0 {
		return fmt.Errorf("获取企业微信 access_token 失败: %s", tokenResp.ErrMsg)
	}

	var sendResp dto.WechatWorkSendMessageResponse
	httpResp, err = client.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetQueryParam("access_token", tokenResp.AccessToken).
		SetBody(dto.WechatWorkSendMessageRequest{
			ToUser:  toUser,
			MsgType: "text",
			AgentID: agentID,
			Text: dto.WechatWorkTextMessage{
				Content: content,
			},
			Safe: 0,
		}).
		SetResult(&sendResp).
		Post("https://qyapi.weixin.qq.com/cgi-bin/message/send")
	if err != nil {
		return fmt.Errorf("发送企业微信应用通知失败: %w", err)
	}
	if httpResp.StatusCode() != 200 {
		return fmt.Errorf("发送企业微信应用通知 HTTP 状态码异常: %d", httpResp.StatusCode())
	}
	if sendResp.ErrCode != 0 {
		return fmt.Errorf("发送企业微信应用通知失败: %s", sendResp.ErrMsg)
	}
	return nil
}

func buildOfflineNotificationMessage(req dto.LogoutNotificationRequest) (string, string) {
	if req.Type == "offline" {
		return "机器人掉线通知", fmt.Sprintf("您的机器人（%s）掉线啦~~~", vars.RobotRuntime.WxID)
	}
	return "机器人发送心跳失败", fmt.Sprintf("您的机器人（%s）第%d次发送心跳失败了~~~", vars.RobotRuntime.WxID, req.RetryCount)
}

func getTrimmedString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
