export default function DashboardPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold">控制台</h1>
      <p className="mt-1 text-sm text-muted-foreground">查看用量概览和账户状态</p>

      <div className="mt-8 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard label="本月请求" value="0" />
        <StatsCard label="Token 消耗" value="0" />
        <StatsCard label="本月费用" value="¥0.00" />
        <StatsCard label="账户余额" value="¥0.00" />
      </div>

      <div className="mt-8 rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold">快速开始</h2>
        <div className="mt-4 space-y-3 text-sm text-muted-foreground">
          <p>1. 在 <a href="/dashboard/keys" className="text-primary hover:underline">API Keys</a> 页面创建一个密钥</p>
          <p>2. 使用 OpenAI SDK 或 HTTP 请求调用 API</p>
          <p>3. 在 <a href="/dashboard/logs" className="text-primary hover:underline">调用日志</a> 查看请求记录</p>
        </div>
        <div className="mt-6 overflow-hidden rounded-lg border border-border bg-background">
          <pre className="p-4 text-xs leading-relaxed overflow-x-auto">
            <code>{`curl https://api.aitoken.dev/v1/chat/completions \\
  -H "Authorization: Bearer sk-your-key" \\
  -H "Content-Type: application/json" \\
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Hi"}]}'`}</code>
          </pre>
        </div>
      </div>
    </div>
  );
}

function StatsCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-card p-5">
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="mt-1 text-2xl font-bold">{value}</div>
    </div>
  );
}
