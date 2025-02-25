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

	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// EinoAgentBuildConfig 定义了构建Eino Agent所需的各组件配置
type EinoAgentBuildConfig struct {
	// 聊天模板配置，用于格式化对话内容
	ChatTemplateKeyOfChatTemplate *ChatTemplateConfig
	// ReAct Agent配置，用于实现推理和行动的智能体
	ReactAgentKeyOfLambda *react.AgentConfig
	// Redis检索器配置，用于从知识库中检索相关文档
	RedisRetrieverKeyOfRetriever *redis.RetrieverConfig
}

// BuildConfig 是整体构建配置的包装结构
type BuildConfig struct {
	EinoAgent *EinoAgentBuildConfig
}

// BuildEinoAgent 构建一个完整的Eino Agent流程图
// 该方法创建并连接各个组件节点，形成一个完整的处理流程
// 输入为用户消息(UserMessage)，输出为模型响应消息(schema.Message)
func BuildEinoAgent(ctx context.Context, config *BuildConfig) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	// 定义流程图中各节点的名称常量
	const (
		InputToQuery   = "InputToQuery"   // 将用户输入转换为查询
		ChatTemplate   = "ChatTemplate"   // 聊天模板处理
		ReactAgent     = "ReactAgent"     // ReAct智能体
		RedisRetriever = "RedisRetriever" // Redis检索器
		InputToHistory = "InputToHistory" // 将用户输入添加到历史记录
	)

	// 创建一个新的计算图，输入类型为UserMessage，输出类型为schema.Message
	g := compose.NewGraph[*UserMessage, *schema.Message]()

	// 添加InputToQuery节点：将用户消息转换为查询字符串
	// 这个节点负责提取用户消息中的查询内容，为后续的检索做准备
	_ = g.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(NewInputToQuery),
		compose.WithNodeName("UserMessageToQuery"))

	// 创建并添加ChatTemplate节点：处理聊天模板
	// 这个节点负责格式化对话内容，包括系统提示、用户输入和检索到的文档
	chatTemplateKeyOfChatTemplate, err := NewChatTemplate(ctx, config.EinoAgent.ChatTemplateKeyOfChatTemplate)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplateKeyOfChatTemplate)

	// 创建并添加ReactAgent节点：实现ReAct智能体
	// 这个节点是核心智能体，负责理解用户意图并生成响应
	reactAgentKeyOfLambda, err := NewReactAgent(ctx, config.EinoAgent.ReactAgentKeyOfLambda)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(ReactAgent, reactAgentKeyOfLambda, compose.WithNodeName("ReAct Agent"))

	// 创建并添加RedisRetriever节点：从Redis中检索相关文档
	// 这个节点负责根据用户查询，从知识库中检索最相关的文档
	redisRetrieverKeyOfRetriever, err := NewRedisRetriever(ctx, config.EinoAgent.RedisRetrieverKeyOfRetriever)
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(RedisRetriever, redisRetrieverKeyOfRetriever, compose.WithOutputKey("documents"))

	// 添加InputToHistory节点：将用户消息添加到对话历史
	// 这个节点负责维护对话上下文，记录用户的输入历史
	_ = g.AddLambdaNode(InputToHistory, compose.InvokableLambdaWithOption(NewInputToHistory),
		compose.WithNodeName("UserMessageToVariables"))

	// 添加边连接各节点，构建完整的处理流程图

	// 从起点连接到InputToQuery和InputToHistory
	_ = g.AddEdge(compose.START, InputToQuery)
	_ = g.AddEdge(compose.START, InputToHistory)

	// 将ReactAgent连接到终点，表示处理完成
	_ = g.AddEdge(ReactAgent, compose.END)

	// 构建主要处理流程：查询->检索->模板->智能体
	_ = g.AddEdge(InputToQuery, RedisRetriever) // 查询传递给检索器
	_ = g.AddEdge(RedisRetriever, ChatTemplate) // 检索结果传递给聊天模板
	_ = g.AddEdge(InputToHistory, ChatTemplate) // 历史记录传递给聊天模板
	_ = g.AddEdge(ChatTemplate, ReactAgent)     // 格式化后的内容传递给智能体

	// 编译计算图，设置图名称为"EinoAgent"，触发模式为所有前置节点完成后触发
	r, err = g.Compile(ctx, compose.WithGraphName("EinoAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
