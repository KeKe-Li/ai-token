export default function HomePage() {
  return (
    <div className="flex flex-col overflow-hidden">
      {/* Hero */}
      <section className="relative min-h-[85vh] flex items-center justify-center hero-grid">
        <div className="absolute inset-0 hero-glow" />
        {/* 装饰光球 */}
        <div className="absolute top-20 left-1/4 w-72 h-72 bg-purple-500/10 rounded-full blur-[100px] animate-float" />
        <div className="absolute bottom-20 right-1/4 w-96 h-96 bg-cyan-500/8 rounded-full blur-[120px] animate-float" style={{ animationDelay: "3s" }} />

        <div className="relative z-10 mx-auto max-w-5xl px-6 text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-border bg-muted/50 px-4 py-1.5 text-sm text-muted-foreground mb-8 backdrop-blur-sm">
            <span className="inline-block w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            全部模型可用 · 99.9% 正常运行
          </div>

          <h1 className="text-5xl font-bold tracking-tight md:text-7xl lg:text-8xl leading-tight">
            一个 API，接入
            <br />
            <span className="gradient-text">所有 AI 模型</span>
          </h1>

          <p className="mx-auto mt-8 max-w-2xl text-lg md:text-xl text-muted-foreground leading-relaxed">
            统一 OpenAI 兼容接口，智能路由 <span className="text-foreground">OpenAI</span>、<span className="text-foreground">Claude</span>、<span className="text-foreground">Gemini</span>、<span className="text-foreground">DeepSeek</span>。
            <br className="hidden md:block" />
            自动故障切换、实时用量追踪、成本完全可控。
          </p>

          <div className="mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
            <a href="/register" className="btn-glow rounded-xl px-8 py-3.5 text-sm font-semibold text-white transition-all">
              免费获取 API Key →
            </a>
            <a href="/models" className="group rounded-xl border border-border px-8 py-3.5 text-sm font-medium text-foreground hover:border-primary/50 hover:bg-muted/50 transition-all backdrop-blur-sm">
              探索模型
              <span className="inline-block ml-1 group-hover:translate-x-1 transition-transform">→</span>
            </a>
          </div>

          {/* 供应商 logos */}
          <div className="mt-16 flex items-center justify-center gap-6 flex-wrap">
            {["OpenAI", "Anthropic", "Google", "DeepSeek"].map((name) => (
              <span key={name} className="provider-badge rounded-lg px-4 py-2 text-sm text-muted-foreground">
                {name}
              </span>
            ))}
          </div>
        </div>
      </section>

      {/* 统计数字 */}
      <section className="relative border-y border-border py-16">
        <div className="absolute inset-0 bg-gradient-to-r from-purple-500/3 via-transparent to-cyan-500/3" />
        <div className="relative mx-auto max-w-7xl px-6">
          <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
            <StatItem value="100+" label="可用模型" />
            <StatItem value="4" label="主流供应商" />
            <StatItem value="99.9%" label="正常运行时间" />
            <StatItem value="<100ms" label="路由延迟" />
          </div>
        </div>
      </section>

      {/* 核心能力 */}
      <section className="relative py-28">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[600px] bg-purple-500/5 rounded-full blur-[150px]" />
        <div className="relative mx-auto max-w-7xl px-6">
          <div className="text-center">
            <h2 className="text-3xl font-bold md:text-4xl">
              为什么选择 <span className="gradient-text">AI Token</span>
            </h2>
            <p className="mx-auto mt-4 max-w-2xl text-muted-foreground">
              不只是 API 代理，更是完整的 AI 模型管理和智能路由平台
            </p>
          </div>
          <div className="mt-16 grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <FeatureCard
              icon="⚡"
              title="统一接口"
              description="OpenAI 兼容格式，一行代码切换 GPT、Claude、Gemini、DeepSeek，无需重写逻辑。"
              gradient="from-purple-500/20 to-transparent"
            />
            <FeatureCard
              icon="🧠"
              title="智能路由"
              description="优先级调度 + 加权随机负载均衡，毫秒级选出最优渠道，请求永远走最快通道。"
              gradient="from-cyan-500/20 to-transparent"
            />
            <FeatureCard
              icon="🛡️"
              title="故障自愈"
              description="供应商异常自动冷却并无缝切换到备用渠道，5 分钟后探测恢复，零停机时间。"
              gradient="from-emerald-500/20 to-transparent"
            />
            <FeatureCard
              icon="🌊"
              title="流式转发"
              description="完整 SSE 流式响应，逐 token 实时输出，与直连供应商体验完全一致。"
              gradient="from-blue-500/20 to-transparent"
            />
            <FeatureCard
              icon="📊"
              title="实时洞察"
              description="每次调用自动记录 token 消耗、延迟、费用，按模型分组可视化，成本一目了然。"
              gradient="from-orange-500/20 to-transparent"
            />
            <FeatureCard
              icon="🔌"
              title="多供应商"
              description="首批支持 OpenAI、Anthropic、Google、DeepSeek，适配器架构轻松扩展更多。"
              gradient="from-pink-500/20 to-transparent"
            />
          </div>
        </div>
      </section>

      {/* 代码示例 */}
      <section className="relative border-t border-border py-28">
        <div className="absolute bottom-0 right-0 w-[500px] h-[500px] bg-cyan-500/5 rounded-full blur-[150px]" />
        <div className="relative mx-auto max-w-7xl px-6">
          <div className="grid gap-12 lg:grid-cols-2 lg:items-center">
            <div>
              <h2 className="text-3xl font-bold md:text-4xl">
                三行代码<span className="gradient-text">即刻接入</span>
              </h2>
              <p className="mt-4 text-muted-foreground leading-relaxed">
                使用任何 OpenAI 兼容 SDK（Python、TypeScript、Go、Rust...），只需替换 base_url，零学习成本。
              </p>
              <ul className="mt-8 space-y-4">
                <CheckItem text="兼容所有 OpenAI SDK 和工具链" />
                <CheckItem text="支持流式和非流式两种模式" />
                <CheckItem text="自动处理重试和故障切换" />
                <CheckItem text="实时记录用量和费用" />
              </ul>
            </div>
            <div className="code-block rounded-2xl overflow-hidden">
              <div className="flex items-center gap-2 border-b border-border/50 px-5 py-3">
                <div className="flex gap-1.5">
                  <span className="w-3 h-3 rounded-full bg-red-500/60" />
                  <span className="w-3 h-3 rounded-full bg-yellow-500/60" />
                  <span className="w-3 h-3 rounded-full bg-green-500/60" />
                </div>
                <span className="ml-3 text-xs text-muted-foreground">main.py</span>
              </div>
              <pre className="p-6 text-sm leading-7 overflow-x-auto">
                <code>
                  <span className="text-purple-400">from</span> <span className="text-cyan-300">openai</span> <span className="text-purple-400">import</span> OpenAI{"\n\n"}
                  <span className="text-muted-foreground"># 只需替换 base_url 即可</span>{"\n"}
                  client = OpenAI({"\n"}
                  {"    "}api_key=<span className="text-emerald-400">&quot;sk-your-key&quot;</span>,{"\n"}
                  {"    "}base_url=<span className="text-emerald-400">&quot;https://api.aitoken.dev/v1&quot;</span>{"\n"}
                  ){"\n\n"}
                  response = client.chat.completions.create({"\n"}
                  {"    "}model=<span className="text-emerald-400">&quot;gpt-4o&quot;</span>,{"\n"}
                  {"    "}messages=[&#123;<span className="text-emerald-400">&quot;role&quot;</span>: <span className="text-emerald-400">&quot;user&quot;</span>, <span className="text-emerald-400">&quot;content&quot;</span>: <span className="text-emerald-400">&quot;你好&quot;</span>&#125;],{"\n"}
                  {"    "}stream=<span className="text-orange-400">True</span>{"\n"}
                  ){"\n\n"}
                  <span className="text-purple-400">for</span> chunk <span className="text-purple-400">in</span> response:{"\n"}
                  {"    "}print(chunk.choices[<span className="text-orange-400">0</span>].delta.content, end=<span className="text-emerald-400">&quot;&quot;</span>)
                </code>
              </pre>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="relative py-28">
        <div className="absolute inset-0 hero-glow opacity-50" />
        <div className="relative mx-auto max-w-7xl px-6 text-center">
          <h2 className="text-3xl font-bold md:text-5xl">
            准备好<span className="gradient-text">释放 AI 潜力</span>了吗？
          </h2>
          <p className="mt-4 text-lg text-muted-foreground">
            注册即送免费额度，无需信用卡，30 秒上手
          </p>
          <div className="mt-10">
            <a href="/register" className="btn-glow inline-block rounded-xl px-10 py-4 text-base font-semibold text-white">
              立即免费开始 →
            </a>
          </div>
          <p className="mt-6 text-sm text-muted-foreground">
            已有 <span className="text-foreground">1,000+</span> 开发者在使用 AI Token
          </p>
        </div>
      </section>
    </div>
  );
}

function StatItem({ value, label }: { value: string; label: string }) {
  return (
    <div className="text-center">
      <div className="text-3xl md:text-4xl font-bold gradient-text stat-glow">{value}</div>
      <div className="mt-2 text-sm text-muted-foreground">{label}</div>
    </div>
  );
}

function FeatureCard({ icon, title, description, gradient }: { icon: string; title: string; description: string; gradient: string }) {
  return (
    <div className="card-glow rounded-2xl p-6 relative overflow-hidden">
      <div className={`absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl ${gradient} rounded-full blur-2xl -translate-y-1/2 translate-x-1/2`} />
      <div className="relative">
        <span className="text-2xl">{icon}</span>
        <h3 className="mt-4 text-lg font-semibold">{title}</h3>
        <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}

function CheckItem({ text }: { text: string }) {
  return (
    <li className="flex items-center gap-3">
      <span className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-400 text-xs">✓</span>
      <span className="text-sm text-muted-foreground">{text}</span>
    </li>
  );
}
