package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
)

// 防抖逻辑：在 5 秒窗口内同一群聊只同步一次
const syncChatRoomMemberDebounceInterval = 5 * time.Second

var (
	syncChatRoomMemberMu     sync.Mutex
	syncChatRoomMemberTimers = make(map[string]*time.Timer)
)

type ChatRoomService struct {
	ctx                context.Context
	msgRepo            *repository.Message
	ctRepo             *repository.Contact
	gsRepo             *repository.GlobalSettings
	crsRepo            *repository.ChatRoomSettings
	crmRepo            *repository.ChatRoomMember
	sysmsgRepo         *repository.SystemMessage
	systemSettingsRepo *repository.SystemSettings
}

func NewChatRoomService(ctx context.Context) *ChatRoomService {
	return &ChatRoomService{
		ctx:                ctx,
		msgRepo:            repository.NewMessageRepo(ctx, vars.DB),
		ctRepo:             repository.NewContactRepo(ctx, vars.DB),
		gsRepo:             repository.NewGlobalSettingsRepo(ctx, vars.DB),
		crsRepo:            repository.NewChatRoomSettingsRepo(ctx, vars.DB),
		crmRepo:            repository.NewChatRoomMemberRepo(ctx, vars.DB),
		sysmsgRepo:         repository.NewSystemMessageRepo(ctx, vars.DB),
		systemSettingsRepo: repository.NewSystemSettingsRepo(ctx, vars.DB),
	}
}

func (s *ChatRoomService) DebounceSyncChatRoomMember(chatRoomID string) {
	syncChatRoomMemberMu.Lock()
	defer syncChatRoomMemberMu.Unlock()

	if timer, ok := syncChatRoomMemberTimers[chatRoomID]; ok {
		timer.Stop()
	}

	var newTimer *time.Timer
	newTimer = time.AfterFunc(syncChatRoomMemberDebounceInterval, func() {
		syncChatRoomMemberMu.Lock()
		currentTimer, ok := syncChatRoomMemberTimers[chatRoomID]
		if !ok || currentTimer != newTimer {
			syncChatRoomMemberMu.Unlock()
			return
		}
		delete(syncChatRoomMemberTimers, chatRoomID)
		syncChatRoomMemberMu.Unlock()

		// 真正执行同步逻辑（放在锁外避免长时间持锁）
		s.SyncChatRoomMember(chatRoomID)
	})

	syncChatRoomMemberTimers[chatRoomID] = newTimer
}

func (s *ChatRoomService) SyncChatRoomMember(chatRoomID string) {
	var chatRoomMembers []robot.ChatRoomMember
	var err error
	chatRoomMembers, err = vars.RobotRuntime.GetChatRoomMemberDetail(chatRoomID)
	if err != nil {
		log.Printf("获取群[%s]成员失败: %v", chatRoomID, err)
		return
	}
	// 遍历获取到的群成员列表，如果数据库存在，则更新，数据库不存在则新增
	if len(chatRoomMembers) > 0 {
		now := time.Now().Unix()
		// 获取当前成员的微信ID列表，用于后续比对
		currentMemberIDs := make([]string, 0, len(chatRoomMembers))
		for _, member := range chatRoomMembers {
			currentMemberIDs = append(currentMemberIDs, member.UserName)
		}
		for _, member := range chatRoomMembers {
			// 检查成员是否已存在
			existMember, err := s.crmRepo.GetChatRoomMember(chatRoomID, member.UserName)
			if err != nil {
				log.Printf("查询群[%s]成员[%s]失败: %v", chatRoomID, member.UserName, err)
				continue
			}
			if existMember != nil {
				// 更新现有成员
				isLeaved := false
				updateMember := map[string]any{
					"nickname":          member.NickName,
					"avatar":            member.SmallHeadImgUrl,
					"inviter_wechat_id": member.InviterUserName,
					"is_leaved":         &isLeaved, // 确保标记为未离开
					"leaved_at":         nil,       // 清除离开时间
				}
				if member.DisplayName != nil && *member.DisplayName != "" {
					updateMember["remark"] = *member.DisplayName
				}
				// 已经离开，重新加入群聊的人
				if existMember.IsLeaved != nil && *existMember.IsLeaved {
					updateMember["joined_at"] = now
				}
				// 更新数据库中已有的记录
				err = s.crmRepo.UpdateByID(existMember.ID, updateMember)
				if err != nil {
					log.Printf("更新群[%s]成员[%s]失败: %v", chatRoomID, member.UserName, err)
					continue
				}
			} else {
				// 创建新成员
				isLeaved := false
				newMember := model.ChatRoomMember{
					ChatRoomID:      chatRoomID,
					WechatID:        member.UserName,
					Nickname:        member.NickName,
					Avatar:          member.SmallHeadImgUrl,
					InviterWechatID: member.InviterUserName,
					IsLeaved:        &isLeaved,
					JoinedAt:        now,
					LastActiveAt:    now,
				}
				if member.DisplayName != nil && *member.DisplayName != "" {
					newMember.Remark = *member.DisplayName
				}
				err = s.crmRepo.Create(&newMember)
				if err != nil {
					log.Printf("创建群[%s]成员[%s]失败: %v", chatRoomID, member.UserName, err)
					continue
				}
			}
		}
		// 查询数据库中该群的所有成员
		dbMembers, err := s.crmRepo.GetChatRoomMembers(chatRoomID)
		if err != nil {
			log.Printf("获取群[%s]成员失败: %v", chatRoomID, err)
			return
		}
		var leavedMembers []string
		// 标记已离开的成员
		for _, dbMember := range dbMembers {
			if dbMember.IsLeaved != nil && !*dbMember.IsLeaved && !slices.Contains(currentMemberIDs, dbMember.WechatID) {
				// 数据库有记录但当前群成员列表中不存在，标记为已离开
				leaveTime := now
				isLeaved := true
				updateMember := model.ChatRoomMember{
					ID:       dbMember.ID,
					IsLeaved: &isLeaved,
					LeavedAt: &leaveTime,
				}
				if dbMember.Remark != "" {
					leavedMembers = append(leavedMembers, dbMember.Remark)
				} else if dbMember.Nickname != "" {
					leavedMembers = append(leavedMembers, dbMember.Nickname)
				}
				err = s.crmRepo.Update(&updateMember)
				if err != nil {
					log.Printf("标记群[%s]成员[%s]为已离开失败: %v", chatRoomID, dbMember.WechatID, err)
					continue
				}
			}
		}
		if len(leavedMembers) > 0 {
			leaveChatRoomConfig := NewChatRoomSettingsService(s.ctx).GetLeaveChatRoomConfig(chatRoomID)
			if leaveChatRoomConfig == nil || leaveChatRoomConfig.LeaveChatRoomAlertEnabled == nil || !*leaveChatRoomConfig.LeaveChatRoomAlertEnabled {
				return
			}
			if strings.TrimSpace(leaveChatRoomConfig.LeaveChatRoomAlertText) == "" {
				return
			}
			if len(leavedMembers) <= 10 {
				NewMessageService(s.ctx).SendTextMessage(
					chatRoomID,
					strings.Replace(leaveChatRoomConfig.LeaveChatRoomAlertText, "{placeholder}", strings.Join(leavedMembers, "、"), 1),
				)
			} else {
				NewMessageService(s.ctx).SendTextMessage(
					chatRoomID,
					strings.Replace(leaveChatRoomConfig.LeaveChatRoomAlertText, "{placeholder}", fmt.Sprintf("%s等%d位", leavedMembers[0], len(leavedMembers)), 1),
				)
			}
		}
	}
}

func (s *ChatRoomService) CreateChatRoom(contactIDs []string) error {
	chatRoom, err := vars.RobotRuntime.CreateChatRoom(contactIDs)
	if err != nil {
		return err
	}
	if chatRoom.ChatRoomName == nil || chatRoom.ChatRoomName.String == nil || *chatRoom.ChatRoomName.String == "" {
		return errors.New("创建群聊失败，返回了空群聊名称")
	}
	NewContactService(s.ctx).DebounceSyncContact(*chatRoom.ChatRoomName.String)
	return nil
}

func (s *ChatRoomService) AutoInviteChatRoomMember(chatRoomName string, contactIDs []string) error {
	systemSettings, err := s.systemSettingsRepo.GetSystemSettings()
	if err != nil {
		return err
	}
	if systemSettings == nil || systemSettings.AutoChatroomInvite == nil || !*systemSettings.AutoChatroomInvite {
		return fmt.Errorf("自动邀请入群功能未开启")
	}
	chatRoom, err := s.ctRepo.GetByChatRoomNickname(chatRoomName)
	if err != nil {
		return err
	}
	if chatRoom == nil {
		return fmt.Errorf("群聊不存在: %s", chatRoomName)
	}
	return s.InviteChatRoomMember(chatRoom.WechatID, contactIDs)
}

func (s *ChatRoomService) InviteChatRoomMember(chatRoomID string, contactIDs []string) error {
	if len(contactIDs) == 0 {
		return fmt.Errorf("无效的联系人ID")
	}
	// 重新获取一下群成员
	contacts, err := s.ctRepo.GetFriendsByWechatIDs(contactIDs)
	if err != nil {
		return err
	}
	var currentContactIDs []string
	for _, contact := range contacts {
		currentContactIDs = append(currentContactIDs, contact.WechatID)
	}
	// 当前群成员
	currentMembers, err := s.GetNotLeftMembers(dto.ChatRoomMemberRequest{ChatRoomID: chatRoomID})
	if err != nil {
		return err
	}
	// 过滤掉已经在群里的成员
	var newMembers []string
	var currentMemberSet = make(map[string]bool, len(currentMembers))
	for _, member := range currentMembers {
		currentMemberSet[member.WechatID] = true
	}
	for _, contactID := range currentContactIDs {
		if !currentMemberSet[contactID] {
			if strings.HasSuffix(contactID, "@chatroom") {
				return fmt.Errorf("参数异常: %s", contactID)
			}
			newMembers = append(newMembers, contactID)
		}
	}
	if len(newMembers) == 0 {
		return fmt.Errorf("所有联系人都已在群中")
	}
	if len(currentMembers)+len(newMembers) > 500 {
		return fmt.Errorf("群成员数量超过上限")
	}
	if len(currentMembers) <= 40 {
		return vars.RobotRuntime.GroupAddChatRoomMember(chatRoomID, newMembers)
	}
	return vars.RobotRuntime.GroupInviteChatRoomMember(chatRoomID, newMembers)
}

func (s *ChatRoomService) GroupConsentToJoin(systemMessageID int64) error {
	systemMessage, err := s.sysmsgRepo.GetByID(systemMessageID)
	if err != nil {
		return err
	}
	if systemMessage == nil {
		return fmt.Errorf("系统消息不存在: %d", systemMessageID)
	}
	if systemMessage.Type != model.SystemMessageTypeJoinChatRoom {
		return fmt.Errorf("系统消息类型错误: %d", systemMessage.Type)
	}
	var xmlMessage robot.XmlMessage
	err = vars.RobotRuntime.XmlDecoder(systemMessage.Content, &xmlMessage)
	if err != nil {
		return fmt.Errorf("解析邀请入群请求消息失败: %v", err)
	}
	chatRoomID, err := vars.RobotRuntime.GroupConsentToJoin(xmlMessage.AppMsg.URL)
	if err != nil {
		return err
	}
	if chatRoomID == "" {
		return fmt.Errorf("加入群聊失败: %s，接口返回了空", systemMessage.FromWxid)
	}
	err = s.sysmsgRepo.Update(&model.SystemMessage{
		ID:     systemMessage.ID,
		IsRead: true,
		Status: 1,
	})
	if err != nil {
		// 忽略错误
		log.Println("更新系统消息状态失败:", err)
	}

	NewContactService(s.ctx).DebounceSyncContact(chatRoomID)

	return nil
}

func (s *ChatRoomService) GroupSetChatRoomName(chatRoomID, content string) error {
	err := vars.RobotRuntime.GroupSetChatRoomName(chatRoomID, content)
	if err != nil {
		return err
	}
	return s.ctRepo.UpdateNicknameByContactID(chatRoomID, content)
}

func (s *ChatRoomService) GroupSetChatRoomRemarks(chatRoomID, content string) error {
	err := vars.RobotRuntime.GroupSetChatRoomRemarks(chatRoomID, content)
	if err != nil {
		return err
	}
	return s.ctRepo.UpdateRemarkByContactID(chatRoomID, content)
}

func (s *ChatRoomService) GroupSetChatRoomAnnouncement(chatRoomID, content string) error {
	return vars.RobotRuntime.GroupSetChatRoomAnnouncement(chatRoomID, content)
}

func (s *ChatRoomService) GroupDelChatRoomMember(chatRoomID string, memberIDs []string) error {
	err := vars.RobotRuntime.GroupDelChatRoomMember(chatRoomID, memberIDs)
	if err != nil {
		return err
	}
	return s.crmRepo.DeleteChatRoomMembers(memberIDs)
}

func (s *ChatRoomService) GroupQuit(chatRoomID string) error {
	err := vars.RobotRuntime.GroupQuit(chatRoomID)
	if err != nil {
		return err
	}
	return s.ctRepo.DeleteByContactID(chatRoomID)
}

func (s *ChatRoomService) UpdateChatRoomMembersOnNewMemberJoinIn(chatRoomID string, memberWeChatIDs []string) ([]*model.ChatRoomMember, error) {
	now := time.Now().Unix()
	// 将ids拆分成二十个一个的数组之后再获取详情
	var newMembers = make([]robot.Contact, 0)
	chunker := slices.Chunk(memberWeChatIDs, 20)
	processChunk := func(chunk []string) bool {
		// 获取昵称等详细信息
		var r robot.GetContactResponse
		r, err := vars.RobotRuntime.GetContactDetail("", chunk)
		if err != nil {
			// 处理错误
			log.Printf("获取联系人详情失败: %v", err)
			return true
		}
		newMembers = append(newMembers, r.ContactList...)
		return true
	}
	chunker(processChunk)
	for _, member := range newMembers {
		if member.UserName.String == nil {
			continue
		}
		if strings.TrimSpace(*member.UserName.String) == "" {
			continue
		}
		memberUserName := *member.UserName.String
		// 检查成员是否已存在
		existMember, err := s.crmRepo.GetChatRoomMember(chatRoomID, memberUserName)
		if err != nil {
			log.Printf("查询群[%s]成员[%s]失败: %v", chatRoomID, memberUserName, err)
			continue
		}
		if existMember != nil {
			// 更新现有成员
			isLeaved := false
			updateMember := map[string]any{
				"wechat_id": memberUserName,
				"alias":     member.Alias,
				"nickname":  *member.NickName.String,
				"avatar":    member.SmallHeadImgUrl,
				"is_leaved": &isLeaved, // 确保标记为未离开
				"leaved_at": nil,       // 清除离开时间
			}
			// 更新数据库中已有的记录
			err = s.crmRepo.UpdateByID(existMember.ID, updateMember)
			if err != nil {
				log.Printf("更新群[%s]成员[%s]失败: %v", chatRoomID, memberUserName, err)
				continue
			}
		} else {
			// 创建新成员
			isLeaved := false
			newMember := model.ChatRoomMember{
				ChatRoomID:      chatRoomID,
				WechatID:        memberUserName,
				Alias:           member.Alias,
				Nickname:        *member.NickName.String,
				Avatar:          member.SmallHeadImgUrl,
				InviterWechatID: "",
				IsLeaved:        &isLeaved,
				JoinedAt:        now,
				LastActiveAt:    now,
			}
			err = s.crmRepo.Create(&newMember)
			if err != nil {
				log.Printf("创建群[%s]成员[%s]失败: %v", chatRoomID, memberUserName, err)
				continue
			}
		}
	}
	return s.crmRepo.GetChatRoomMemberByWeChatIDs(chatRoomID, memberWeChatIDs)
}

func (s *ChatRoomService) GetChatRoomMembers(req dto.ChatRoomMemberRequest, pager appx.Pager) ([]*model.ChatRoomMember, int64, error) {
	return s.crmRepo.GetByChatRoomID(req, pager)
}

func (s *ChatRoomService) GetNotLeftMembers(req dto.ChatRoomMemberRequest) ([]*model.ChatRoomMember, error) {
	return s.crmRepo.GetNotLeftMemberByChatRoomID(req)
}

func (s *ChatRoomService) GetChatRoomMemberCount(chatRoomID string) (int64, error) {
	return s.crmRepo.GetChatRoomMemberCount(chatRoomID)
}

func (s *ChatRoomService) GetChatRoomSummary(chatRoomID string) (dto.ChatRoomSummary, error) {
	summary := dto.ChatRoomSummary{ChatRoomID: chatRoomID}
	memberCount, err := s.crmRepo.GetChatRoomMemberCount(chatRoomID)
	if err != nil {
		return summary, err
	}
	joinCount, err := s.crmRepo.GetYesterdayJoinCount(chatRoomID)
	if err != nil {
		return summary, err
	}
	leaveCount, err := s.crmRepo.GetYesterdayLeaveCount(chatRoomID)
	if err != nil {
		return summary, err
	}
	summary.MemberTotalCount = int(memberCount)
	summary.MemberJoinCount = int(joinCount)
	summary.MemberLeaveCount = int(leaveCount)

	chatInfo, err := s.msgRepo.GetYesterdayChatInfo(chatRoomID)
	if err != nil {
		return summary, err
	}
	summary.MemberChatCount = len(chatInfo)
	summary.MessageCount = 0
	for _, info := range chatInfo {
		summary.MessageCount += info.MessageCount
	}

	return summary, nil
}

func (s *ChatRoomService) ChatRoomAISummaryByChatRoomID(globalSettings *model.GlobalSettings, setting *model.ChatRoomSettings, startTime, endTime int64) error {
	msgService := NewMessageService(context.Background())
	chatRoomName := setting.ChatRoomID
	chatRoom, err := s.ctRepo.GetByWechatID(setting.ChatRoomID)
	if err != nil {
		return err
	}

	if chatRoom != nil && chatRoom.Nickname != nil && *chatRoom.Nickname != "" {
		chatRoomName = *chatRoom.Nickname
	}

	messages, err := s.msgRepo.GetMessagesByTimeRange(vars.RobotRuntime.WxID, setting.ChatRoomID, startTime, endTime)
	if err != nil {
		return err
	}
	if len(messages) < 100 {
		err := msgService.SendTextMessage(setting.ChatRoomID, "聊天不够活跃啊~~~")
		if err != nil {
			log.Printf("发送消息失败: %v", err)
		}
		return nil
	}

	// 组装对话记录为字符串
	var content []string
	for _, message := range messages {
		// 将时间戳秒格式化为时间YYYY-MM-DD HH:MM:SS 字符串
		timeStr := time.Unix(message.CreatedAt, 0).Format("2006-01-02 15:04:05")
		content = append(content, fmt.Sprintf(`[%s] {"%s": "%s"}--end--`, timeStr, message.Nickname, strings.ReplaceAll(message.Message, "\n", "。。")))
	}

	maxCompletionTokens := 2000
	prompt := `你是一个中文的群聊总结的助手，你可以为一个微信的群聊记录，提取并总结每个时间段大家在重点讨论的话题内容。

每一行代表一个人的发言，每一行的的格式为： {"[time] {nickname}": "{content}"}--end--

请帮我将给出的群聊内容总结成一个今日的群聊报告，包含不多于10个的话题的总结（如果还有更多话题，可以在后面简单补充）。每个话题包含以下内容：
- 话题名(50字以内，带序号1️⃣2️⃣3️⃣，同时附带热度，以🔥数量表示）
- 参与者(不超过5个人，将重复的人名去重)
- 时间段(从几点到几点)
- 过程(50到200字左右）
- 评价(50字以下)
- 分割线： ------------

另外有以下要求：
1. 每个话题结束使用 ------------ 分割
2. 使用中文冒号
3. 无需大标题
4. 开始给出本群讨论风格的整体评价，例如活跃、太水、太黄、太暴力、话题不集中、无聊诸如此类
5. 群友分享的链接资源要提取出来，并附加在总结的最后`
	prompt = fmt.Sprintf("%s\n6. 总结结果不得超过%d字符", prompt, maxCompletionTokens)
	msg := fmt.Sprintf("群名称: %s\n聊天记录如下:\n%s", chatRoomName, strings.Join(content, "\n"))
	// AI总结
	aiMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: msg,
		},
	}

	// 默认使用AI回复
	aiApiKey := globalSettings.ChatAPIKey
	if *setting.ChatAPIKey != "" {
		aiApiKey = *setting.ChatAPIKey
	}
	aiConfig := openai.DefaultConfig(aiApiKey)
	aiApiBaseURL := strings.TrimRight(globalSettings.ChatBaseURL, "/")
	if setting.ChatBaseURL != nil && *setting.ChatBaseURL != "" {
		aiApiBaseURL = strings.TrimRight(*setting.ChatBaseURL, "/")
	}
	aiConfig.BaseURL = utils.NormalizeAIBaseURL(aiApiBaseURL)
	model := globalSettings.ChatRoomSummaryModel
	if setting.ChatRoomSummaryModel != nil && *setting.ChatRoomSummaryModel != "" {
		model = *setting.ChatRoomSummaryModel
	}
	ai := openai.NewClientWithConfig(aiConfig)
	var resp openai.ChatCompletionResponse
	resp, err = ai.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:               model,
			Messages:            aiMessages,
			Stream:              false,
			MaxCompletionTokens: maxCompletionTokens,
		},
	)
	if err != nil {
		log.Printf("群聊记录总结失败: %v", err.Error())
		msgService.SendTextMessage(setting.ChatRoomID, "#昨日消息总结\n\n群聊消息总结失败，错误信息: "+err.Error())
		return err
	}
	// 返回消息为空
	if resp.Choices[0].Message.Content == "" {
		msgService.SendTextMessage(setting.ChatRoomID, "#昨日消息总结\n\n群聊消息总结失败，AI返回结果为空")
		return nil
	}
	replyMsg := fmt.Sprintf("#消息总结\n让我们一起来看看群友们都聊了什么有趣的话题吧~\n\n本次总结由**%s**加持\n\n%s", model, resp.Choices[0].Message.Content)
	msgService.SendLongTextMessage(setting.ChatRoomID, replyMsg)
	return nil
}

func (s *ChatRoomService) ChatRoomAISummary() error {
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()

	globalSettings, err := s.gsRepo.GetGlobalSettings()
	if err != nil {
		return err
	}

	if globalSettings == nil || globalSettings.ChatAIEnabled == nil || !*globalSettings.ChatAIEnabled || globalSettings.ChatAPIKey == "" || globalSettings.ChatBaseURL == "" {
		log.Printf("全局设置未开启AI，跳过群聊总结")
		return nil
	}

	settings, err := NewChatRoomSettingsService(s.ctx).GetAllEnableAISummary()
	if err != nil {
		return err
	}

	for _, setting := range settings {
		if setting == nil || setting.ChatRoomSummaryEnabled == nil || !*setting.ChatRoomSummaryEnabled {
			log.Printf("群聊 %s 的 AI 总结模型未配置，跳过处理\n", setting.ChatRoomID)
			continue
		}
		err := s.ChatRoomAISummaryByChatRoomID(globalSettings, setting, yesterdayStartTimestamp, todayStartTimestamp)
		if err != nil {
			log.Printf("处理群聊 %s 的 AI 总结失败: %v\n", setting.ChatRoomID, err)
			continue
		}
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingDaily() error {
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	settings, err := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
	if err != nil {
		return err
	}

	msgService := NewMessageService(context.Background())

	for _, setting := range settings {
		notifyMsgs := []string{"#昨日水群排行榜"}

		ranks, err := msgService.GetYesterdayChatRommRank(setting.ChatRoomID)
		if err != nil {
			log.Printf("获取群聊 %s 的排行榜失败: %v\n", setting.ChatRoomID, err)
			continue
		}
		if len(ranks) == 0 {
			log.Printf("群聊 %s 昨天没有聊天记录，跳过排行榜更新\n", setting.ChatRoomID)
			continue
		}
		chatRoomMemberCount, err := s.GetChatRoomMemberCount(setting.ChatRoomID)
		if err != nil {
			log.Printf("获取群聊 %s 成员数量失败: %v\n", setting.ChatRoomID, err)
		}
		// 计算活跃度
		showActivity := err == nil && chatRoomMemberCount > 0
		activity := "0.00"
		if chatRoomMemberCount > 0 {
			activity = fmt.Sprintf("%.2f", (float64(len(ranks))/float64(chatRoomMemberCount))*100)
		}
		// 计算消息总数、中位数、前十位消息总数
		var msgCount, medianCount, topTenCount int64
		for idx, v := range ranks {
			msgCount += v.Count
			if idx == (len(ranks)/2)-1 {
				medianCount = v.Count
			}
			if len(ranks) > 10 && idx < 10 {
				topTenCount += v.Count
			}
		}
		// 计算活跃用户人均消息条数
		avgMsgCount := int(float64(msgCount) / float64(len(ranks)))
		// 组装消息总数推送信息
		notifyMsgs = append(notifyMsgs, " ")
		notifyMsgs = append(notifyMsgs, fmt.Sprintf("🗣️ 昨日本群 %d 位朋友共产生 %d 条发言", len(ranks), msgCount))
		if showActivity {
			m := fmt.Sprintf("🎭 活跃度: %s%%，人均消息条数: %d，中位数: %d", activity, avgMsgCount, medianCount)
			// 计算前十占比
			if topTenCount > 0 {
				m += fmt.Sprintf("，前十名占比: %.2f%%", float64(topTenCount)/float64(msgCount)*100)
			}
			notifyMsgs = append(notifyMsgs, m)
		}
		notifyMsgs = append(notifyMsgs, "\n🏵 活跃用户排行榜 🏵")
		notifyMsgs = append(notifyMsgs, " ")
		for i, r := range ranks {
			// 只取前十条
			if i >= 10 {
				break
			}
			log.Printf("账号: %s[%s] -> %d", r.ChatRoomMemberNickname, r.SenderWxID, r.Count)
			badge := "🏆"
			switch i {
			case 0:
				badge = "🥇"
			case 1:
				badge = "🥈"
			case 2:
				badge = "🥉"
			}
			notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条", badge, r.ChatRoomMemberNickname, r.Count))
		}
		notifyMsgs = append(notifyMsgs, " \n🎉感谢以上群友昨日对群活跃做出的卓越贡献，也请未上榜的群友多多反思。")
		msgService.SendTextMessage(setting.ChatRoomID, strings.Join(notifyMsgs, "\n"))
		// 发送词云图片
		wordCloudCacheDir := filepath.Join(string(filepath.Separator), "app", "word_cloud_cache")
		dateStr := yesterdayStart.Format("2006-01-02")
		filename := fmt.Sprintf("%s_%s.png", setting.ChatRoomID, dateStr)
		filePath := filepath.Join(wordCloudCacheDir, filename)
		imageFile, err := os.Open(filePath)
		if err != nil {
			log.Printf("群聊 %s 打开词云图片文件失败: %v", setting.ChatRoomID, err)
			continue
		}
		defer imageFile.Close()
		_, err = msgService.MsgUploadImg(setting.ChatRoomID, imageFile)
		if err != nil {
			log.Printf("群聊 %s 词云图片发送失败: %v", setting.ChatRoomID, err)
			continue
		}
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingWeekly() error {
	settings, err := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
	if err != nil {
		return err
	}

	msgService := NewMessageService(context.Background())

	for _, setting := range settings {
		notifyMsgs := []string{"#上周水群排行榜"}

		ranks, err := msgService.GetLastWeekChatRommRank(setting.ChatRoomID)
		if err != nil {
			log.Printf("获取群聊 %s 的排行榜失败: %v\n", setting.ChatRoomID, err)
			continue
		}
		if len(ranks) == 0 {
			log.Printf("群聊 %s 上周没有聊天记录，跳过排行榜更新\n", setting.ChatRoomID)
			continue
		}
		chatRoomMemberCount, err := s.GetChatRoomMemberCount(setting.ChatRoomID)
		if err != nil {
			log.Printf("获取群聊 %s 成员数量失败: %v\n", setting.ChatRoomID, err)
		}
		// 计算活跃度
		showActivity := err == nil && chatRoomMemberCount > 0
		activity := "0.00"
		if chatRoomMemberCount > 0 {
			activity = fmt.Sprintf("%.2f", (float64(len(ranks))/float64(chatRoomMemberCount))*100)
		}

		// 计算消息总数、中位数、前十位消息总数
		var msgCount, medianCount, topTenCount int64
		for idx, v := range ranks {
			msgCount += v.Count
			if idx == (len(ranks)/2)-1 {
				medianCount = v.Count
			}
			if len(ranks) > 10 && idx < 10 {
				topTenCount += v.Count
			}
		}
		// 计算活跃用户人均消息条数
		avgMsgCount := int(float64(msgCount) / float64(len(ranks)))
		// 组装消息总数推送信息
		notifyMsgs = append(notifyMsgs, " ")
		notifyMsgs = append(notifyMsgs, fmt.Sprintf("🗣️ 上周本群 %d 位朋友共产生 %d 条发言", len(ranks), msgCount))
		if showActivity {
			m := fmt.Sprintf("🎭 活跃度: %s%%，人均消息条数: %d，中位数: %d", activity, avgMsgCount, medianCount)
			// 计算前十占比
			if topTenCount > 0 {
				m += fmt.Sprintf("，前十名占比: %.2f%%", float64(topTenCount)/float64(msgCount)*100)
			}
			notifyMsgs = append(notifyMsgs, m)
		}
		notifyMsgs = append(notifyMsgs, "\n🏵 活跃用户排行榜 🏵")
		notifyMsgs = append(notifyMsgs, " ")
		for i, r := range ranks {
			// 只取前十条
			if i >= 10 {
				break
			}
			log.Printf("账号: %s[%s] -> %d", r.ChatRoomMemberNickname, r.SenderWxID, r.Count)
			badge := "🏆"
			switch i {
			case 0:
				badge = "🥇"
			case 1:
				badge = "🥈"
			case 2:
				badge = "🥉"
			}
			notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条", badge, r.ChatRoomMemberNickname, r.Count))
		}
		notifyMsgs = append(notifyMsgs, " \n🎉感谢以上群友上周对群活跃做出的卓越贡献，也请未上榜的群友多多反思。")
		msgService.SendTextMessage(setting.ChatRoomID, strings.Join(notifyMsgs, "\n"))
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingMonthly() error {
	monthStr := time.Now().Local().AddDate(0, 0, -1).Format("2006年01月")

	settings, err := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
	if err != nil {
		return err
	}

	msgService := NewMessageService(context.Background())

	for _, setting := range settings {
		notifyMsgs := []string{fmt.Sprintf("#%s水群排行榜", monthStr)}

		ranks, err := msgService.GetLastMonthChatRommRank(setting.ChatRoomID)
		if err != nil {
			log.Printf("获取群聊 %s 的排行榜失败: %v\n", setting.ChatRoomID, err)
			continue
		}
		if len(ranks) == 0 {
			log.Printf("群聊 %s 上个月没有聊天记录，跳过排行榜更新\n", setting.ChatRoomID)
			continue
		}
		chatRoomMemberCount, err := s.GetChatRoomMemberCount(setting.ChatRoomID)
		if err != nil {
			log.Printf("获取群聊 %s 成员数量失败: %v\n", setting.ChatRoomID, err)
		}
		// 计算活跃度
		showActivity := err == nil && chatRoomMemberCount > 0
		activity := "0.00"
		if chatRoomMemberCount > 0 {
			activity = fmt.Sprintf("%.2f", (float64(len(ranks))/float64(chatRoomMemberCount))*100)
		}
		// 计算消息总数、中位数、前十位消息总数
		var msgCount, medianCount, topTenCount int64
		for idx, v := range ranks {
			msgCount += v.Count
			if idx == (len(ranks)/2)-1 {
				medianCount = v.Count
			}
			if len(ranks) > 10 && idx < 10 {
				topTenCount += v.Count
			}
		}
		// 计算活跃用户人均消息条数
		avgMsgCount := int(float64(msgCount) / float64(len(ranks)))
		// 组装消息总数推送信息
		notifyMsgs = append(notifyMsgs, " ")
		notifyMsgs = append(notifyMsgs, fmt.Sprintf("🗣️ %s本群 %d 位朋友共产生 %d 条发言", monthStr, len(ranks), msgCount))
		if showActivity {
			m := fmt.Sprintf("🎭 活跃度: %s%%，人均消息条数: %d，中位数: %d", activity, avgMsgCount, medianCount)
			// 计算前十占比
			if topTenCount > 0 {
				m += fmt.Sprintf("，前十名占比: %.2f%%", float64(topTenCount)/float64(msgCount)*100)
			}
			notifyMsgs = append(notifyMsgs, m)
		}
		notifyMsgs = append(notifyMsgs, "\n🏵 活跃用户排行榜 🏵")
		notifyMsgs = append(notifyMsgs, " ")
		for i, r := range ranks {
			// 只取前十条
			if i >= 10 {
				break
			}
			log.Printf("账号: %s[%s] -> %d", r.ChatRoomMemberNickname, r.SenderWxID, r.Count)
			badge := "🏆"
			switch i {
			case 0:
				badge = "🥇"
			case 1:
				badge = "🥈"
			case 2:
				badge = "🥉"
			}
			notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条", badge, r.ChatRoomMemberNickname, r.Count))
		}
		notifyMsgs = append(notifyMsgs, fmt.Sprintf(" \n🎉感谢以上群友%s对群活跃做出的卓越贡献，也请未上榜的群友多多反思。", monthStr))
		msgService.SendTextMessage(setting.ChatRoomID, strings.Join(notifyMsgs, "\n"))
	}
	return nil
}
