package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
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
func (c *MCPToolConverter) convertSingleTool(serverName string, mcpTool *sdkmcp.Tool) (openai.Tool, error) {
	// 为工具名称添加服务器前缀以避免冲突
	toolName := fmt.Sprintf("%s__%s", serverName, mcpTool.Name)

	// 转换inputSchema到OpenAI的参数格式
	params, noParams, err := c.convertInputSchemaToParameters(mcpTool.InputSchema)
	if err != nil {
		return openai.Tool{}, fmt.Errorf("failed to convert input schema: %w", err)
	}

	fn := &openai.FunctionDefinition{
		Name:        toolName,
		Description: mcpTool.Description,
	}

	// 只有在不是“无参数工具”时才设置 Parameters
	if !noParams {
		fn.Parameters = params
	}

	return openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: fn,
	}, nil
}

// convertInputSchemaToParameters 转换InputSchema到OpenAI参数格式
func (c *MCPToolConverter) convertInputSchemaToParameters(inputSchema any) (jsonschema.Definition, bool, error) {
	if inputSchema == nil {
		return jsonschema.Definition{}, true, nil
	}

	schemaBytes, err := json.Marshal(inputSchema)
	if err != nil {
		return jsonschema.Definition{}, false, err
	}

	var params jsonschema.Definition
	if err := json.Unmarshal(schemaBytes, &params); err != nil {
		return jsonschema.Definition{}, false, err
	}

	// 如果解析后啥都没有（没有 type / properties / required），也视为无参数
	isEmpty := (params.Type == "" || params.Type == "object") &&
		len(params.Properties) == 0 &&
		len(params.Required) == 0
	if isEmpty {
		return jsonschema.Definition{}, true, nil
	}

	if params.Type == "" {
		params.Type = jsonschema.Object
	}
	if params.Type == jsonschema.Object && params.Properties == nil {
		params.Properties = make(map[string]jsonschema.Definition)
	}

	return params, false, nil
}
