package openaitools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/openai/openai-go/v3"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type SearchChatRoomMemoryTool struct {
	db *gorm.DB
}

type chatRoomMemoryToolArgs struct {
	MemberNames []string `json:"member_names"`
	Query       string   `json:"query"`
}

type resolvedChatRoomMember struct {
	InputName string
	MatchMode string
	Members   []*model.ChatRoomMember
}

func NewSearchChatRoomMemoryTool(db *gorm.DB) OpenAITool {
	return &SearchChatRoomMemoryTool{db: db}
}

func (t *SearchChatRoomMemoryTool) GetOpenAITool(robotCtx *robotctx.RobotContext) *openai.ChatCompletionToolUnionParam {
	systemPrompt, err := t.BuildSystemPrompt(context.Background(), robotCtx)
	if err != nil {
		fmt.Printf("构建系统提示词失败: %v\n", err)
		return nil
	}
	if systemPrompt == "" {
		return nil
	}
	tool := openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "search_chat_room_memory",
		Description: openai.String("查询当前微信群成员的长期记忆、成员画像，以及两个群成员之间的关系。适用于用户询问某个群成员是怎样的人，或询问两个群成员关系、熟悉程度、互动模式等问题。"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"member_names": map[string]any{
					"type":        "array",
					"description": "用户问题中出现的当前微信群成员昵称、备注、微信号或微信ID。询问一个人时传 1 个，询问两人关系时传 2 个。",
					"items":       map[string]string{"type": "string"},
					"minItems":    1,
					"maxItems":    2,
				},
				"query": map[string]string{
					"type":        "string",
					"description": "用户的原始问题，用于保留查询意图。",
				},
			},
			"required": []string{"member_names"},
		},
	})
	return &tool
}

func (t *SearchChatRoomMemoryTool) BuildSystemPrompt(ctx context.Context, robotCtx *robotctx.RobotContext) (string, error) {
	if t.db == nil || robotCtx == nil || !strings.HasSuffix(robotCtx.FromWxID, "@chatroom") {
		return "", nil
	}
	return "微信群成员画像/关系查询工具", nil
}

func (t *SearchChatRoomMemoryTool) ExecuteToolCall(ctx context.Context, robotCtx *robotctx.RobotContext, toolCall openai.ChatCompletionMessageToolCallUnion) (string, bool, error) {
	var args chatRoomMemoryToolArgs
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("解析参数失败: %w", err)
	}
	args.MemberNames = normalizeMemberNames(args.MemberNames)
	if len(args.MemberNames) == 0 {
		return "member_names 不能为空", false, nil
	}
	if len(args.MemberNames) > 2 {
		return "一次最多查询两个群成员", false, nil
	}
	if t.db == nil {
		return "群成员记忆查询工具不可用", false, nil
	}
	if robotCtx == nil || !strings.HasSuffix(robotCtx.FromWxID, "@chatroom") {
		return "该工具只能在微信群聊中查询当前群成员信息", false, nil
	}
	if vars.RobotRuntime.RobotCode == "" {
		return "机器人编码为空，无法查询长期记忆", false, nil
	}

	crmRepo := repository.NewChatRoomMemberRepo(ctx, t.db)
	members, err := crmRepo.GetChatRoomMembers(robotCtx.FromWxID)
	if err != nil {
		return "", false, fmt.Errorf("查询群成员失败: %w", err)
	}

	resolved := make([]resolvedChatRoomMember, 0, len(args.MemberNames))
	for _, name := range args.MemberNames {
		resolved = append(resolved, resolveChatRoomMember(members, name))
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[用户问题]\n%s\n\n", strings.TrimSpace(args.Query))
	sb.WriteString("[群成员匹配]\n")
	allResolved := true
	for _, item := range resolved {
		fmt.Fprintf(&sb, "- 输入 %q：", item.InputName)
		switch len(item.Members) {
		case 0:
			sb.WriteString("未找到匹配成员\n")
			allResolved = false
		case 1:
			fmt.Fprintf(&sb, "%s（%s匹配）\n", formatChatRoomMemberName(item.Members[0]), item.MatchMode)
		default:
			allResolved = false
			fmt.Fprintf(&sb, "匹配到多个候选（%s匹配），需要用户进一步确认：\n", item.MatchMode)
			for _, member := range item.Members {
				fmt.Fprintf(&sb, "  - %s\n", formatChatRoomMemberName(member))
			}
		}
	}
	if !allResolved {
		return strings.TrimSpace(sb.String()), false, nil
	}

	memoryRepo := repository.NewMemoryRepo(ctx, t.db)
	memberByWxID := buildChatRoomMemberByWxID(members)
	if len(resolved) == 1 {
		member := resolved[0].Members[0]
		t.renderMemberProfileAndMemories(&sb, memoryRepo, robotCtx.FromWxID, member)
		t.renderMemberRelationships(&sb, memoryRepo, robotCtx.FromWxID, member, memberByWxID)
		return strings.TrimSpace(sb.String()), false, nil
	}

	firstMember := resolved[0].Members[0]
	secondMember := resolved[1].Members[0]
	t.renderRelationshipBetween(&sb, memoryRepo, robotCtx.FromWxID, firstMember, secondMember)
	t.renderMemberProfileAndMemories(&sb, memoryRepo, robotCtx.FromWxID, firstMember)
	t.renderMemberProfileAndMemories(&sb, memoryRepo, robotCtx.FromWxID, secondMember)
	return strings.TrimSpace(sb.String()), false, nil
}

func (t *SearchChatRoomMemoryTool) renderMemberProfileAndMemories(sb *strings.Builder, memoryRepo *repository.Memory, chatRoomID string, member *model.ChatRoomMember) {
	fmt.Fprintf(sb, "\n[成员画像：%s]\n", formatChatRoomMemberName(member))
	profile, err := memoryRepo.GetMemberProfile(vars.RobotRuntime.RobotCode, chatRoomID, member.WechatID)
	if err != nil {
		fmt.Fprintf(sb, "查询成员画像失败：%v\n", err)
	} else if profile == nil {
		sb.WriteString("暂无成员画像\n")
	} else {
		writeNonEmptyLine(sb, "画像摘要", replaceWxIDs(profile.Summary, member))
		writeNonEmptyLine(sb, "性格特点", replaceWxIDs(profile.Personality, member))
		writeNonEmptyLine(sb, "沟通风格", replaceWxIDs(profile.CommunicationStyle, member))
		writeNonEmptyLine(sb, "对机器人的态度", replaceWxIDs(profile.AttitudeToBot, member))
		writeStringListLine(sb, "兴趣", parseJSONStrings(profile.Interests))
		writeStringListLine(sb, "常聊话题", parseJSONStrings(profile.FrequentTopics))
		fmt.Fprintf(sb, "置信度：%d\n", profile.Confidence)
	}

	memories, err := memoryRepo.ListMemberMemories(vars.RobotRuntime.RobotCode, chatRoomID, member.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询成员长期记忆失败：%v\n", err)
		return
	}
	if len(memories) == 0 {
		return
	}
	sb.WriteString("成员长期记忆：\n")
	for _, memory := range memories {
		fmt.Fprintf(sb, "- %s（重要性%d，置信度%d）\n", replaceWxIDs(memory.Content, member), memory.Importance, memory.Confidence)
	}
}

func (t *SearchChatRoomMemoryTool) renderMemberRelationships(sb *strings.Builder, memoryRepo *repository.Memory, chatRoomID string, member *model.ChatRoomMember, memberByWxID map[string]*model.ChatRoomMember) {
	relationships, err := memoryRepo.ListMemberRelationships(vars.RobotRuntime.RobotCode, chatRoomID, member.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "\n[成员关系]\n查询成员关系失败：%v\n", err)
		return
	}
	if len(relationships) == 0 {
		return
	}
	fmt.Fprintf(sb, "\n[成员关系：%s]\n", formatChatRoomMemberName(member))
	for _, rel := range relationships {
		fmt.Fprintf(sb, "- %s，强度%d：%s\n", rel.RelationType, rel.Strength, replaceWxIDs(rel.Summary, memberByWxID[rel.FromWxID], memberByWxID[rel.ToWxID]))
	}
}

func (t *SearchChatRoomMemoryTool) renderRelationshipBetween(sb *strings.Builder, memoryRepo *repository.Memory, chatRoomID string, firstMember, secondMember *model.ChatRoomMember) {
	fmt.Fprintf(sb, "\n[两人关系：%s 与 %s]\n", formatChatRoomMemberName(firstMember), formatChatRoomMemberName(secondMember))
	relationships, err := memoryRepo.ListMemberRelationshipsBetween(vars.RobotRuntime.RobotCode, chatRoomID, firstMember.WechatID, secondMember.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询两人关系失败：%v\n", err)
	} else if len(relationships) == 0 {
		sb.WriteString("关系表暂无两人的直接关系记录\n")
	} else {
		for _, rel := range relationships {
			fmt.Fprintf(sb, "- 关系类型：%s，强度：%d，摘要：%s\n", rel.RelationType, rel.Strength, replaceWxIDs(rel.Summary, firstMember, secondMember))
		}
	}

	memories, err := memoryRepo.ListRelationMemoriesBetween(vars.RobotRuntime.RobotCode, chatRoomID, firstMember.WechatID, secondMember.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询两人关系记忆失败：%v\n", err)
		return
	}
	if len(memories) == 0 {
		return
	}
	sb.WriteString("关系长期记忆：\n")
	for _, memory := range memories {
		fmt.Fprintf(sb, "- %s（重要性%d，置信度%d）\n", replaceWxIDs(memory.Content, firstMember, secondMember), memory.Importance, memory.Confidence)
	}
}

func resolveChatRoomMember(members []*model.ChatRoomMember, inputName string) resolvedChatRoomMember {
	inputName = strings.TrimSpace(inputName)
	result := resolvedChatRoomMember{InputName: inputName}
	if inputName == "" {
		return result
	}

	exactMatches := filterChatRoomMembers(members, inputName, true)
	if len(exactMatches) > 0 {
		result.MatchMode = "完全"
		result.Members = exactMatches
		return result
	}

	containsMatches := filterChatRoomMembers(members, inputName, false)
	if len(containsMatches) > 5 {
		containsMatches = containsMatches[:5]
	}
	result.MatchMode = "包含"
	result.Members = containsMatches
	return result
}

func filterChatRoomMembers(members []*model.ChatRoomMember, inputName string, exact bool) []*model.ChatRoomMember {
	matched := make([]*model.ChatRoomMember, 0)
	for _, member := range members {
		if member == nil {
			continue
		}
		if exact && chatRoomMemberExactRank(member, inputName) >= 0 {
			matched = append(matched, member)
			continue
		}
		if !exact && chatRoomMemberContainsRank(member, inputName) >= 0 {
			matched = append(matched, member)
		}
	}
	sort.SliceStable(matched, func(i, j int) bool {
		leftRank := chatRoomMemberMatchRank(matched[i], inputName, exact)
		rightRank := chatRoomMemberMatchRank(matched[j], inputName, exact)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if matched[i].LastActiveAt != matched[j].LastActiveAt {
			return matched[i].LastActiveAt > matched[j].LastActiveAt
		}
		return matched[i].ID > matched[j].ID
	})
	return matched
}

func chatRoomMemberMatchRank(member *model.ChatRoomMember, inputName string, exact bool) int {
	if exact {
		return chatRoomMemberExactRank(member, inputName)
	}
	return chatRoomMemberContainsRank(member, inputName)
}

func chatRoomMemberExactRank(member *model.ChatRoomMember, inputName string) int {
	fields := []string{member.Nickname, member.Remark, member.Alias, member.WechatID}
	for i, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" && (field == inputName || strings.EqualFold(field, inputName)) {
			return i
		}
	}
	return -1
}

func chatRoomMemberContainsRank(member *model.ChatRoomMember, inputName string) int {
	inputName = strings.ToLower(inputName)
	fields := []string{member.Nickname, member.Remark, member.Alias, member.WechatID}
	for i, field := range fields {
		field = strings.ToLower(strings.TrimSpace(field))
		if field != "" && strings.Contains(field, inputName) {
			return i
		}
	}
	return -1
}

func normalizeMemberNames(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func buildChatRoomMemberByWxID(members []*model.ChatRoomMember) map[string]*model.ChatRoomMember {
	result := make(map[string]*model.ChatRoomMember, len(members))
	for _, member := range members {
		if member == nil || member.WechatID == "" {
			continue
		}
		result[member.WechatID] = member
	}
	return result
}

func formatChatRoomMemberName(member *model.ChatRoomMember) string {
	if member == nil {
		return ""
	}
	name := firstNonEmpty(member.Remark, member.Nickname, member.Alias, member.WechatID)
	parts := []string{name}
	if member.Nickname != "" && member.Nickname != name {
		parts = append(parts, "昵称: "+member.Nickname)
	}
	if member.Remark != "" && member.Remark != name {
		parts = append(parts, "备注: "+member.Remark)
	}
	if member.Alias != "" && member.Alias != name {
		parts = append(parts, "微信号: "+member.Alias)
	}
	return strings.Join(parts, "，")
}

func replaceWxIDs(content string, members ...*model.ChatRoomMember) string {
	result := content
	for _, member := range members {
		if member == nil || member.WechatID == "" {
			continue
		}
		result = strings.ReplaceAll(result, member.WechatID, firstNonEmpty(member.Remark, member.Nickname, member.Alias, member.WechatID))
	}
	return result
}

func writeNonEmptyLine(sb *strings.Builder, label, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	fmt.Fprintf(sb, "%s：%s\n", label, value)
}

func writeStringListLine(sb *strings.Builder, label string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(sb, "%s：%s\n", label, strings.Join(values, "、"))
}

func parseJSONStrings(data datatypes.JSON) []string {
	var values []string
	if len(data) == 0 {
		return values
	}
	if err := json.Unmarshal(data, &values); err != nil {
		return nil
	}
	return values
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
