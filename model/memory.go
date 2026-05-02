package model

import "gorm.io/datatypes"

type MemoryScope string

const (
	MemoryScopeFriend      MemoryScope = "friend"
	MemoryScopeGroupMember MemoryScope = "group_member"
	MemoryScopeGroup       MemoryScope = "group"
	MemoryScopeRelation    MemoryScope = "relation"
)

type MemoryCategory string

const (
	MemoryCategoryProfile    MemoryCategory = "profile"
	MemoryCategoryPreference MemoryCategory = "preference"
	MemoryCategoryFact       MemoryCategory = "fact"
	MemoryCategoryEvent      MemoryCategory = "event"
	MemoryCategoryRelation   MemoryCategory = "relation"
	MemoryCategoryEmotion    MemoryCategory = "emotion"
	MemoryCategoryTopic      MemoryCategory = "topic"
	MemoryCategoryReminder   MemoryCategory = "reminder"
)

type Memory struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode      string         `gorm:"column:robot_code;type:varchar(64);not null;index:idx_memory_scope,priority:1;uniqueIndex:idx_memory_hash,priority:1" json:"robot_code"`
	Scope          MemoryScope    `gorm:"column:scope;type:varchar(32);not null;index:idx_memory_scope,priority:2" json:"scope"`
	ContactWxID    string         `gorm:"column:contact_wxid;type:varchar(128);not null;default:'';index:idx_memory_scope,priority:3" json:"contact_wxid"`
	ChatRoomID     string         `gorm:"column:chat_room_id;type:varchar(128);not null;default:'';index:idx_memory_scope,priority:4" json:"chat_room_id"`
	Category       MemoryCategory `gorm:"column:category;type:varchar(32);not null;index:idx_memory_scope,priority:5" json:"category"`
	Content        string         `gorm:"column:content;type:text;not null" json:"content"`
	Summary        string         `gorm:"column:summary;type:text" json:"summary"`
	Keywords       datatypes.JSON `gorm:"column:keywords;type:json" json:"keywords"`
	Participants   datatypes.JSON `gorm:"column:participants;type:json" json:"participants"`
	Importance     int            `gorm:"column:importance;not null;default:5" json:"importance"`
	Confidence     int            `gorm:"column:confidence;not null;default:70" json:"confidence"`
	Source         string         `gorm:"column:source;type:varchar(32);not null;default:'chat'" json:"source"`
	EvidenceMsgIDs datatypes.JSON `gorm:"column:evidence_msg_ids;type:json" json:"evidence_msg_ids"`
	Hash           string         `gorm:"column:hash;type:char(64);not null;uniqueIndex:idx_memory_hash,priority:2" json:"hash"`
	VectorID       string         `gorm:"column:vector_id;type:varchar(64);index:idx_memory_vector_id" json:"vector_id"`
	OccurredAt     int64          `gorm:"column:occurred_at;not null;default:0" json:"occurred_at"`
	LastSeenAt     int64          `gorm:"column:last_seen_at;not null;default:0" json:"last_seen_at"`
	ExpiresAt      int64          `gorm:"column:expires_at;not null;default:0" json:"expires_at"`
	CreatedAt      int64          `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt      int64          `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (Memory) TableName() string {
	return "memories_v3"
}

type MemoryExtractionState struct {
	ID                 int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode          string `gorm:"column:robot_code;type:varchar(64);not null;uniqueIndex:idx_memory_extraction_scope,priority:1" json:"robot_code"`
	Scope              string `gorm:"column:scope;type:varchar(32);not null;uniqueIndex:idx_memory_extraction_scope,priority:2" json:"scope"`
	ContactWxID        string `gorm:"column:contact_wxid;type:varchar(128);not null;default:'';uniqueIndex:idx_memory_extraction_scope,priority:3" json:"contact_wxid"`
	ChatRoomID         string `gorm:"column:chat_room_id;type:varchar(128);not null;default:'';uniqueIndex:idx_memory_extraction_scope,priority:4" json:"chat_room_id"`
	WindowStartMsgID   int64  `gorm:"column:window_start_msg_id;not null;default:0" json:"window_start_msg_id"`
	WindowStartedAt    int64  `gorm:"column:window_started_at;not null;default:0" json:"window_started_at"`
	PendingCount       *int   `gorm:"column:pending_count;not null;default:0" json:"pending_count"`
	LastExtractedMsgID int64  `gorm:"column:last_extracted_msg_id;not null;default:0" json:"last_extracted_msg_id"`
	LastExtractedAt    int64  `gorm:"column:last_extracted_at;not null;default:0" json:"last_extracted_at"`
	CreatedAt          int64  `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt          int64  `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemoryExtractionState) TableName() string {
	return "memory_extraction_states_v3"
}

type MemberProfile struct {
	ID                 int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode          string         `gorm:"column:robot_code;type:varchar(64);not null;uniqueIndex:idx_member_profile,priority:1" json:"robot_code"`
	ChatRoomID         string         `gorm:"column:chat_room_id;type:varchar(128);not null;uniqueIndex:idx_member_profile,priority:2" json:"chat_room_id"`
	MemberWxID         string         `gorm:"column:member_wxid;type:varchar(128);not null;uniqueIndex:idx_member_profile,priority:3" json:"member_wxid"`
	Personality        string         `gorm:"column:personality;type:text" json:"personality"`
	Interests          datatypes.JSON `gorm:"column:interests;type:json" json:"interests"`
	CommunicationStyle string         `gorm:"column:communication_style;type:text" json:"communication_style"`
	FrequentTopics     datatypes.JSON `gorm:"column:frequent_topics;type:json" json:"frequent_topics"`
	AttitudeToBot      string         `gorm:"column:attitude_to_bot;type:text" json:"attitude_to_bot"`
	Summary            string         `gorm:"column:summary;type:text" json:"summary"`
	Confidence         int            `gorm:"column:confidence;not null;default:70" json:"confidence"`
	EvidenceMemoryIDs  datatypes.JSON `gorm:"column:evidence_memory_ids;type:json" json:"evidence_memory_ids"`
	LastAnalyzedAt     int64          `gorm:"column:last_analyzed_at;not null;default:0" json:"last_analyzed_at"`
	CreatedAt          int64          `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt          int64          `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemberProfile) TableName() string {
	return "member_profiles_v3"
}

type MemberRelationship struct {
	ID                int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RobotCode         string         `gorm:"column:robot_code;type:varchar(64);not null;uniqueIndex:idx_member_relationship,priority:1;index:idx_member_relationship_from,priority:1;index:idx_member_relationship_to,priority:1" json:"robot_code"`
	ChatRoomID        string         `gorm:"column:chat_room_id;type:varchar(128);not null;uniqueIndex:idx_member_relationship,priority:2;index:idx_member_relationship_from,priority:2;index:idx_member_relationship_to,priority:2" json:"chat_room_id"`
	FromWxID          string         `gorm:"column:from_wxid;type:varchar(128);not null;uniqueIndex:idx_member_relationship,priority:3;index:idx_member_relationship_from,priority:3" json:"from_wxid"`
	ToWxID            string         `gorm:"column:to_wxid;type:varchar(128);not null;uniqueIndex:idx_member_relationship,priority:4;index:idx_member_relationship_to,priority:3" json:"to_wxid"`
	RelationType      string         `gorm:"column:relation_type;type:varchar(64);not null;default:'';uniqueIndex:idx_member_relationship,priority:5" json:"relation_type"`
	Strength          int            `gorm:"column:strength;not null;default:50" json:"strength"`
	Summary           string         `gorm:"column:summary;type:text" json:"summary"`
	EvidenceMemoryIDs datatypes.JSON `gorm:"column:evidence_memory_ids;type:json" json:"evidence_memory_ids"`
	LastSeenAt        int64          `gorm:"column:last_seen_at;not null;default:0" json:"last_seen_at"`
	CreatedAt         int64          `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt         int64          `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
}

func (MemberRelationship) TableName() string {
	return "member_relationships_v3"
}
