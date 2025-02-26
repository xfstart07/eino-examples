-- 创建pgvector扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 创建默认表
CREATE TABLE IF NOT EXISTS eino_doc_vectors (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    metadata JSONB,
    embedding vector(1536) NOT NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS eino_doc_vectors_vector_idx 
ON eino_doc_vectors 
USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);

-- 创建示例数据
INSERT INTO eino_doc_vectors (id, content, metadata, embedding)
VALUES 
    ('example1', '这是一个示例文档', '{"source": "init-script", "category": "example"}', '[0.1, 0.2, 0.3]'),
    ('example2', '这是另一个示例文档', '{"source": "init-script", "category": "example"}', '[0.2, 0.3, 0.4]')
ON CONFLICT (id) DO NOTHING;

-- 授予权限
GRANT ALL PRIVILEGES ON TABLE eino_doc_vectors TO postgres; 