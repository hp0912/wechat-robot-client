package common_cron

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/dto"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type ChatRoomRankingDailyCron struct {
	CronManager *CronManager
}

func NewChatRoomRankingDailyCron(cronManager *CronManager) vars.CommonCronInstance {
	return &ChatRoomRankingDailyCron{
		CronManager: cronManager,
	}
}

func (cron *ChatRoomRankingDailyCron) IsActive() bool {
	if cron.CronManager.globalSettings.ChatRoomRankingEnabled != nil && *cron.CronManager.globalSettings.ChatRoomRankingEnabled {
		return true
	}
	return false
}

func (cron *ChatRoomRankingDailyCron) Register() {
	if !cron.IsActive() {
		log.Println("每日群聊排行榜任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomRankingDailyCron, cron.CronManager.globalSettings.ChatRoomRankingDailyCron, func(params ...any) error {
		log.Println("开始执行每日群聊排行榜任务")

		notifyMsgs := []string{"#昨日水群排行榜"}

		settings := service.NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
		msgService := service.NewMessageService(context.Background())
		crService := service.NewChatRoomService(context.Background())

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
			chatRoomMemberCount, err := crService.GetChatRoomMemberCount(setting.ChatRoomID)
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
	})
	log.Println("每日群聊排行榜任务初始化成功")
}
