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
	"time"
)

// NewInputToQuery 将用户消息转换为查询字符串
// 这个函数非常简单，它从UserMessage结构体中提取Query字段并返回
// 在Eino Agent流程图中，这个函数用于InputToQuery节点，将用户输入转换为可用于检索的查询
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//   - input: 用户消息，包含查询内容和历史记录
//   - opts: 可选参数，当前未使用
//
// 返回:
//   - output: 提取的查询字符串
//   - err: 错误信息，如果有的话
func NewInputToQuery(ctx context.Context, input *UserMessage, opts ...any) (output string, err error) {
	return input.Query, nil
}

// NewInputToHistory 将用户消息转换为包含历史记录的映射
// 这个函数创建一个包含用户查询、历史记录和当前时间的映射
// 在Eino Agent流程图中，这个函数用于InputToHistory节点，为聊天模板提供上下文信息
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//   - input: 用户消息，包含查询内容和历史记录
//   - opts: 可选参数，当前未使用
//
// 返回:
//   - output: 包含查询内容、历史记录和时间的映射
//   - err: 错误信息，如果有的话
func NewInputToHistory(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"content": input.Query,                              // 用户当前的查询内容
		"history": input.History,                            // 用户的历史消息记录
		"date":    time.Now().Format("2006-01-02 15:04:05"), // 当前时间，格式化为易读形式
	}, nil
}
