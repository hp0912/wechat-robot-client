package service

import (
	"errors"
	"log"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/gin-gonic/gin"
)

type AICallbackService struct {
	ctx        *gin.Context
	aiTaskRepo *repository.AITask
	msgRespo   *repository.Message
}

func NewAICallbackService(ctx *gin.Context) *AICallbackService {
	return &AICallbackService{
		ctx:        ctx,
		aiTaskRepo: repository.NewAITaskRepo(ctx, vars.DB),
		msgRespo:   repository.NewMessageRepo(ctx, vars.DB),
	}
}

func (s *AICallbackService) SendMessage(msgService *MessageService, message *model.Message, msg string) {
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
		if req.TaskStatus == 1 {
			aiTask.AITaskStatus = model.AITaskStatusCompleted
			s.SendMessage(msgService, message, "任务已完成，正在发送音频消息...")
		} else if req.TaskStatus == 2 {
			aiTask.AITaskStatus = model.AITaskStatusFailed
			s.SendMessage(msgService, message, req.Message)
		} else {
			aiTask.AITaskStatus = model.AITaskStatusProcessing
			s.SendMessage(msgService, message, "任务处理中，请耐心等待...")
		}
		return s.aiTaskRepo.Update(aiTask)
	}

	aiTask.AITaskStatus = model.AITaskStatusFailed
	s.SendMessage(msgService, message, req.Message)

	return s.aiTaskRepo.Update(aiTask)
}
