package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type AITaskService struct {
	ctx        context.Context
	aiTaskRepo *repository.AITask
}

func NewAITaskService(ctx context.Context) *AITaskService {
	return &AITaskService{
		ctx:        ctx,
		aiTaskRepo: repository.NewAITaskRepo(ctx, vars.DB),
	}
}

func (s *AITaskService) CreateAITask(aiTask *model.AITask) error {
	return s.aiTaskRepo.Create(aiTask)
}

func (s *AITaskService) UpdateAITask(aiTask *model.AITask) error {
	return s.aiTaskRepo.Update(aiTask)
}

func (s *AITaskService) GetByID(id int64) (*model.AITask, error) {
	return s.aiTaskRepo.GetByID(id)
}

// 获取进行中的ai任务
func (s *AITaskService) GetOngoingByWeChatID(wxID string) ([]*model.AITask, error) {
	return s.aiTaskRepo.GetOngoingByWeChatID(wxID)
}
