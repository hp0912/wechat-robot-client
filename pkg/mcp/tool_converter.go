package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// MCPToolConverter MCP工具到OpenAI工具格式的转换器
type MCPToolConverter struct {
	manager *MCPManager
}

// NewMCPToolConverter 创建转换器
func NewMCPToolConverter(manager *MCPManager) *MCPToolConverter {
	return &MCPToolConverter{
		manager: manager,
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
func (c *MCPToolConverter) convertSingleTool(serverName string, mcpTool MCPTool) (openai.Tool, error) {
	// 为工具名称添加服务器前缀以避免冲突
	toolName := fmt.Sprintf("%s__%s", serverName, mcpTool.Name)

	// 转换inputSchema到OpenAI的参数格式
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
func (c *MCPToolConverter) convertInputSchemaToParameters(inputSchema map[string]any) (jsonschema.Definition, error) {
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
func (c *MCPToolConverter) ExecuteOpenAIToolCall(ctx context.Context, toolCall openai.ToolCall) (string, error) {
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

	// 构建MCP调用参数
	params := MCPCallToolParams{
		Name:      toolName,
		Arguments: args,
	}

	// 调用MCP工具
	result, err := c.manager.CallToolByName(ctx, serverName, params)
	if err != nil {
		return "", fmt.Errorf("failed to call mcp tool: %w", err)
	}

	// 转换结果为字符串
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

// formatToolResult 格式化工具结果
func (c *MCPToolConverter) formatToolResult(result *MCPCallToolResult) (string, error) {
	if result.IsError {
		return "", fmt.Errorf("tool execution error: %v", result.Content)
	}

	// 合并所有内容
	var output string
	for i, content := range result.Content {
		if i > 0 {
			output += "\n\n"
		}

		switch content.Type {
		case "text":
			output += content.Text
		case "image":
			// 图片内容可能需要特殊处理
			output += fmt.Sprintf("[Image: %v]", content.Data)
		case "resource":
			// 资源内容
			output += fmt.Sprintf("[Resource: %v]", content.Data)
		default:
			// 其他类型
			dataBytes, err := json.Marshal(content.Data)
			if err != nil {
				output += fmt.Sprintf("[Data: %v]", content.Data)
			} else {
				output += string(dataBytes)
			}
		}
	}

	return output, nil
}

// BuildSystemPromptWithMCPTools 构建包含MCP工具描述的系统提示词
func (c *MCPToolConverter) BuildSystemPromptWithMCPTools(ctx context.Context, basePrompt string) (string, error) {
	// 获取所有工具
	allTools, err := c.manager.GetAllTools(ctx)
	if err != nil {
		return basePrompt, err
	}

	if len(allTools) == 0 {
		return basePrompt, nil
	}

	// 构建工具描述
	toolsDesc := "\n\n## 可用工具\n\n你可以调用以下工具来帮助回答用户的问题：\n\n"

	for serverName, tools := range allTools {
		toolsDesc += fmt.Sprintf("### 来自 %s 的工具：\n\n", serverName)
		for _, tool := range tools {
			toolsDesc += fmt.Sprintf("- **%s**: %s\n", tool.Name, tool.Description)
		}
		toolsDesc += "\n"
	}

	toolsDesc += "请根据用户的需求，合理选择和使用这些工具。\n"

	return basePrompt + toolsDesc, nil
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
