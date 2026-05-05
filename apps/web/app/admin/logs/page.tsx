"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type LogEntry = { id: number; user_id: number; model: string; status_code: number; input_tokens: number; output_tokens: number; cost: number; reserved_amount?: number; latency_ms: number; ip_address: string; billing_status: string; billing_note?: string; estimated_tokens: boolean; created_at: string };

export default function AdminLogsPage() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<{ data: LogEntry[] }>("/api/admin/logs?limit=50", token)
      .then((res) => setLogs(res.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <h2 className="text-2xl font-bold">全局日志</h2>
      <p className="mt-1 text-sm text-muted-foreground">查看所有用户的 API 调用记录</p>

      <div className="mt-6">
        {loading ? (
          <div className="py-16 text-center text-muted-foreground">加载中...</div>
        ) : logs.length === 0 ? (
          <div className="card-glow rounded-xl p-16 text-center">
            <div className="text-4xl mb-4">📋</div>
            <p className="text-foreground font-medium">暂无调用记录</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="px-3 py-3 text-left font-medium">时间</th>
                  <th className="px-3 py-3 text-left font-medium">用户</th>
                  <th className="px-3 py-3 text-left font-medium">模型</th>
                  <th className="px-3 py-3 text-right font-medium">Tokens</th>
                  <th className="px-3 py-3 text-right font-medium">费用</th>
                  <th className="px-3 py-3 text-right font-medium">预授权</th>
                  <th className="px-3 py-3 text-right font-medium">延迟</th>
                  <th className="px-3 py-3 text-center font-medium">计费</th>
                  <th className="px-3 py-3 text-left font-medium">IP</th>
                  <th className="px-3 py-3 text-center font-medium">状态</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <tr key={log.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                    <td className="px-3 py-3 text-muted-foreground whitespace-nowrap text-xs">
                      {new Date(log.created_at).toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" })}
                    </td>
                    <td className="px-3 py-3 text-xs">#{log.user_id}</td>
                    <td className="px-3 py-3 font-mono text-xs">{log.model}</td>
                    <td className="px-3 py-3 text-right text-muted-foreground text-xs">{(log.input_tokens + log.output_tokens).toLocaleString()}</td>
                    <td className="px-3 py-3 text-right text-xs">¥{(log.cost / 1000).toFixed(4)}</td>
                    <td className="px-3 py-3 text-right text-muted-foreground text-xs">¥{((log.reserved_amount ?? 0) / 1000).toFixed(4)}</td>
                    <td className="px-3 py-3 text-right text-muted-foreground text-xs">{log.latency_ms}ms</td>
                    <td className="px-3 py-3 text-center"><BillingBadge status={log.billing_status} estimated={log.estimated_tokens} note={log.billing_note} /></td>
                    <td className="px-3 py-3 text-muted-foreground text-xs">{log.ip_address}</td>
                    <td className="px-3 py-3 text-center">
                      <span className={`rounded-full px-2 py-0.5 text-xs ${log.status_code === 200 ? "text-emerald-400 bg-emerald-500/10" : "text-red-400 bg-red-500/10"}`}>
                        {log.status_code}
                      </span>
                    </td>
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

function BillingBadge({ status, estimated, note }: { status?: string; estimated?: boolean; note?: string }) {
  const normalized = status || "unknown";
  const ok = normalized === "charged" || normalized === "zero_cost";
  const warn = normalized === "unpriced" || normalized === "charge_failed";
  return (
    <span
      title={note || normalized}
      className={`rounded-full px-2 py-0.5 text-xs ${
        ok ? "text-emerald-400 bg-emerald-500/10" :
        warn ? "text-yellow-400 bg-yellow-500/10" :
        "text-muted-foreground bg-muted"
      }`}
    >
      {normalized}{estimated ? " · 估" : ""}
    </span>
  );
}
