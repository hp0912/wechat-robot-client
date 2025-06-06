package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
	"github.com/sashabaranov/go-openai"
)

type MessageService struct {
	ctx context.Context
}

func NewMessageService(ctx context.Context) *MessageService {
	return &MessageService{
		ctx: ctx,
	}
}

func (s *MessageService) ProcessAI(message *model.Message) {
	aiService := NewAIService(s.ctx, message)
	chatIntention := aiService.ChatIntention(message)
	switch chatIntention {
	case ChatIntentionChat:
		aiContext, err := s.GetAIMessageContext(message)
		if err != nil {
			s.SendTextMessage(message.FromWxID, err.Error())
			return
		}
		aiReply, err := aiService.Chat(aiContext)
		if err != nil {
			s.SendTextMessage(message.FromWxID, err.Error())
			return
		}
		if message.IsChatRoom {
			s.SendTextMessage(message.FromWxID, aiReply, message.SenderWxID)
		} else {
			s.SendTextMessage(message.FromWxID, aiReply)
		}
	case ChatIntentionSing:
		s.SendTextMessage(message.FromWxID, "唱歌功能正在开发中，敬请期待！")
	case ChatIntentionSongRequest:
		title := aiService.GetSongRequestTitle(message)
		if title == "" {
			s.SendTextMessage(message.FromWxID, "抱歉，我无法识别您想要点的歌曲。")
			return
		}
		err := s.SendMusicMessage(message.FromWxID, title)
		if err != nil {
			s.SendTextMessage(message.FromWxID, err.Error())
		}
	case ChatIntentionDrawAPicture:
		s.SendTextMessage(message.FromWxID, "绘画功能正在开发中，敬请期待！")
	case ChatIntentionEditPictures:
		s.SendTextMessage(message.FromWxID, "修图功能正在开发中，敬请期待！")
	default:
		s.SendTextMessage(message.FromWxID, "更多功能正在开发中，敬请期待！")
	}
}

// ProcessTextMessage 处理文本消息
func (s *MessageService) ProcessTextMessage(message *model.Message) {
	aiService := NewAIService(s.ctx, message)
	if message.IsChatRoom {
		if aiService.IsAISessionStart(message) {
			s.SendTextMessage(message.FromWxID, "AI会话已开始，请输入您的问题。10分钟不说话会话将自动结束，您也可以输入 #退出AI会话 来结束会话。", message.SenderWxID)
			return
		}
		isInSession, err := aiService.IsInAISession(message)
		if err != nil {
			log.Printf("检查AI会话失败: %v", err)
			return
		}
		if isInSession {
			fmt.Println(isInSession)
			return
		}
		if aiService.IsAISessionEnd(message) {
			s.SendTextMessage(message.FromWxID, "AI会话已结束，您可以输入 #进入AI会话 来重新开始。", message.SenderWxID)
			return
		}
		if aiService.IsAITrigger(message) {
			s.ProcessAI(message)
			return
		}
	} else if aiService.IsAIEnabled() {
		s.ProcessAI(message)
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
}

// ProcessAppMessage 处理应用消息
func (s *MessageService) ProcessAppMessage(message *model.Message) {
	if message.AppMsgType == model.AppMsgTypequote {
		s.ProcessReferMessage(message)
		return
	}
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

func (s *MessageService) ProcessNewChatRoomMemberMessage(message *model.Message, msgXml robot.SystemMessage) {
	welcomeConfig, err := NewChatRoomSettingsService(s.ctx).GetChatRoomWelcomeConfig(message.FromWxID)
	if err != nil {
		log.Printf("获取群聊欢迎配置失败: %v", err)
		return
	}
	if welcomeConfig.WelcomeEnabled != nil && !*welcomeConfig.WelcomeEnabled {
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
		err = s.MsgUploadImg(message.FromWxID, tempFile)
		if err != nil {
			log.Println("发送欢迎图片消息失败: ", err)
			return
		}
	}
	if welcomeConfig.WelcomeType == model.WelcomeTypeURL {
		var newMemberWechatIds []string
		if len(msgXml.SysMsgTemplate.ContentTemplate.LinkList.Links) > 0 {
			newMembers := msgXml.SysMsgTemplate.ContentTemplate.LinkList.Links[0]
			if newMembers.MemberList != nil {
				for _, member := range newMembers.MemberList.Members {
					newMemberWechatIds = append(newMemberWechatIds, member.Username)
				}
			}
		}
		// 将ids拆分成二十个一个的数组之后再获取详情
		var newMembers = make([]robot.Contact, 0)
		chunker := slices.Chunk(newMemberWechatIds, 20)
		processChunk := func(chunk []string) bool {
			// 获取昵称等详细信息
			var c = make([]robot.Contact, 0)
			c, err = vars.RobotRuntime.GetContactDetail(chunk)
			if err != nil {
				// 处理错误
				log.Printf("获取联系人详情失败: %v", err)
				return false
			}
			newMembers = append(newMembers, c...)
			return true
		}
		chunker(processChunk)
		if len(newMembers) == 0 {
			return
		}
		shareLinkInfo := robot.ShareLinkInfo{
			Desc:     welcomeConfig.WelcomeText,
			Url:      welcomeConfig.WelcomeURL,
			ThumbUrl: newMembers[0].SmallHeadImgUrl,
		}
		if len(newMembers) > 1 {
			shareLinkInfo.Title = fmt.Sprintf("欢迎%d位家人进群", len(newMembers))
		} else if newMembers[0].NickName.String != nil && *newMembers[0].NickName.String != "" {
			shareLinkInfo.Title = fmt.Sprintf("欢迎%s进群", *newMembers[0].NickName.String)
		} else {
			shareLinkInfo.Title = "欢迎新成员进群"
		}
		err := s.ShareLink(message.FromWxID, shareLinkInfo)
		if err != nil {
			log.Println("发送欢迎链接消息失败: ", err)
			return
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
		s.ProcessPatMessage(message)
		return
	}
	if msgXml.SysMsgTemplate.ContentTemplate.Type == "tmpl_type_profilewithrevoke" && strings.Contains(msgXml.SysMsgTemplate.ContentTemplate.Template, "加入了群聊") {
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
			m.IsChatRoom = true
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
			m.IsChatRoom = false
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

func (s *MessageService) SendTextMessage(toWxID, content string, at ...string) error {
	newMessages, newMessageContent, err := vars.RobotRuntime.SendTextMessage(toWxID, content, at...)
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
	var resp robot.MusicSearchResponse
	_, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("msg", songTitle).
		SetQueryParam("type", "json").
		SetQueryParam("n", "1").
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
	songInfo.AppID = "wx79f2c4418704b4f8"
	songInfo.Title = *result.Title
	songInfo.Singer = result.Singer
	songInfo.Url = result.Link
	songInfo.MusicUrl = result.Url
	if result.Cover != nil {
		songInfo.CoverUrl = *result.Cover
	}
	if result.Lyric != nil {
		songInfo.Lyric = *result.Lyric
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
			IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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

func (s *MessageService) ShareLink(toWxID string, shareLinkInfo robot.ShareLinkInfo) error {
	message, xmlStr, err := vars.RobotRuntime.ShareLink(toWxID, shareLinkInfo)
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
		IsChatRoom:         strings.HasSuffix(toWxID, "@chatroom"),
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
					respo := repository.NewMessageRepo(s.ctx, vars.DB)
					referMsgIDStr := xmlMessage.AppMsg.ReferMsg.SvrID
					// 字符串转int64
					referMsgID, err := strconv.ParseInt(referMsgIDStr, 10, 64)
					if err != nil {
						continue
					}
					refreMsg, err := respo.GetByID(referMsgID)
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
					respo := repository.NewMessageRepo(s.ctx, vars.DB)
					referMsgIDStr := xmlMessage.AppMsg.ReferMsg.SvrID
					// 字符串转int64
					referMsgID, err := strconv.ParseInt(referMsgIDStr, 10, 64)
					if err != nil {
						continue
					}
					refreMsg, err := respo.GetByID(referMsgID)
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

func (s *MessageService) GetFriendAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error) {
	messages, err := repository.NewMessageRepo(s.ctx, vars.DB).GetFriendAIMessageContext(message)
	if err != nil {
		return nil, err
	}
	return s.ProcessAIMessageContext(messages), nil
}

func (s *MessageService) GetChatRoomAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error) {
	messages, err := repository.NewMessageRepo(s.ctx, vars.DB).GetChatRoomAIMessageContext(message)
	if err != nil {
		return nil, err
	}
	return s.ProcessAIMessageContext(messages), nil
}

func (s *MessageService) GetAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error) {
	if message.IsChatRoom {
		return s.GetChatRoomAIMessageContext(message)
	}
	return s.GetFriendAIMessageContext(message)
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
