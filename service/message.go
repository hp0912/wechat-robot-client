package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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
	defer func() {
		if err := recover(); err != nil {
			log.Printf("消息入库出错了: %v", err)
		}
	}()
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
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
			var msgXml robot.SystemMessage
			err := vars.RobotRuntime.XmlDecoder(m.Content, &msgXml)
			if err == nil {
				oldMsg := respo.GetByMsgID(msgXml.RevokeMsg.NewMsgID)
				if oldMsg != nil {
					oldMsg.IsRecalled = true
					respo.Update(oldMsg)
					continue
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
		respo.Create(&m)
		// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
		NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)
	}
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
	message := respo.GetByID(req.MessageID)
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
	// 如果消息内容超过2000字，直接截断
	if len(req.Content) > 2000 {
		req.Content = req.Content[:2000]
	}

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
				respo.Create(&m)
				// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
				NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)
			}
		}
	}

	return nil
}

func (s *MessageService) MsgUploadImg(toWxID string, image multipart.File) error {
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
	respo.Create(&m)
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) MsgSendVideo(toWxID string, video multipart.File, videoExt string) error {
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
	respo.Create(&m)
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}

func (s *MessageService) MsgSendVoice(toWxID string, voice multipart.File, voiceExt string) error {
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
	respo.Create(&m)
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
		return errors.New(fmt.Sprintf("没有搜索到歌曲 %s", songTitle))
	}
	songInfo := robot.SongInfo{}
	songInfo.FromUsername = vars.RobotRuntime.WxID
	songInfo.AppID = "wx8dd6ecd81906fd84"
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
	respo.Create(&m)
	// 插入一条联系人记录，获取联系人列表接口获取不到未保存到通讯录的群聊
	NewContactService(s.ctx).InsertOrUpdateContactActiveTime(m.FromWxID)

	return nil
}
