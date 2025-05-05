package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
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
	if vars.RobotRuntime.SyncMessageCancel != nil {
		vars.RobotRuntime.SyncMessageCancel()
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

func (r *RobotService) Login() (uuid string, awkenLogin, autoLogin bool, err error) {
	if vars.RobotRuntime.Status == model.RobotStatusOnline {
		err = errors.New("您已经登陆，可以尝试刷新机器人状态")
		return
	}
	uuid, awkenLogin, autoLogin, err = vars.RobotRuntime.Login()
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
					r.Offline()
					return
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
			ToWxID:             message.ToUserName.String,
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
			if m.FromWxID == self {
				m.FromWxID = m.ToWxID
				m.ToWxID = self
			}
		}
		if m.Type == model.MsgTypeInit || m.Type == model.MsgTypeUnknow {
			continue
		}
		if m.Type == model.MsgTypeRecalled && m.SenderWxID == "weixin" {
			continue
		}
		if m.Type == model.MsgTypeApp {
			subTypeStr := vars.RobotRuntime.XmlFastDecoder(m.Content, "type")
			if subTypeStr != "" {
				subType, err := strconv.Atoi(subTypeStr)
				if err != nil {
					continue
				}
				m.AppMsgType = model.AppMessageType(subType)
				if m.AppMsgType == model.AppMsgTypeAttachUploading {
					continue
				}
			}
		}
		// 正常撤回的消息
		if m.Type == model.MsgTypeRecalled {
			oldMsg := respo.GetByMsgID(m.MsgId)
			if oldMsg != nil {
				oldMsg.IsRecalled = true
				respo.Update(oldMsg)
				continue
			}
		}
		// 是否艾特我的消息
		ats := vars.RobotRuntime.XmlFastDecoder(message.MsgSource, "atuserlist")
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
		// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
		if strings.HasSuffix(m.FromWxID, "@chatroom") {
			contactRespo := repository.NewContactRepo(r.ctx, vars.DB)
			isExist := contactRespo.ExistsByWeChatID(m.FromWxID)
			if !isExist {
				contactGroup := model.Contact{
					WechatID:  m.FromWxID,
					Owner:     vars.RobotRuntime.WxID,
					Type:      model.ContactTypeGroup,
					CreatedAt: time.Now().Unix(),
					UpdatedAt: time.Now().Unix(),
				}
				contactRespo.Create(&contactGroup)
			} else {
				// 存在，更新一下活跃时间
				contactGroup := model.Contact{
					Owner:     vars.RobotRuntime.WxID,
					UpdatedAt: time.Now().Unix(),
				}
				contactRespo.UpdateColumnsByWhere(&contactGroup, map[string]any{
					"wechat_id": m.FromWxID,
				})
			}
		}
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
		vars.RobotRuntime.Status = model.RobotStatusOnline
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

func (r *RobotService) SyncChatRoomMember(chatRoomID string) {
	var chatRoomMembers []robot.ChatRoomMember
	var err error
	chatRoomMembers, err = vars.RobotRuntime.GetChatRoomMemberDetail(chatRoomID)
	if err != nil {
		log.Printf("获取群[%s]成员失败: %v", chatRoomID, err)
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf("获取群[%s]成员失败: %v", chatRoomID, err)
		}
	}()
	// 遍历获取到的群成员列表，如果数据库存在，则更新，数据库不存在则新增
	if len(chatRoomMembers) > 0 {
		memberRepo := repository.NewChatRoomMemberRepo(r.ctx, vars.DB)
		now := time.Now().Unix()

		// 获取当前成员的微信ID列表，用于后续比对
		currentMemberIDs := make([]string, 0, len(chatRoomMembers))
		for _, member := range chatRoomMembers {
			currentMemberIDs = append(currentMemberIDs, member.UserName)
		}

		for _, member := range chatRoomMembers {
			// 检查成员是否已存在
			exists := memberRepo.ExistsByWhere(map[string]any{
				"chat_room_id": chatRoomID,
				"wechat_id":    member.UserName,
			})
			if exists {
				// 更新现有成员
				updateMember := map[string]any{
					"nickname":  member.NickName,
					"avatar":    member.SmallHeadImgUrl,
					"is_leaved": false, // 确保标记为未离开
					"leaved_at": nil,   // 清除离开时间
				}
				// 更新数据库中已有的记录
				memberRepo.UpdateColumnsByWhere(&updateMember, map[string]any{
					"chat_room_id": chatRoomID,
					"wechat_id":    member.UserName,
				})
			} else {
				// 创建新成员
				newMember := model.ChatRoomMember{
					ChatRoomID:      chatRoomID,
					WechatID:        member.UserName,
					Nickname:        member.NickName,
					Avatar:          member.SmallHeadImgUrl,
					InviterWechatID: member.InviterUserName,
					IsLeaved:        false,
					JoinedAt:        now,
					LastActiveAt:    now,
				}
				memberRepo.Create(&newMember)
			}
		}
		// 查询数据库中该群的所有成员
		dbMembers := memberRepo.ListByWhere(nil, map[string]any{
			"chat_room_id": chatRoomID,
			"is_leaved":    false, // 只处理未离开的成员
		})
		// 标记已离开的成员
		for _, dbMember := range dbMembers {
			if !slices.Contains(currentMemberIDs, dbMember.WechatID) {
				// 数据库有记录但当前群成员列表中不存在，标记为已离开
				leaveTime := now
				updateMember := model.ChatRoomMember{
					IsLeaved: true,
					LeavedAt: &leaveTime,
				}
				memberRepo.UpdateColumnsByWhere(&updateMember, map[string]any{
					"chat_room_id": chatRoomID,
					"wechat_id":    dbMember.WechatID,
				})
			}
		}
	}
}

func (r *RobotService) SyncContact(syncChatRoomMember bool) (err error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return
	}
	// 先获取全部id
	var contactIds []string
	contactIds, err = vars.RobotRuntime.GetContactList()
	if err != nil {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf("同步联系人出错了: %v", err)
		}
	}()
	respo := repository.NewContactRepo(r.ctx, vars.DB)
	recentGroupContacts := respo.FindRecentGroupContacts()
	for _, groupContact := range recentGroupContacts {
		if !slices.Contains(contactIds, groupContact.WechatID) {
			contactIds = append(contactIds, groupContact.WechatID)
		}
	}
	// 同步群成员
	if syncChatRoomMember {
		for _, groupContact := range recentGroupContacts {
			r.SyncChatRoomMember(groupContact.WechatID)
		}
	}

	// 将ids拆分成二十个一个的数组之后再获取详情
	var contacts = make([]robot.Contact, 0)
	chunker := slices.Chunk(contactIds, 20)
	processChunk := func(chunk []string) bool {
		// 获取昵称等详细信息
		var c = make([]robot.Contact, 0)
		c, err = vars.RobotRuntime.GetContactDetail(chunk)
		if err != nil {
			// 处理错误
			log.Printf("获取联系人详情失败: %v", err)
			return false
		}
		contacts = append(contacts, c...)
		return true
	}
	chunker(processChunk)
	validContactIds := make([]string, 0)
	for _, contact := range contacts {
		if strings.TrimSpace(contact.UserName.String) == "" {
			continue
		}
		validContactIds = append(validContactIds, contact.UserName.String)
		// 判断数据库是否存在当前数据，不存在就新建，存在就更新
		isExist := respo.ExistsByWeChatID(contact.UserName.String)
		if isExist {
			// 存在，修改
			contactPerson := model.Contact{
				Owner:         vars.RobotRuntime.WxID,
				Alias:         contact.Alias,
				Nickname:      contact.NickName.String,
				Avatar:        contact.BigHeadImgUrl,
				Pyinitial:     contact.Pyinitial.String,
				QuanPin:       contact.QuanPin.String,
				Sex:           contact.Sex,
				Country:       contact.Country,
				Province:      contact.Province,
				City:          contact.City,
				Signature:     contact.Signature,
				SnsBackground: contact.SnsUserInfo.SnsBgimgId,
			}
			if contact.BigHeadImgUrl == "" {
				contactPerson.Avatar = contact.SmallHeadImgUrl
			}
			respo.UpdateColumnsByWhere(&contactPerson, map[string]any{
				"wechat_id": contact.UserName.String,
			})
		} else {
			contactPerson := model.Contact{
				WechatID:      contact.UserName.String,
				Owner:         vars.RobotRuntime.WxID,
				Alias:         contact.Alias,
				Nickname:      contact.NickName.String,
				Avatar:        contact.BigHeadImgUrl,
				Type:          model.ContactTypeFriend,
				Pyinitial:     contact.Pyinitial.String,
				QuanPin:       contact.QuanPin.String,
				Sex:           contact.Sex,
				Country:       contact.Country,
				Province:      contact.Province,
				City:          contact.City,
				Signature:     contact.Signature,
				SnsBackground: contact.SnsUserInfo.SnsBgimgId,
				CreatedAt:     time.Now().Unix(),
				UpdatedAt:     time.Now().Unix(),
			}
			if contact.BigHeadImgUrl == "" {
				contactPerson.Avatar = contact.SmallHeadImgUrl
			}
			if strings.HasSuffix(contact.UserName.String, "@chatroom") {
				contactPerson.Type = model.ContactTypeGroup
			}
			respo.Create(&contactPerson)
		}
	}
	return
}

func (r *RobotService) GetContacts(req dto.ContactListRequest, pager appx.Pager) ([]*model.Contact, int64, error) {
	req.Owner = vars.RobotRuntime.WxID
	respo := repository.NewContactRepo(r.ctx, vars.DB)
	return respo.GetByOwner(req, pager)
}

func (r *RobotService) GetChatRoomMembers(req dto.ChatRoomMemberRequest, pager appx.Pager) ([]*model.ChatRoomMember, int64, error) {
	respo := repository.NewChatRoomMemberRepo(r.ctx, vars.DB)
	return respo.GetByChatRoomID(req, pager)
}

func (r *RobotService) GetChatHistory(req dto.ChatHistoryRequest, pager appx.Pager) ([]*model.Message, int64, error) {
	req.Owner = vars.RobotRuntime.WxID
	respo := repository.NewMessageRepo(r.ctx, vars.DB)
	return respo.GetByContactID(req, pager)
}
