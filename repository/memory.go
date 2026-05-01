package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type Memory struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewMemoryRepo(ctx context.Context, db *gorm.DB) *Memory {
	return &Memory{Ctx: ctx, DB: db}
}

func (r *Memory) GetState(robotCode, scope, contactWxID, chatRoomID string) (*model.MemoryExtractionState, error) {
	var state model.MemoryExtractionState
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND scope = ? AND contact_wxid = ? AND chat_room_id = ?", robotCode, scope, contactWxID, chatRoomID).
		First(&state).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *Memory) SaveState(state *model.MemoryExtractionState) error {
	if state.ID == 0 {
		return r.DB.WithContext(r.Ctx).Create(state).Error
	}
	return r.DB.WithContext(r.Ctx).Where("id = ?", state.ID).Updates(state).Error
}

func (r *Memory) GetByHash(robotCode, hash string) (*model.Memory, error) {
	var memory model.Memory
	err := r.DB.WithContext(r.Ctx).Where("robot_code = ? AND hash = ?", robotCode, hash).First(&memory).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &memory, nil
}

func (r *Memory) CreateMemory(memory *model.Memory) error {
	return r.DB.WithContext(r.Ctx).Create(memory).Error
}

func (r *Memory) UpdateMemory(memory *model.Memory) error {
	return r.DB.WithContext(r.Ctx).Where("id = ?", memory.ID).Updates(memory).Error
}

func (r *Memory) GetMemoriesByVectorIDs(vectorIDs []string) ([]*model.Memory, error) {
	if len(vectorIDs) == 0 {
		return nil, nil
	}
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).Where("vector_id IN ?", vectorIDs).Find(&memories).Error
	return memories, err
}

func (r *Memory) ListRelationMemories(robotCode, chatRoomID, participantWxID string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	query := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND scope = ?", robotCode, chatRoomID, model.MemoryScopeRelation).
		Where("JSON_CONTAINS(participants, JSON_QUOTE(?))", participantWxID).
		Order("importance DESC, last_seen_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&memories).Error; err != nil {
		return nil, err
	}
	return memories, nil
}

func (r *Memory) UpsertMemberProfile(profile *model.MemberProfile) error {
	var existing model.MemberProfile
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND member_wxid = ?", profile.RobotCode, profile.ChatRoomID, profile.MemberWxID).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(r.Ctx).Create(profile).Error
	}
	if err != nil {
		return err
	}
	profile.ID = existing.ID
	profile.CreatedAt = existing.CreatedAt
	return r.DB.WithContext(r.Ctx).Where("id = ?", existing.ID).Updates(profile).Error
}

func (r *Memory) GetMemberProfile(robotCode, chatRoomID, memberWxID string) (*model.MemberProfile, error) {
	var profile model.MemberProfile
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND member_wxid = ?", robotCode, chatRoomID, memberWxID).
		First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *Memory) UpsertMemberRelationship(rel *model.MemberRelationship) error {
	var existing model.MemberRelationship
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND from_wxid = ? AND to_wxid = ? AND relation_type = ?", rel.RobotCode, rel.ChatRoomID, rel.FromWxID, rel.ToWxID, rel.RelationType).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(r.Ctx).Create(rel).Error
	}
	if err != nil {
		return err
	}
	rel.ID = existing.ID
	rel.CreatedAt = existing.CreatedAt
	return r.DB.WithContext(r.Ctx).Where("id = ?", existing.ID).Updates(rel).Error
}

func (r *Memory) ListMemberRelationships(robotCode, chatRoomID, memberWxID string, limit int) ([]*model.MemberRelationship, error) {
	var relationships []*model.MemberRelationship
	query := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND (from_wxid = ? OR to_wxid = ?)", robotCode, chatRoomID, memberWxID, memberWxID).
		Order("strength DESC, last_seen_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&relationships).Error; err != nil {
		return nil, err
	}
	return relationships, nil
}
