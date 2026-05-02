"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type Stats = { total_requests: number; total_cost: number; total_input_tokens: number; total_output_tokens: number; unique_users: number };

export default function AdminPage() {
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<Stats>("/api/admin/stats", token).then(setStats).catch(() => {});
  }, []);

  return (
    <div>
      <h2 className="text-2xl font-bold">系统概览</h2>
      <p className="mt-1 text-sm text-muted-foreground">过去 24 小时全局统计</p>

      <div className="mt-8 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard label="请求总量" value={stats?.total_requests?.toLocaleString() ?? "0"} />
        <StatsCard label="活跃用户" value={stats?.unique_users?.toLocaleString() ?? "0"} />
        <StatsCard label="总收入" value={`¥${stats ? (stats.total_cost / 1000).toFixed(2) : "0.00"}`} accent />
        <StatsCard label="总 Tokens" value={stats ? ((stats.total_input_tokens + stats.total_output_tokens) / 1000).toFixed(0) + "K" : "0"} />
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
