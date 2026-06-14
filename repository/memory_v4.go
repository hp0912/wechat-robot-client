package repository

import (
	"context"

	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type MemoryV4 struct {
	Ctx context.Context
	DB  *gorm.DB
}

type MemberEventQuery struct {
	RobotCode  string
	ChatRoomID string
	ActorWxIDs []string
	StartAt    int64
	EndAt      int64
	Limit      int
}

func NewMemoryV4Repo(ctx context.Context, db *gorm.DB) *MemoryV4 {
	return &MemoryV4{Ctx: ctx, DB: db}
}

func (r *MemoryV4) UpsertMemberAlias(alias *model.MemberAlias) error {
	var existing model.MemberAlias
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND member_wxid = ? AND alias = ? AND alias_type = ?",
			alias.RobotCode, alias.ChatRoomID, alias.MemberWxID, alias.Alias, alias.AliasType).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(r.Ctx).Create(alias).Error
	}
	if err != nil {
		return err
	}
	alias.ID = existing.ID
	alias.CreatedAt = existing.CreatedAt
	if alias.FirstSeenAt == 0 || (existing.FirstSeenAt > 0 && existing.FirstSeenAt < alias.FirstSeenAt) {
		alias.FirstSeenAt = existing.FirstSeenAt
	}
	if alias.LastSeenAt < existing.LastSeenAt {
		alias.LastSeenAt = existing.LastSeenAt
	}
	if alias.Confidence < existing.Confidence {
		alias.Confidence = existing.Confidence
	}
	return r.DB.WithContext(r.Ctx).Where("id = ?", existing.ID).Updates(alias).Error
}

func (r *MemoryV4) ListMemberAliases(robotCode, chatRoomID string) ([]*model.MemberAlias, error) {
	var aliases []*model.MemberAlias
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ?", robotCode, chatRoomID).
		Order("confidence DESC, last_seen_at DESC, id DESC").
		Find(&aliases).Error
	return aliases, err
}

func (r *MemoryV4) UpsertMemberFact(fact *model.MemberFact) error {
	var existing model.MemberFact
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND hash = ?", fact.RobotCode, fact.Hash).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(r.Ctx).Create(fact).Error
	}
	if err != nil {
		return err
	}
	fact.ID = existing.ID
	fact.CreatedAt = existing.CreatedAt
	if fact.Confidence < existing.Confidence {
		fact.Confidence = existing.Confidence
	}
	return r.DB.WithContext(r.Ctx).Where("id = ?", existing.ID).Updates(fact).Error
}

func (r *MemoryV4) ListMemberFacts(robotCode, chatRoomID, memberWxID string, limit int) ([]*model.MemberFact, error) {
	var facts []*model.MemberFact
	query := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND subject_wxid = ? AND status = ?", robotCode, chatRoomID, memberWxID, model.MemberFactStatusActive).
		Order("confidence DESC, observed_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&facts).Error; err != nil {
		return nil, err
	}
	return facts, nil
}

func (r *MemoryV4) UpsertMemberEvent(event *model.MemberEvent) error {
	var existing model.MemberEvent
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND hash = ?", event.RobotCode, event.Hash).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(r.Ctx).Create(event).Error
	}
	if err != nil {
		return err
	}
	event.ID = existing.ID
	event.CreatedAt = existing.CreatedAt
	if event.Confidence < existing.Confidence {
		event.Confidence = existing.Confidence
	}
	return r.DB.WithContext(r.Ctx).Where("id = ?", existing.ID).Updates(event).Error
}

func (r *MemoryV4) ListMemberEvents(req MemberEventQuery) ([]*model.MemberEvent, error) {
	var events []*model.MemberEvent
	query := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ?", req.RobotCode, req.ChatRoomID).
		Order("mentioned_at DESC, time_start DESC, id DESC")
	for _, wxID := range req.ActorWxIDs {
		if wxID == "" {
			continue
		}
		query = query.Where("JSON_CONTAINS(actor_wxids, JSON_QUOTE(?))", wxID)
	}
	if req.StartAt > 0 {
		query = query.Where("(time_start = 0 OR time_start >= ? OR mentioned_at >= ?)", req.StartAt, req.StartAt)
	}
	if req.EndAt > 0 {
		query = query.Where("(time_start = 0 OR time_start < ? OR mentioned_at < ?)", req.EndAt, req.EndAt)
	}
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *MemoryV4) UpsertRelationshipAssertion(assertion *model.MemberRelationshipAssertion) error {
	var existing model.MemberRelationshipAssertion
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND hash = ?", assertion.RobotCode, assertion.Hash).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(r.Ctx).Create(assertion).Error
	}
	if err != nil {
		return err
	}
	assertion.ID = existing.ID
	assertion.CreatedAt = existing.CreatedAt
	if assertion.Confidence < existing.Confidence {
		assertion.Confidence = existing.Confidence
	}
	return r.DB.WithContext(r.Ctx).Where("id = ?", existing.ID).Updates(assertion).Error
}

func (r *MemoryV4) ListRelationshipAssertionsBetween(robotCode, chatRoomID, firstWxID, secondWxID string, limit int) ([]*model.MemberRelationshipAssertion, error) {
	var assertions []*model.MemberRelationshipAssertion
	query := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ?", robotCode, chatRoomID).
		Where("((from_wxid = ? AND to_wxid = ?) OR (from_wxid = ? AND to_wxid = ?))", firstWxID, secondWxID, secondWxID, firstWxID).
		Order("confidence DESC, observed_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&assertions).Error; err != nil {
		return nil, err
	}
	return assertions, nil
}

func (r *MemoryV4) UpsertRelationshipEdge(edge *model.MemberRelationshipEdge) error {
	var existing model.MemberRelationshipEdge
	err := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND from_wxid = ? AND to_wxid = ? AND relation_type = ? AND direction = ?",
			edge.RobotCode, edge.ChatRoomID, edge.FromWxID, edge.ToWxID, edge.RelationType, edge.Direction).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.WithContext(r.Ctx).Create(edge).Error
	}
	if err != nil {
		return err
	}
	edge.ID = existing.ID
	edge.CreatedAt = existing.CreatedAt
	if edge.Strength < existing.Strength {
		edge.Strength = existing.Strength
	}
	if edge.LastSeenAt < existing.LastSeenAt {
		edge.LastSeenAt = existing.LastSeenAt
	}
	return r.DB.WithContext(r.Ctx).Where("id = ?", existing.ID).Updates(edge).Error
}

func (r *MemoryV4) ListRelationshipEdges(robotCode, chatRoomID, memberWxID string, limit int) ([]*model.MemberRelationshipEdge, error) {
	var edges []*model.MemberRelationshipEdge
	query := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ? AND (from_wxid = ? OR to_wxid = ?)", robotCode, chatRoomID, memberWxID, memberWxID).
		Order("strength DESC, last_seen_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&edges).Error; err != nil {
		return nil, err
	}
	return edges, nil
}

func (r *MemoryV4) ListRelationshipEdgesBetween(robotCode, chatRoomID, firstWxID, secondWxID string, limit int) ([]*model.MemberRelationshipEdge, error) {
	var edges []*model.MemberRelationshipEdge
	query := r.DB.WithContext(r.Ctx).
		Where("robot_code = ? AND chat_room_id = ?", robotCode, chatRoomID).
		Where("((from_wxid = ? AND to_wxid = ?) OR (from_wxid = ? AND to_wxid = ?))", firstWxID, secondWxID, secondWxID, firstWxID).
		Order("strength DESC, last_seen_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&edges).Error; err != nil {
		return nil, err
	}
	return edges, nil
}
