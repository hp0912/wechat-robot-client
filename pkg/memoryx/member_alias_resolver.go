package memoryx

import (
	"sort"
	"strings"

	"wechat-robot-client/model"
)

type MemberAliasResolver struct {
	members []*model.ChatRoomMember
	aliases []*model.MemberAlias
}

type MemberAliasCandidate struct {
	MemberWxID  string
	DisplayName string
	MatchedName string
	MatchMode   string
	Confidence  int
	LastSeenAt  int64
	Member      *model.ChatRoomMember
}

type currentNameMatch struct {
	Member          *model.ChatRoomMember
	Name            string
	Mode            string
	NormalizedInput string
	OriginalInput   string
	Confidence      int
}

func NewMemberAliasResolver(members []*model.ChatRoomMember, aliases []*model.MemberAlias) *MemberAliasResolver {
	return &MemberAliasResolver{
		members: members,
		aliases: aliases,
	}
}

func (r *MemberAliasResolver) Resolve(inputName string) []MemberAliasCandidate {
	inputName = strings.TrimSpace(inputName)
	if inputName == "" || r == nil {
		return nil
	}
	normalizedInput := normalizeAliasForMatch(inputName)
	memberByWxID := r.memberByWxID()
	bestByMember := make(map[string]MemberAliasCandidate)

	for _, member := range r.members {
		if member == nil || strings.TrimSpace(member.WechatID) == "" {
			continue
		}
		r.addCurrentNameMatch(bestByMember, currentNameMatch{
			Member:          member,
			Name:            member.WechatID,
			Mode:            "wechat_id",
			NormalizedInput: normalizedInput,
			OriginalInput:   inputName,
			Confidence:      120,
		})
		r.addCurrentNameMatch(bestByMember, currentNameMatch{
			Member:          member,
			Name:            member.Remark,
			Mode:            string(model.MemberAliasTypeCurrentRemark),
			NormalizedInput: normalizedInput,
			OriginalInput:   inputName,
			Confidence:      110,
		})
		r.addCurrentNameMatch(bestByMember, currentNameMatch{
			Member:          member,
			Name:            member.Nickname,
			Mode:            string(model.MemberAliasTypeCurrentNickname),
			NormalizedInput: normalizedInput,
			OriginalInput:   inputName,
			Confidence:      105,
		})
		r.addCurrentNameMatch(bestByMember, currentNameMatch{
			Member:          member,
			Name:            member.Alias,
			Mode:            string(model.MemberAliasTypeWechatAlias),
			NormalizedInput: normalizedInput,
			OriginalInput:   inputName,
			Confidence:      100,
		})
	}

	for _, alias := range r.aliases {
		if alias == nil || strings.TrimSpace(alias.MemberWxID) == "" || strings.TrimSpace(alias.Alias) == "" {
			continue
		}
		if normalizeAliasForMatch(alias.Alias) != normalizedInput {
			continue
		}
		member := memberByWxID[alias.MemberWxID]
		if member == nil {
			continue
		}
		confidence := clampInt(defaultInt(alias.Confidence, 70), 1, 100)
		if alias.IsActive {
			confidence += 5
		}
		r.keepBest(bestByMember, MemberAliasCandidate{
			MemberWxID:  alias.MemberWxID,
			DisplayName: DisplayNameForMember(member),
			MatchedName: alias.Alias,
			MatchMode:   string(alias.AliasType),
			Confidence:  clampInt(confidence, 1, 100),
			LastSeenAt:  alias.LastSeenAt,
			Member:      member,
		})
	}

	matches := make([]MemberAliasCandidate, 0, len(bestByMember))
	for _, candidate := range bestByMember {
		matches = append(matches, candidate)
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Confidence != matches[j].Confidence {
			return matches[i].Confidence > matches[j].Confidence
		}
		if matches[i].LastSeenAt != matches[j].LastSeenAt {
			return matches[i].LastSeenAt > matches[j].LastSeenAt
		}
		return matches[i].MemberWxID < matches[j].MemberWxID
	})
	return matches
}

func (r *MemberAliasResolver) memberByWxID() map[string]*model.ChatRoomMember {
	memberByWxID := make(map[string]*model.ChatRoomMember, len(r.members))
	for _, member := range r.members {
		if member == nil || strings.TrimSpace(member.WechatID) == "" {
			continue
		}
		memberByWxID[member.WechatID] = member
	}
	return memberByWxID
}

func (r *MemberAliasResolver) addCurrentNameMatch(bestByMember map[string]MemberAliasCandidate, match currentNameMatch) {
	if match.Member == nil {
		return
	}
	name := strings.TrimSpace(match.Name)
	if name == "" || normalizeAliasForMatch(name) != match.NormalizedInput {
		return
	}
	r.keepBest(bestByMember, MemberAliasCandidate{
		MemberWxID:  match.Member.WechatID,
		DisplayName: DisplayNameForMember(match.Member),
		MatchedName: match.OriginalInput,
		MatchMode:   match.Mode,
		Confidence:  match.Confidence,
		LastSeenAt:  match.Member.LastActiveAt,
		Member:      match.Member,
	})
}

func (r *MemberAliasResolver) keepBest(bestByMember map[string]MemberAliasCandidate, candidate MemberAliasCandidate) {
	existing, ok := bestByMember[candidate.MemberWxID]
	if !ok || candidate.Confidence > existing.Confidence || (candidate.Confidence == existing.Confidence && candidate.LastSeenAt > existing.LastSeenAt) {
		bestByMember[candidate.MemberWxID] = candidate
	}
}

func DisplayNameForMember(member *model.ChatRoomMember) string {
	if member == nil {
		return ""
	}
	for _, value := range []string{member.Remark, member.Nickname, member.Alias, member.WechatID} {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeAliasForMatch(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func defaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
