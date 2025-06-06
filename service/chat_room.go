package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
)

type ChatRoomService struct {
	ctx context.Context
}

func NewChatRoomService(ctx context.Context) *ChatRoomService {
	return &ChatRoomService{
		ctx: ctx,
	}
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
		memberRepo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
		now := time.Now().Unix()

		// 获取当前成员的微信ID列表，用于后续比对
		currentMemberIDs := make([]string, 0, len(chatRoomMembers))
		for _, member := range chatRoomMembers {
			currentMemberIDs = append(currentMemberIDs, member.UserName)
		}

		for _, member := range chatRoomMembers {
			// 检查成员是否已存在
			existMember, err := memberRepo.GetChatRoomMember(chatRoomID, member.UserName)
			if err != nil {
				log.Printf("查询群[%s]成员[%s]失败: %v", chatRoomID, member.UserName, err)
				continue
			}
			if existMember != nil {
				// 更新现有成员
				isLeaved := false
				updateMember := model.ChatRoomMember{
					ID:       existMember.ID,
					Nickname: member.NickName,
					Avatar:   member.SmallHeadImgUrl,
					IsLeaved: &isLeaved, // 确保标记为未离开
					LeavedAt: nil,       // 清除离开时间
				}
				// 更新数据库中已有的记录
				err = memberRepo.Update(&updateMember)
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
				err = memberRepo.Create(&newMember)
				if err != nil {
					log.Printf("创建群[%s]成员[%s]失败: %v", chatRoomID, member.UserName, err)
					continue
				}
			}
		}
		// 查询数据库中该群的所有成员
		dbMembers, err := memberRepo.GetChatRoomMembers(chatRoomID)
		if err != nil {
			log.Printf("获取群[%s]成员失败: %v", chatRoomID, err)
			return
		}
		// 标记已离开的成员
		for _, dbMember := range dbMembers {
			if !slices.Contains(currentMemberIDs, dbMember.WechatID) {
				// 数据库有记录但当前群成员列表中不存在，标记为已离开
				leaveTime := now
				isLeaved := true
				updateMember := model.ChatRoomMember{
					ID:       dbMember.ID,
					IsLeaved: &isLeaved,
					LeavedAt: &leaveTime,
				}
				err = memberRepo.Update(&updateMember)
				if err != nil {
					log.Printf("标记群[%s]成员[%s]为已离开失败: %v", chatRoomID, dbMember.WechatID, err)
					continue
				}
			}
		}
	}
}

func (s *ChatRoomService) GetChatRoomMembers(req dto.ChatRoomMemberRequest, pager appx.Pager) ([]*model.ChatRoomMember, int64, error) {
	respo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	return respo.GetByChatRoomID(req, pager)
}

func (s *ChatRoomService) GetChatRoomMemberCount(chatRoomID string) (int64, error) {
	respo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	return respo.GetChatRoomMemberCount(chatRoomID)
}

func (s *ChatRoomService) GetChatRoomSummary(chatRoomID string) (dto.ChatRoomSummary, error) {
	summary := dto.ChatRoomSummary{ChatRoomID: chatRoomID}

	crmRespo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	memberCount, err := crmRespo.GetChatRoomMemberCount(chatRoomID)
	if err != nil {
		return summary, err
	}
	joinCount, err := crmRespo.GetYesterdayJoinCount(chatRoomID)
	if err != nil {
		return summary, err
	}
	leaveCount, err := crmRespo.GetYesterdayLeaveCount(chatRoomID)
	if err != nil {
		return summary, err
	}
	summary.MemberTotalCount = int(memberCount)
	summary.MemberJoinCount = int(joinCount)
	summary.MemberLeaveCount = int(leaveCount)

	messageRepo := repository.NewMessageRepo(s.ctx, vars.DB)
	chatInfo, err := messageRepo.GetYesterdayChatInfo(chatRoomID)
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
	msgRespo := repository.NewMessageRepo(s.ctx, vars.DB)
	ctRespo := repository.NewContactRepo(s.ctx, vars.DB)

	chatRoomName := setting.ChatRoomID
	chatRoom, err := ctRespo.GetByWechatID(setting.ChatRoomID)
	if err != nil {
		return err
	}

	if chatRoom != nil && chatRoom.Nickname != nil && *chatRoom.Nickname != "" {
		chatRoomName = *chatRoom.Nickname
	}

	messages, err := msgRespo.GetMessagesByTimeRange(setting.ChatRoomID, startTime, endTime)
	if err != nil {
		return err
	}
	if len(messages) < 100 {
		err := msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: "聊天不够活跃啊~~~",
		})
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
`
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
	aiConfig.BaseURL = aiApiBaseURL
	if !strings.HasSuffix(aiConfig.BaseURL, "/v1") {
		aiConfig.BaseURL += "/v1"
	}
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
			MaxCompletionTokens: 2000,
		},
	)
	if err != nil {
		log.Printf("群聊记录总结失败: %v", err.Error())
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: "#昨日消息总结\n\n群聊消息总结失败，错误信息: " + err.Error(),
		})
		return err
	}
	// 返回消息为空
	if resp.Choices[0].Message.Content == "" {
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: "#昨日消息总结\n\n群聊消息总结失败，AI返回结果为空",
		})
		return nil
	}
	replyMsg := fmt.Sprintf("#消息总结\n让我们一起来看看群友们都聊了什么有趣的话题吧~\n\n%s", resp.Choices[0].Message.Content)
	msgService.SendTextMessage(dto.SendTextMessageRequest{
		SendMessageCommonRequest: dto.SendMessageCommonRequest{
			ToWxid: setting.ChatRoomID,
		},
		Content: replyMsg,
	})
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

	globalSettings, err := repository.NewGlobalSettingsRepo(s.ctx, vars.DB).GetGlobalSettings()
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
		// 休眠一秒，防止频繁发送
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingDaily() error {
	notifyMsgs := []string{"#昨日水群排行榜"}

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
			log.Printf("账号: %s[%s] -> %d", r.Nickname, r.WechatID, r.Count)
			badge := "🏆"
			switch i {
			case 0:
				badge = "🥇"
			case 1:
				badge = "🥈"
			case 2:
				badge = "🥉"
			}
			notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条", badge, r.Nickname, r.Count))
		}
		notifyMsgs = append(notifyMsgs, " \n🎉感谢以上群友昨日对群活跃做出的卓越贡献，也请未上榜的群友多多反思。")
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: strings.Join(notifyMsgs, "\n"),
		})
		// 休眠一秒，防止频繁发送
		time.Sleep(1 * time.Second)
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
		err = msgService.MsgUploadImg(setting.ChatRoomID, imageFile)
		if err != nil {
			log.Printf("群聊 %s 词云图片发送失败: %v", setting.ChatRoomID, err)
			continue
		}
		// 休眠一秒，防止频繁发送
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingWeekly() error {
	notifyMsgs := []string{"#上周水群排行榜"}

	settings, err := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
	if err != nil {
		return err
	}

	msgService := NewMessageService(context.Background())

	for _, setting := range settings {
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
			log.Printf("账号: %s[%s] -> %d", r.Nickname, r.WechatID, r.Count)
			badge := "🏆"
			switch i {
			case 0:
				badge = "🥇"
			case 1:
				badge = "🥈"
			case 2:
				badge = "🥉"
			}
			notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条", badge, r.Nickname, r.Count))
		}
		notifyMsgs = append(notifyMsgs, " \n🎉感谢以上群友上周对群活跃做出的卓越贡献，也请未上榜的群友多多反思。")
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: strings.Join(notifyMsgs, "\n"),
		})
		// 休眠一秒，防止频繁发送
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingMonthly() error {
	monthStr := time.Now().Local().AddDate(0, 0, -1).Format("2006年01月")
	notifyMsgs := []string{fmt.Sprintf("#%s水群排行榜", monthStr)}

	settings, err := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
	if err != nil {
		return err
	}

	msgService := NewMessageService(context.Background())

	for _, setting := range settings {
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
			log.Printf("账号: %s[%s] -> %d", r.Nickname, r.WechatID, r.Count)
			badge := "🏆"
			switch i {
			case 0:
				badge = "🥇"
			case 1:
				badge = "🥈"
			case 2:
				badge = "🥉"
			}
			notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条", badge, r.Nickname, r.Count))
		}
		notifyMsgs = append(notifyMsgs, fmt.Sprintf(" \n🎉感谢以上群友%s对群活跃做出的卓越贡献，也请未上榜的群友多多反思。", monthStr))
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: strings.Join(notifyMsgs, "\n"),
		})
		// 休眠一秒，防止频繁发送
		time.Sleep(1 * time.Second)
	}
	return nil
}
