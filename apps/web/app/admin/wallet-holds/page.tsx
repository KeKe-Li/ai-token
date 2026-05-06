"use client";

import { useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";

type WalletHold = {
  id: number;
  user_id: number;
  api_key_id: number;
  request_log_id?: number;
  model?: string;
  amount: number;
  captured_amount: number;
  released_amount: number;
  status: string;
  reason: string;
  created_at: string;
  updated_at: string;
  expires_at?: string;
};

const STATUS_OPTIONS = [
  { value: "", label: "全部" },
  { value: "held", label: "Held" },
  { value: "captured", label: "Captured" },
  { value: "released", label: "Released" },
  { value: "failed", label: "Failed" },
];

export default function AdminWalletHoldsPage() {
  const [holds, setHolds] = useState<WalletHold[]>([]);
  const [loading, setLoading] = useState(true);
  const [userID, setUserID] = useState("");
  const [status, setStatus] = useState("held");
  const [releasingID, setReleasingID] = useState<number | null>(null);

  const load = () => {
    const token = localStorage.getItem("token");
    if (!token) return;
    setLoading(true);
    const params = new URLSearchParams({ limit: "100" });
    if (userID.trim()) params.set("user_id", userID.trim());
    if (status) params.set("status", status);
    api.get<{ data: WalletHold[] }>(`/api/admin/wallet/holds?${params.toString()}`, token)
      .then((res) => setHolds(res.data || []))
      .catch(() => setHolds([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const summary = useMemo(() => {
    return holds.reduce(
      (acc, hold) => {
        acc.total += hold.amount;
        if (hold.status === "held") acc.held += hold.amount;
        if (hold.status === "captured") acc.captured += hold.captured_amount;
        if (hold.status === "released") acc.released += hold.released_amount;
        return acc;
      },
      { total: 0, held: 0, captured: 0, released: 0 },
    );
  }, [holds]);

  const releaseHold = async (holdID: number) => {
    const token = localStorage.getItem("token");
    if (!token) return;
    const reason = window.prompt("请输入手动释放原因", "异常冻结修复") || "";
    setReleasingID(holdID);
    try {
      await api.post(`/api/admin/wallet/holds/${holdID}/release`, { reason }, token);
      await Promise.resolve(load());
    } finally {
      setReleasingID(null);
    }
  };

  return (
    <div>
      <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <h2 className="text-2xl font-bold">预授权 Hold</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            查看请求级资金冻结状态，并可手动释放异常 held hold。
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <input
            value={userID}
            onChange={(event) => setUserID(event.target.value)}
            placeholder="按用户 ID 过滤"
            className="w-36 rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
          />
          <select
            value={status}
            onChange={(event) => setStatus(event.target.value)}
            className="rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
          >
            {STATUS_OPTIONS.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
          </select>
          <button onClick={load} className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white">查询</button>
        </div>
      </div>

      <div className="mt-6 grid gap-4 md:grid-cols-4">
        <MetricCard label="Hold 总额" value={formatMoney(summary.total)} />
        <MetricCard label="当前冻结" value={formatMoney(summary.held)} tone="yellow" />
        <MetricCard label="已 Capture" value={formatMoney(summary.captured)} tone="green" />
        <MetricCard label="已 Release" value={formatMoney(summary.released)} />
      </div>

      <div className="mt-6 overflow-x-auto">
        {loading ? (
          <div className="py-16 text-center text-muted-foreground">加载中...</div>
        ) : holds.length === 0 ? (
          <div className="card-glow rounded-xl p-16 text-center">
            <div className="text-4xl mb-4">🧊</div>
            <p className="font-medium text-foreground">暂无预授权 Hold</p>
            <p className="mt-1 text-sm text-muted-foreground">创建预授权、结算或释放后会显示在这里。</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-3 py-3 text-left font-medium">创建时间</th>
                <th className="px-3 py-3 text-left font-medium">用户</th>
                <th className="px-3 py-3 text-left font-medium">状态</th>
                <th className="px-3 py-3 text-left font-medium">模型</th>
                <th className="px-3 py-3 text-right font-medium">Hold</th>
                <th className="px-3 py-3 text-right font-medium">Capture</th>
                <th className="px-3 py-3 text-right font-medium">Release</th>
                <th className="px-3 py-3 text-left font-medium">关联</th>
                <th className="px-3 py-3 text-left font-medium">过期</th>
                <th className="px-3 py-3 text-right font-medium">操作</th>
              </tr>
            </thead>
            <tbody>
              {holds.map((hold) => (
                <tr key={hold.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                  <td className="px-3 py-3 text-xs text-muted-foreground whitespace-nowrap">{formatTime(hold.created_at)}</td>
                  <td className="px-3 py-3 text-xs">#{hold.user_id}</td>
                  <td className="px-3 py-3"><StatusBadge status={hold.status} /></td>
                  <td className="px-3 py-3 font-mono text-xs text-muted-foreground">{hold.model || "-"}</td>
                  <td className="px-3 py-3 text-right font-mono">{formatMoney(hold.amount)}</td>
                  <td className="px-3 py-3 text-right text-emerald-400">{formatMoney(hold.captured_amount)}</td>
                  <td className="px-3 py-3 text-right text-yellow-400">{formatMoney(hold.released_amount)}</td>
                  <td className="px-3 py-3 text-xs text-muted-foreground">
                    {hold.request_log_id ? <a className="text-primary hover:underline" href={`/admin/logs?request_log_id=${hold.request_log_id}`}>请求 #{hold.request_log_id}</a> : "-"}
                    <div><a className="text-primary hover:underline" href={`/admin/billing-events?wallet_hold_id=${hold.id}`}>事件</a></div>
                  </td>
                  <td className="px-3 py-3 text-xs text-muted-foreground whitespace-nowrap">{hold.expires_at ? formatTime(hold.expires_at) : "-"}</td>
                  <td className="px-3 py-3 text-right">
                    {hold.status === "held" ? (
                      <button
                        onClick={() => releaseHold(hold.id)}
                        disabled={releasingID === hold.id}
                        className="rounded-lg border border-yellow-500/40 px-3 py-1 text-xs text-yellow-400 hover:bg-yellow-500/10 disabled:opacity-50"
                      >
                        {releasingID === hold.id ? "释放中" : "手动释放"}
                      </button>
                    ) : (
                      <span className="text-xs text-muted-foreground">-</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

function MetricCard({ label, value, tone }: { label: string; value: string; tone?: "green" | "yellow" }) {
  const color = tone === "green" ? "text-emerald-400" : tone === "yellow" ? "text-yellow-400" : "text-foreground";
  return (
    <div className="card-glow rounded-xl p-4">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className={`mt-2 text-xl font-semibold ${color}`}>{value}</div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const color = status === "held" ? "bg-yellow-500/10 text-yellow-400" : status === "captured" ? "bg-emerald-500/10 text-emerald-400" : status === "failed" ? "bg-red-500/10 text-red-400" : "bg-muted text-muted-foreground";
  return <span className={`rounded-full px-2 py-0.5 text-xs ${color}`}>{status}</span>;
}

function formatMoney(value: number) {
  return `¥${(value / 1000).toFixed(4)}`;
}

function formatTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}
