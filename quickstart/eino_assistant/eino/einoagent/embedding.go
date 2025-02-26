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
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
)

// defaultArkEmbeddingConfig 创建默认的火山引擎Ark嵌入模型配置
// 这个函数从环境变量中读取模型名称和API密钥，创建一个默认的配置
// 在Eino Agent中，这个配置用于连接火山引擎的文本嵌入服务
//
// 环境变量:
//   - ARK_EMBEDDING_MODEL: 火山引擎Ark嵌入模型的名称/ID
//   - ARK_API_KEY: 访问火山引擎API的密钥
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//
// 返回:
//   - *ark.EmbeddingConfig: 默认的嵌入模型配置
//   - error: 错误信息，如果有的话
func defaultArkEmbeddingConfig(ctx context.Context) (*ark.EmbeddingConfig, error) {
	config := &ark.EmbeddingConfig{
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"), // 从环境变量获取嵌入模型名称
		APIKey: os.Getenv("ARK_API_KEY"),         // 从环境变量获取API密钥
	}
	return config, nil
}

// NewArkEmbedding 创建一个新的火山引擎Ark嵌入模型实例
// 这个函数根据提供的配置创建一个Embedder实例，用于将文本转换为向量表示
// 如果没有提供配置，则使用默认配置
//
// 在Eino Agent流程图中，这个函数用于创建嵌入模型实例，该实例被RedisRetriever使用
// 用于将用户查询转换为向量，以便在Redis中进行相似性搜索
//
// 参数:
//   - ctx: 上下文，用于传递请求范围的值
//   - config: 嵌入模型配置，如果为nil则使用默认配置
//
// 返回:
//   - eb: 创建的嵌入模型实例
//   - err: 错误信息，如果有的话
func NewArkEmbedding(ctx context.Context, config *ark.EmbeddingConfig) (eb embedding.Embedder, err error) {
	if config == nil {
		config, err = defaultArkEmbeddingConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	eb, err = ark.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
