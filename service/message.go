package service

import (
	"context"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type MessageService struct {
	ctx             context.Context
	settings        settings.Settings
	msgRespo        *repository.Message
	crmRespo        *repository.ChatRoomMember
	sysmsgRespo     *repository.SystemMessage
	robotAdminRespo *repository.RobotAdmin
}

var _ plugin.MessageServiceIface = (*MessageService)(nil)

func NewMessageService(ctx context.Context) *MessageService {
	return &MessageService{
		ctx:             ctx,
		msgRespo:        repository.NewMessageRepo(ctx, vars.DB),
		crmRespo:        repository.NewChatRoomMemberRepo(ctx, vars.DB),
		sysmsgRespo:     repository.NewSystemMessageRepo(ctx, vars.DB),
		robotAdminRespo: repository.NewRobotAdminRepo(ctx, vars.AdminDB),
	}
}

// ProcessTextMessage 处理文本消息
func (s *MessageService) ProcessTextMessage(message *model.Message) {
	msgCtx := &plugin.MessageContext{
		Context:        s.ctx,
		Settings:       s.settings,
		Message:        message,
		MessageContent: message.Content,
		MessageService: s,
	}
	for _, messagePlugin := range vars.MessagePlugin.Plugins {
		abort := messagePlugin.Run(msgCtx)
		if abort {
			return
		}
	}
}

// ProcessImageMessage 处理图片消息
func (s *MessageService) ProcessImageMessage(message *model.Message) {

}

// ProcessVoiceMessage 处理语音消息
func (s *MessageService) ProcessVoiceMessage(message *model.Message) {

}

// ProcessVideoMessage 处理视频消息
func (s *MessageService) ProcessVideoMessage(message *model.Message) {

}

// ProcessEmojiMessage 处理表情消息
func (s *MessageService) ProcessEmojiMessage(message *model.Message) {

}

// ProcessReferMessage 处理引用消息
func (s *MessageService) ProcessReferMessage(message *model.Message) {
	var xmlMessage robot.XmlMessage
	err := vars.RobotRuntime.XmlDecoder(message.Content, &xmlMessage)
	if err != nil {
		log.Printf("解析引用消息失败: %v", err)
		return
	}
	referMessageID, err := strconv.ParseInt(xmlMessage.AppMsg.ReferMsg.SvrID, 10, 64)
	if err != nil {
		log.Printf("解析引用消息ID失败: %v", err)
		return
	}
	referMessage, err := s.msgRespo.GetByMsgID(referMessageID)
	if err != nil {
		log.Printf("获取引用消息失败: %v", err)
		return
	}
	if referMessage == nil {
		log.Printf("获取引用消息为空")
		return
	}
	msgCtx := &plugin.MessageContext{
		Context:        s.ctx,
		Settings:       s.settings,
		Message:        message,
		MessageContent: xmlMessage.AppMsg.Title,
		ReferMessage:   referMessage,
		MessageService: s,
	}
	for _, messagePlugin := range vars.MessagePlugin.Plugins {
		abort := messagePlugin.Run(msgCtx)
		if abort {
			return
		}
	}
}

// ProcessAppMessage 处理应用消息
func (s *MessageService) ProcessAppMessage(message *model.Message) {
	if message.AppMsgType == model.AppMsgTypequote {
		s.ProcessReferMessage(message)
		return
	}
	if message.AppMsgType == model.AppMsgTypeUrl {
		xmlMessage, err := s.XmlDecoder(message.Content)
		if err != nil {
			log.Printf("解析应用消息失败: %v", err)
			return
		}
		if xmlMessage.AppMsg.Title == "邀请你加入群聊" || xmlMessage.AppMsg.Title == "Group Chat Invitation" {
			now := time.Now().Unix()
			err := s.sysmsgRespo.Create(&model.SystemMessage{
				MsgID:       message.MsgId,
				ClientMsgID: message.ClientMsgId,
				Type:        model.SystemMessageTypeJoinChatRoom,
				ImageURL:    xmlMessage.AppMsg.ThumbURL,
				Description: xmlMessage.AppMsg.Des,
				Content:     message.Content,
				FromWxid:    message.FromWxID,
				ToWxid:      message.ToWxID,
				Status:      0,
				IsRead:      false,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
			if err != nil {
				log.Printf("入库邀请进群通知消息失败: %v", err)
				return
			}
			if message.ID > 0 {
				// 消息已经没什么用了，删除掉
				err := s.msgRespo.Delete(message)
				if err != nil {
					log.Printf("删除消息失败: %v", err)
					return
				}
			}
			return
		}
		return
	}
}

// ProcessShareCardMessage 处理分享名片消息
func (s *MessageService) ProcessShareCardMessage(message *model.Message) {

}

// ProcessFriendVerifyMessage 处理好友添加请求通知消息
func (s *MessageService) ProcessFriendVerifyMessage(message *model.Message) {
	now := time.Now().Unix()
	var xmlMessage robot.NewFriendMessage
	err := vars.RobotRuntime.XmlDecoder(message.Content, &xmlMessage)
	if err != nil {
		log.Printf("解析好友添加请求消息失败: %v", err)
		return
	}

	systeMessage := model.SystemMessage{
		MsgID:       message.MsgId,
		ClientMsgID: message.ClientMsgId,
		Type:        model.SystemMessageTypeVerify,
		ImageURL:    xmlMessage.BigHeadImgURL,
		Description: xmlMessage.Content,
		Content:     message.Content,
		FromWxid:    message.FromWxID,
		ToWxid:      message.ToWxID,
		Status:      0,
		IsRead:      false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err = s.sysmsgRespo.Create(&systeMessage)
	if err != nil {
		log.Printf("入库好友添加请求通知消息失败: %v", err)
		return
	}

	// 自动通过好友
	go func(systemSettingsID int64) {
		err := NewContactService(context.Background()).FriendAutoPassVerify(systemSettingsID)
		if err != nil {
			log.Printf("自动通过好友验证失败: %v", err)
		}
	}(systeMessage.ID)

	if message.ID > 0 {
		// 消息已经没什么用了，删除掉
		err := s.msgRespo.Delete(message)
		if err != nil {
			log.Printf("删除消息失败: %v", err)
			return
		}
	}
}

// ProcessRecalledMessage 处理撤回消息
func (s *MessageService) ProcessRecalledMessage(message *model.Message, msgXml robot.SystemMessage) {
	oldMsg, err := s.msgRespo.GetByMsgID(msgXml.RevokeMsg.NewMsgID)
	if err != nil {
		log.Printf("获取撤回的消息失败: %v", err)
		return
	}
	if oldMsg != nil {
		oldMsg.IsRecalled = true
		err = s.msgRespo.Update(oldMsg)
		if err != nil {
			log.Printf("标记撤回消息失败: %v", err)
		} else {
			if message.ID > 0 {
				// 消息已经没什么用了，删除掉
				err := s.msgRespo.Delete(message)
				if err != nil {
					log.Printf("删除消息失败: %v", err)
					return
				}
			}
		}
		return
	}
}

// ProcessPatMessage 处理拍一拍消息
func (s *MessageService) ProcessPatMessage(message *model.Message, msgXml robot.SystemMessage) {
	msgCtx := &plugin.MessageContext{
		Context:        s.ctx,
		Settings:       s.settings,
		Message:        message,
		MessageContent: message.Content,
		Pat:            message.IsChatRoom && msgXml.Pat.PattedUsername == vars.RobotRuntime.WxID,
		MessageService: s,
	}
	for _, messagePlugin := range vars.MessagePlugin.Plugins {
		if slices.Contains(messagePlugin.GetLabels(), "pat") {
			abort := messagePlugin.Run(msgCtx)
			if abort {
				return
			}
		}
	}
}

func (s *MessageService) ProcessNewChatRoomMemberMessage(message *model.Message, msgXml robot.SystemMessage) {
	var newMemberWechatIds []string
	if len(msgXml.SysMsgTemplate.ContentTemplate.LinkList.Links) > 0 {
		links := msgXml.SysMsgTemplate.ContentTemplate.LinkList.Links
		for _, link := range links {
			if link.Name == "names" || link.Name == "adder" {
				if link.MemberList != nil {
					for _, member := range link.MemberList.Members {
						newMemberWechatIds = append(newMemberWechatIds, member.Username)
					}
				}
			}
		}
	}
	newMembers, err := NewChatRoomService(s.ctx).UpdateChatRoomMembersOnNewMemberJoinIn(message.FromWxID, newMemberWechatIds)
	if err != nil {
		log.Printf("邀请新成员加入群聊时，更新群成员失败: %v", err)
	}
	if len(newMembers) == 0 {
		log.Println("根据新成员微信ID获取群成员信息失败，没查询到有效的成员信息")
	}
	welcomeConfig, err := NewChatRoomSettingsService(s.ctx).GetChatRoomWelcomeConfig(message.FromWxID)
	if err != nil {
		log.Printf("获取群聊欢迎配置失败: %v", err)
		return
	}
	if welcomeConfig.WelcomeEnabled != nil && !*welcomeConfig.WelcomeEnabled {
		log.Printf("[%s]群聊欢迎消息未启用", message.FromWxID)
		return
	}
	if welcomeConfig.WelcomeType == model.WelcomeTypeText {
		s.SendTextMessage(message.FromWxID, welcomeConfig.WelcomeText)
	}
	if welcomeConfig.WelcomeType == model.WelcomeTypeEmoji {
		s.SendEmoji(message.FromWxID, welcomeConfig.WelcomeEmojiMD5, int32(welcomeConfig.WelcomeEmojiLen))
	}
	if welcomeConfig.WelcomeType == model.WelcomeTypeImage {
		resp, err := resty.New().R().SetDoNotParseResponse(true).Get(welcomeConfig.WelcomeImageURL)
		if err != nil {
			log.Println("获取欢迎图片失败: ", err)
			return
		}
		defer resp.RawBody().Close()
		// 创建临时文件
		tempFile, err := os.CreateTemp("", "welcome_image_*")
		if err != nil {
			log.Println("创建临时文件失败: ", err)
			return
		}
		defer tempFile.Close()
		defer os.Remove(tempFile.Name()) // 清理临时文件
		// 将图片数据写入临时文件
		_, err = io.Copy(tempFile, resp.RawBody())
		if err != nil {
			log.Println("将图片数据写入临时文件失败: ", err)
			return
		}
		_, err = s.MsgUploadImg(message.FromWxID, tempFile)
		if err != nil {
			log.Println("发送欢迎图片消息失败: ", err)
			return
		}
	}
	if welcomeConfig.WelcomeType == model.WelcomeTypeURL {
		if len(newMembers) == 0 {
			return
		}
		var title string
		if len(newMembers) > 1 {
			title = fmt.Sprintf("欢迎%d位家人加入群聊", len(newMembers))
		} else if newMembers[0].Nickname != "" {
			title = fmt.Sprintf("欢迎%s加入群聊", newMembers[0].Nickname)
		} else {
			title = "欢迎新成员加入群聊"
		}
		err := s.ShareLink(message.FromWxID, robot.ShareLinkMessage{
			Title:    title,
			Des:      welcomeConfig.WelcomeText,
			Url:      welcomeConfig.WelcomeURL,
			ThumbUrl: robot.CDATAString(newMembers[0].Avatar),
		})
		if err != nil {
			log.Println("发送欢迎链接消息失败: ", err)
		}
	}
}

// ProcessSystemMessage 处理系统消息
func (s *MessageService) ProcessSystemMessage(message *model.Message) {
	var msgXml robot.SystemMessage
	err := vars.RobotRuntime.XmlDecoder(message.Content, &msgXml)
	if err != nil {
		return
	}
	if msgXml.Type == "revokemsg" {
		s.ProcessRecalledMessage(message, msgXml)
		return
	}
	if msgXml.Type == "pat" {
		s.ProcessPatMessage(message, msgXml)
		return
	}
	if msgXml.Type == "sysmsgtemplate" &&
		(strings.Contains(msgXml.SysMsgTemplate.ContentTemplate.Template, "加入了群聊") ||
			strings.Contains(msgXml.SysMsgTemplate.ContentTemplate.Template, "分享的二维码加入群聊") ||
			strings.Contains(msgXml.SysMsgTemplate.ContentTemplate.Template, "joined group chat")) {
		s.ProcessNewChatRoomMemberMessage(message, msgXml)
		return
	}
}

// ProcessLocationMessage 处理位置消息
func (s *MessageService) ProcessLocationMessage(message *model.Message) {

}

// ProcessPromptMessage 处理提示消息
func (s *MessageService) ProcessPromptMessage(message *model.Message) {

}

func (s *MessageService) ProcessMessageSender(message *model.Message) {
	self := vars.RobotRuntime.WxID
	// 处理一下自己发的消息
	// 自己发发到群聊
	if message.FromWxID == self && strings.HasSuffix(message.ToWxID, "@chatroom") {
		from := message.FromWxID
		to := message.ToWxID
		message.FromWxID = to
		message.ToWxID = from
	}
	// 群聊消息
	if strings.HasSuffix(message.FromWxID, "@chatroom") {
		message.IsChatRoom = true
		splitContents := strings.SplitN(message.Content, ":\n", 2)
		if len(splitContents) > 1 {
			message.Content = splitContents[1]
			message.SenderWxID = splitContents[0]
		} else {
			// 绝对是自己发的消息! qwq
			message.Content = splitContents[0]
			message.SenderWxID = self
		}
	} else {
		message.IsChatRoom = false
		message.SenderWxID = message.FromWxID
		if message.FromWxID == self {
			message.FromWxID = message.ToWxID
			message.ToWxID = self
		}
	}
}

func (s *MessageService) ProcessMessageShouldInsertToDB(message *model.Message) bool {
	if message.Type == model.MsgTypeInit || message.Type == model.MsgTypeUnknow {
		return false
	}
	if message.Type == model.MsgTypeSystem && message.SenderWxID == "weixin" {
		return false
	}
	if message.Type == model.MsgTypeApp {
		subTypeStr := vars.RobotRuntime.XmlFastDecoder(message.Content, "type")
		if subTypeStr != "" {
			subType, err := strconv.Atoi(subTypeStr)
			if err == nil {
				message.AppMsgType = model.AppMessageType(subType)
				if message.AppMsgType == model.AppMsgTypeAttachUploading {
					// 如果是上传中的应用消息，则不入库
					return false
				}
			}
		}
	}
	return true
}

// ProcessMentionedMeMessage 处理下艾特我的消息
func (s *MessageService) ProcessMentionedMeMessage(message *model.Message, msgSource string) {
	self := vars.RobotRuntime.WxID
	// 是否艾特我的消息
	ats := vars.RobotRuntime.XmlFastDecoder(msgSource, "atuserlist")
	if ats != "" {
		atMembers := strings.Split(ats, ",")
		for _, at := range atMembers {
			if strings.Trim(at, " ") == self {
				message.IsAtMe = true
				break
			}
		}
	}
}

func (s *MessageService) InitSettingsByMessage(message *model.Message) (settings settings.Settings) {
	if message.IsChatRoom {
		settings = NewChatRoomSettingsService(s.ctx)
	} else {
		settings = NewFriendSettingsService(s.ctx)
	}
	err := settings.InitByMessage(message)
	if err != nil {
		log.Println("初始化设置失败: ", err)
		return nil
	}
	return
}

func (s *MessageService) ProcessMessage(syncResp robot.SyncMessage) {
	for _, message := range syncResp.AddMsgs {
		m := model.Message{
			MsgId:              message.NewMsgId,
			ClientMsgId:        message.MsgId,
			Type:               message.MsgType,
			Content:            *message.Content.String,
			DisplayFullContent: message.PushContent,
			MessageSource:      message.MsgSource,
			FromWxID:           *message.FromUserName.String,
			ToWxID:             *message.ToUserName.String,
			CreatedAt:          message.CreateTime,
			UpdatedAt:          time.Now().Unix(),
		}
		s.ProcessMessageSender(&m)
		if !s.ProcessMessageShouldInsertToDB(&m) {
			continue
		}
		s.ProcessMentionedMeMessage(&m, message.MsgSource)
		settings := s.InitSettingsByMessage(&m)
		if settings == nil {
			continue
		}
		s.settings = settings
		err := s.msgRespo.Create(&m)
		if err != nil {
			log.Printf("入库消息失败: %v", err)
			continue
		}
		switch m.Type {
		case model.MsgTypeText:
			go s.ProcessTextMessage(&m)
		case model.MsgTypeImage:
			go s.ProcessImageMessage(&m)
		case model.MsgTypeVoice:
			go s.ProcessVoiceMessage(&m)
		case model.MsgTypeVideo:
			go s.ProcessVideoMessage(&m)
		case model.MsgTypeEmoticon:
			go s.ProcessEmojiMessage(&m)
		case model.MsgTypeApp:
			go s.ProcessAppMessage(&m)
		case model.MsgTypeShareCard:
			go s.ProcessShareCardMessage(&m)
		case model.MsgTypeVerify:
			// 好友添加请求通知消息
			go s.ProcessFriendVerifyMessage(&m)
		case model.MsgTypeSystem:
			go s.ProcessSystemMessage(&m)
		case model.MsgTypeLocation:
			go s.ProcessLocationMessage(&m)
		case model.MsgTypePrompt:
			go s.ProcessPromptMessage(&m)
		default:
			// 未知消息类型
			log.Printf("未知消息类型: %d, 内容: %s", m.Type, m.Content)
		}
		go func() {
			// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
			NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)
		}()
	}
	for _, contact := range syncResp.ModContacts {
		if contact.UserName.String != nil {
			if strings.HasSuffix(*contact.UserName.String, "@chatroom") {
				// 群成员信息有变化，更新群聊成员（防抖，5 秒内只执行最后一次）
				NewChatRoomService(context.Background()).DebounceSyncChatRoomMember(*contact.UserName.String)
			} else {
				// 更新联系人信息
				NewContactService(context.Background()).DebounceSyncContact(*contact.UserName.String)
			}
		}
	}
	for _, contact := range syncResp.DelContacts {
		if contact.UserName.String != nil {
			err := NewContactService(context.Background()).DeleteContactByContactID(*contact.UserName.String)
			if err != nil {
				log.Println("删除联系人失败: ", err)
			}
		}
	}
}

func (s *MessageService) SyncMessage() {
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
	s.ProcessMessage(syncResp)
}

func (s *MessageService) XmlDecoder(content string) (robot.XmlMessage, error) {
	var xmlMessage robot.XmlMessage
	err := vars.RobotRuntime.XmlDecoder(content, &xmlMessage)
	if err != nil {
		return xmlMessage, err
	}
	return xmlMessage, nil
}

func (s *MessageService) MessageRevoke(req dto.MessageCommonRequest) error {
	message, err := s.msgRespo.GetByID(req.MessageID)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	if message == nil {
		return errors.New("消息不存在")
	}
	// 两分钟前
	if message.CreatedAt+120 < time.Now().Unix() {
		return errors.New("消息已过期")
	}
	return vars.RobotRuntime.MessageRevoke(*message)
}

func (s *MessageService) SendTextMessage(toWxID, content string, at ...string) error {
	atContent := ""
	if len(at) > 0 {
		// 手动拼接上 @ 符号和昵称
		for index, wxid := range at {
			var targetNickname string

			if strings.HasSuffix(toWxID, "@chatroom") {
				// 群聊消息，昵称优先取群备注，备注取不到或者取失败了，再去取联系人的昵称
				chatRoomMember, err := s.crmRespo.GetChatRoomMember(toWxID, wxid)
				if err != nil || chatRoomMember == nil {
					r, err := vars.RobotRuntime.GetContactDetail("", []string{wxid})
					if err != nil || len(r.ContactList) == 0 {
						continue
					}
					if r.ContactList[0].NickName.String == nil {
						continue
					}
					targetNickname = *r.ContactList[0].NickName.String
				} else {
					if chatRoomMember.Remark != "" {
						targetNickname = chatRoomMember.Remark
					} else {
						targetNickname = chatRoomMember.Nickname
					}
				}
			} else {
				// 私聊消息
				r, err := vars.RobotRuntime.GetContactDetail("", []string{wxid})
				if err != nil || len(r.ContactList) == 0 {
					continue
				}
				if r.ContactList[0].NickName.String == nil {
					continue
				}
				targetNickname = *r.ContactList[0].NickName.String
			}

			if targetNickname == "" {
				continue
			}
			if index > 0 {
				atContent += " "
			}
			atContent += fmt.Sprintf("@%s%s", targetNickname, "\u2005")
		}
	}
	content = atContent + content
	newMessages, err := vars.RobotRuntime.SendTextMessage(toWxID, content, at...)
	if err != nil {
		return err
	}

	// 通过机器人发送的消息，消息同步接口获取不到，所以这里需要手动入库
	if len(newMessages.List) > 0 {
		for _, message := range newMessages.List {
			if message.Ret == 0 {
				m := model.Message{
					MsgId:              message.NewMsgId,
					ClientMsgId:        message.ClientMsgid,
					Type:               model.MsgTypeText,
					Content:            content,
					DisplayFullContent: "",
					MessageSource:      "",
					FromWxID:           toWxID,
					ToWxID:             vars.RobotRuntime.WxID,
					SenderWxID:         vars.RobotRuntime.WxID,
					IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
					CreatedAt:          message.Createtime,
					UpdatedAt:          time.Now().Unix(),
				}
				if m.IsChatRoom && len(at) > 0 {
					m.ReplyWxID = at[0]
				}
				err = s.msgRespo.Create(&m)
				if err != nil {
					log.Printf("入库消息失败: %v", err)
				}
				// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
				NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)
			}
		}
	}

	return nil
}

// MsgSendGroupMassMsgText 文本消息群发接口
func (s *MessageService) MsgSendGroupMassMsgText(toWxID []string, content string) error {
	_, err := vars.RobotRuntime.MsgSendGroupMassMsgText(robot.MsgSendGroupMassMsgTextRequest{
		ToWxid:  toWxID,
		Content: content,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *MessageService) MsgUploadImg(toWxID string, image io.Reader) (*model.Message, error) {
	imageBytes, err := io.ReadAll(image)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}
	message, err := vars.RobotRuntime.MsgUploadImg(toWxID, imageBytes)
	if err != nil {
		return nil, err
	}

	m := model.Message{
		MsgId:              message.Newmsgid,
		ClientMsgId:        message.Msgid,
		Type:               model.MsgTypeImage,
		Content:            "", // 获取不到图片的 xml 内容
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return &m, nil
}

func (s *MessageService) MsgSendVideo(toWxID string, video io.Reader, videoExt string) error {
	videoBytes, err := io.ReadAll(video)
	if err != nil {
		return fmt.Errorf("读取文件内容失败: %w", err)
	}
	_, err = vars.RobotRuntime.MsgSendVideo(toWxID, videoBytes, videoExt)
	if err != nil {
		return err
	}

	msgid := time.Now().UnixNano()
	m := model.Message{
		MsgId:              msgid,
		ClientMsgId:        msgid,
		Type:               model.MsgTypeVideo,
		Content:            "", // 获取不到视频的 xml 内容
		DisplayFullContent: "",
		MessageSource:      "",
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          time.Now().Unix(),
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) MsgSendVoice(toWxID string, voice io.Reader, voiceExt string) error {
	videoBytes, err := io.ReadAll(voice)
	if err != nil {
		return fmt.Errorf("读取文件内容失败: %w", err)
	}
	message, err := vars.RobotRuntime.MsgSendVoice(toWxID, videoBytes, voiceExt)
	if err != nil {
		return err
	}

	clientMsgId, _ := strconv.ParseInt(message.ClientMsgId, 10, 64)
	m := model.Message{
		MsgId:              message.NewMsgId,
		ClientMsgId:        clientMsgId,
		Type:               model.MsgTypeVoice,
		Content:            "", // 获取不到音频的 xml 内容
		DisplayFullContent: "",
		MessageSource:      "",
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) SendLongTextMessage(toWxID string, longText string) error {
	currentRobot, err := s.robotAdminRespo.GetByWeChatID(vars.RobotRuntime.WxID)
	if err != nil {
		return err
	}
	if currentRobot == nil || currentRobot.Nickname == nil {
		return fmt.Errorf("未找到机器人信息")
	}

	dataID := uuid.New().String()
	fiveMinuteAgo := time.Now().Add(-5 * time.Minute)

	recordInfo := robot.RecordInfo{
		Info:       fmt.Sprintf("%s: %s", *currentRobot.Nickname, longText),
		IsChatRoom: 1,
		Desc:       fmt.Sprintf("%s: %s", *currentRobot.Nickname, longText),
		FromScene:  3,
		DataList: robot.DataList{
			Count: 1,
			Items: []robot.DataItem{
				{
					DataType:         1,
					DataID:           strings.ReplaceAll(dataID, "-", ""),
					SrcMsgLocalID:    rand.Intn(90000) + 10000,
					SourceTime:       fiveMinuteAgo.Format("2006-1-2 15:04"),
					FromNewMsgID:     time.Now().UnixNano() / 100,
					SrcMsgCreateTime: fiveMinuteAgo.Unix(),
					DataDesc:         longText,
					DataItemSource: &robot.DataItemSource{
						HashUsername: fmt.Sprintf("%x", sha256.Sum256([]byte(vars.RobotRuntime.WxID))),
					},
					SourceName:    *currentRobot.Nickname,
					SourceHeadURL: *currentRobot.Avatar,
				},
			},
		},
	}

	recordInfoBytes, err := xml.MarshalIndent(recordInfo, "", "  ")
	if err != nil {
		return err
	}

	newMsg := robot.ChatHistoryMessage{
		AppMsg: robot.ChatHistoryAppMsg{
			AppID:  "",
			SDKVer: "0",
			Title:  "群聊的聊天记录",
			Type:   19,
			URL:    "https://support.weixin.qq.com/cgi-bin/mmsupport-bin/readtemplate?t=page/favorite_record__w_unsupport",
			Des:    fmt.Sprintf("%s: %s", *currentRobot.Nickname, longText),
			RecordItem: robot.ChatHistoryRecordItem{XML: fmt.Sprintf(`<![CDATA[
%s
]]>`, string(recordInfoBytes))},
		},
	}
	message, err := vars.RobotRuntime.SendChatHistoryMessage(toWxID, newMsg)
	if err != nil {
		return err
	}

	m := model.Message{
		MsgId:              message.NewMsgId,
		ClientMsgId:        message.MsgId,
		Type:               model.MsgTypeApp,
		Content:            message.Content,
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) SendMusicMessage(toWxID string, songTitle string) error {
	var resp robot.MusicSearchResponse
	_, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("msg", songTitle).
		SetQueryParam("type", "json").
		SetQueryParam("n", "1").
		SetQueryParam("br", "7").
		SetResult(&resp).
		Get(vars.MusicSearchApi)
	if err != nil {
		return fmt.Errorf("获取歌曲信息失败: %w", err)
	}
	result := resp.Data
	if result.Title == nil {
		return fmt.Errorf("没有搜索到歌曲 %s", songTitle)
	}

	songInfo := robot.SongInfo{}
	songInfo.FromUsername = vars.RobotRuntime.WxID
	songInfo.AppID = "wx8dd6ecd81906fd84"
	songInfo.Title = *result.Title
	songInfo.Singer = result.Singer
	songInfo.Url = result.Link
	songInfo.MusicUrl = result.MusicURL
	if result.Cover != nil {
		songInfo.CoverUrl = *result.Cover
	}
	if result.Lrc != nil {
		songInfo.Lyric = *result.Lrc
	}

	message, err := vars.RobotRuntime.SendMusicMessage(toWxID, songInfo)
	if err != nil {
		return err
	}

	m := model.Message{
		MsgId:              message.NewMsgId,
		ClientMsgId:        message.MsgId,
		Type:               model.MsgTypeApp,
		Content:            message.Content,
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) SendFileMessage(ctx context.Context, req dto.SendFileMessageRequest, file io.Reader, fileHeader *multipart.FileHeader) error {
	message, err := vars.RobotRuntime.MsgSendFile(robot.SendFileMessageRequest{
		ToWxid:          req.ToWxid,
		ClientAppDataId: req.ClientAppDataId,
		Filename:        req.Filename,
		FileMD5:         req.FileHash,
		TotalLen:        req.FileSize,
		StartPos:        req.ChunkIndex * vars.UploadFileChunkSize,
		TotalChunks:     req.TotalChunks,
	}, file, fileHeader)
	if err != nil {
		return err
	}
	// 文件还没上传完
	if message == nil {
		return nil
	}

	clientMsgId, _ := strconv.ParseInt(message.ClientMsgId, 10, 64)
	m := model.Message{
		MsgId:              message.NewMsgId,
		ClientMsgId:        clientMsgId,
		Type:               model.MsgTypeApp,
		AppMsgType:         model.AppMsgTypeAttach,
		Content:            message.Content,
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           req.ToWxid,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(req.ToWxid, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) SendEmoji(toWxID string, md5 string, totalLen int32) error {
	message, err := vars.RobotRuntime.SendEmoji(robot.SendEmojiRequest{
		ToWxid:   toWxID,
		Md5:      md5,
		TotalLen: totalLen,
	})
	if err != nil {
		return err
	}

	for _, emojiItem := range message.EmojiItem {
		if emojiItem.Ret != 0 {
			continue
		}
		m := model.Message{
			MsgId:              emojiItem.NewMsgId,
			ClientMsgId:        emojiItem.MsgId,
			Type:               model.MsgTypeEmoticon,
			Content:            "",
			DisplayFullContent: "",
			MessageSource:      "",
			FromWxID:           toWxID,
			ToWxID:             vars.RobotRuntime.WxID,
			SenderWxID:         vars.RobotRuntime.WxID,
			IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
			CreatedAt:          time.Now().Unix(),
			UpdatedAt:          time.Now().Unix(),
		}
		err = s.msgRespo.Create(&m)
		if err != nil {
			log.Println("入库消息失败: ", err)
		}
		// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
		NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)
	}

	return nil
}

func (s *MessageService) ShareLink(toWxID string, shareLinkInfo robot.ShareLinkMessage) error {
	message, xmlStr, err := vars.RobotRuntime.ShareLink(toWxID, shareLinkInfo)
	if err != nil {
		return err
	}
	m := model.Message{
		MsgId:              message.NewMsgId,
		ClientMsgId:        message.MsgId,
		Type:               model.MsgTypeApp,
		Content:            xmlStr,
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)
	return nil
}

func (s *MessageService) SendCDNFile(toWxID string, content string) error {
	message, err := vars.RobotRuntime.SendCDNFile(robot.SendCDNAttachmentRequest{
		ToWxid:  toWxID,
		Content: content,
	})
	if err != nil {
		return err
	}

	m := model.Message{
		MsgId:              message.NewMsgId,
		ClientMsgId:        message.MsgId,
		Type:               model.MsgTypeApp,
		Content:            "",
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) SendCDNImg(toWxID string, content string) error {
	message, err := vars.RobotRuntime.SendCDNImg(robot.SendCDNAttachmentRequest{
		ToWxid:  toWxID,
		Content: content,
	})
	if err != nil {
		return err
	}

	m := model.Message{
		MsgId:              message.Newmsgid,
		ClientMsgId:        message.Msgid,
		Type:               model.MsgTypeImage,
		Content:            "",
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) SendCDNVideo(toWxID string, content string) error {
	message, err := vars.RobotRuntime.SendCDNVideo(robot.SendCDNAttachmentRequest{
		ToWxid:  toWxID,
		Content: content,
	})
	if err != nil {
		return err
	}

	m := model.Message{
		MsgId:              message.NewMsgId,
		ClientMsgId:        message.MsgId,
		Type:               model.MsgTypeVideo,
		Content:            "",
		DisplayFullContent: "",
		MessageSource:      message.MsgSource,
		FromWxID:           toWxID,
		ToWxID:             vars.RobotRuntime.WxID,
		SenderWxID:         vars.RobotRuntime.WxID,
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          time.Now().Unix(),
		UpdatedAt:          time.Now().Unix(),
	}
	err = s.msgRespo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) ProcessAIMessageContext(messages []*model.Message) []openai.ChatCompletionMessage {
	var aiMessages []openai.ChatCompletionMessage
	re := regexp.MustCompile(vars.TrimAtRegexp)
	for _, msg := range messages {
		aiMessage := openai.ChatCompletionMessage{}
		if msg.SenderWxID == vars.RobotRuntime.WxID {
			aiMessage.Role = openai.ChatMessageRoleAssistant
		} else {
			aiMessage.Role = openai.ChatMessageRoleUser
		}
		if msg.Type == model.MsgTypeText {
			aiMessage.Content = re.ReplaceAllString(msg.Content, "")
		}
		if msg.Type == model.MsgTypeImage {
			aiMessage.MultiContent = []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: msg.AttachmentUrl,
					},
				},
			}
		}
		if msg.Type == model.MsgTypeApp && msg.AppMsgType == model.AppMsgTypequote {
			var xmlMessage robot.XmlMessage
			err := vars.RobotRuntime.XmlDecoder(msg.Content, &xmlMessage)
			if err != nil {
				continue
			}
			referUser := xmlMessage.AppMsg.ReferMsg.ChatUsr
			// 如果引用的消息不是自己发的，也不是机器人发的，将消息内容添加到上下文
			if referUser != msg.SenderWxID && referUser != vars.RobotRuntime.WxID {
				// 引用的是第三人的文本消息，将引用的消息内容添加到上下文
				if xmlMessage.AppMsg.ReferMsg.Type == int(model.MsgTypeText) {
					aiMessage.MultiContent = []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: re.ReplaceAllString(xmlMessage.AppMsg.ReferMsg.Content, ""),
						},
						{
							Type: openai.ChatMessagePartTypeText,
							Text: re.ReplaceAllString(xmlMessage.AppMsg.Title, ""),
						},
					}
				}
				if xmlMessage.AppMsg.ReferMsg.Type == int(model.MsgTypeImage) {
					referMsgIDStr := xmlMessage.AppMsg.ReferMsg.SvrID
					// 字符串转int64
					referMsgID, err := strconv.ParseInt(referMsgIDStr, 10, 64)
					if err != nil {
						continue
					}
					refreMsg, err := s.msgRespo.GetByID(referMsgID)
					if err != nil {
						continue
					}
					if refreMsg == nil {
						continue
					}
					aiMessage.MultiContent = []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: refreMsg.AttachmentUrl,
							},
						},
						{
							Type: openai.ChatMessagePartTypeText,
							Text: re.ReplaceAllString(xmlMessage.AppMsg.Title, ""),
						},
					}
				}
				if xmlMessage.AppMsg.ReferMsg.Type == int(model.AppMsgTypequote) {
					referMsgIDStr := xmlMessage.AppMsg.ReferMsg.SvrID
					// 字符串转int64
					referMsgID, err := strconv.ParseInt(referMsgIDStr, 10, 64)
					if err != nil {
						continue
					}
					refreMsg, err := s.msgRespo.GetByID(referMsgID)
					if err != nil {
						continue
					}
					if refreMsg == nil {
						continue
					}
					var subXmlMessage robot.XmlMessage
					err = vars.RobotRuntime.XmlDecoder(refreMsg.Content, &subXmlMessage)
					if err != nil {
						continue
					}
					aiMessage.MultiContent = []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: re.ReplaceAllString(subXmlMessage.AppMsg.Title, ""),
						},
						{
							Type: openai.ChatMessagePartTypeText,
							Text: re.ReplaceAllString(xmlMessage.AppMsg.Title, ""),
						},
					}
				}
			} else {
				aiMessage.Content = re.ReplaceAllString(xmlMessage.AppMsg.Title, "")
			}
		}
		aiMessages = append(aiMessages, aiMessage)
	}
	return aiMessages
}

func (s *MessageService) SetMessageIsInContext(message *model.Message) error {
	return s.msgRespo.SetMessageIsInContext(message)
}

func (s *MessageService) GetFriendAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error) {
	messages, err := s.msgRespo.GetFriendAIMessageContext(message)
	if err != nil {
		return nil, err
	}
	if !slices.ContainsFunc(messages, func(m *model.Message) bool {
		return m.ID == message.ID
	}) {
		messages = append(messages, message)
	}
	return s.ProcessAIMessageContext(messages), nil
}

func (s *MessageService) ResetFriendAIMessageContext(message *model.Message) error {
	return s.msgRespo.ResetFriendAIMessageContext(message)
}

func (s *MessageService) GetChatRoomAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error) {
	messages, err := s.msgRespo.GetChatRoomAIMessageContext(message)
	if err != nil {
		return nil, err
	}
	if !slices.ContainsFunc(messages, func(m *model.Message) bool {
		return m.ID == message.ID
	}) {
		messages = append(messages, message)
	}
	return s.ProcessAIMessageContext(messages), nil
}

func (s *MessageService) UpdateMessage(message *model.Message) error {
	return s.msgRespo.Update(message)
}

func (s *MessageService) ResetChatRoomAIMessageContext(message *model.Message) error {
	return s.msgRespo.ResetChatRoomAIMessageContext(message)
}

func (s *MessageService) GetAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error) {
	if message.IsChatRoom {
		return s.GetChatRoomAIMessageContext(message)
	}
	return s.GetFriendAIMessageContext(message)
}

func (s *MessageService) GetYesterdayChatRommRank(chatRoomID string) ([]*dto.ChatRoomRank, error) {
	return s.msgRespo.GetYesterdayChatRommRank(vars.RobotRuntime.WxID, chatRoomID)
}

func (s *MessageService) GetLastWeekChatRommRank(chatRoomID string) ([]*dto.ChatRoomRank, error) {
	return s.msgRespo.GetLastWeekChatRommRank(vars.RobotRuntime.WxID, chatRoomID)
}

func (s *MessageService) GetLastMonthChatRommRank(chatRoomID string) ([]*dto.ChatRoomRank, error) {
	return s.msgRespo.GetLastMonthChatRommRank(vars.RobotRuntime.WxID, chatRoomID)
}

func (s *MessageService) ChatRoomAIDisabled(chatRoomID string) error {
	chatRoomSettingsSvc := NewChatRoomSettingsService(s.ctx)
	chatRoomSettings, err := chatRoomSettingsSvc.GetChatRoomSettings(chatRoomID)
	if err != nil {
		return err
	}
	if chatRoomSettings == nil || chatRoomSettings.ChatAIEnabled == nil || !*chatRoomSettings.ChatAIEnabled {
		return nil
	}
	disabled := false
	chatRoomSettings.ChatAIEnabled = &disabled
	err = chatRoomSettingsSvc.SaveChatRoomSettings(chatRoomSettings)
	if err != nil {
		return err
	}
	return nil
}
