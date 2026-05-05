"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { UsageChart } from "@/components/dashboard/usage-chart";

type DashboardData = {
  balance: number;
  reserved?: number;
  used_amount: number;
  request_count: number;
  usage: { model: string; total_requests: number; input_tokens: number; output_tokens: number; total_cost: number }[] | null;
};

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<DashboardData>("/api/user/dashboard", token).then(setData).catch(() => {});
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-bold">控制台</h1>
      <p className="mt-1 text-sm text-muted-foreground">查看用量概览和账户状态</p>

      <div className="mt-8 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard label="账户余额" value={`¥${data ? (data.balance / 1000).toFixed(2) : "0.00"}`} accent />
        <StatsCard label="预授权冻结" value={`¥${data ? ((data.reserved ?? 0) / 1000).toFixed(2) : "0.00"}`} />
        <StatsCard label="累计消费" value={`¥${data ? (data.used_amount / 1000).toFixed(2) : "0.00"}`} />
        <StatsCard label="累计请求" value={String(data?.request_count ?? 0)} />
      </div>

      <div className="mt-8">
        <UsageChart />
      </div>

      {data?.usage && data.usage.length > 0 && (
        <div className="mt-8">
          <h2 className="text-lg font-semibold">本月用量</h2>
          <div className="mt-4 grid gap-3 md:grid-cols-2">
            {data.usage.map((u) => (
              <div key={u.model} className="card-glow rounded-xl p-4 flex items-center justify-between">
                <div>
                  <div className="font-medium text-sm">{u.model}</div>
                  <div className="text-xs text-muted-foreground mt-1">{u.total_requests} 次请求</div>
                </div>
                <div className="text-right">
                  <div className="text-sm font-medium">¥{(u.total_cost / 1000).toFixed(3)}</div>
                  <div className="text-xs text-muted-foreground">{(u.input_tokens + u.output_tokens).toLocaleString()} tokens</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="mt-8 card-glow rounded-xl p-6">
        <h2 className="text-lg font-semibold">快速开始</h2>
        <div className="mt-4 space-y-3 text-sm text-muted-foreground">
          <p>1. 在 <a href="/dashboard/keys" className="text-primary hover:underline">API Keys</a> 页面创建一个密钥</p>
          <p>2. 使用 OpenAI SDK 或 HTTP 请求调用 API</p>
          <p>3. 在 <a href="/dashboard/logs" className="text-primary hover:underline">调用日志</a> 查看请求记录</p>
        </div>
        <div className="mt-6 code-block rounded-lg overflow-hidden">
          <pre className="p-4 text-xs leading-relaxed overflow-x-auto">
            <code>{`curl http://localhost:8080/v1/chat/completions \\
  -H "Authorization: Bearer sk-your-key" \\
  -H "Content-Type: application/json" \\
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Hi"}]}'`}</code>
          </pre>
        </div>
      </div>
    </div>
  );
}

function StatsCard({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
  return (
    <div className="card-glow rounded-xl p-5">
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className={`mt-1 text-2xl font-bold ${accent ? "gradient-text" : ""}`}>{value}</div>
    </div>
  );
}
