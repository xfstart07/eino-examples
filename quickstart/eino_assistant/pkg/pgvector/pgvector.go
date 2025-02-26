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

package pgvector

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const (
	// 表名前缀
	TablePrefix = "eino_doc_"

	// 默认表名
	DefaultTableName = "vectors"

	// 字段名
	IDField       = "id"
	ContentField  = "content"
	MetadataField = "metadata"
	VectorField   = "embedding"

	// 创建表的SQL模板
	createTableSQL = `
CREATE TABLE IF NOT EXISTS %s (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    metadata JSONB,
    embedding vector(%d) NOT NULL
);
`

	// 创建索引的SQL模板
	createIndexSQL = `
CREATE INDEX IF NOT EXISTS %s_vector_idx ON %s USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
`

	// 插入数据的SQL模板
	insertSQL = `
INSERT INTO %s (id, content, metadata, embedding) 
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE 
SET content = EXCLUDED.content, 
    metadata = EXCLUDED.metadata, 
    embedding = EXCLUDED.embedding;
`

	// 相似性搜索的SQL模板
	searchSQL = `
SELECT id, content, metadata, 1 - (embedding <=> $1) as similarity 
FROM %s 
ORDER BY embedding <=> $1 
LIMIT $2;
`

	// 数据库驱动名称
	driverName = "postgres"
)

var initOnce sync.Once

// Config 定义了PGVector的配置参数
type Config struct {
	// PostgreSQL连接信息
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// 向量维度
	Dimension int

	// 表名
	TableName string
}

// Client 是PGVector客户端
type Client struct {
	db        *sql.DB
	tableName string
	dimension int
}

// Document 表示存储在PGVector中的文档
type Document struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata"`
	Embedding  []float32              `json:"embedding"`
	Similarity float64                `json:"similarity,omitempty"`
}

// Init 初始化PGVector，创建必要的表和索引
func Init() error {
	var err error
	initOnce.Do(func() {
		err = InitPGVectorDB(context.Background(), &Config{
			Host:      "localhost",
			Port:      5432,
			User:      "postgres",
			Password:  "postgres",
			DBName:    "vectordb",
			SSLMode:   "disable",
			Dimension: 1536,
			TableName: DefaultTableName,
		})
	})
	return err
}

// InitPGVectorDB 初始化PGVector数据库
func InitPGVectorDB(ctx context.Context, config *Config) error {
	if config.Dimension <= 0 {
		return fmt.Errorf("dimension must be positive")
	}

	if config.TableName == "" {
		config.TableName = DefaultTableName
	}

	// 构建连接字符串
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// 连接数据库
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	// 检查连接
	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// 启用pgvector扩展
	_, err = db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS vector;")
	if err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}

	// 创建表
	tableName := TablePrefix + config.TableName
	_, err = db.ExecContext(ctx, fmt.Sprintf(createTableSQL, tableName, config.Dimension))
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// 创建索引
	_, err = db.ExecContext(ctx, fmt.Sprintf(createIndexSQL, tableName, tableName))
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// NewClient 创建一个新的PGVector客户端
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	if config.Dimension <= 0 {
		return nil, fmt.Errorf("dimension must be positive")
	}

	if config.TableName == "" {
		config.TableName = DefaultTableName
	}

	// 构建连接字符串
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// 连接数据库
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 检查连接
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &Client{
		db:        db,
		tableName: TablePrefix + config.TableName,
		dimension: config.Dimension,
	}, nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.db.Close()
}

// UpsertDocument 插入或更新文档
func (c *Client) UpsertDocument(ctx context.Context, doc *Document) error {
	if len(doc.Embedding) != c.dimension {
		return fmt.Errorf("embedding dimension mismatch: expected %d, got %d", c.dimension, len(doc.Embedding))
	}

	// 将Go的map转换为PostgreSQL的jsonb
	metadataBytes, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	// 执行插入或更新
	_, err = c.db.ExecContext(
		ctx,
		fmt.Sprintf(insertSQL, c.tableName),
		doc.ID,
		doc.Content,
		metadataBytes,
		encodeVector(doc.Embedding),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert document: %w", err)
	}

	return nil
}

// BatchUpsertDocuments 批量插入或更新文档
func (c *Client) BatchUpsertDocuments(ctx context.Context, docs []*Document) error {
	// 开始事务
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 准备语句
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(insertSQL, c.tableName))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 批量执行
	for _, doc := range docs {
		if len(doc.Embedding) != c.dimension {
			tx.Rollback()
			return fmt.Errorf("embedding dimension mismatch for doc %s: expected %d, got %d",
				doc.ID, c.dimension, len(doc.Embedding))
		}

		// 将Go的map转换为PostgreSQL的jsonb
		metadataBytes, err := json.Marshal(doc.Metadata)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to encode metadata for doc %s: %w", doc.ID, err)
		}

		_, err = stmt.ExecContext(
			ctx,
			doc.ID,
			doc.Content,
			metadataBytes,
			encodeVector(doc.Embedding),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to upsert document %s: %w", doc.ID, err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SearchSimilar 搜索相似文档
func (c *Client) SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]*Document, error) {
	if len(embedding) != c.dimension {
		return nil, fmt.Errorf("embedding dimension mismatch: expected %d, got %d", c.dimension, len(embedding))
	}

	if limit <= 0 {
		limit = 10 // 默认返回10条结果
	}

	// 执行查询
	rows, err := c.db.QueryContext(
		ctx,
		fmt.Sprintf(searchSQL, c.tableName),
		encodeVector(embedding),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar documents: %w", err)
	}
	defer rows.Close()

	// 处理结果
	var results []*Document
	for rows.Next() {
		var (
			id            string
			content       string
			metadataBytes []byte
			similarity    float64
		)

		if err := rows.Scan(&id, &content, &metadataBytes, &similarity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 解析metadata
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, fmt.Errorf("failed to decode metadata: %w", err)
		}

		results = append(results, &Document{
			ID:         id,
			Content:    content,
			Metadata:   metadata,
			Similarity: similarity,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// DeleteDocument 删除文档
func (c *Client) DeleteDocument(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(
		ctx,
		fmt.Sprintf("DELETE FROM %s WHERE id = $1", c.tableName),
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// GetDocument 获取文档
func (c *Client) GetDocument(ctx context.Context, id string) (*Document, error) {
	var (
		content       string
		metadataBytes []byte
	)

	err := c.db.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT content, metadata FROM %s WHERE id = $1", c.tableName),
		id,
	).Scan(&content, &metadataBytes)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// 解析metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &Document{
		ID:       id,
		Content:  content,
		Metadata: metadata,
	}, nil
}

// CountDocuments 获取文档总数
func (c *Client) CountDocuments(ctx context.Context) (int, error) {
	var count int
	err := c.db.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", c.tableName),
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// encodeVector 将float32切片编码为PostgreSQL向量格式
func encodeVector(vec []float32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range vec {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%f", v))
	}
	sb.WriteString("]")
	return sb.String()
}
