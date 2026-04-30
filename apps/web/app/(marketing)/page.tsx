export default function HomePage() {
  return (
    <div className="flex flex-col">
      {/* Hero */}
      <section className="relative overflow-hidden py-24 md:py-32">
        <div className="mx-auto max-w-7xl px-6 text-center">
          <h1 className="text-4xl font-bold tracking-tight md:text-6xl">
            一个 API，接入
            <span className="text-primary"> 所有 AI 模型</span>
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground">
            统一 OpenAI 兼容接口，智能路由多家供应商。自动故障切换、用量追踪、成本可控。
          </p>
          <div className="mt-10 flex items-center justify-center gap-4">
            <a
              href="/register"
              className="rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              免费获取 API Key
            </a>
            <a
              href="/models"
              className="rounded-lg border border-border px-6 py-3 text-sm font-medium text-foreground hover:bg-muted transition-colors"
            >
              探索模型
            </a>
          </div>
        </div>
      </section>

      {/* 统计数字 */}
      <section className="border-y border-border bg-card py-12">
        <div className="mx-auto max-w-7xl px-6">
          <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
            <StatItem value="100+" label="可用模型" />
            <StatItem value="4" label="主流供应商" />
            <StatItem value="99.9%" label="正常运行时间" />
            <StatItem value="<100ms" label="路由延迟" />
          </div>
        </div>
      </section>

      {/* 核心能力 */}
      <section className="py-24">
        <div className="mx-auto max-w-7xl px-6">
          <h2 className="text-center text-3xl font-bold">为什么选择 AI Token</h2>
          <p className="mx-auto mt-4 max-w-2xl text-center text-muted-foreground">
            不只是 API 代理，更是完整的 AI 模型管理平台
          </p>
          <div className="mt-16 grid gap-8 md:grid-cols-2 lg:grid-cols-3">
            <FeatureCard
              title="统一接口"
              description="OpenAI 兼容格式，无缝切换 GPT、Claude、Gemini、DeepSeek 等模型，无需修改代码。"
            />
            <FeatureCard
              title="智能路由"
              description="优先级调度 + 加权随机负载均衡，自动选择最优渠道，确保每次请求都走最快通道。"
            />
            <FeatureCard
              title="故障切换"
              description="供应商异常时自动冷却并切换到备用渠道，5分钟后自动恢复，保证服务不中断。"
            />
            <FeatureCard
              title="流式转发"
              description="完整支持 SSE 流式响应，逐 token 实时输出，与直连供应商体验一致。"
            />
            <FeatureCard
              title="用量追踪"
              description="实时记录每次调用的 token 消耗、延迟、费用，按模型分组统计，成本一目了然。"
            />
            <FeatureCard
              title="多供应商"
              description="首批支持 OpenAI、Anthropic、Google、DeepSeek，持续接入更多供应商。"
            />
          </div>
        </div>
      </section>

      {/* 代码示例 */}
      <section className="border-t border-border bg-card py-24">
        <div className="mx-auto max-w-7xl px-6">
          <h2 className="text-center text-3xl font-bold">即刻开始</h2>
          <p className="mx-auto mt-4 max-w-2xl text-center text-muted-foreground">
            使用任何 OpenAI SDK 即可调用，零学习成本
          </p>
          <div className="mx-auto mt-12 max-w-2xl overflow-hidden rounded-xl border border-border bg-background">
            <div className="flex items-center gap-2 border-b border-border px-4 py-3">
              <span className="text-xs text-muted-foreground">Python</span>
            </div>
            <pre className="overflow-x-auto p-6 text-sm leading-relaxed text-foreground">
              <code>{`from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="https://api.aitoken.dev/v1"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "你好"}],
    stream=True
)

for chunk in response:
    print(chunk.choices[0].delta.content, end="")`}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-24">
        <div className="mx-auto max-w-7xl px-6 text-center">
          <h2 className="text-3xl font-bold">准备好了吗？</h2>
          <p className="mt-4 text-muted-foreground">
            注册即送免费额度，无需信用卡
          </p>
          <a
            href="/register"
            className="mt-8 inline-block rounded-lg bg-primary px-8 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            立即免费开始
          </a>
        </div>
      </section>
    </div>
  );
}

function StatItem({ value, label }: { value: string; label: string }) {
  return (
    <div className="text-center">
      <div className="text-3xl font-bold text-primary">{value}</div>
      <div className="mt-1 text-sm text-muted-foreground">{label}</div>
    </div>
  );
}

function FeatureCard({ title, description }: { title: string; description: string }) {
  return (
    <div className="rounded-xl border border-border bg-card p-6 transition-colors hover:border-primary/50">
      <h3 className="text-lg font-semibold">{title}</h3>
      <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{description}</p>
    </div>
  );
}
