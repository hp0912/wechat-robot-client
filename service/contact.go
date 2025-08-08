package service

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

// 防抖逻辑：在 5 秒窗口内同一群聊只同步一次
const syncContactDebounceInterval = 5 * time.Second

var (
	syncContactMu     sync.Mutex
	syncContactTimers = make(map[string]*time.Timer)
)

type ContactService struct {
	ctx                 context.Context
	ctRespo             *repository.Contact
	crmRespo            *repository.ChatRoomMember
	sysmsgRespo         *repository.SystemMessage
	systemSettingsRespo *repository.SystemSettings
}

func NewContactService(ctx context.Context) *ContactService {
	return &ContactService{
		ctx:                 ctx,
		ctRespo:             repository.NewContactRepo(ctx, vars.DB),
		crmRespo:            repository.NewChatRoomMemberRepo(ctx, vars.DB),
		sysmsgRespo:         repository.NewSystemMessageRepo(ctx, vars.DB),
		systemSettingsRespo: repository.NewSystemSettingsRepo(ctx, vars.DB),
	}
}

func (s *ContactService) FriendSearch(req dto.FriendSearchRequest) (dto.FriendSearchResponse, error) {
	friend, err := vars.RobotRuntime.FriendSearch(robot.FriendSearchRequest{
		ToUserName:  req.ToUserName,
		FromScene:   0,
		SearchScene: 1,
	})
	if err != nil {
		return dto.FriendSearchResponse{}, err
	}
	if friend.UserName == nil || friend.UserName.String == nil || *friend.UserName.String == "" {
		return dto.FriendSearchResponse{}, fmt.Errorf("用户不存在")
	}
	if friend.AntispamTicket == nil || *friend.AntispamTicket == "" {
		return dto.FriendSearchResponse{}, fmt.Errorf("搜索用户失败，可能你们已经是好友了")
	}
	resp := dto.FriendSearchResponse{
		UserName:       *friend.UserName.String,
		NickName:       *friend.NickName.String,
		Avatar:         *friend.BigHeadImgUrl,
		AntispamTicket: *friend.AntispamTicket,
	}
	if friend.BigHeadImgUrl != nil && *friend.BigHeadImgUrl != "" {
		resp.Avatar = *friend.BigHeadImgUrl
	} else if friend.SmallHeadImgUrl != nil && *friend.SmallHeadImgUrl != "" {
		resp.Avatar = *friend.SmallHeadImgUrl
	}
	return resp, nil
}

func (s *ContactService) FriendSendRequest(req dto.FriendSendRequestRequest) error {
	_, err := vars.RobotRuntime.FriendSendRequest(robot.FriendSendRequestParam{
		Opcode:        2,
		V1:            req.V1,
		V2:            req.V2,
		VerifyContent: req.VerifyContent,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ContactService) FriendSendRequestFromChatRoom(req dto.FriendSendRequestFromChatRoomRequest) error {
	chatRoomMember, err := s.crmRespo.GetByID(req.ChatRoomMemberID)
	if err != nil {
		return fmt.Errorf("获取群成员信息失败: %v", err)
	}
	if chatRoomMember == nil {
		return fmt.Errorf("群成员不存在: %d", req.ChatRoomMemberID)
	}
	contact, err := s.ctRespo.GetContact(chatRoomMember.WechatID)
	if err != nil {
		return fmt.Errorf("获取联系人信息失败: %v", err)
	}
	if contact != nil {
		return fmt.Errorf("你们已经是好友了: %s", contact.WechatID)
	}
	c, err := vars.RobotRuntime.GetContactDetail(chatRoomMember.ChatRoomID, []string{chatRoomMember.WechatID})
	if err != nil {
		return fmt.Errorf("获取联系人详情失败: %v", err)
	}
	if len(c.Ticket) == 0 || c.Ticket[0].AntispamTicket == nil || *c.Ticket[0].AntispamTicket == "" {
		return fmt.Errorf("获取联系人密钥失败: %s", chatRoomMember.WechatID)
	}
	_, err = vars.RobotRuntime.FriendSendRequest(robot.FriendSendRequestParam{
		Opcode:        2,
		Scene:         14,
		V1:            *c.Ticket[0].Username,
		V2:            *c.Ticket[0].AntispamTicket,
		VerifyContent: req.VerifyContent,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ContactService) FriendSetRemarks(req dto.FriendSetRemarksRequest) error {
	_, err := vars.RobotRuntime.FriendSetRemarks(req.ToWxid, req.Remarks)
	if err != nil {
		return err
	}
	return s.ctRespo.UpdateRemarkByContactID(req.ToWxid, req.Remarks)
}

func (s *ContactService) FriendAutoPassVerify(systemMessageID int64) error {
	systemSettings, err := s.systemSettingsRespo.GetSystemSettings()
	if err != nil {
		return fmt.Errorf("获取系统设置失败: %v", err)
	}
	if systemSettings == nil || systemSettings.AutoVerifyUser == nil || !*systemSettings.AutoVerifyUser {
		return fmt.Errorf("系统设置未开启好友验证")
	}
	if systemSettings.VerifyUserDelay != nil && *systemSettings.VerifyUserDelay > 0 {
		// 延迟处理
		time.Sleep(time.Duration(*systemSettings.VerifyUserDelay) * time.Second)
	}
	return s.FriendPassVerify(systemMessageID)
}

func (s *ContactService) FriendPassVerify(systemMessageID int64) error {
	systemMessage, err := s.sysmsgRespo.GetByID(systemMessageID)
	if err != nil {
		return err
	}
	if systemMessage == nil {
		return fmt.Errorf("系统消息不存在: %d", systemMessageID)
	}
	if systemMessage.Type != model.SystemMessageTypeVerify {
		return fmt.Errorf("系统消息类型错误: %d", systemMessage.Type)
	}
	var xmlMessage robot.NewFriendMessage
	err = vars.RobotRuntime.XmlDecoder(systemMessage.Content, &xmlMessage)
	if err != nil {
		return fmt.Errorf("解析好友添加请求消息失败: %v", err)
	}
	scene, err := strconv.Atoi(xmlMessage.Scene)
	if err != nil {
		return fmt.Errorf("解析好友添加请求场景失败: %v", err)
	}
	userVerify, err := vars.RobotRuntime.FriendPassVerify(robot.FriendPassVerifyRequest{
		Scene: scene,
		V1:    xmlMessage.FromUsername,
		V2:    xmlMessage.Ticket,
	})
	if err != nil {
		return err
	}
	if userVerify.Username == nil || *userVerify.Username == "" {
		return fmt.Errorf("通过好友验证失败: %s，接口返回了空", systemMessage.FromWxid)
	}
	s.DebounceSyncContact(*userVerify.Username)
	err = s.sysmsgRespo.Update(&model.SystemMessage{
		ID:     systemMessage.ID,
		IsRead: true,
		Status: 1,
	})
	if err != nil {
		// 忽略错误
		log.Println("更新系统消息状态失败:", err)
	}
	return nil
}

func (s *ContactService) FriendDelete(contactID string) error {
	_, err := vars.RobotRuntime.FriendDelete(contactID)
	if err != nil {
		log.Printf("删除好友失败: %v", err)
		return err
	}
	return s.DeleteContactByContactID(contactID)
}

func (s *ContactService) GetContactType(contact model.Contact) model.ContactType {
	if strings.HasSuffix(contact.WechatID, "@chatroom") {
		return model.ContactTypeChatRoom
	}
	if _, ok := vars.OfficialAccount[contact.WechatID]; ok {
		return model.ContactTypeOfficialAccount
	}
	if strings.HasPrefix(contact.WechatID, "gh_") && contact.Sex == 0 {
		return model.ContactTypeOfficialAccount
	}
	return model.ContactTypeFriend
}

func (s *ContactService) SyncContact(syncChatRoomMember bool) error {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return nil
	}
	// 先获取全部好友id，没有保存到通讯录的群聊不会在这里
	var contactIds []string
	contactIds, err := vars.RobotRuntime.GetContactList()
	if err != nil {
		return err
	}
	// 查询没有保存到通讯录的群聊，只同步一天内活跃的群聊
	recentChatRoomContacts, err := s.ctRespo.FindRecentChatRoomContacts()
	if err != nil {
		return err
	}
	for _, chatRoomContact := range recentChatRoomContacts {
		if !slices.Contains(contactIds, chatRoomContact.WechatID) {
			contactIds = append(contactIds, chatRoomContact.WechatID)
		}
	}
	// 同步群成员
	cmService := NewChatRoomService(s.ctx)
	if syncChatRoomMember {
		for _, chatRoomContact := range recentChatRoomContacts {
			cmService.SyncChatRoomMember(chatRoomContact.WechatID)
		}
	}
	err = s.SyncContactByContactIDs(contactIds)
	if err != nil {
		log.Printf("同步联系人失败: %v", err)
		return err
	}
	return nil
}

func (s *ContactService) DebounceSyncContact(contactID string) {
	syncContactMu.Lock()
	defer syncContactMu.Unlock()

	if timer, ok := syncContactTimers[contactID]; ok {
		timer.Stop()
	}

	var newTimer *time.Timer
	newTimer = time.AfterFunc(syncContactDebounceInterval, func() {
		syncContactMu.Lock()
		currentTimer, ok := syncContactTimers[contactID]
		if !ok || currentTimer != newTimer {
			syncContactMu.Unlock()
			return
		}
		delete(syncContactTimers, contactID)
		syncContactMu.Unlock()

		// 真正执行同步逻辑（放在锁外避免长时间持锁）
		s.SyncContactByContactIDs([]string{contactID})
	})

	syncContactTimers[contactID] = newTimer
}

func (s *ContactService) SyncContactByContactIDs(contactIDs []string) error {
	now := time.Now().Unix()
	// 将ids拆分成二十个一个的数组之后再获取详情
	var contacts = make([]robot.Contact, 0)
	chunker := slices.Chunk(contactIDs, 20)
	processChunk := func(chunk []string) bool {
		// 获取昵称等详细信息
		r, err := vars.RobotRuntime.GetContactDetail("", chunk)
		if err != nil {
			// 处理错误
			log.Printf("获取联系人详情失败: %v", err)
			return true
		}
		contacts = append(contacts, r.ContactList...)
		return true
	}
	chunker(processChunk)
	for _, contact := range contacts {
		if contact.UserName.String == nil {
			continue
		}
		if strings.TrimSpace(*contact.UserName.String) == "" {
			continue
		}
		// 判断数据库是否存在当前数据，不存在就新建，存在就更新
		existContact, err := s.ctRespo.GetContact(*contact.UserName.String)
		if err != nil {
			log.Printf("获取联系人失败: %v", err)
			continue
		}
		if existContact != nil {
			// 存在，修改
			contactPerson := model.Contact{
				ID:            existContact.ID,
				WechatID:      *contact.UserName.String,
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
			if contact.Remark.String != nil && *contact.Remark.String != "" {
				contactPerson.Remark = *contact.Remark.String
			}
			contactPerson.Type = s.GetContactType(contactPerson)
			if contact.BigHeadImgUrl == "" {
				contactPerson.Avatar = contact.SmallHeadImgUrl
			}
			if contactPerson.Type == model.ContactTypeChatRoom {
				if contact.ChatRoomOwner != nil && *contact.ChatRoomOwner != "" {
					contactPerson.ChatRoomOwner = *contact.ChatRoomOwner
				}
			}
			err = s.ctRespo.Update(&contactPerson)
			if err != nil {
				log.Printf("更新联系人失败: %v", err)
				continue
			}
		} else {
			contactPerson := model.Contact{
				WechatID:      *contact.UserName.String,
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
				CreatedAt:     now,
				LastActiveAt:  now,
				UpdatedAt:     now,
			}
			if contact.BigHeadImgUrl == "" {
				contactPerson.Avatar = contact.SmallHeadImgUrl
			}
			if contact.Remark.String != nil && *contact.Remark.String != "" {
				contactPerson.Remark = *contact.Remark.String
			}
			contactPerson.Type = s.GetContactType(contactPerson)
			if contactPerson.Type == model.ContactTypeChatRoom {
				if contact.ChatRoomOwner != nil && *contact.ChatRoomOwner != "" {
					contactPerson.ChatRoomOwner = *contact.ChatRoomOwner
				}
			}
			err = s.ctRespo.Create(&contactPerson)
			if err != nil {
				log.Printf("创建联系人失败: %v", err)
				continue
			}
		}
	}
	return nil
}

func (s *ContactService) GetContacts(req dto.ContactListRequest, pager appx.Pager) ([]*model.Contact, int64, error) {
	return s.ctRespo.GetContacts(req, pager)
}

func (s *ContactService) DeleteContactByContactID(contactID string) error {
	return s.ctRespo.DeleteByContactID(contactID)
}

func (s *ContactService) InsertOrUpdateContactActiveTime(contactID string) {
	now := time.Now().Unix()
	existContact, err := s.ctRespo.GetContact(contactID)
	if err != nil {
		log.Printf("获取联系人失败: %v", err)
		return
	}
	// 群聊类型的联系人
	if strings.HasSuffix(contactID, "@chatroom") {
		if existContact == nil {
			contactChatRoom := model.Contact{
				WechatID:     contactID,
				Type:         model.ContactTypeChatRoom,
				CreatedAt:    now,
				LastActiveAt: now,
				UpdatedAt:    now,
			}
			err = s.ctRespo.Create(&contactChatRoom)
			if err != nil {
				log.Printf("创建群聊联系人失败: %v", err)
				return
			}
		} else {
			// 存在，更新一下活跃时间
			contactChatRoom := model.Contact{
				ID:           existContact.ID,
				LastActiveAt: now,
			}
			err = s.ctRespo.Update(&contactChatRoom)
			if err != nil {
				log.Printf("更新群聊联系人失败: %v", err)
				return
			}
		}
		return
	}
	// 好友类型的联系人
	if existContact == nil {
		contact := model.Contact{
			WechatID:     contactID,
			Type:         model.ContactTypeFriend,
			CreatedAt:    now,
			LastActiveAt: now,
			UpdatedAt:    now,
		}
		contact.Type = s.GetContactType(contact)
		err = s.ctRespo.Create(&contact)
		if err != nil {
			log.Printf("创建好友失败: %v", err)
			return
		}
	} else {
		contact := model.Contact{
			ID:           existContact.ID,
			LastActiveAt: now,
		}
		err = s.ctRespo.Update(&contact)
		if err != nil {
			log.Printf("更新好友失败: %v", err)
			return
		}
	}
}
