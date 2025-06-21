package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type AITask struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewAITaskRepo(ctx context.Context, db *gorm.DB) *AITask {
	return &AITask{
		Ctx: ctx,
		DB:  db,
	}
}

func (repo *AITask) GetByID(id int64) (*model.AITask, error) {
	var task model.AITask
	err := repo.DB.WithContext(repo.Ctx).Where("id = ?", id).First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (repo *AITask) GetOngoingByWeChatID(wxID string) ([]*model.AITask, error) {
	var tasks []*model.AITask
	// 查询进行中的任务，状态为Pending或Processing，并且创建时间在3小时内
	err := repo.DB.WithContext(repo.Ctx).Where("contact_id = ? AND ai_task_status in (?) AND created_at > ?", wxID, []model.AITaskStatus{model.AITaskStatusPending, model.AITaskStatusProcessing}, time.Now().Add(-3*time.Hour)).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (repo *AITask) GetByMessageID(id int64) (*model.AITask, error) {
	var task model.AITask
	err := repo.DB.WithContext(repo.Ctx).Where("message_id = ?", id).First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (repo *AITask) GetByAIProviderTaskID(id string) (*model.AITask, error) {
	var task model.AITask
	err := repo.DB.WithContext(repo.Ctx).Where("ai_provider_task_id = ?", id).First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (repo *AITask) Create(data *model.AITask) error {
	return repo.DB.WithContext(repo.Ctx).Create(data).Error
}

func (repo *AITask) Update(data *model.AITask) error {
	return repo.DB.WithContext(repo.Ctx).Where("id = ?", data.ID).Updates(data).Error
}
