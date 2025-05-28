package service

import (
	"context"
	"log"
	"slices"
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
	respo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	return respo.GetByChatRoomID(req, pager)
}

func (s *ChatRoomService) GetChatRoomSummary(chatRoomID string) (dto.ChatRoomSummary, error) {
	summary := dto.ChatRoomSummary{}
	//
	return summary, nil
}
