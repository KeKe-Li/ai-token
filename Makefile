.PHONY: dev dev-gateway dev-web build up down logs migrate seed clean

# 开发环境
dev:
	@echo "启动开发环境..."
	docker compose up postgres redis -d
	@echo "等待数据库就绪..."
	@sleep 3
	@echo "PostgreSQL 和 Redis 已就绪"
	@echo "运行 make dev-gateway 和 make dev-web 分别启动后端和前端"

dev-gateway:
	cd services/gateway && go run ./cmd/server

dev-web:
	cd apps/web && pnpm dev

# 构建
build:
	docker compose build

# Docker Compose 操作
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

# 数据库
migrate:
	docker compose exec postgres psql -U aitoken -d aitoken -f /docker-entrypoint-initdb.d/001_init.sql

seed:
	@echo "初始数据已通过迁移脚本自动填充"

# 清理
clean:
	docker compose down -v
	@echo "已清理所有容器和数据卷"

# Go 后端
gateway-build:
	cd services/gateway && go build -o bin/gateway ./cmd/server

gateway-test:
	cd services/gateway && go test ./...

# 前端
web-build:
	cd apps/web && pnpm build

web-lint:
	cd apps/web && pnpm lint
