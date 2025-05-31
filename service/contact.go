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

func (s *ContactService) SyncContact(syncChatRoomMember bool) (err error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return
	}
	// 先获取全部id
	var contactIds []string
	contactIds, err = vars.RobotRuntime.GetContactList()
	if err != nil {
		return
	}
	respo := repository.NewContactRepo(s.ctx, vars.DB)
	recentGroupContacts := respo.FindRecentGroupContacts()
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
		isExist := respo.ExistsByWeChatID(*contact.UserName.String)
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
				"wechat_id": *contact.UserName.String,
			})
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
			respo.Create(&contactPerson)
		}
	}
	return
}

func (s *ContactService) GetContacts(req dto.ContactListRequest, pager appx.Pager) ([]*model.Contact, int64, error) {
	req.Owner = vars.RobotRuntime.WxID
	respo := repository.NewContactRepo(s.ctx, vars.DB)
	return respo.GetByOwner(req, pager)
}

func (s *ContactService) InsertOrUpdateContactActiveTime(contactID string) {
	contactRespo := repository.NewContactRepo(s.ctx, vars.DB)
	if strings.HasSuffix(contactID, "@chatroom") {
		isExist := contactRespo.ExistsByWeChatID(contactID)
		if !isExist {
			contactGroup := model.Contact{
				WechatID:  contactID,
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
				"wechat_id": contactID,
			})
		}
	} else {
		// 普通联系人肯定存在，更新一下活跃时间就好了
		contactGroup := model.Contact{
			Owner:     vars.RobotRuntime.WxID,
			UpdatedAt: time.Now().Unix(),
		}
		contactRespo.UpdateColumnsByWhere(&contactGroup, map[string]any{
			"wechat_id": contactID,
		})
	}
}
