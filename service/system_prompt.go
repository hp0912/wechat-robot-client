package service

import (
	"context"
	"errors"
	"strings"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type SystemPromptService struct {
	ctx              context.Context
	systemPromptRepo *repository.SystemPrompt
}

func NewSystemPromptService(ctx context.Context) *SystemPromptService {
	return &SystemPromptService{
		ctx:              ctx,
		systemPromptRepo: repository.NewSystemPromptRepo(ctx, vars.DB),
	}
}

func validateSystemPrompt(title, content string) (string, string, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return "", "", errors.New("标题不能为空")
	}
	if content == "" {
		return "", "", errors.New("提示词内容不能为空")
	}
	return title, content, nil
}

func normalizeSystemPromptKeyword(keyword string) string {
	return strings.TrimSpace(keyword)
}

func (s *SystemPromptService) Create(title, content string) (*model.SystemPrompt, error) {
	title, content, err := validateSystemPrompt(title, content)
	if err != nil {
		return nil, err
	}
	prompt := &model.SystemPrompt{
		Title:   title,
		Content: content,
	}
	if err := s.systemPromptRepo.Create(prompt); err != nil {
		return nil, err
	}
	return prompt, nil
}

func (s *SystemPromptService) Update(id int64, title, content string) error {
	if id <= 0 {
		return errors.New("id 参数错误")
	}
	title, content, err := validateSystemPrompt(title, content)
	if err != nil {
		return err
	}
	existing, err := s.systemPromptRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("系统提示词不存在")
	}
	return s.systemPromptRepo.Update(id, title, content)
}

func (s *SystemPromptService) Delete(id int64) error {
	if id <= 0 {
		return errors.New("id 参数错误")
	}
	existing, err := s.systemPromptRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("系统提示词不存在")
	}
	return s.systemPromptRepo.Delete(id)
}

func (s *SystemPromptService) GetByID(id int64) (*model.SystemPrompt, error) {
	if id <= 0 {
		return nil, errors.New("id 参数错误")
	}
	prompt, err := s.systemPromptRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if prompt == nil {
		return nil, errors.New("系统提示词不存在")
	}
	return prompt, nil
}

func (s *SystemPromptService) List(keyword string) ([]*model.SystemPrompt, error) {
	return s.systemPromptRepo.List(normalizeSystemPromptKeyword(keyword))
}
