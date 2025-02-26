# PGVector 模块

本模块提供了使用PostgreSQL的pgvector扩展进行向量存储和相似性搜索的功能。pgvector是一个PostgreSQL的扩展，支持向量相似性搜索，适用于存储和检索嵌入向量。

## 功能特点

- 向量数据的存储和检索
- 基于余弦相似度的向量搜索
- 支持批量插入和更新操作
- 支持元数据存储和检索

## 前置条件

使用本模块前，需要确保：

1. 已安装PostgreSQL数据库（推荐版本13或更高）
2. 已安装pgvector扩展（可通过`CREATE EXTENSION vector;`命令安装）

## 使用方法

### 初始化

```go
import (
    "context"
    "github.com/cloudwego/eino/quickstart/eino_assistant/pkg/pgvector"
)

// 使用默认配置初始化
err := pgvector.Init()
if err != nil {
    // 处理错误
}

// 或者使用自定义配置初始化
config := &pgvector.Config{
    Host:      "localhost",
    Port:      5432,
    User:      "postgres",
    Password:  "postgres",
    DBName:    "vectordb",
    SSLMode:   "disable",
    Dimension: 1536,  // 向量维度
    TableName: "my_vectors",
}

err = pgvector.InitPGVectorDB(context.Background(), config)
if err != nil {
    // 处理错误
}
```

### 创建客户端

```go
client, err := pgvector.NewClient(context.Background(), config)
if err != nil {
    // 处理错误
}
defer client.Close()
```

### 存储文档

```go
// 单个文档存储
doc := &pgvector.Document{
    ID:        "doc1",
    Content:   "这是一个示例文档",
    Metadata:  map[string]interface{}{"source": "example", "category": "test"},
    Embedding: []float32{0.1, 0.2, 0.3, ...}, // 1536维向量
}

err = client.UpsertDocument(context.Background(), doc)
if err != nil {
    // 处理错误
}

// 批量文档存储
docs := []*pgvector.Document{
    {
        ID:        "doc2",
        Content:   "第二个示例文档",
        Metadata:  map[string]interface{}{"source": "example", "category": "test"},
        Embedding: []float32{0.2, 0.3, 0.4, ...},
    },
    {
        ID:        "doc3",
        Content:   "第三个示例文档",
        Metadata:  map[string]interface{}{"source": "example", "category": "production"},
        Embedding: []float32{0.3, 0.4, 0.5, ...},
    },
}

err = client.BatchUpsertDocuments(context.Background(), docs)
if err != nil {
    // 处理错误
}
```

### 相似性搜索

```go
// 搜索与给定向量最相似的10个文档
queryVector := []float32{0.1, 0.2, 0.3, ...} // 1536维查询向量
results, err := client.SearchSimilar(context.Background(), queryVector, 10)
if err != nil {
    // 处理错误
}

// 处理搜索结果
for _, result := range results {
    fmt.Printf("ID: %s, Content: %s, Similarity: %.4f\n", 
        result.ID, result.Content, result.Similarity)
    
    // 访问元数据
    source := result.Metadata["source"].(string)
    category := result.Metadata["category"].(string)
}
```

### 获取文档

```go
doc, err := client.GetDocument(context.Background(), "doc1")
if err != nil {
    // 处理错误
}

fmt.Printf("Content: %s\n", doc.Content)
```

### 删除文档

```go
err = client.DeleteDocument(context.Background(), "doc1")
if err != nil {
    // 处理错误
}
```

### 获取文档数量

```go
count, err := client.CountDocuments(context.Background())
if err != nil {
    // 处理错误
}

fmt.Printf("Total documents: %d\n", count)
```

## 性能优化

1. 对于大批量插入操作，建议使用`BatchUpsertDocuments`方法，它会在一个事务中完成所有插入，提高性能。
2. pgvector使用IVFFlat索引进行向量搜索，这种索引在大规模数据集上性能更好，但可能会牺牲一些准确性。
3. 如果需要更高的搜索精度，可以考虑修改索引创建SQL，使用HNSW索引。

## 注意事项

1. 确保插入的向量维度与配置的维度一致，否则会返回错误。
2. 本模块需要PostgreSQL数据库支持pgvector扩展，请确保已正确安装。
3. 对于生产环境，建议配置适当的连接池大小和超时设置。
4. 元数据字段使用JSONB类型存储，支持复杂的嵌套结构，但查询效率可能会受到影响。 