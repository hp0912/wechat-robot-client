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
	ctx context.Context
}

func NewContactService(ctx context.Context) *ContactService {
	return &ContactService{
		ctx: ctx,
	}
}

func (s *ContactService) SyncContact(syncChatRoomMember bool) error {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return nil
	}
	// 先获取全部id
	var contactIds []string
	contactIds, err := vars.RobotRuntime.GetContactList()
	if err != nil {
		return err
	}

	respo := repository.NewContactRepo(s.ctx, vars.DB)
	recentGroupContacts, err := respo.FindRecentGroupContacts()
	if err != nil {
		return err
	}

	for _, groupContact := range recentGroupContacts {
		if !slices.Contains(contactIds, groupContact.WechatID) {
			contactIds = append(contactIds, groupContact.WechatID)
		}
	}
	// 同步群成员
	cmService := NewChatRoomService(s.ctx)
	if syncChatRoomMember {
		for _, groupContact := range recentGroupContacts {
			cmService.SyncChatRoomMember(groupContact.WechatID)
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
		if contact.UserName.String == nil {
			continue
		}
		if strings.TrimSpace(*contact.UserName.String) == "" {
			continue
		}
		validContactIds = append(validContactIds, *contact.UserName.String)
		// 判断数据库是否存在当前数据，不存在就新建，存在就更新
		existContact, err := respo.GetContact(vars.RobotRuntime.WxID, *contact.UserName.String)
		if err != nil {
			log.Printf("获取联系人失败: %v", err)
			continue
		}
		if existContact != nil {
			// 存在，修改
			contactPerson := model.Contact{
				ID:            existContact.ID,
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
			err = respo.Update(&contactPerson)
			if err != nil {
				log.Printf("更新联系人失败: %v", err)
				continue
			}
		} else {
			contactPerson := model.Contact{
				WechatID:      *contact.UserName.String,
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
			if strings.HasSuffix(*contact.UserName.String, "@chatroom") {
				contactPerson.Type = model.ContactTypeGroup
			}
			err = respo.Create(&contactPerson)
			if err != nil {
				log.Printf("创建联系人失败: %v", err)
				continue
			}
		}
	}
	return nil
}

func (s *ContactService) GetContacts(req dto.ContactListRequest, pager appx.Pager) ([]*model.Contact, int64, error) {
	req.Owner = vars.RobotRuntime.WxID
	respo := repository.NewContactRepo(s.ctx, vars.DB)
	return respo.GetByOwner(req, pager)
}

func (s *ContactService) InsertOrUpdateContactActiveTime(contactID string) {
	contactRespo := repository.NewContactRepo(s.ctx, vars.DB)
	existContact, err := contactRespo.GetContact(vars.RobotRuntime.WxID, contactID)
	if err != nil {
		log.Printf("获取联系人失败: %v", err)
		return
	}
	if strings.HasSuffix(contactID, "@chatroom") {
		if existContact == nil {
			contactGroup := model.Contact{
				WechatID:  contactID,
				Owner:     vars.RobotRuntime.WxID,
				Type:      model.ContactTypeGroup,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			}
			err = contactRespo.Create(&contactGroup)
			if err != nil {
				log.Printf("创建群聊联系人失败: %v", err)
				return
			}
		} else {
			// 存在，更新一下活跃时间
			contactGroup := model.Contact{
				ID:        existContact.ID,
				Owner:     vars.RobotRuntime.WxID,
				UpdatedAt: time.Now().Unix(),
			}
			err = contactRespo.Update(&contactGroup)
			if err != nil {
				log.Printf("更新群聊联系人失败: %v", err)
				return
			}
		}
	} else {
		// 普通联系人肯定存在，更新一下活跃时间就好了
		contactGroup := model.Contact{
			ID:        existContact.ID,
			Owner:     vars.RobotRuntime.WxID,
			UpdatedAt: time.Now().Unix(),
		}
		err = contactRespo.Update(&contactGroup)
		if err != nil {
			log.Printf("更新联系人活跃时间失败: %v", err)
			return
		}
	}
}
