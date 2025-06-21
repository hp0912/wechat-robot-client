package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type AICallbackService struct {
	ctx        context.Context
	aiTaskRepo *repository.AITask
	msgRespo   *repository.Message
}

func NewAICallbackService(ctx context.Context) *AICallbackService {
	return &AICallbackService{
		ctx:        ctx,
		aiTaskRepo: repository.NewAITaskRepo(ctx, vars.DB),
		msgRespo:   repository.NewMessageRepo(ctx, vars.DB),
	}
}

func (s *AICallbackService) SendTextMessage(msgService *MessageService, message *model.Message, msg string) {
	if message.IsChatRoom {
		err := msgService.SendTextMessage(message.FromWxID, msg, message.SenderWxID)
		if err != nil {
			log.Println("发送消息失败: ", message.FromWxID, msg, err)
			return
		}
	} else {
		err := msgService.SendTextMessage(message.FromWxID, msg)
		if err != nil {
			log.Println("发送消息失败: ", message.FromWxID, msg, err)
			return
		}
	}
}

func (s *AICallbackService) DoubaoTTS(req dto.DoubaoTTSCallbackRequest) error {
	aiTask, err := s.aiTaskRepo.GetByAIProviderTaskID(req.TaskID)
	if err != nil {
		log.Println("获取任务失败: ", req.TaskID, err)
		return err
	}
	if aiTask == nil {
		log.Println("任务不存在: ", req.TaskID)
		return errors.New("任务不存在")
	}

	message, err := s.msgRespo.GetByID(aiTask.MessageID)
	if err != nil {
		log.Println("获取消息失败: ", aiTask.MessageID, err)
		return err
	}
	if message == nil {
		log.Println("消息不存在: ", aiTask.MessageID)
		return errors.New("消息不存在")
	}

	msgService := NewMessageService(s.ctx)
	aiTask.UpdatedAt = time.Now().Unix()
	if req.Code == 0 {
		switch req.TaskStatus {
		case 1:
			aiTask.AITaskStatus = model.AITaskStatusCompleted
			s.SendTextMessage(msgService, message, fmt.Sprintf("豆包长文本转语音任务完成，有效期24小时，请及时下载，音频地址: %s", req.AudioURL))
		case 2:
			aiTask.AITaskStatus = model.AITaskStatusFailed
			s.SendTextMessage(msgService, message, req.Message)
		default:
			aiTask.AITaskStatus = model.AITaskStatusProcessing
			s.SendTextMessage(msgService, message, "任务处理中，请耐心等待...")
		}
		return s.aiTaskRepo.Update(aiTask)
	}

	aiTask.AITaskStatus = model.AITaskStatusFailed
	s.SendTextMessage(msgService, message, req.Message)

	return s.aiTaskRepo.Update(aiTask)
}
