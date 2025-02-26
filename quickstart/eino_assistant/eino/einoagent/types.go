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

import "github.com/cloudwego/eino/schema"

// UserMessage 表示用户发送的消息
// 这个结构体是Eino Agent的输入，包含了用户查询和对话历史
// 在整个处理流程中，UserMessage会被转换为查询字符串用于检索，
// 同时其中的历史记录会被用于维护对话上下文
type UserMessage struct {
	// ID 是消息的唯一标识符
	// 用于跟踪和关联消息，特别是在异步处理场景中
	ID string `json:"id"`

	// Query 是用户的当前查询文本
	// 这是用户输入的实际内容，将被用于知识库检索和生成回复
	Query string `json:"query"`

	// History 是之前的对话历史记录
	// 包含了用户和系统之间的历史消息，用于维护对话上下文
	// 使用schema.Message类型，可以包含角色、内容等信息
	History []*schema.Message `json:"history"`
}
