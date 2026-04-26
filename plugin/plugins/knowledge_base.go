package plugins

import (
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type KnowledgeBasePlugin struct{}

func NewKnowledgeBasePlugin() plugin.MessageHandler {
	return &KnowledgeBasePlugin{}
}

func (p *KnowledgeBasePlugin) GetName() string {
	return "KnowledgeBase"
}

func (p *KnowledgeBasePlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *KnowledgeBasePlugin) Match(ctx *plugin.MessageContext) bool {
	return ctx.Message.IsChatRoom && ctx.ReferMessage != nil && strings.HasPrefix(ctx.MessageContent, "#录入知识库")
}

func (p *KnowledgeBasePlugin) PreAction(ctx *plugin.MessageContext) bool {
	chatRoomMember, err := ctx.MessageService.GetChatRoomMember(ctx.Message.FromWxID, ctx.Message.SenderWxID)
	if err != nil {
		log.Printf("获取群成员信息失败: %v", err)
		return false
	}
	if chatRoomMember == nil {
		log.Printf("群成员信息不存在: 群ID=%s, 成员微信ID=%s", ctx.Message.FromWxID, ctx.Message.SenderWxID)
		return false
	}
	if chatRoomMember.IsBlacklisted != nil && *chatRoomMember.IsBlacklisted {
		log.Printf("群成员[%s]在黑名单中，跳过AI回复", chatRoomMember.Nickname)
		return false
	}
	if chatRoomMember.IsAdmin == nil || !*chatRoomMember.IsAdmin {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "您配使用这个指令吗？", ctx.Message.SenderWxID)
		return false
	}
	return true
}

func (p *KnowledgeBasePlugin) PostAction(ctx *plugin.MessageContext) {
}

func (p *KnowledgeBasePlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}
	if vars.KnowledgeService == nil || vars.DB == nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "知识库服务未初始化", ctx.Message.SenderWxID)
		return
	}
	parts := strings.SplitN(ctx.MessageContent, " ", 2)
	if len(parts) < 2 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "格式: #录入知识库 知识库名称", ctx.Message.SenderWxID)
		return
	}
	knowledgeBaseName := strings.TrimSpace(parts[1])
	if knowledgeBaseName == "" {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "格式: #录入知识库 知识库名称", ctx.Message.SenderWxID)
		return
	}
	knowledgeConfigs := strings.SplitN(ctx.ReferMessage.Content, "--#--", 2)
	if len(knowledgeConfigs) < 2 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, `格式:
知识文档名称
--#--
知识文档内容
`, ctx.Message.SenderWxID)
		return
	}
	title := strings.TrimSpace(knowledgeConfigs[0])
	content := strings.TrimSpace(knowledgeConfigs[1])
	if title == "" || content == "" {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "知识文档名称和内容都不能为空", ctx.Message.SenderWxID)
		return
	}
	category, err := p.findKnowledgeCategory(ctx, knowledgeBaseName)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error(), ctx.Message.SenderWxID)
		return
	}
	if err := vars.KnowledgeService.AddDocument(ctx.Context, title, content, "manual", category.Code); err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("录入知识库失败: %v", err), ctx.Message.SenderWxID)
		return
	}
	ctx.MessageService.SendTextMessage(
		ctx.Message.FromWxID,
		fmt.Sprintf("已录入知识库[%s]\n文档: %s", category.Name, title),
		ctx.Message.SenderWxID,
	)
}

func (p *KnowledgeBasePlugin) findKnowledgeCategory(ctx *plugin.MessageContext, knowledgeBaseName string) (*model.KnowledgeCategory, error) {
	repo := repository.NewKnowledgeCategoryRepo(ctx.Context, vars.DB)
	categories, err := repo.List(model.KnowledgeCategoryTypeText)
	if err != nil {
		log.Printf("查询知识库分类失败: %v", err)
		return nil, fmt.Errorf("查询知识库失败，请稍后重试")
	}
	trimmedName := strings.TrimSpace(knowledgeBaseName)
	for _, category := range categories {
		if strings.EqualFold(category.Name, trimmedName) || strings.EqualFold(category.Code, trimmedName) {
			return category, nil
		}
	}
	return nil, fmt.Errorf("未找到知识库[%s]", knowledgeBaseName)
}
