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

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// defaultReactAgentConfig 创建默认的ReAct智能体配置
// ReAct (Reasoning and Acting) 是一种结合推理和行动的智能体框架
// 这个函数设置智能体的最大步骤数、工具配置和使用的大语言模型
//
// 在Eino Agent流程图中，这个配置用于创建ReactAgent节点，该节点是核心智能体
// 负责理解用户意图、推理并生成响应
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//
// 返回:
//   - *react.AgentConfig: 默认的ReAct智能体配置
//   - error: 错误信息，如果有的话
func defaultReactAgentConfig(ctx context.Context) (*react.AgentConfig, error) {
	config := &react.AgentConfig{
		MaxStep:            25,                    // 最大执行步骤数，防止无限循环
		ToolReturnDirectly: map[string]struct{}{}, // 设置哪些工具的结果可以直接返回
	}

	// 创建并设置聊天模型
	chatModelCfg11, err := defaultArkChatModelConfig(ctx)
	if err != nil {
		return nil, err
	}
	chatModelIns11, err := NewArkChatModel(ctx, chatModelCfg11)
	if err != nil {
		return nil, err
	}
	config.Model = chatModelIns11

	// 获取并设置可用工具
	tools, err := GetTools(ctx)
	if err != nil {
		return nil, err
	}

	config.ToolsConfig.Tools = tools
	return config, nil
}

// NewReactAgent 创建一个新的ReAct智能体
// 这个函数根据提供的配置创建一个ReAct智能体实例，并将其包装为Lambda函数
// 如果没有提供配置，则使用默认配置
//
// 在Eino Agent流程图中，这个函数用于创建ReactAgent节点
// 该节点接收格式化后的提示，进行推理并生成最终响应
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//   - config: ReAct智能体配置，如果为nil则使用默认配置
//
// 返回:
//   - lba: 包装了ReAct智能体的Lambda函数
//   - err: 错误信息，如果有的话
func NewReactAgent(ctx context.Context, config *react.AgentConfig) (lba *compose.Lambda, err error) {
	if config == nil {
		config, err = defaultReactAgentConfig(ctx)
		if err != nil {
			return nil, err
		}
	}

	// 创建ReAct智能体实例
	ins, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, err
	}

	// 将智能体包装为Lambda函数
	// 参数分别是：生成函数、流式处理函数、批处理函数、取消函数
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lba, nil
}
