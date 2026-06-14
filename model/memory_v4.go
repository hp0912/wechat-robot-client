package model

import "gorm.io/datatypes"

type MemberAliasType string

const (
	MemberAliasTypeCurrentNickname  MemberAliasType = "current_nickname"
	MemberAliasTypeCurrentRemark    MemberAliasType = "current_remark"
	MemberAliasTypeWechatAlias      MemberAliasType = "wechat_alias"
	MemberAliasTypeOldNickname      MemberAliasType = "old_nickname"
	MemberAliasTypeOldRemark        MemberAliasType = "old_remark"
	MemberAliasTypeOldWechatAlias   MemberAliasType = "old_wechat_alias"
	MemberAliasTypeObservedCallName MemberAliasType = "observed_call_name"
	MemberAliasTypeSelfClaimed      MemberAliasType = "self_claimed"
)

type MemberAlias struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode   string          `gorm:"column:robot_code;type:varchar(64);not null;uniqueIndex:idx_member_alias,priority:1;index:idx_member_alias_lookup,priority:1" json:"robot_code"`
	ChatRoomID  string          `gorm:"column:chat_room_id;type:varchar(128);not null;uniqueIndex:idx_member_alias,priority:2;index:idx_member_alias_lookup,priority:2" json:"chat_room_id"`
	MemberWxID  string          `gorm:"column:member_wxid;type:varchar(128);not null;uniqueIndex:idx_member_alias,priority:3;index:idx_member_alias_member,priority:1" json:"member_wxid"`
	Alias       string          `gorm:"column:alias;type:varchar(255);not null;uniqueIndex:idx_member_alias,priority:4;index:idx_member_alias_lookup,priority:3" json:"alias"`
	AliasType   MemberAliasType `gorm:"column:alias_type;type:varchar(64);not null;uniqueIndex:idx_member_alias,priority:5" json:"alias_type"`
	Confidence  int             `gorm:"column:confidence;not null;default:70" json:"confidence"`
	Source      string          `gorm:"column:source;type:varchar(64);not null;default:''" json:"source"`
	SourceMsgID int64           `gorm:"column:source_msg_id;not null;default:0" json:"source_msg_id"`
	ObservedBy  string          `gorm:"column:observed_by;type:varchar(128);not null;default:''" json:"observed_by"`
	IsActive    bool            `gorm:"column:is_active;not null;default:true;index:idx_member_alias_lookup,priority:4" json:"is_active"`
	FirstSeenAt int64           `gorm:"column:first_seen_at;not null;default:0" json:"first_seen_at"`
	LastSeenAt  int64           `gorm:"column:last_seen_at;not null;default:0" json:"last_seen_at"`
	CreatedAt   int64           `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt   int64           `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemberAlias) TableName() string {
	return "member_aliases_v4"
}

type MemberFactStatus string

const (
	MemberFactStatusActive       MemberFactStatus = "active"
	MemberFactStatusContradicted MemberFactStatus = "contradicted"
	MemberFactStatusExpired      MemberFactStatus = "expired"
)

type MemberFact struct {
	ID             int64            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode      string           `gorm:"column:robot_code;type:varchar(64);not null;index:idx_member_fact_subject,priority:1;uniqueIndex:idx_member_fact_hash,priority:1" json:"robot_code"`
	ChatRoomID     string           `gorm:"column:chat_room_id;type:varchar(128);not null;index:idx_member_fact_subject,priority:2" json:"chat_room_id"`
	SubjectWxID    string           `gorm:"column:subject_wxid;type:varchar(128);not null;index:idx_member_fact_subject,priority:3" json:"subject_wxid"`
	Predicate      string           `gorm:"column:predicate;type:varchar(64);not null;index:idx_member_fact_subject,priority:4" json:"predicate"`
	ObjectText     string           `gorm:"column:object_text;type:text;not null" json:"object_text"`
	ObjectJSON     datatypes.JSON   `gorm:"column:object_json;type:json" json:"object_json"`
	Polarity       int              `gorm:"column:polarity;not null;default:1" json:"polarity"`
	Confidence     int              `gorm:"column:confidence;not null;default:70" json:"confidence"`
	Source         string           `gorm:"column:source;type:varchar(64);not null;default:'chat'" json:"source"`
	EvidenceMsgIDs datatypes.JSON   `gorm:"column:evidence_msg_ids;type:json" json:"evidence_msg_ids"`
	Hash           string           `gorm:"column:hash;type:char(64);not null;uniqueIndex:idx_member_fact_hash,priority:2" json:"hash"`
	ObservedAt     int64            `gorm:"column:observed_at;not null;default:0;index:idx_member_fact_observed_at" json:"observed_at"`
	ValidFrom      int64            `gorm:"column:valid_from;not null;default:0" json:"valid_from"`
	ValidUntil     int64            `gorm:"column:valid_until;not null;default:0" json:"valid_until"`
	Status         MemberFactStatus `gorm:"column:status;type:varchar(32);not null;default:'active';index:idx_member_fact_subject,priority:5" json:"status"`
	CreatedAt      int64            `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt      int64            `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemberFact) TableName() string {
	return "member_facts_v4"
}

type MemberEvent struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode      string         `gorm:"column:robot_code;type:varchar(64);not null;index:idx_member_event_scope,priority:1;uniqueIndex:idx_member_event_hash,priority:1" json:"robot_code"`
	ChatRoomID     string         `gorm:"column:chat_room_id;type:varchar(128);not null;index:idx_member_event_scope,priority:2" json:"chat_room_id"`
	EventType      string         `gorm:"column:event_type;type:varchar(64);not null;index:idx_member_event_scope,priority:3" json:"event_type"`
	Summary        string         `gorm:"column:summary;type:text;not null" json:"summary"`
	ActorWxIDs     datatypes.JSON `gorm:"column:actor_wxids;type:json" json:"actor_wxids"`
	TargetWxIDs    datatypes.JSON `gorm:"column:target_wxids;type:json" json:"target_wxids"`
	Confidence     int            `gorm:"column:confidence;not null;default:70" json:"confidence"`
	Source         string         `gorm:"column:source;type:varchar(64);not null;default:'chat'" json:"source"`
	EvidenceMsgIDs datatypes.JSON `gorm:"column:evidence_msg_ids;type:json" json:"evidence_msg_ids"`
	Hash           string         `gorm:"column:hash;type:char(64);not null;uniqueIndex:idx_member_event_hash,priority:2" json:"hash"`
	TimeStart      int64          `gorm:"column:time_start;not null;default:0;index:idx_member_event_time" json:"time_start"`
	TimeEnd        int64          `gorm:"column:time_end;not null;default:0;index:idx_member_event_time" json:"time_end"`
	MentionedAt    int64          `gorm:"column:mentioned_at;not null;default:0;index:idx_member_event_time" json:"mentioned_at"`
	CreatedAt      int64          `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt      int64          `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemberEvent) TableName() string {
	return "member_events_v4"
}

type RelationshipDirection string

const (
	RelationshipDirectionUndirected RelationshipDirection = "undirected"
	RelationshipDirectionDirected   RelationshipDirection = "directed"
)

type MemberRelationshipAssertion struct {
	ID             int64                 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode      string                `gorm:"column:robot_code;type:varchar(64);not null;index:idx_rel_assertion_scope,priority:1;uniqueIndex:idx_rel_assertion_hash,priority:1" json:"robot_code"`
	ChatRoomID     string                `gorm:"column:chat_room_id;type:varchar(128);not null;index:idx_rel_assertion_scope,priority:2" json:"chat_room_id"`
	FromWxID       string                `gorm:"column:from_wxid;type:varchar(128);not null;index:idx_rel_assertion_from,priority:1" json:"from_wxid"`
	ToWxID         string                `gorm:"column:to_wxid;type:varchar(128);not null;index:idx_rel_assertion_to,priority:1" json:"to_wxid"`
	RelationType   string                `gorm:"column:relation_type;type:varchar(64);not null;index:idx_rel_assertion_scope,priority:3" json:"relation_type"`
	Direction      RelationshipDirection `gorm:"column:direction;type:varchar(32);not null;default:'undirected'" json:"direction"`
	Summary        string                `gorm:"column:summary;type:text;not null" json:"summary"`
	Confidence     int                   `gorm:"column:confidence;not null;default:70" json:"confidence"`
	EvidenceMsgIDs datatypes.JSON        `gorm:"column:evidence_msg_ids;type:json" json:"evidence_msg_ids"`
	Hash           string                `gorm:"column:hash;type:char(64);not null;uniqueIndex:idx_rel_assertion_hash,priority:2" json:"hash"`
	ObservedAt     int64                 `gorm:"column:observed_at;not null;default:0" json:"observed_at"`
	CreatedAt      int64                 `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt      int64                 `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemberRelationshipAssertion) TableName() string {
	return "member_relationship_assertions_v4"
}

type MemberRelationshipEdge struct {
	ID                int64                 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode         string                `gorm:"column:robot_code;type:varchar(64);not null;uniqueIndex:idx_rel_edge,priority:1;index:idx_rel_edge_from,priority:1;index:idx_rel_edge_to,priority:1" json:"robot_code"`
	ChatRoomID        string                `gorm:"column:chat_room_id;type:varchar(128);not null;uniqueIndex:idx_rel_edge,priority:2;index:idx_rel_edge_from,priority:2;index:idx_rel_edge_to,priority:2" json:"chat_room_id"`
	FromWxID          string                `gorm:"column:from_wxid;type:varchar(128);not null;uniqueIndex:idx_rel_edge,priority:3;index:idx_rel_edge_from,priority:3" json:"from_wxid"`
	ToWxID            string                `gorm:"column:to_wxid;type:varchar(128);not null;uniqueIndex:idx_rel_edge,priority:4;index:idx_rel_edge_to,priority:3" json:"to_wxid"`
	RelationType      string                `gorm:"column:relation_type;type:varchar(64);not null;uniqueIndex:idx_rel_edge,priority:5" json:"relation_type"`
	Direction         RelationshipDirection `gorm:"column:direction;type:varchar(32);not null;default:'undirected';uniqueIndex:idx_rel_edge,priority:6" json:"direction"`
	Strength          int                   `gorm:"column:strength;not null;default:50" json:"strength"`
	Summary           string                `gorm:"column:summary;type:text" json:"summary"`
	EvidenceAssertIDs datatypes.JSON        `gorm:"column:evidence_assert_ids;type:json" json:"evidence_assert_ids"`
	LastSeenAt        int64                 `gorm:"column:last_seen_at;not null;default:0" json:"last_seen_at"`
	CreatedAt         int64                 `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt         int64                 `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemberRelationshipEdge) TableName() string {
	return "member_relationship_edges_v4"
}
