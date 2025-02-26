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
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

// defaultArkChatModelConfig 创建默认的火山引擎Ark聊天模型配置
// 这个函数从环境变量中读取模型名称和API密钥，创建一个默认的配置
// 在Eino Agent中，这个配置用于连接火山引擎的大语言模型服务
//
// 环境变量:
//   - ARK_CHAT_MODEL: 火山引擎Ark聊天模型的名称/ID
//   - ARK_API_KEY: 访问火山引擎API的密钥
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//
// 返回:
//   - *ark.ChatModelConfig: 默认的聊天模型配置
//   - error: 错误信息，如果有的话
func defaultArkChatModelConfig(ctx context.Context) (*ark.ChatModelConfig, error) {
	log.Printf("model: %s", os.Getenv("ARK_CHAT_MODEL"))

	config := &ark.ChatModelConfig{
		Model:  os.Getenv("ARK_CHAT_MODEL"), // 从环境变量获取模型名称
		APIKey: os.Getenv("ARK_API_KEY"),    // 从环境变量获取API密钥
	}
	return config, nil
}

// NewArkChatModel 创建一个新的火山引擎Ark聊天模型实例
// 这个函数根据提供的配置创建一个ChatModel实例，用于与大语言模型服务交互
// 如果没有提供配置，则使用默认配置
//
// 在Eino Agent流程图中，这个函数用于创建聊天模型实例，该实例被ReactAgent使用
// 用于生成对用户查询的响应
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//   - config: 聊天模型配置，如果为nil则使用默认配置
//
// 返回:
//   - cm: 创建的聊天模型实例
//   - err: 错误信息，如果有的话
func NewArkChatModel(ctx context.Context, config *ark.ChatModelConfig) (cm model.ChatModel, err error) {
	if config == nil {
		config, err = defaultArkChatModelConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
