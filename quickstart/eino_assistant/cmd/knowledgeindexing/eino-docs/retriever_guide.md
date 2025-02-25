# Eino: Retriever 使用说明

## 1. 基本介绍
Retriever 是 Eino 框架中的一个组件，用于从各种数据源检索文档。它根据用户的查询（query）从文档库中检索出最相关的文档，适用于以下场景：
- 基于向量相似度的文档检索。
- 基于关键词的文档搜索。
- 知识库问答系统（RAG）。

## 2. 组件定义

### 2.1 接口定义
Retriever 组件的核心接口定义如下：
```go
type Retriever interface {
    Retrieve(ctx context.Context, query string, opts ...Option) ([]*schema.Document, error)
}
```
- **功能**：根据查询检索相关文档。
- **参数**：
  - `ctx`：上下文对象，用于传递请求级别的信息和 Callback Manager。
  - `query`：查询字符串。
  - `opts`：检索选项，用于配置检索行为。
- **返回值**：
  - `[]*schema.Document`：检索到的文档列表。
  - `error`：检索过程中的错误信息。

### 2.2 Document 结构体
检索返回的文档结构体定义如下：
```go
type Document struct {
    ID        string
    Content   string
    MetaData  map[string]any
}
```
- `ID`：文档的唯一标识符。
- `Content`：文档的内容。
- `MetaData`：用于存储文档的元数据信息。

### 2.3 公共 Option
Retriever 组件使用 `RetrieverOption` 来定义可选参数，包括以下公共选项：
```go
type Options struct {
    Index           *string
    SubIndex        *string
    TopK            *int
    ScoreThreshold  *float64
    Embedding       embedding.Embedder
    DSLInfo         map[string]interface{}
}
```
- `Index`：检索器使用的索引。
- `SubIndex`：检索器使用的子索引。
- `TopK`：检索的文档数量上限。
- `ScoreThreshold`：文档相似度的阈值。
- `Embedding`：用于生成查询向量的组件。
- `DSLInfo`：用于检索的 DSL 信息（仅在 VikingDB 类型的检索器中使用）。

公共 Option 的设置方法：
```go
WithIndex(index string) Option
WithSubIndex(subIndex string) Option
WithTopK(topK int) Option
WithScoreThreshold(threshold float64) Option
WithEmbedding(emb embedding.Embedder) Option
WithDSLInfo(dsl map[string]any) Option
```

## 3. 使用方式

### 3.1 单独使用
Retriever 可以单独使用，示例代码如下：
```go
import (
    "github.com/cloudwego/eino/components/retriever"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino-ext/components/retriever/volc_vikingdb"
)

collectionName := "eino_test"
indexName := "test_index_1"

cfg := &volc_vikingdb.RetrieverConfig{
    Host:              "api-vikingdb.volces.com",
    Region:            "cn-beijing",
    AK:                ak,
    SK:                sk,
    Scheme:            "https",
    ConnectionTimeout: 0,
    Collection:        collectionName,
    Index:             indexName,
    EmbeddingConfig: volc_vikingdb.EmbeddingConfig{
        UseBuiltin:  true,
        ModelName:   "bge-m3",
        UseSparse:   true,
        DenseWeight: 0.4,
    },
    Partition:      "",
    TopK:           of(10),
    ScoreThreshold: of(0.1),
    FilterDSL:      nil,
}

volcRetriever, _ := volc_vikingdb.NewRetriever(ctx, cfg)

query := "tourist attraction"
docs, _ := volcRetriever.Retrieve(ctx, query)

log.Printf("vikingDB retrieve success, query=%v, docs=%v", query, docs)
```

### 3.2 在编排中使用
Retriever 也可以在 Eino 的 Chain 或 Graph 编排中使用：
```go
// 在 Chain 中使用
chain := compose.NewChain[string, []*schema.Document]()
chain.AppendRetriever(retriever)

// 在 Graph 中使用
graph := compose.NewGraph[string, []*schema.Document]()
graph.AddRetrieverNode("retriever_node", retriever)
```

## 4. Option 和 Callback 使用

### 4.1 Callback 使用示例
Retriever 支持在检索过程中触发回调，示例代码如下：
```go
import (
    "github.com/cloudwego/eino/callbacks"
    "github.com/cloudwego/eino/components/retriever"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino/utils/callbacks"
    "github.com/cloudwego/eino-ext/components/retriever/volc_vikingdb"
)

handler := &callbacks.RetrieverCallbackHandler{
    OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *retriever.CallbackInput) context.Context {
        log.Printf("input access, content: %s\n", input.Query)
        return ctx
    },
    OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *retriever.CallbackOutput) context.Context {
        log.Printf("output finished, len: %v\n", len(output.Docs))
        return ctx
    },
}

helper := callbacks.NewHandlerHelper().Retriever(handler).Handler()

chain := compose.NewChain[string, []*schema.Document]()
chain.AppendRetriever(volcRetriever)

run, _ := chain.Compile(ctx)

outDocs, _ := run.Invoke(ctx, query, compose.WithCallbacks(helper))

log.Printf("vikingDB retrieve success, query=%v, docs=%v", query, outDocs)
```

## 5. 已有实现
目前 Eino 提供了以下 Retriever 的实现：
- **Volc VikingDB Retriever**：基于火山引擎 VikingDB 的检索实现，详情见 [Retriever - VikingDB](https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/retriever/retriever_volc_vikingdb)。

## 6. 自行实现参考

### 6.1 Option 机制
实现自定义 Retriever 时，需要正确处理公共 Option：
```go
func (r *MyRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
    options := &retriever.Options{
        Index:      &r.index,
        TopK:       &r.topK,
        Embedding:  r.embedder,
    }
    options = retriever.GetCommonOptions(options, opts...)
    // ...
}
```

### 6.2 Callback 处理
Retriever 需要在适当的时机触发回调，回调输入输出结构体定义如下：
```go
type CallbackInput struct {
    Query          string
    TopK           int
    Filter         string
    ScoreThreshold *float64
    Extra          map[string]any
}

type CallbackOutput struct {
    Docs []*schema.Document
    Extra map[string]any
}
```

### 6.3 完整实现示例
以下是一个自定义 Retriever 的完整实现示例：
```go
type MyRetriever struct {
    embedder embedding.Embedder
    index    string
    topK     int
}

func NewMyRetriever(config *MyRetrieverConfig) (*MyRetriever, error) {
    return &MyRetriever{
        embedder: config.Embedder,
        index:    config.Index,
        topK:     config.DefaultTopK,
    }, nil
}

func (r *MyRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
    options := &retriever.Options{
        Index:      &r.index,
        TopK:       &r.topK,
        Embedding:  r.embedder,
    }
    options = retriever.GetCommonOptions(options, opts...)

    cm := callbacks.ManagerFromContext(ctx)
    ctx = cm.OnStart(ctx, info, &retriever.CallbackInput{
        Query: query,
        TopK:  *options.TopK,
    })

    docs, err := r.doRetrieve(ctx, query, options)
    if err != nil {
        ctx = cm.OnError(ctx, info, err)
        return nil, err
    }

    ctx = cm.OnEnd(ctx, info, &retriever.CallbackOutput{
        Docs: docs,
    })

    return docs, nil
}

func (r *MyRetriever) doRetrieve(ctx context.Context, query string, opts *retriever.Options) ([]*schema.Document, error) {
    var queryVector []float64
    if opts.Embedding != nil {
        vectors, err := opts.Embedding.Embed