package service

import (
	"context"
	"fmt"
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
	defer func() {
		if err := recover(); err != nil {
			log.Printf("获取群[%s]成员失败: %v", chatRoomID, err)
		}
	}()
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
			exists := memberRepo.ExistsByWhere(map[string]any{
				"chat_room_id": chatRoomID,
				"wechat_id":    member.UserName,
			})
			if exists {
				// 更新现有成员
				updateMember := map[string]any{
					"nickname":  member.NickName,
					"avatar":    member.SmallHeadImgUrl,
					"is_leaved": false, // 确保标记为未离开
					"leaved_at": nil,   // 清除离开时间
				}
				// 更新数据库中已有的记录
				memberRepo.UpdateColumnsByWhere(&updateMember, map[string]any{
					"chat_room_id": chatRoomID,
					"wechat_id":    member.UserName,
				})
			} else {
				// 创建新成员
				newMember := model.ChatRoomMember{
					Owner:           vars.RobotRuntime.WxID,
					ChatRoomID:      chatRoomID,
					WechatID:        member.UserName,
					Nickname:        member.NickName,
					Avatar:          member.SmallHeadImgUrl,
					InviterWechatID: member.InviterUserName,
					IsLeaved:        false,
					JoinedAt:        now,
					LastActiveAt:    now,
				}
				memberRepo.Create(&newMember)
			}
		}
		// 查询数据库中该群的所有成员
		dbMembers := memberRepo.ListByWhere(nil, map[string]any{
			"chat_room_id": chatRoomID,
			"is_leaved":    false, // 只处理未离开的成员
		})
		// 标记已离开的成员
		for _, dbMember := range dbMembers {
			if !slices.Contains(currentMemberIDs, dbMember.WechatID) {
				// 数据库有记录但当前群成员列表中不存在，标记为已离开
				leaveTime := now
				updateMember := model.ChatRoomMember{
					IsLeaved: true,
					LeavedAt: &leaveTime,
				}
				memberRepo.UpdateColumnsByWhere(&updateMember, map[string]any{
					"chat_room_id": chatRoomID,
					"wechat_id":    dbMember.WechatID,
				})
			}
		}
	}
}

func (s *ChatRoomService) GetChatRoomMembers(req dto.ChatRoomMemberRequest, pager appx.Pager) ([]*model.ChatRoomMember, int64, error) {
	req.Owner = vars.RobotRuntime.WxID
	respo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	return respo.GetByChatRoomID(req, pager)
}

func (s *ChatRoomService) GetChatRoomMemberCount(chatRoomID string) (int64, error) {
	respo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	return respo.GetChatRoomMemberCount(vars.RobotRuntime.WxID, chatRoomID)
}

func (s *ChatRoomService) GetChatRoomSummary(chatRoomID string) (dto.ChatRoomSummary, error) {
	summary := dto.ChatRoomSummary{ChatRoomID: chatRoomID}

	owner := vars.RobotRuntime.WxID
	crmRespo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	memberCount, err := crmRespo.GetChatRoomMemberCount(owner, chatRoomID)
	if err != nil {
		return summary, err
	}
	joinCount, err := crmRespo.GetYesterdayJoinCount(owner, chatRoomID)
	if err != nil {
		return summary, err
	}
	leaveCount, err := crmRespo.GetYesterdayLeaveCount(owner, chatRoomID)
	if err != nil {
		return summary, err
	}
	summary.MemberTotalCount = int(memberCount)
	summary.MemberJoinCount = int(joinCount)
	summary.MemberLeaveCount = int(leaveCount)

	messageRepo := repository.NewMessageRepo(s.ctx, vars.DB)
	chatInfo, err := messageRepo.GetYesterdayChatInfo(owner, chatRoomID)
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

func (s *ChatRoomService) ChatRoomAISummaryByChatRoomID(chatRoomID string, startTime, endTime int64) error {
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
	settings := NewChatRoomSettingsService(s.ctx).GetAllEnableAISummary()
	for _, setting := range settings {
		err := s.ChatRoomAISummaryByChatRoomID(setting.ChatRoomID, yesterdayStartTimestamp, todayStartTimestamp)
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

	settings := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
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
		log.Printf("排行榜: \n%s", strings.Join(notifyMsgs, "\n"))
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: strings.Join(notifyMsgs, "\n"),
		})
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingWeekly() error {
	notifyMsgs := []string{"#上周水群排行榜"}

	settings := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
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
		log.Printf("排行榜: \n%s", strings.Join(notifyMsgs, "\n"))
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: strings.Join(notifyMsgs, "\n"),
		})
	}
	return nil
}

func (s *ChatRoomService) ChatRoomRankingMonthly() error {
	monthStr := time.Now().Local().AddDate(0, 0, -1).Format("2006年01月")
	notifyMsgs := []string{fmt.Sprintf("#%s水群排行榜", monthStr)}

	settings := NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
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
		log.Printf("排行榜: \n%s", strings.Join(notifyMsgs, "\n"))
		msgService.SendTextMessage(dto.SendTextMessageRequest{
			SendMessageCommonRequest: dto.SendMessageCommonRequest{
				ToWxid: setting.ChatRoomID,
			},
			Content: strings.Join(notifyMsgs, "\n"),
		})
	}
	return nil
}
