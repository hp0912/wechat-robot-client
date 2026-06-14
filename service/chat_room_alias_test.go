package service

import (
	"testing"

	"wechat-robot-client/model"
)

func TestBuildMemberNameAliasRecordsDistinguishesMissingAndEmptyFields(t *testing.T) {
	emptyRemark := ""
	newNickname := "新牛"
	existing := &model.ChatRoomMember{
		WechatID: "wxid_boss",
		Nickname: "牛老大",
		Remark:   "牛群主",
		Alias:    "niu",
	}

	records := buildMemberNameAliasRecords(existing, memberNameObservation{
		Nickname: &newNickname,
		Remark:   &emptyRemark,
	})

	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d: %#v", len(records), records)
	}
	assertAliasRecord(t, records[0], "牛老大", model.MemberAliasTypeOldNickname, false)
	assertAliasRecord(t, records[1], "新牛", model.MemberAliasTypeCurrentNickname, true)
	assertAliasRecord(t, records[2], "牛群主", model.MemberAliasTypeOldRemark, false)
}

func TestBuildMemberNameAliasRecordsRefreshesUnchangedObservedAlias(t *testing.T) {
	alias := "niu"
	existing := &model.ChatRoomMember{
		WechatID: "wxid_boss",
		Alias:    "niu",
	}

	records := buildMemberNameAliasRecords(existing, memberNameObservation{
		Alias: &alias,
	})

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d: %#v", len(records), records)
	}
	assertAliasRecord(t, records[0], "niu", model.MemberAliasTypeWechatAlias, true)
}

func assertAliasRecord(t *testing.T, record memberAliasRecord, alias string, aliasType model.MemberAliasType, isActive bool) {
	t.Helper()
	if record.Alias != alias {
		t.Fatalf("expected alias %q, got %q", alias, record.Alias)
	}
	if record.AliasType != aliasType {
		t.Fatalf("expected alias type %q, got %q", aliasType, record.AliasType)
	}
	if record.IsActive != isActive {
		t.Fatalf("expected is_active %t, got %t", isActive, record.IsActive)
	}
}
