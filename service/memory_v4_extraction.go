package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type extractedMemberAlias struct {
	ChatRoomID  string `json:"chat_room_id"`
	MemberWxID  string `json:"member_wxid"`
	Alias       string `json:"alias"`
	AliasType   string `json:"alias_type"`
	Confidence  int    `json:"confidence"`
	SourceMsgID int64  `json:"source_msg_id"`
	ObservedBy  string `json:"observed_by"`
	IsActive    bool   `json:"is_active"`
}

type extractedMemberFact struct {
	ChatRoomID     string         `json:"chat_room_id"`
	SubjectWxID    string         `json:"subject_wxid"`
	Predicate      string         `json:"predicate"`
	ObjectText     string         `json:"object_text"`
	ObjectJSON     map[string]any `json:"object_json"`
	Polarity       int            `json:"polarity"`
	Confidence     int            `json:"confidence"`
	EvidenceMsgIDs []int64        `json:"evidence_msg_ids"`
	ObservedAt     int64          `json:"observed_at"`
	ValidFrom      int64          `json:"valid_from"`
	ValidUntil     int64          `json:"valid_until"`
	Status         string         `json:"status"`
}

type extractedMemberEvent struct {
	ChatRoomID     string   `json:"chat_room_id"`
	ActorWxIDs     []string `json:"actor_wxids"`
	TargetWxIDs    []string `json:"target_wxids"`
	EventType      string   `json:"event_type"`
	Summary        string   `json:"summary"`
	TimeStart      int64    `json:"time_start"`
	TimeEnd        int64    `json:"time_end"`
	MentionedAt    int64    `json:"mentioned_at"`
	Confidence     int      `json:"confidence"`
	EvidenceMsgIDs []int64  `json:"evidence_msg_ids"`
}

type extractedRelationshipAssertion struct {
	ChatRoomID     string  `json:"chat_room_id"`
	FromWxID       string  `json:"from_wxid"`
	ToWxID         string  `json:"to_wxid"`
	RelationType   string  `json:"relation_type"`
	Direction      string  `json:"direction"`
	Summary        string  `json:"summary"`
	Confidence     int     `json:"confidence"`
	EvidenceMsgIDs []int64 `json:"evidence_msg_ids"`
	ObservedAt     int64   `json:"observed_at"`
}

type extractedRelationshipEdge struct {
	ChatRoomID   string  `json:"chat_room_id"`
	FromWxID     string  `json:"from_wxid"`
	ToWxID       string  `json:"to_wxid"`
	RelationType string  `json:"relation_type"`
	Direction    string  `json:"direction"`
	Strength     int     `json:"strength"`
	Summary      string  `json:"summary"`
	EvidenceIDs  []int64 `json:"evidence_assert_ids"`
}

func (s *MemoryService) saveV4Extraction(ctx context.Context, result *memoryExtractionResult, chatRoomID string) error {
	if result == nil || s.memoryV4Repo == nil || vars.RobotRuntime.RobotCode == "" || strings.TrimSpace(chatRoomID) == "" {
		return nil
	}
	for _, item := range result.MemberAliases {
		if err := s.saveExtractedMemberAlias(item, chatRoomID); err != nil {
			return err
		}
	}
	for _, item := range result.MemberFacts {
		if err := s.saveExtractedMemberFact(item, chatRoomID); err != nil {
			return err
		}
	}
	for _, item := range result.MemberEvents {
		if err := s.saveExtractedMemberEvent(item, chatRoomID); err != nil {
			return err
		}
	}
	for _, item := range result.RelationshipAssertions {
		if err := s.saveExtractedRelationshipAssertion(item, chatRoomID); err != nil {
			return err
		}
	}
	for _, item := range result.RelationshipEdges {
		if err := s.saveExtractedRelationshipEdge(ctx, item, chatRoomID); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryService) saveExtractedMemberAlias(item extractedMemberAlias, chatRoomID string) error {
	item.ChatRoomID = defaultChatRoomID(item.ChatRoomID, chatRoomID)
	item.MemberWxID = strings.TrimSpace(item.MemberWxID)
	item.Alias = strings.TrimSpace(item.Alias)
	if item.ChatRoomID != chatRoomID || item.MemberWxID == "" || item.Alias == "" {
		return nil
	}
	now := time.Now().Unix()
	aliasType := model.MemberAliasType(strings.TrimSpace(item.AliasType))
	if aliasType == "" {
		aliasType = model.MemberAliasTypeObservedCallName
	}
	record := &model.MemberAlias{
		RobotCode:   vars.RobotRuntime.RobotCode,
		ChatRoomID:  chatRoomID,
		MemberWxID:  item.MemberWxID,
		Alias:       item.Alias,
		AliasType:   aliasType,
		Confidence:  clampInt(defaultInt(item.Confidence, 70), 1, 100),
		Source:      "chat_extraction",
		SourceMsgID: item.SourceMsgID,
		ObservedBy:  strings.TrimSpace(item.ObservedBy),
		IsActive:    item.IsActive,
		FirstSeenAt: now,
		LastSeenAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.memoryV4Repo.UpsertMemberAlias(record)
}

func (s *MemoryService) saveExtractedMemberFact(item extractedMemberFact, chatRoomID string) error {
	item.ChatRoomID = defaultChatRoomID(item.ChatRoomID, chatRoomID)
	item.SubjectWxID = strings.TrimSpace(item.SubjectWxID)
	item.Predicate = strings.TrimSpace(item.Predicate)
	item.ObjectText = strings.TrimSpace(item.ObjectText)
	if item.ChatRoomID != chatRoomID || item.SubjectWxID == "" || item.Predicate == "" || item.ObjectText == "" {
		return nil
	}
	now := time.Now().Unix()
	status := model.MemberFactStatus(strings.TrimSpace(item.Status))
	if status == "" {
		status = model.MemberFactStatusActive
	}
	fact := &model.MemberFact{
		RobotCode:      vars.RobotRuntime.RobotCode,
		ChatRoomID:     chatRoomID,
		SubjectWxID:    item.SubjectWxID,
		Predicate:      item.Predicate,
		ObjectText:     item.ObjectText,
		ObjectJSON:     jsonData(defaultObjectJSON(item.ObjectJSON)),
		Polarity:       defaultInt(item.Polarity, 1),
		Confidence:     clampInt(defaultInt(item.Confidence, 70), 1, 100),
		Source:         "chat",
		EvidenceMsgIDs: jsonData(item.EvidenceMsgIDs),
		Hash:           stableMemoryV4Hash(vars.RobotRuntime.RobotCode, chatRoomID, item.SubjectWxID, item.Predicate, item.ObjectText),
		ObservedAt:     defaultInt64(item.ObservedAt, now),
		ValidFrom:      item.ValidFrom,
		ValidUntil:     item.ValidUntil,
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return s.memoryV4Repo.UpsertMemberFact(fact)
}

func (s *MemoryService) saveExtractedMemberEvent(item extractedMemberEvent, chatRoomID string) error {
	item.ChatRoomID = defaultChatRoomID(item.ChatRoomID, chatRoomID)
	item.ActorWxIDs = normalizeStrings(item.ActorWxIDs)
	item.TargetWxIDs = normalizeStrings(item.TargetWxIDs)
	item.EventType = strings.TrimSpace(item.EventType)
	item.Summary = strings.TrimSpace(item.Summary)
	if item.ChatRoomID != chatRoomID || len(item.ActorWxIDs) == 0 || item.Summary == "" {
		return nil
	}
	if item.EventType == "" {
		item.EventType = "other"
	}
	now := time.Now().Unix()
	event := &model.MemberEvent{
		RobotCode:      vars.RobotRuntime.RobotCode,
		ChatRoomID:     chatRoomID,
		EventType:      item.EventType,
		Summary:        item.Summary,
		ActorWxIDs:     jsonData(item.ActorWxIDs),
		TargetWxIDs:    jsonData(item.TargetWxIDs),
		Confidence:     clampInt(defaultInt(item.Confidence, 70), 1, 100),
		Source:         "chat",
		EvidenceMsgIDs: jsonData(item.EvidenceMsgIDs),
		Hash:           stableMemoryV4Hash(vars.RobotRuntime.RobotCode, chatRoomID, strings.Join(item.ActorWxIDs, ","), item.EventType, item.Summary, fmt.Sprint(item.TimeStart), fmt.Sprint(item.TimeEnd)),
		TimeStart:      item.TimeStart,
		TimeEnd:        item.TimeEnd,
		MentionedAt:    defaultInt64(item.MentionedAt, now),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return s.memoryV4Repo.UpsertMemberEvent(event)
}

func (s *MemoryService) saveExtractedRelationshipAssertion(item extractedRelationshipAssertion, chatRoomID string) error {
	item.ChatRoomID = defaultChatRoomID(item.ChatRoomID, chatRoomID)
	item.FromWxID = strings.TrimSpace(item.FromWxID)
	item.ToWxID = strings.TrimSpace(item.ToWxID)
	item.RelationType = strings.TrimSpace(item.RelationType)
	item.Summary = strings.TrimSpace(item.Summary)
	if item.ChatRoomID != chatRoomID || item.FromWxID == "" || item.ToWxID == "" || item.FromWxID == item.ToWxID || item.Summary == "" {
		return nil
	}
	direction := relationshipDirection(item.Direction)
	fromWxID, toWxID := canonicalRelationshipPair(item.FromWxID, item.ToWxID, direction)
	if item.RelationType == "" {
		item.RelationType = "other"
	}
	now := time.Now().Unix()
	assertion := &model.MemberRelationshipAssertion{
		RobotCode:      vars.RobotRuntime.RobotCode,
		ChatRoomID:     chatRoomID,
		FromWxID:       fromWxID,
		ToWxID:         toWxID,
		RelationType:   item.RelationType,
		Direction:      direction,
		Summary:        item.Summary,
		Confidence:     clampInt(defaultInt(item.Confidence, 70), 1, 100),
		EvidenceMsgIDs: jsonData(item.EvidenceMsgIDs),
		Hash:           stableMemoryV4Hash(vars.RobotRuntime.RobotCode, chatRoomID, fromWxID, toWxID, item.RelationType, string(direction), item.Summary),
		ObservedAt:     defaultInt64(item.ObservedAt, now),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return s.memoryV4Repo.UpsertRelationshipAssertion(assertion)
}

func (s *MemoryService) saveExtractedRelationshipEdge(ctx context.Context, item extractedRelationshipEdge, chatRoomID string) error {
	item.ChatRoomID = defaultChatRoomID(item.ChatRoomID, chatRoomID)
	item.FromWxID = strings.TrimSpace(item.FromWxID)
	item.ToWxID = strings.TrimSpace(item.ToWxID)
	item.RelationType = strings.TrimSpace(item.RelationType)
	item.Summary = strings.TrimSpace(item.Summary)
	if item.ChatRoomID != chatRoomID || item.FromWxID == "" || item.ToWxID == "" || item.FromWxID == item.ToWxID {
		return nil
	}
	if item.RelationType == "" {
		item.RelationType = "other"
	}
	direction := relationshipDirection(item.Direction)
	fromWxID, toWxID := canonicalRelationshipPair(item.FromWxID, item.ToWxID, direction)
	now := time.Now().Unix()
	edge := &model.MemberRelationshipEdge{
		RobotCode:         vars.RobotRuntime.RobotCode,
		ChatRoomID:        chatRoomID,
		FromWxID:          fromWxID,
		ToWxID:            toWxID,
		RelationType:      item.RelationType,
		Direction:         direction,
		Strength:          clampInt(defaultInt(item.Strength, 50), 1, 100),
		Summary:           item.Summary,
		EvidenceAssertIDs: jsonData(item.EvidenceIDs),
		LastSeenAt:        now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.memoryV4Repo.UpsertRelationshipEdge(edge); err != nil {
		return err
	}
	_, err := s.saveExtractedMemory(ctx, extractedMemory{
		Scope:        string(model.MemoryScopeRelation),
		ChatRoomID:   chatRoomID,
		Category:     string(model.MemoryCategoryRelation),
		Content:      item.Summary,
		Summary:      item.Summary,
		Participants: []string{fromWxID, toWxID},
		Importance:   7,
		Confidence:   clampInt(defaultInt(item.Strength, 70), 1, 100),
	}, true, chatRoomID, "")
	return err
}

func defaultChatRoomID(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func relationshipDirection(value string) model.RelationshipDirection {
	switch model.RelationshipDirection(strings.TrimSpace(value)) {
	case model.RelationshipDirectionDirected:
		return model.RelationshipDirectionDirected
	default:
		return model.RelationshipDirectionUndirected
	}
}

func canonicalRelationshipPair(fromWxID, toWxID string, direction model.RelationshipDirection) (string, string) {
	if direction == model.RelationshipDirectionUndirected && fromWxID > toWxID {
		return toWxID, fromWxID
	}
	return fromWxID, toWxID
}

func defaultObjectJSON(value map[string]any) any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func defaultInt64(value, defaultValue int64) int64 {
	if value == 0 {
		return defaultValue
	}
	return value
}

func stableMemoryV4Hash(parts ...string) string {
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized = append(normalized, strings.TrimSpace(part))
	}
	data, _ := json.Marshal(normalized)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
