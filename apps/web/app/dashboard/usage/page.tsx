"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type UsageItem = { model: string; total_requests: number; input_tokens: number; output_tokens: number; total_cost: number };

export default function UsagePage() {
  const [usage, setUsage] = useState<UsageItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<{ data: UsageItem[] }>("/api/user/usage?days=30", token)
      .then((res) => setUsage(res.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const totalCost = usage.reduce((s, u) => s + u.total_cost, 0);
  const totalRequests = usage.reduce((s, u) => s + u.total_requests, 0);
  const totalTokens = usage.reduce((s, u) => s + u.input_tokens + u.output_tokens, 0);

  return (
    <div>
      <h1 className="text-2xl font-bold">用量统计</h1>
      <p className="mt-1 text-sm text-muted-foreground">近 30 天按模型查看消耗明细</p>

      <div className="mt-8 grid gap-4 md:grid-cols-3">
        <div className="card-glow rounded-xl p-5">
          <div className="text-sm text-muted-foreground">总费用</div>
          <div className="mt-1 text-2xl font-bold gradient-text">¥{(totalCost / 1000).toFixed(2)}</div>
        </div>
        <div className="card-glow rounded-xl p-5">
          <div className="text-sm text-muted-foreground">总请求</div>
          <div className="mt-1 text-2xl font-bold">{totalRequests.toLocaleString()}</div>
        </div>
        <div className="card-glow rounded-xl p-5">
          <div className="text-sm text-muted-foreground">总 Tokens</div>
          <div className="mt-1 text-2xl font-bold">{totalTokens.toLocaleString()}</div>
        </div>
      </div>

      <div className="mt-8">
        <h2 className="text-lg font-semibold">按模型分组</h2>
        {loading ? (
          <div className="mt-4 py-12 text-center text-muted-foreground">加载中...</div>
        ) : usage.length === 0 ? (
          <div className="mt-4 card-glow rounded-xl p-12 text-center">
            <p className="text-muted-foreground">暂无用量数据</p>
          </div>
        ) : (
          <div className="mt-4 overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="px-4 py-3 text-left font-medium">模型</th>
                  <th className="px-4 py-3 text-right font-medium">请求数</th>
                  <th className="px-4 py-3 text-right font-medium">输入 Tokens</th>
                  <th className="px-4 py-3 text-right font-medium">输出 Tokens</th>
                  <th className="px-4 py-3 text-right font-medium">费用</th>
                </tr>
              </thead>
              <tbody>
                {usage.map((u) => (
                  <tr key={u.model} className="border-b border-border hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono text-xs">{u.model}</td>
                    <td className="px-4 py-3 text-right">{u.total_requests.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right text-muted-foreground">{u.input_tokens.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right text-muted-foreground">{u.output_tokens.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right font-medium">¥{(u.total_cost / 1000).toFixed(4)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
