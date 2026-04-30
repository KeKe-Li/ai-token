# AI Token

面向开发者的 AI 模型聚合中转平台。一个 API，接入所有主流 AI 模型。

## 功能

- **统一接口** — OpenAI Compatible API，无缝切换 GPT、Claude、Gemini、DeepSeek
- **智能路由** — 优先级调度 + 加权随机负载均衡
- **故障切换** — 供应商异常自动冷却并切换备用渠道
- **流式转发** — 完整支持 SSE 流式响应
- **用量追踪** — 实时记录 token 消耗、延迟、费用
- **多供应商** — OpenAI、Anthropic、Google、DeepSeek

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Next.js 15 + TypeScript + Tailwind CSS 4 |
| 后端 | Go 1.22 + Gin |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 部署 | Docker Compose |

## 快速开始

### 开发环境

```bash
# 启动数据库和缓存
make dev

# 启动后端 (新终端)
make dev-gateway

# 启动前端 (新终端)
make dev-web
```

### Docker 部署

```bash
docker compose up -d
```

访问:
- 前端: http://localhost:3000
- API: http://localhost:8080
- 健康检查: http://localhost:8080/health

## API 使用

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "你好"}],
    stream=True
)

for chunk in response:
    print(chunk.choices[0].delta.content, end="")
```

## 项目结构

```
ai-token/
├── apps/web/              # Next.js 前端
├── services/gateway/      # Go API Gateway
├── deploy/                # Docker 和部署配置
├── scripts/               # 开发脚本
└── docker-compose.yml     # 一键启动
```

## License

MIT
