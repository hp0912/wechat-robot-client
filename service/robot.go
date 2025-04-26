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
	vars.RobotRuntime.Status = robot.RobotStatusOnline
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: robot.RobotStatusOnline,
	}
	respo.Update(&robot)
}

func (r *RobotService) Offline() {
	vars.RobotRuntime.Status = robot.RobotStatusOffline
	if vars.RobotRuntime.HeartbeatCancel != nil {
		vars.RobotRuntime.HeartbeatCancel()
	}
	if vars.RobotRuntime.SyncMessageCancel != nil {
		vars.RobotRuntime.SyncMessageCancel()
	}
	respo := repository.NewRobotAdminRepo(r.ctx, vars.AdminDB)
	robot := model.RobotAdmin{
		ID:     vars.RobotRuntime.RobotID,
		Status: robot.RobotStatusOffline,
	}
	respo.Update(&robot)
}

func (r *RobotService) IsRunning() (result bool) {
	result = vars.RobotRuntime.IsRunning()
	if !result && vars.RobotRuntime.Status != robot.RobotStatusOffline {
		r.Offline()
	}
	return
}

func (r *RobotService) IsLoggedIn() (result bool) {
	result = vars.RobotRuntime.IsLoggedIn()
	if !result && vars.RobotRuntime.Status != robot.RobotStatusOffline {
		r.Offline()
	}
	return
}

func (r *RobotService) Login() (uuid string, awken bool, err error) {
	if vars.RobotRuntime.Status == robot.RobotStatusOnline {
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

func (r *RobotService) SyncMessage() {
	// 获取新消息
	syncResp, err := vars.RobotRuntime.SyncMessage()
	if err != nil {
		// 有可能是用户退出了，或者掉线了，这里不处理，由心跳机制处理机器人在线/离线状态
		log.Println("获取新消息失败: ", err)
		return
	}
	if len(syncResp.AddMsgs) == 0 {
		// 没有消息，直接返回
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf("消息入库出错了: %v", err)
		}
	}()
	respo := repository.NewMessageRepo(r.ctx, vars.DB)
	for _, message := range syncResp.AddMsgs {
		m := model.Message{
			MsgId:              message.NewMsgId,
			ClientMsgId:        message.MsgId,
			Type:               message.MsgType,
			Content:            message.Content.String,
			DisplayFullContent: message.PushContent,
			MessageSource:      message.MsgSource,
			FromWxID:           message.FromUserName.String,
			ToWxID:             message.ToWxid.String,
			CreatedAt:          time.Now().Unix(),
			UpdatedAt:          time.Now().Unix(),
		}
		self := vars.RobotRuntime.WxID
		// 处理一下自己发的消息
		// 自己发发到群聊
		if m.FromWxID == self && strings.HasSuffix(m.ToWxID, "@chatroom") {
			from := m.FromWxID
			to := m.ToWxID
			m.FromWxID = to
			m.ToWxID = from
		}
		// 群聊消息
		if strings.HasSuffix(m.FromWxID, "@chatroom") {
			m.IsGroup = true
			splitContents := strings.SplitN(m.Content, ":\n", 2)
			if len(splitContents) > 1 {
				m.Content = splitContents[1]
				m.SenderWxID = splitContents[0]
			} else {
				// 绝对是自己发的消息! qwq
				m.Content = splitContents[0]
				m.SenderWxID = self
			}
		} else {
			m.IsGroup = false
			m.SenderWxID = m.FromWxID
			m.FromWxID = m.ToWxID
		}
		// 是否艾特我的消息
		ats := vars.RobotRuntime.AtListDecoder(message.MsgSource)
		if ats != "" {
			atMembers := strings.Split(ats, ",")
			for _, at := range atMembers {
				if strings.Trim(at, " ") == self {
					m.IsAtMe = true
					break
				}
			}
		}
		respo.Create(&m)
	}
}

func (r *RobotService) SyncMessageStart() {
	ctx := context.Background()
	vars.RobotRuntime.SyncMessageContext, vars.RobotRuntime.SyncMessageCancel = context.WithCancel(ctx)
	for {
		select {
		case <-vars.RobotRuntime.SyncMessageContext.Done():
			return
		case <-time.After(1 * time.Second):
			r.SyncMessage()
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
		vars.RobotRuntime.Status = robot.RobotStatusOnline
		// 开启心跳
		go r.HeartbeatStart()
		// 开启消息同步
		go r.SyncMessageStart()
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
			Status:      robot.RobotStatusOnline,
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
