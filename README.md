# AI Token

A unified AI model aggregation gateway for developers. One API to access all mainstream AI models.

## Features

- **Unified API** — OpenAI-compatible interface, seamlessly switch between GPT, Claude, Gemini, DeepSeek
- **Smart Routing** — Priority-based scheduling with weighted random load balancing
- **Auto Failover** — Automatic cooldown and channel switching on provider failures
- **Stream Forwarding** — Full SSE streaming support, token-by-token real-time output
- **Usage Tracking** — Real-time logging of token consumption, latency, and costs
- **Multi-Provider** — OpenAI, Anthropic, Google, DeepSeek out of the box

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15 + TypeScript + Tailwind CSS 4 |
| Backend | Go 1.22+ + Gin |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Deploy | Docker Compose |

## Quick Start

### Docker (Recommended)

```bash
cp .env.example .env
# Edit .env to add your provider API keys
docker compose up -d
```

Access:
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- Health Check: http://localhost:8080/health

### Local Development

```bash
# Start PostgreSQL and Redis
make dev

# Start backend (new terminal)
make dev-gateway

# Start frontend (new terminal)
make dev-web
```

## API Usage

Use any OpenAI-compatible SDK — just replace the `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello"}],
    stream=True
)

for chunk in response:
    print(chunk.choices[0].delta.content, end="")
```

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: 'sk-your-api-key',
  baseURL: 'http://localhost:8080/v1',
});

const stream = await client.chat.completions.create({
  model: 'claude-sonnet-4-6',
  messages: [{ role: 'user', content: 'Hello' }],
  stream: true,
});

for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || '');
}
```

## Project Structure

```
ai-token/
├── apps/web/                # Next.js frontend
│   ├── app/                 # Pages (App Router)
│   ├── components/          # Shared components
│   ├── lib/                 # Utilities & API client
│   └── stores/              # Zustand state management
├── services/gateway/        # Go API Gateway
│   ├── cmd/server/          # Entry point
│   ├── internal/
│   │   ├── config/          # Configuration
│   │   ├── handler/         # HTTP handlers
│   │   ├── middleware/      # Auth, rate limit, CORS
│   │   ├── model/           # Database models & stores
│   │   ├── relay/           # Core relay engine
│   │   │   ├── adaptor/     # Provider adaptors
│   │   │   ├── router.go    # Channel selection algorithm
│   │   │   ├── relay.go     # Request orchestration
│   │   │   └── stream.go    # SSE stream forwarding
│   │   └── router/          # Route registration
│   └── migrations/          # SQL migrations
├── deploy/                  # Dockerfiles & nginx config
├── docker-compose.yml       # One-command full stack
└── .env.example             # Environment variables template
```

## Supported Providers

| Provider | Models | Format |
|----------|--------|--------|
| OpenAI | GPT-5.5, GPT-5.5 | Native (passthrough) |
| Anthropic | Claude Sonnet 4.8, Claude opu5 | Messages API → OpenAI |
| Google | Gemini 2.5 Pro, Gemini 2.5 Flash | Gemini API → OpenAI |
| DeepSeek | deepseek-v4-flash,deepseek-v4-pro | OpenAI-compatible |

## Environment Variables

See [`.env.example`](.env.example) for the full list. Key variables:

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GOOGLE_API_KEY` | Google AI API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `JWT_SECRET` | JWT signing secret (change in production) |
| `DATABASE_URL` | PostgreSQL connection string |

## License

MIT
