package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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

type MessageService struct {
	ctx context.Context
}

func NewMessageService(ctx context.Context) *MessageService {
	return &MessageService{
		ctx: ctx,
	}
}

// ProcessTextMessage 处理文本消息
func (s *MessageService) ProcessTextMessage(message *model.Message) {

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

// ProcessAppMessage 处理应用消息
func (s *MessageService) ProcessAppMessage(message *model.Message) {

}

// ProcessShareCardMessage 处理分享名片消息
func (s *MessageService) ProcessShareCardMessage(message *model.Message) {

}

// ProcessFriendVerifyMessage 处理好友添加请求通知消息
func (s *MessageService) ProcessFriendVerifyMessage(message *model.Message) {

}

// ProcessRecalledMessage 处理撤回消息
func (s *MessageService) ProcessRecalledMessage(message *model.Message, msgXml robot.SystemMessage) {
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
	oldMsg, err := respo.GetByMsgID(msgXml.RevokeMsg.NewMsgID)
	if err != nil {
		log.Printf("获取撤回的消息失败: %v", err)
		return
	}
	if oldMsg != nil {
		oldMsg.IsRecalled = true
		err = respo.Update(oldMsg)
		if err != nil {
			log.Printf("标记撤回消息失败: %v", err)
		}
		return
	}
}

// ProcessPatMessage 处理拍一拍消息
func (s *MessageService) ProcessPatMessage(message *model.Message) {

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
		s.ProcessPatMessage(message)
		return
	}
}

// ProcessLocationMessage 处理位置消息
func (s *MessageService) ProcessLocationMessage(message *model.Message) {

}

// ProcessPromptMessage 处理提示消息
func (s *MessageService) ProcessPromptMessage(message *model.Message) {

}

func (s *MessageService) ProcessMessage(syncResp robot.SyncMessage) {
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		if m.Type == model.MsgTypeSystem && m.SenderWxID == "weixin" {
			continue
		}
		if m.Type == model.MsgTypeApp {
			subTypeStr := vars.RobotRuntime.XmlFastDecoder(m.Content, "type")
			if subTypeStr != "" {
				subType, err := strconv.Atoi(subTypeStr)
				if err == nil {
					m.AppMsgType = model.AppMessageType(subType)
					if m.AppMsgType == model.AppMsgTypeAttachUploading {
						// 消息不入库
						continue
					}
				}
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
		err := respo.Create(&m)
		if err != nil {
			log.Printf("入库消息失败: %v", err)
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

func (s *MessageService) SyncMessageStart() {
	ctx := context.Background()
	vars.RobotRuntime.SyncMessageContext, vars.RobotRuntime.SyncMessageCancel = context.WithCancel(ctx)
	for {
		select {
		case <-vars.RobotRuntime.SyncMessageContext.Done():
			return
		case <-time.After(1 * time.Second):
			s.SyncMessage()
		}
	}
}

func (s *MessageService) MessageRevoke(req dto.MessageCommonRequest) error {
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
	message, err := respo.GetByID(req.MessageID)
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

func (s *MessageService) SendTextMessage(req dto.SendTextMessageRequest) error {
	newMessages, newMessageContent, err := vars.RobotRuntime.SendTextMessage(req.ToWxid, req.Content, req.At...)
	if err != nil {
		return err
	}

	// 通过机器人发送的消息，消息同步接口获取不到，所以这里需要手动入库
	if len(newMessages.List) > 0 {
		respo := repository.NewMessageRepo(s.ctx, vars.DB)
		for _, message := range newMessages.List {
			if message.Ret == 0 {
				m := model.Message{
					MsgId:              message.NewMsgId,
					ClientMsgId:        message.ClientMsgid,
					Type:               model.MsgTypeText,
					Content:            newMessageContent,
					DisplayFullContent: "",
					MessageSource:      "",
					FromWxID:           req.ToWxid,
					ToWxID:             vars.RobotRuntime.WxID,
					SenderWxID:         vars.RobotRuntime.WxID,
					IsGroup:            strings.HasSuffix(req.ToWxid, "@chatroom"),
					CreatedAt:          message.Createtime,
					UpdatedAt:          time.Now().Unix(),
				}
				err = respo.Create(&m)
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

func (s *MessageService) MsgUploadImg(toWxID string, image io.Reader) error {
	imageBytes, err := io.ReadAll(image)
	if err != nil {
		return fmt.Errorf("读取文件内容失败: %w", err)
	}
	message, err := vars.RobotRuntime.MsgUploadImg(toWxID, imageBytes)
	if err != nil {
		return err
	}

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
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

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          time.Now().Unix(),
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
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

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) SendMusicMessage(toWxID string, songTitle string) error {
	var result robot.MusicSearchResponse
	_, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("gm", songTitle).
		SetQueryParam("type", "json").
		SetQueryParam("n", "1").
		SetResult(&result).
		Get(vars.MusicSearchApi)
	if err != nil {
		return fmt.Errorf("获取歌曲信息失败: %w", err)
	}
	if result.Title == nil {
		return fmt.Errorf("没有搜索到歌曲 %s", songTitle)
	}

	songInfo := robot.SongInfo{}
	songInfo.FromUsername = vars.RobotRuntime.WxID
	songInfo.AppID = "wx79f2c4418704b4f8"
	songInfo.Title = *result.Title
	songInfo.Singer = result.Singer
	songInfo.Url = result.Link
	songInfo.MusicUrl = result.MusicUrl
	if result.Cover != nil {
		songInfo.CoverUrl = *result.Cover
	}
	if result.Lrc != nil {
		songInfo.Lyric = *result.Lrc
	}

	message, xmlStr, err := vars.RobotRuntime.SendMusicMessage(toWxID, songInfo)
	if err != nil {
		return err
	}

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
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

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
			IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
			CreatedAt:          time.Now().Unix(),
			UpdatedAt:          time.Now().Unix(),
		}
		err = respo.Create(&m)
		if err != nil {
			log.Println("入库消息失败: ", err)
		}
		// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
		NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)
	}

	return nil
}

func (s *MessageService) ShareLink(toWxID string, linkXml string) error {
	message, err := vars.RobotRuntime.ShareLink(robot.ShareLinkRequest{
		ToWxid: toWxID,
		Type:   5,
		Xml:    linkXml,
	})
	if err != nil {
		return err
	}

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
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

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
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

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          message.CreateTime,
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
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

	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
		IsGroup:            strings.HasSuffix(toWxID, "@chatroom"),
		CreatedAt:          time.Now().Unix(),
		UpdatedAt:          time.Now().Unix(),
	}
	err = respo.Create(&m)
	if err != nil {
		log.Println("入库消息失败: ", err)
	}
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) GetYesterdayChatRommRank(chatRoomID string) ([]*dto.ChatRoomRank, error) {
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
	return respo.GetYesterdayChatRommRank(chatRoomID)
}

func (s *MessageService) GetLastWeekChatRommRank(chatRoomID string) ([]*dto.ChatRoomRank, error) {
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
	return respo.GetLastWeekChatRommRank(chatRoomID)
}

func (s *MessageService) GetLastMonthChatRommRank(chatRoomID string) ([]*dto.ChatRoomRank, error) {
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
	return respo.GetLastMonthChatRommRank(chatRoomID)
}
