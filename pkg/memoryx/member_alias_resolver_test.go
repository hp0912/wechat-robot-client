package memoryx

import (
	"testing"

	"wechat-robot-client/model"
)

func TestMemberAliasResolverExactCurrentNickname(t *testing.T) {
	resolver := NewMemberAliasResolver(
		[]*model.ChatRoomMember{
			{WechatID: "wxid_boss", Nickname: "牛老大", LastActiveAt: 100},
			{WechatID: "wxid_other", Nickname: "小王", LastActiveAt: 200},
		},
		nil,
	)

	matches := resolver.Resolve("牛老大")

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %#v", len(matches), matches)
	}
	if matches[0].MemberWxID != "wxid_boss" {
		t.Fatalf("expected wxid_boss, got %q", matches[0].MemberWxID)
	}
	if matches[0].MatchMode != "current_nickname" {
		t.Fatalf("expected current_nickname match mode, got %q", matches[0].MatchMode)
	}
}

func TestMemberAliasResolverObservedAlias(t *testing.T) {
	resolver := NewMemberAliasResolver(
		[]*model.ChatRoomMember{
			{WechatID: "wxid_boss", Nickname: "牛老大", LastActiveAt: 100},
			{WechatID: "wxid_other", Nickname: "小王", LastActiveAt: 200},
		},
		[]*model.MemberAlias{
			{
				MemberWxID: "wxid_boss",
				Alias:      "老牛",
				AliasType:  model.MemberAliasTypeObservedCallName,
				Confidence: 86,
				IsActive:   true,
				LastSeenAt: 300,
			},
		},
	)

	matches := resolver.Resolve("老牛")

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %#v", len(matches), matches)
	}
	if matches[0].MemberWxID != "wxid_boss" {
		t.Fatalf("expected wxid_boss, got %q", matches[0].MemberWxID)
	}
	if matches[0].MatchMode != string(model.MemberAliasTypeObservedCallName) {
		t.Fatalf("expected observed alias match mode, got %q", matches[0].MatchMode)
	}
}

func TestMemberAliasResolverAmbiguousAlias(t *testing.T) {
	resolver := NewMemberAliasResolver(
		[]*model.ChatRoomMember{
			{WechatID: "wxid_first", Nickname: "老张", LastActiveAt: 100},
			{WechatID: "wxid_second", Nickname: "张哥", LastActiveAt: 200},
		},
		[]*model.MemberAlias{
			{
				MemberWxID: "wxid_first",
				Alias:      "张总",
				AliasType:  model.MemberAliasTypeObservedCallName,
				Confidence: 70,
				IsActive:   true,
				LastSeenAt: 100,
			},
			{
				MemberWxID: "wxid_second",
				Alias:      "张总",
				AliasType:  model.MemberAliasTypeObservedCallName,
				Confidence: 90,
				IsActive:   true,
				LastSeenAt: 200,
			},
		},
	)

	matches := resolver.Resolve("张总")

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d: %#v", len(matches), matches)
	}
	if matches[0].MemberWxID != "wxid_second" {
		t.Fatalf("expected highest confidence match first, got %q", matches[0].MemberWxID)
	}
	if matches[1].MemberWxID != "wxid_first" {
		t.Fatalf("expected lower confidence match second, got %q", matches[1].MemberWxID)
	}
}

func TestMemberAliasResolverOldNicknameStillMatchesAfterRename(t *testing.T) {
	resolver := NewMemberAliasResolver(
		[]*model.ChatRoomMember{
			{WechatID: "wxid_boss", Nickname: "新牛", LastActiveAt: 500},
		},
		[]*model.MemberAlias{
			{
				MemberWxID: "wxid_boss",
				Alias:      "牛老大",
				AliasType:  model.MemberAliasTypeOldNickname,
				Confidence: 95,
				IsActive:   false,
				LastSeenAt: 400,
			},
		},
	)

	matches := resolver.Resolve("牛老大")

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d: %#v", len(matches), matches)
	}
	if matches[0].MemberWxID != "wxid_boss" {
		t.Fatalf("expected wxid_boss, got %q", matches[0].MemberWxID)
	}
	if matches[0].DisplayName != "新牛" {
		t.Fatalf("expected current display name 新牛, got %q", matches[0].DisplayName)
	}
}
