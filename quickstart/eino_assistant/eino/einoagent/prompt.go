/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package einoagent

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// ChatTemplateConfig 定义了聊天模板的配置
// 聊天模板用于格式化发送给大模型的提示信息，包括系统提示、历史记录和用户查询
// 在Eino Agent流程图中，这个配置用于创建ChatTemplate节点
type ChatTemplateConfig struct {
	// FormatType 指定模板的格式类型
	// 例如：字符串格式(FString)、JSON格式(FJson)等
	FormatType schema.FormatType

	// Templates 是一组消息模板
	// 定义了发送给大模型的消息结构，包括系统消息、用户消息等
	Templates []schema.MessagesTemplate
}

// systemPrompt 是系统提示模板
// 这是发送给大模型的系统指令，定义了AI助手的角色、能力和行为准则
// 模板中的{date}和{documents}是占位符，会在运行时被实际值替换：
// - {date}: 当前日期时间
// - {documents}: 从知识库检索到的相关文档内容
var systemPrompt = `
# Role: Eino Expert Assistant

Always in 中文

## Core Competencies
- knowledge of Eino framework and ecosystem
- Project scaffolding and best practices consultation
- Documentation navigation and implementation guidance
- Search web, clone github repo, open file/url, task management

## Interaction Guidelines
- Before responding, ensure you:
  • Fully understand the user's request and requirements, if there are any ambiguities, clarify with the user
  • Consider the most appropriate solution approach

- When providing assistance:
  • Be clear and concise
  • Include practical examples when relevant
  • Reference documentation when helpful
  • Suggest improvements or next steps if applicable

- If a request exceeds your capabilities:
  • Clearly communicate your limitations, suggest alternative approaches if possible

- If the question is compound or complex, you need to think step by step, avoiding giving low-quality answers directly.

## Context Information
- Current Date: {date}
- Related Documents: |-
==== doc start ====
  {documents}
==== doc end ====
`

// defaultPromptTemplateConfig 创建默认的聊天模板配置
// 如果用户没有提供自定义配置，则使用这个默认配置
//
// 默认配置包括：
// 1. 系统提示消息：定义AI助手的角色和行为
// 2. 历史消息占位符：插入对话历史记录
// 3. 用户消息：插入当前用户查询
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//
// 返回:
//   - *ChatTemplateConfig: 默认的聊天模板配置
//   - error: 错误信息，如果有的话
func defaultPromptTemplateConfig(ctx context.Context) (*ChatTemplateConfig, error) {
	config := &ChatTemplateConfig{
		FormatType: schema.FString, // 使用字符串格式
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),          // 系统提示消息
			schema.MessagesPlaceholder("history", true), // 历史消息占位符，true表示展开消息
			schema.UserMessage("{content}"),             // 用户消息，{content}会被替换为实际查询内容
		},
	}
	return config, nil
}

// NewChatTemplate 创建一个新的聊天模板
// 这个函数根据提供的配置创建一个ChatTemplate实例
// 如果没有提供配置，则使用默认配置
//
// 在Eino Agent流程图中，这个函数用于创建ChatTemplate节点
// 该节点负责将用户查询、历史记录和检索到的文档组合成完整的提示
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//   - config: 聊天模板配置，如果为nil则使用默认配置
//
// 返回:
//   - ct: 创建的聊天模板实例
//   - err: 错误信息，如果有的话
func NewChatTemplate(ctx context.Context, config *ChatTemplateConfig) (ct prompt.ChatTemplate, err error) {
	if config == nil {
		config, err = defaultPromptTemplateConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	ct = prompt.FromMessages(config.FormatType, config.Templates...)
	return ct, nil
}
