package service

import (
	"context"
	"log"
	"slices"
	"strings"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type ContactService struct {
	ctx     context.Context
	ctRespo *repository.Contact
}

func NewContactService(ctx context.Context) *ContactService {
	return &ContactService{
		ctx:     ctx,
		ctRespo: repository.NewContactRepo(ctx, vars.DB),
	}
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

func (s *ContactService) SyncContactByContactIDs(contactIDs []string) error {
	// 将ids拆分成二十个一个的数组之后再获取详情
	var contacts = make([]robot.Contact, 0)
	chunker := slices.Chunk(contactIDs, 20)
	processChunk := func(chunk []string) bool {
		// 获取昵称等详细信息
		var c = make([]robot.Contact, 0)
		c, err := vars.RobotRuntime.GetContactDetail(chunk)
		if err != nil {
			// 处理错误
			log.Printf("获取联系人详情失败: %v", err)
			return true
		}
		contacts = append(contacts, c...)
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
			contactPerson.Type = s.GetContactType(contactPerson)
			if contact.BigHeadImgUrl == "" {
				contactPerson.Avatar = contact.SmallHeadImgUrl
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
				CreatedAt:     time.Now().Unix(),
				UpdatedAt:     time.Now().Unix(),
			}
			if contact.BigHeadImgUrl == "" {
				contactPerson.Avatar = contact.SmallHeadImgUrl
			}
			contactPerson.Type = s.GetContactType(contactPerson)
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

func (s *ContactService) InsertOrUpdateContactActiveTime(contactID string) {
	existContact, err := s.ctRespo.GetContact(contactID)
	if err != nil {
		log.Printf("获取联系人失败: %v", err)
		return
	}
	// 群聊类型的联系人
	if strings.HasSuffix(contactID, "@chatroom") {
		if existContact == nil {
			contactChatRoom := model.Contact{
				WechatID:  contactID,
				Type:      model.ContactTypeChatRoom,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			}
			err = s.ctRespo.Create(&contactChatRoom)
			if err != nil {
				log.Printf("创建群聊联系人失败: %v", err)
				return
			}
		} else {
			// 存在，更新一下活跃时间
			contactChatRoom := model.Contact{
				ID:        existContact.ID,
				UpdatedAt: time.Now().Unix(),
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
			WechatID:  contactID,
			Type:      model.ContactTypeFriend,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		err = s.ctRespo.Create(&contact)
		if err != nil {
			log.Printf("创建好友失败: %v", err)
			return
		}
	} else {
		contact := model.Contact{
			ID:        existContact.ID,
			UpdatedAt: time.Now().Unix(),
		}
		err = s.ctRespo.Update(&contact)
		if err != nil {
			log.Printf("更新好友失败: %v", err)
			return
		}
	}
}
