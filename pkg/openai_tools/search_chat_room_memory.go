package openaitools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/memoryx"
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
	InputName  string
	MatchMode  string
	Members    []*model.ChatRoomMember
	Candidates []memoryx.MemberAliasCandidate
}

type memberV4RenderRequest struct {
	Builder      *strings.Builder
	Repo         *repository.MemoryV4
	Aliases      []*model.MemberAlias
	ChatRoomID   string
	Member       *model.ChatRoomMember
	MemberByWxID map[string]*model.ChatRoomMember
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
		Description: openai.String("查询当前微信群成员的长期记忆、基本情况、兴趣偏好、过去动态，以及两个群成员之间的关系。适用于用户询问某个群成员是怎样的人、最近做过什么，或询问两个群成员关系、熟悉程度、互动模式等问题。"),
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
	return "微信群成员记忆和关系查询工具", nil
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
	memoryV4Repo := repository.NewMemoryV4Repo(ctx, t.db)
	aliases, err := memoryV4Repo.ListMemberAliases(vars.RobotRuntime.RobotCode, robotCtx.FromWxID)
	if err != nil {
		return "", false, fmt.Errorf("查询群成员别称失败: %w", err)
	}
	aliasResolver := memoryx.NewMemberAliasResolver(members, aliases)

	resolved := make([]resolvedChatRoomMember, 0, len(args.MemberNames))
	for _, name := range args.MemberNames {
		resolved = append(resolved, t.resolveChatRoomMemberWithAliases(aliasResolver, name))
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
			fmt.Fprintf(&sb, "%s（通过%s匹配，%s）\n", formatChatRoomMemberName(item.Members[0]), humanAliasType(item.MatchMode), humanConfidence(item.Candidates[0].Confidence))
		default:
			allResolved = false
			fmt.Fprintf(&sb, "匹配到多个候选（通过%s匹配），需要用户进一步确认：\n", humanAliasType(item.MatchMode))
			for _, candidate := range item.Candidates {
				fmt.Fprintf(&sb, "  - %s（匹配到的叫法：%s，%s）\n", formatChatRoomMemberName(candidate.Member), candidate.MatchedName, humanConfidence(candidate.Confidence))
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
		t.renderMemberV4Context(memberV4RenderRequest{
			Builder:      &sb,
			Repo:         memoryV4Repo,
			Aliases:      aliases,
			ChatRoomID:   robotCtx.FromWxID,
			Member:       member,
			MemberByWxID: memberByWxID,
		})
		t.renderMemberProfileAndMemories(&sb, memoryRepo, robotCtx.FromWxID, member)
		t.renderMemberRelationships(&sb, memoryRepo, robotCtx.FromWxID, member, memberByWxID)
		return strings.TrimSpace(sb.String()), false, nil
	}

	firstMember := resolved[0].Members[0]
	secondMember := resolved[1].Members[0]
	t.renderV4RelationshipBetween(&sb, memoryV4Repo, robotCtx.FromWxID, firstMember, secondMember)
	t.renderRelationshipBetween(&sb, memoryRepo, robotCtx.FromWxID, firstMember, secondMember)
	t.renderMemberV4Context(memberV4RenderRequest{
		Builder:      &sb,
		Repo:         memoryV4Repo,
		Aliases:      aliases,
		ChatRoomID:   robotCtx.FromWxID,
		Member:       firstMember,
		MemberByWxID: memberByWxID,
	})
	t.renderMemberV4Context(memberV4RenderRequest{
		Builder:      &sb,
		Repo:         memoryV4Repo,
		Aliases:      aliases,
		ChatRoomID:   robotCtx.FromWxID,
		Member:       secondMember,
		MemberByWxID: memberByWxID,
	})
	t.renderMemberProfileAndMemories(&sb, memoryRepo, robotCtx.FromWxID, firstMember)
	t.renderMemberProfileAndMemories(&sb, memoryRepo, robotCtx.FromWxID, secondMember)
	return strings.TrimSpace(sb.String()), false, nil
}

func (t *SearchChatRoomMemoryTool) resolveChatRoomMemberWithAliases(resolver *memoryx.MemberAliasResolver, inputName string) resolvedChatRoomMember {
	result := resolvedChatRoomMember{InputName: strings.TrimSpace(inputName)}
	if resolver == nil || result.InputName == "" {
		return result
	}
	candidates := resolver.Resolve(result.InputName)
	result.Candidates = candidates
	result.Members = make([]*model.ChatRoomMember, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Member != nil {
			result.Members = append(result.Members, candidate.Member)
		}
	}
	if len(candidates) > 0 {
		result.MatchMode = candidates[0].MatchMode
	}
	return result
}

func (t *SearchChatRoomMemoryTool) renderMemberV4Context(req memberV4RenderRequest) {
	if req.Builder == nil || req.Repo == nil || req.Member == nil {
		return
	}
	fmt.Fprintf(req.Builder, "\n[我记得的成员信息：%s]\n", formatChatRoomMemberName(req.Member))
	t.renderMemberAliases(req.Builder, req.Aliases, req.Member)
	t.renderMemberFacts(req.Builder, req.Repo, req.ChatRoomID, req.Member)
	t.renderMemberEvents(req.Builder, req.Repo, req.ChatRoomID, req.Member)
	t.renderMemberV4Relationships(req.Builder, req.Repo, req.ChatRoomID, req.Member, req.MemberByWxID)
}

func (t *SearchChatRoomMemoryTool) renderMemberAliases(sb *strings.Builder, aliases []*model.MemberAlias, member *model.ChatRoomMember) {
	values := make([]string, 0, 5)
	for _, alias := range aliases {
		if alias == nil || alias.MemberWxID != member.WechatID || alias.Alias == "" {
			continue
		}
		values = append(values, fmt.Sprintf("%s（%s，%s）", alias.Alias, humanAliasType(string(alias.AliasType)), humanConfidence(alias.Confidence)))
		if len(values) >= 5 {
			break
		}
	}
	if len(values) > 0 {
		fmt.Fprintf(sb, "别称：%s\n", strings.Join(values, "、"))
	}
}

func (t *SearchChatRoomMemoryTool) renderMemberFacts(sb *strings.Builder, memoryV4Repo *repository.MemoryV4, chatRoomID string, member *model.ChatRoomMember) {
	facts, err := memoryV4Repo.ListMemberFacts(vars.RobotRuntime.RobotCode, chatRoomID, member.WechatID, 8)
	if err != nil {
		fmt.Fprintf(sb, "查询成员信息失败：%v\n", err)
		return
	}
	if len(facts) == 0 {
		return
	}
	sb.WriteString("成员信息：\n")
	for _, fact := range facts {
		fmt.Fprintf(sb, "- %s：%s（%s）\n", humanFactPredicate(fact.Predicate), fact.ObjectText, humanConfidence(fact.Confidence))
	}
}

func (t *SearchChatRoomMemoryTool) renderMemberEvents(sb *strings.Builder, memoryV4Repo *repository.MemoryV4, chatRoomID string, member *model.ChatRoomMember) {
	events, err := memoryV4Repo.ListMemberEvents(repository.MemberEventQuery{
		RobotCode:  vars.RobotRuntime.RobotCode,
		ChatRoomID: chatRoomID,
		ActorWxIDs: []string{member.WechatID},
		Limit:      5,
	})
	if err != nil {
		fmt.Fprintf(sb, "查询成员事件失败：%v\n", err)
		return
	}
	if len(events) == 0 {
		return
	}
	sb.WriteString("最近事件：\n")
	for _, event := range events {
		fmt.Fprintf(sb, "- %s：%s（%s）\n", humanEventType(event.EventType), event.Summary, humanConfidence(event.Confidence))
	}
}

func (t *SearchChatRoomMemoryTool) renderMemberV4Relationships(sb *strings.Builder, memoryV4Repo *repository.MemoryV4, chatRoomID string, member *model.ChatRoomMember, memberByWxID map[string]*model.ChatRoomMember) {
	edges, err := memoryV4Repo.ListRelationshipEdges(vars.RobotRuntime.RobotCode, chatRoomID, member.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询群友关系失败：%v\n", err)
		return
	}
	if len(edges) == 0 {
		return
	}
	sb.WriteString("和其他群友的关系：\n")
	for _, edge := range edges {
		fmt.Fprintf(sb, "- %s 与 %s：%s，关系强弱%d，%s\n",
			formatChatRoomMemberName(memberByWxID[edge.FromWxID]),
			formatChatRoomMemberName(memberByWxID[edge.ToWxID]),
			humanRelationType(edge.RelationType),
			edge.Strength,
			replaceWxIDs(edge.Summary, memberByWxID[edge.FromWxID], memberByWxID[edge.ToWxID]),
		)
	}
}

func (t *SearchChatRoomMemoryTool) renderV4RelationshipBetween(sb *strings.Builder, memoryV4Repo *repository.MemoryV4, chatRoomID string, firstMember, secondMember *model.ChatRoomMember) {
	if memoryV4Repo == nil || firstMember == nil || secondMember == nil {
		return
	}
	fmt.Fprintf(sb, "\n[我记得的两人关系：%s 与 %s]\n", formatChatRoomMemberName(firstMember), formatChatRoomMemberName(secondMember))
	edges, err := memoryV4Repo.ListRelationshipEdgesBetween(vars.RobotRuntime.RobotCode, chatRoomID, firstMember.WechatID, secondMember.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询两人关系失败：%v\n", err)
	} else if len(edges) == 0 {
		sb.WriteString("暂时没找到两人之间的明确关系记录\n")
	} else {
		for _, edge := range edges {
			fmt.Fprintf(sb, "- %s，%s，关系强弱%d：%s\n", humanRelationType(edge.RelationType), humanRelationDirection(string(edge.Direction)), edge.Strength, replaceWxIDs(edge.Summary, firstMember, secondMember))
		}
	}
	assertions, err := memoryV4Repo.ListRelationshipAssertionsBetween(vars.RobotRuntime.RobotCode, chatRoomID, firstMember.WechatID, secondMember.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询关系线索失败：%v\n", err)
		return
	}
	if len(assertions) == 0 {
		return
	}
	sb.WriteString("关系线索：\n")
	for _, assertion := range assertions {
		fmt.Fprintf(sb, "- %s，%s，%s：%s\n", humanRelationType(assertion.RelationType), humanRelationDirection(string(assertion.Direction)), humanConfidence(assertion.Confidence), replaceWxIDs(assertion.Summary, firstMember, secondMember))
	}
}

func (t *SearchChatRoomMemoryTool) renderMemberProfileAndMemories(sb *strings.Builder, memoryRepo *repository.Memory, chatRoomID string, member *model.ChatRoomMember) {
	fmt.Fprintf(sb, "\n[成员印象：%s]\n", formatChatRoomMemberName(member))
	profile, err := memoryRepo.GetMemberProfile(vars.RobotRuntime.RobotCode, chatRoomID, member.WechatID)
	if err != nil {
		fmt.Fprintf(sb, "查询成员印象失败：%v\n", err)
	} else if profile == nil {
		sb.WriteString("暂时没有这个人的稳定印象\n")
	} else {
		writeNonEmptyLine(sb, "整体印象", replaceWxIDs(profile.Summary, member))
		writeNonEmptyLine(sb, "性格特点", replaceWxIDs(profile.Personality, member))
		writeNonEmptyLine(sb, "沟通风格", replaceWxIDs(profile.CommunicationStyle, member))
		writeNonEmptyLine(sb, "对机器人的态度", replaceWxIDs(profile.AttitudeToBot, member))
		writeStringListLine(sb, "兴趣", parseJSONStrings(profile.Interests))
		writeStringListLine(sb, "常聊话题", parseJSONStrings(profile.FrequentTopics))
		fmt.Fprintf(sb, "%s\n", humanConfidence(profile.Confidence))
	}

	memories, err := memoryRepo.ListMemberMemories(vars.RobotRuntime.RobotCode, chatRoomID, member.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询成员长期记忆失败：%v\n", err)
		return
	}
	if len(memories) == 0 {
		return
	}
	sb.WriteString("其他成员记忆：\n")
	for _, memory := range memories {
		fmt.Fprintf(sb, "- %s（重要程度%d，%s）\n", replaceWxIDs(memory.Content, member), memory.Importance, humanConfidence(memory.Confidence))
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
		fmt.Fprintf(sb, "- %s，关系强弱%d：%s\n", humanRelationType(rel.RelationType), rel.Strength, replaceWxIDs(rel.Summary, memberByWxID[rel.FromWxID], memberByWxID[rel.ToWxID]))
	}
}

func (t *SearchChatRoomMemoryTool) renderRelationshipBetween(sb *strings.Builder, memoryRepo *repository.Memory, chatRoomID string, firstMember, secondMember *model.ChatRoomMember) {
	fmt.Fprintf(sb, "\n[两人关系：%s 与 %s]\n", formatChatRoomMemberName(firstMember), formatChatRoomMemberName(secondMember))
	relationships, err := memoryRepo.ListMemberRelationshipsBetween(vars.RobotRuntime.RobotCode, chatRoomID, firstMember.WechatID, secondMember.WechatID, 5)
	if err != nil {
		fmt.Fprintf(sb, "查询两人关系失败：%v\n", err)
	} else if len(relationships) == 0 {
		sb.WriteString("暂时没找到两人之间的明确关系记录\n")
	} else {
		for _, rel := range relationships {
			fmt.Fprintf(sb, "- 关系：%s，关系强弱%d：%s\n", humanRelationType(rel.RelationType), rel.Strength, replaceWxIDs(rel.Summary, firstMember, secondMember))
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
	sb.WriteString("其他关系记忆：\n")
	for _, memory := range memories {
		fmt.Fprintf(sb, "- %s（重要程度%d，%s）\n", replaceWxIDs(memory.Content, firstMember, secondMember), memory.Importance, humanConfidence(memory.Confidence))
	}
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

func humanConfidence(confidence int) string {
	return fmt.Sprintf("可信度%d", confidence)
}

func humanAliasType(aliasType string) string {
	switch strings.TrimSpace(aliasType) {
	case "wechat_id":
		return "微信ID"
	case "current_nickname":
		return "当前昵称"
	case "current_remark":
		return "当前群备注"
	case "wechat_alias":
		return "微信号"
	case "old_nickname":
		return "以前的昵称"
	case "old_remark":
		return "以前的群备注"
	case "old_wechat_alias":
		return "以前的微信号"
	case "observed_call_name":
		return "群里常用叫法"
	case "self_claimed":
		return "本人自称"
	default:
		if strings.TrimSpace(aliasType) == "" {
			return "名称"
		}
		return aliasType
	}
}

func humanFactPredicate(predicate string) string {
	switch strings.TrimSpace(predicate) {
	case "occupation":
		return "职业"
	case "interest":
		return "兴趣"
	case "dislike":
		return "不喜欢"
	case "preference":
		return "偏好"
	case "personality":
		return "性格"
	case "skill":
		return "技能"
	case "background":
		return "背景"
	case "location":
		return "地点"
	case "habit":
		return "习惯"
	case "other":
		return "其他"
	default:
		if strings.TrimSpace(predicate) == "" {
			return "信息"
		}
		return predicate
	}
}

func humanEventType(eventType string) string {
	switch strings.TrimSpace(eventType) {
	case "activity":
		return "动态"
	case "plan":
		return "计划"
	case "completed":
		return "已完成的事"
	case "location":
		return "去过的地方"
	case "work":
		return "工作相关"
	case "social":
		return "社交相关"
	case "other":
		return "其他"
	default:
		if strings.TrimSpace(eventType) == "" {
			return "动态"
		}
		return eventType
	}
}

func humanRelationType(relationType string) string {
	switch strings.TrimSpace(relationType) {
	case "friend":
		return "朋友"
	case "coworker":
		return "同事"
	case "helper":
		return "常互相帮忙"
	case "familiar":
		return "比较熟"
	case "conflict":
		return "有过冲突"
	case "joke_partner":
		return "经常开玩笑"
	case "mentor":
		return "指导或请教关系"
	case "other":
		return "其他关系"
	default:
		if strings.TrimSpace(relationType) == "" {
			return "关系"
		}
		return relationType
	}
}

func humanRelationDirection(direction string) string {
	switch strings.TrimSpace(direction) {
	case "directed":
		return "单向关系"
	case "undirected":
		return "不分方向"
	default:
		if strings.TrimSpace(direction) == "" {
			return "关系方向不明确"
		}
		return direction
	}
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
