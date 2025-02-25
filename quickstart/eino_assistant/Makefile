# Eino Assistant Makefile

# 设置默认目标
.PHONY: all
all: help

# 设置工作目录
WORKDIR := $(shell pwd)

# 设置环境变量文件
ENV_FILE := .env

# 帮助信息
.PHONY: help
help:
	@echo "Eino Assistant 使用说明:"
	@echo ""
	@echo "可用命令:"
	@echo "  make run-agent        - 启动 Eino Agent 服务器"
	@echo "  make run-indexing     - 运行知识库索引"
	@echo "  make docker-redis     - 启动 Redis Stack 容器"
	@echo "  make docker-stop      - 停止所有 Docker 容器"
	@echo "  make clean            - 清理临时文件"
	@echo "  make help             - 显示此帮助信息"
	@echo ""
	@echo "环境变量配置在 .env 文件中"

# 启动 Eino Agent 服务器
.PHONY: run-agent
run-agent:
	@echo "启动 Eino Agent 服务器..."
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "警告: .env 文件不存在，可能导致配置错误"; \
	fi
	cd $(WORKDIR) && go run cmd/einoagent/*.go

# 运行知识库索引
.PHONY: run-indexing
run-indexing:
	@echo "运行知识库索引..."
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "警告: .env 文件不存在，可能导致配置错误"; \
	fi
	go run cmd/knowledgeindexing/main.go

# 启动 Redis Stack 容器
.PHONY: docker-redis
docker-redis:
	@echo "启动 Redis Stack 容器..."
	docker-compose up -d
	@echo "Redis Stack 已启动"
	@echo "Redis 服务端口: 6379"
	@echo "RedisInsight 界面: http://127.0.0.1:8001"

# 停止所有 Docker 容器
.PHONY: docker-stop
docker-stop:
	@echo "停止所有 Docker 容器..."
	docker-compose down
	@echo "所有容器已停止"

# 清理临时文件
.PHONY: clean
clean:
	@echo "清理临时文件..."
	find . -name "*.log" -type f -delete
	find . -name "*.tmp" -type f -delete
	@echo "清理完成"
