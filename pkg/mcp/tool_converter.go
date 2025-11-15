package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPToolConverter MCP工具到OpenAI工具格式的转换器
type MCPToolConverter struct {
	manager       *MCPManager
	messageSender MessageSender
}

// NewMCPToolConverter 创建转换器
func NewMCPToolConverter(manager *MCPManager, messageSender MessageSender) *MCPToolConverter {
	return &MCPToolConverter{
		manager:       manager,
		messageSender: messageSender,
	}
}

// ConvertMCPToolsToOpenAI 将MCP工具转换为OpenAI工具格式
func (c *MCPToolConverter) ConvertMCPToolsToOpenAI(ctx context.Context) ([]openai.Tool, error) {
	// 获取所有MCP工具
	allTools, err := c.manager.GetAllTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mcp tools: %w", err)
	}

	var openaiTools []openai.Tool

	// 转换每个工具
	for serverName, tools := range allTools {
		for _, mcpTool := range tools {
			openaiTool, err := c.convertSingleTool(serverName, mcpTool)
			if err != nil {
				// 记录错误但继续转换其他工具
				fmt.Printf("Failed to convert tool %s from server %s: %v\n", mcpTool.Name, serverName, err)
				continue
			}
			openaiTools = append(openaiTools, openaiTool)
		}
	}

	return openaiTools, nil
}

// convertSingleTool 转换单个工具
func (c *MCPToolConverter) convertSingleTool(serverName string, mcpTool *sdkmcp.Tool) (openai.Tool, error) {
	// 为工具名称添加服务器前缀以避免冲突
	toolName := fmt.Sprintf("%s__%s", serverName, mcpTool.Name)

	// 转换inputSchema到OpenAI的参数格式
	// mcpTool.InputSchema 是 any，通常为 map[string]any
	parameters, err := c.convertInputSchemaToParameters(mcpTool.InputSchema)
	if err != nil {
		return openai.Tool{}, fmt.Errorf("failed to convert input schema: %w", err)
	}

	// 构建OpenAI工具
	openaiTool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        toolName,
			Description: mcpTool.Description,
			Parameters:  parameters,
		},
	}

	return openaiTool, nil
}

// convertInputSchemaToParameters 转换InputSchema到OpenAI参数格式
func (c *MCPToolConverter) convertInputSchemaToParameters(inputSchema any) (jsonschema.Definition, error) {
	// 将map转换为JSON字符串再解析为jsonschema.Definition
	schemaBytes, err := json.Marshal(inputSchema)
	if err != nil {
		return jsonschema.Definition{}, err
	}

	var params jsonschema.Definition
	if err := json.Unmarshal(schemaBytes, &params); err != nil {
		return jsonschema.Definition{}, err
	}

	// 确保type为object
	if params.Type == "" {
		params.Type = jsonschema.Object
	}

	return params, nil
}

// ExecuteOpenAIToolCall 执行OpenAI函数调用
func (c *MCPToolConverter) ExecuteOpenAIToolCall(ctx context.Context, robotCtx RobotContext, toolCall openai.ToolCall) (string, error) {
	// 解析工具名称，提取服务器名称和原始工具名称
	serverName, toolName, err := c.parseToolName(toolCall.Function.Name)
	if err != nil {
		return "", err
	}

	// 解析参数
	var args map[string]any
	if toolCall.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	}

	// 将 RobotContext 转换为 Meta（map[string]any）
	metaBytes, err := json.Marshal(robotCtx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal robot context: %w", err)
	}
	var meta map[string]any
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return "", fmt.Errorf("failed to unmarshal robot context to meta: %w", err)
	}

	// 构建MCP调用参数
	params := &sdkmcp.CallToolParams{Meta: meta, Name: toolName, Arguments: args}

	// 调用MCP工具
	result, err := c.manager.CallToolByName(ctx, serverName, params)
	if err != nil {
		return "", fmt.Errorf("failed to call mcp tool: %w", err)
	}

	return c.formatToolResult(result)
}

// parseToolName 解析工具名称
func (c *MCPToolConverter) parseToolName(fullName string) (serverName, toolName string, err error) {
	// 工具名称格式：serverName__toolName
	for _, client := range c.manager.GetAllClients() {
		prefix := client.GetConfig().Name + "__"
		if len(fullName) > len(prefix) && fullName[:len(prefix)] == prefix {
			return client.GetConfig().Name, fullName[len(prefix):], nil
		}
	}

	return "", "", fmt.Errorf("invalid tool name format: %s", fullName)
}

func (c *MCPToolConverter) formatToolResult(result *sdkmcp.CallToolResult) (string, error) {
	if result.IsError {
		if len(result.Content) > 0 {
			var errmsgs []string
			for _, content := range result.Content {
				if textContent, ok := content.(*sdkmcp.TextContent); ok {
					if textContent.Text != "" {
						errmsgs = append(errmsgs, textContent.Text)
					}
				}
			}
			if len(errmsgs) > 0 {
				return "", fmt.Errorf("MCP 调用失败: %s", strings.Join(errmsgs, "\n"))
			}
		}
		return "", fmt.Errorf("MCP 调用失败")
	}
	// 直接将结果序列化为字符串返回，交由上层决定发送策略
	b, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// BuildSystemPromptWithMCPTools 构建包含MCP工具描述的系统提示词
func (c *MCPToolConverter) BuildSystemPromptWithMCPTools(ctx context.Context, basePrompt string) (string, error) {
	allTools, err := c.manager.GetAllTools(ctx)
	if err != nil {
		return basePrompt, err
	}

	if len(allTools) == 0 {
		return basePrompt, nil
	}

	intro := `你运行在一个支持 MCP（Model Context Protocol）工具的聊天应用环境中。
当你自身能力不足或需要访问外部数据时，应主动调用这些工具来完成任务。

1. 何时使用工具
- 当需要访问或处理「外部数据」时（例如：查询、统计、搜索、总结群聊内容等）。
- 当用户要求执行你自身无法完成的动作（例如：根据文本生成图片/音频/视频、操作第三方系统等）。
- 当用户明确点名某个工具或某类能力时。
- 当任务涉及具体时间区间、对象范围、过滤条件等，并且需要依赖工具才能得到准确结果时。

2. 工具命名与选择
- 每个工具在调用时的名称格式为：{serverName}__{toolName}。
  例如：服务器名为 "calendar"，工具名为 "create_event"，实际调用名为 "calendar__create_event"。
- 在选择工具时：
  - 先根据工具描述判断是否符合用户意图；
  - 如有多个类似工具，优先选择描述更精确、与当前场景更贴近的工具。

3. 构造工具调用参数
- 在调用工具前，应先向用户澄清目标、范围和约束（如时间区间、数量限制、过滤条件等）。
- 构造参数时：
  - 必须严格遵守工具参数的 JSON Schema 要求；
  - 不要省略必填字段，不要编造关键参数；
  - 对不确定的信息应先向用户确认，再发起调用。

4. 处理工具返回结果
- 工具可能返回结构化 JSON 或文本，你需要先理解其含义，再用自然语言为用户总结。
- 若工具返回错误或空结果：
  - 根据返回信息解释可能原因，不要编造结果；
  - 必要时建议用户调整请求或参数。
- 最终回答时：
  - 使用简洁、结构化的方式呈现结果（例如分点说明、列表）；
  - 除非用户明确要求，否则不要直接原样输出冗长的 JSON。

下面是你当前可以使用的 MCP 工具列表，请在需要时主动选择合适的工具进行调用：
`

	toolsDesc := "\n\n## 可用工具\n\n你可以调用以下工具来帮助回答用户的问题：\n\n"

	for serverName, tools := range allTools {
		toolsDesc += fmt.Sprintf("### 来自 %s 的工具：\n\n", serverName)
		for _, tool := range tools {
			toolsDesc += fmt.Sprintf("- **%s**: %s\n", tool.Name, tool.Description)
		}
		toolsDesc += "\n"
	}

	toolsDesc += "调用工具时，请根据上述规则谨慎选择工具并构造参数。\n"

	return basePrompt + intro + toolsDesc, nil
}

// GetToolsByServer 获取指定服务器的所有工具（OpenAI格式）
func (c *MCPToolConverter) GetToolsByServer(ctx context.Context, serverName string) ([]openai.Tool, error) {
	client, err := c.manager.GetClientByName(serverName)
	if err != nil {
		return nil, err
	}

	mcpTools, err := client.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	var openaiTools []openai.Tool
	for _, mcpTool := range mcpTools {
		openaiTool, err := c.convertSingleTool(serverName, mcpTool)
		if err != nil {
			fmt.Printf("Failed to convert tool %s: %v\n", mcpTool.Name, err)
			continue
		}
		openaiTools = append(openaiTools, openaiTool)
	}

	return openaiTools, nil
}
