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

// knowledgeindexing 包负责知识库索引的构建和管理
package knowledgeindexing

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"                   // 文件加载器组件
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown" // Markdown分割器组件
	"github.com/cloudwego/eino-ext/components/indexer/redis"                          // Redis索引器组件
	"github.com/cloudwego/eino/components/document"                                   // 文档处理核心组件
	"github.com/cloudwego/eino/compose"                                               // 组件编排工具
)

// KnowledgeIndexingBuildConfig 定义了知识索引构建所需的各组件配置
type KnowledgeIndexingBuildConfig struct {
	FileLoaderKeyOfLoader                    *file.FileLoaderConfig // 文件加载器配置
	MarkdownSplitterKeyOfDocumentTransformer *markdown.HeaderConfig // Markdown分割器配置
	RedisIndexerKeyOfIndexer                 *redis.IndexerConfig   // Redis索引器配置
}

// BuildConfig 是整体构建配置的容器
type BuildConfig struct {
	KnowledgeIndexing *KnowledgeIndexingBuildConfig // 知识索引构建配置
}

// BuildKnowledgeIndexing 构建知识索引处理流程
// 参数:
//   - ctx: 上下文
//   - config: 构建配置
//
// 返回:
//   - compose.Runnable: 可执行的处理流程
//   - error: 错误信息
func BuildKnowledgeIndexing(ctx context.Context, config *BuildConfig) (r compose.Runnable[document.Source, []string], err error) {
	// 定义处理节点的名称常量
	const (
		FileLoader       = "FileLoader"       // 文件加载器节点名称
		MarkdownSplitter = "MarkdownSplitter" // Markdown分割器节点名称
		RedisIndexer     = "RedisIndexer"     // Redis索引器节点名称
	)

	// 创建新的处理图，从文档源到字符串数组的转换流程
	g := compose.NewGraph[document.Source, []string]()

	// 初始化文件加载器并添加到图中
	fileLoaderKeyOfLoader, err := NewFileLoader(ctx, config.KnowledgeIndexing.FileLoaderKeyOfLoader)
	if err != nil {
		return nil, err
	}
	_ = g.AddLoaderNode(FileLoader, fileLoaderKeyOfLoader)

	// 初始化Markdown分割器并添加到图中
	markdownSplitterKeyOfDocumentTransformer, err := NewMarkdownSplitter(ctx, config.KnowledgeIndexing.MarkdownSplitterKeyOfDocumentTransformer)
	if err != nil {
		return nil, err
	}
	_ = g.AddDocumentTransformerNode(MarkdownSplitter, markdownSplitterKeyOfDocumentTransformer)

	// 初始化Redis索引器并添加到图中
	redisIndexerKeyOfIndexer, err := NewRedisIndexer(ctx, config.KnowledgeIndexing.RedisIndexerKeyOfIndexer)
	if err != nil {
		return nil, err
	}
	_ = g.AddIndexerNode(RedisIndexer, redisIndexerKeyOfIndexer)

	// 构建处理流程的边，定义数据流向
	_ = g.AddEdge(compose.START, FileLoader)      // 从起点到文件加载器
	_ = g.AddEdge(RedisIndexer, compose.END)      // 从Redis索引器到终点
	_ = g.AddEdge(FileLoader, MarkdownSplitter)   // 从文件加载器到Markdown分割器
	_ = g.AddEdge(MarkdownSplitter, RedisIndexer) // 从Markdown分割器到Redis索引器

	// 编译处理图，生成可执行的处理流程
	// WithGraphName: 设置图的名称为"KnowledgeIndexing"
	// WithNodeTriggerMode: 设置节点触发模式为任一前置节点完成即触发
	r, err = g.Compile(ctx, compose.WithGraphName("KnowledgeIndexing"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}

	return r, err
}
